package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// DealPortalLinksHandler manages deal portal links (authenticated CRUD) and
// serves the public token-based deal snapshot endpoint.
type DealPortalLinksHandler struct {
	links repository.PortalLinkRepository
	deals repository.DealRepository
	notes repository.NoteRepository
	orgs  repository.OrgRepository
}

func NewDealPortalLinksHandler(
	links repository.PortalLinkRepository,
	deals repository.DealRepository,
	notes repository.NoteRepository,
	orgs repository.OrgRepository,
) *DealPortalLinksHandler {
	return &DealPortalLinksHandler{links: links, deals: deals, notes: notes, orgs: orgs}
}

// AuthRouter returns routes that require authentication (mounted under /api/v1).
func (h *DealPortalLinksHandler) AuthRouter() chi.Router {
	r := chi.NewRouter()

	// Routes under /deals/{dealId}/portal-links
	r.Get("/", h.ListLinks)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.CreateLink)
	})

	// Routes under /portal-links/{id}
	return r
}

// RevokeRouter returns the revoke route, mounted at /portal-links/{id}.
func (h *DealPortalLinksHandler) RevokeRouter() chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Delete("/{id}", h.RevokeLink)
	})
	return r
}

// PublicRouter returns the unauthenticated token lookup route.
// Mount this BEFORE the auth middleware.
func (h *DealPortalLinksHandler) PublicRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/{token}", h.GetSnapshot)
	return r
}

// CreateLink handles POST /api/v1/deals/{dealId}/portal-links.
func (h *DealPortalLinksHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "dealId"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid deal id")
		return
	}

	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Verify deal exists and belongs to this org.
	if _, err := h.deals.GetByID(r.Context(), dealID); err != nil {
		handleDomainErr(w, err)
		return
	}

	var req struct {
		Label     *string    `json:"label,omitempty"`
		ExpiresAt *time.Time `json:"expires_at,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	token, err := generateToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	userID := claims.UserID
	link := &domain.PortalLink{
		OrgID:           orgID,
		DealID:          dealID,
		CreatedByUserID: &userID,
		Token:           token,
		Label:           req.Label,
		ExpiresAt:       req.ExpiresAt,
	}

	created, err := h.links.Create(r.Context(), link)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// ListLinks handles GET /api/v1/deals/{dealId}/portal-links.
func (h *DealPortalLinksHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "dealId"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid deal id")
		return
	}

	// Verify deal exists and is accessible in this org context.
	if _, err := h.deals.GetByID(r.Context(), dealID); err != nil {
		handleDomainErr(w, err)
		return
	}

	links, err := h.links.ListByDeal(r.Context(), dealID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if links == nil {
		links = []*domain.PortalLink{}
	}
	writeJSON(w, http.StatusOK, links)
}

// RevokeLink handles DELETE /api/v1/portal-links/{id}.
func (h *DealPortalLinksHandler) RevokeLink(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}

	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	link, err := h.links.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if _, ok := domain.AccessContextFromContext(r.Context()); ok {
		canAccessDeal, err := h.deals.CanAccess(r.Context(), link.DealID, domain.SharingAccessWrite)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		if !canAccessDeal {
			handleDomainErr(w, domain.ErrNotFound)
			return
		}
	}

	if err := h.links.Revoke(r.Context(), id, orgID); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetSnapshot handles GET /api/portal/{token} — public, no auth required.
// Returns a curated deal snapshot; revoked/expired tokens return 404.
func (h *DealPortalLinksHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	// portal_links is not RLS-enforced, so this query works without org context.
	link, err := h.links.GetByToken(r.Context(), token)
	if err != nil {
		// Always 404 — do not leak existence.
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if !link.IsActive() {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	// Bump view counter asynchronously so the response isn't delayed.
	go func() {
		_ = h.links.IncrementView(r.Context(), link.ID)
	}()

	// Scope subsequent queries to the portal link's org.
	ctx := domain.WithOrgID(r.Context(), link.OrgID)

	deal, err := h.deals.GetByID(ctx, link.DealID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	// Fetch public (non-internal) notes for the deal.
	noteFilter := domain.NoteFilter{
		OrgID:      link.OrgID,
		EntityType: domain.NoteEntityDeal,
		EntityID:   link.DealID,
		Limit:      50,
		Page:       1,
	}
	rawNotes, _, err := h.notes.ListByEntity(ctx, noteFilter)
	if err != nil {
		rawNotes = nil
	}

	portalNotes := make([]domain.PortalNote, 0, len(rawNotes))
	for _, n := range rawNotes {
		portalNotes = append(portalNotes, domain.PortalNote{
			Body:      n.Content,
			CreatedAt: n.CreatedAt,
		})
	}

	orgName := ""
	if org, err := h.orgs.GetByID(ctx, link.OrgID); err == nil {
		orgName = org.Name
	}

	snapshot := domain.DealPortalSnapshot{
		Title:             deal.Title,
		Stage:             deal.Stage,
		ValueCents:        deal.ValueCents,
		Currency:          deal.Currency,
		ExpectedCloseDate: deal.ExpectedCloseDate,
		OrgName:           orgName,
		LinkLabel:         link.Label,
		Notes:             portalNotes,
	}

	writeJSON(w, http.StatusOK, snapshot)
}

// generateToken produces a cryptographically random URL-safe base64 token (~44 chars).
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
