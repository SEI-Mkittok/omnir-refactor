package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	"github.com/omnir/crm-api/internal/repository"
)

const inviteTokenTTL = 72 * time.Hour

// OnboardingHandler manages wizard state and team invites for a new org.
type OnboardingHandler struct {
	repo   repository.OnboardingRepository
	users  repository.UserRepository
	orgs   repository.OrgRepository
	mailer *email.Mailer
	appURL string
}

func NewOnboardingHandler(
	repo repository.OnboardingRepository,
	users repository.UserRepository,
	orgs repository.OrgRepository,
	mailer *email.Mailer,
	appURL string,
) *OnboardingHandler {
	return &OnboardingHandler{repo: repo, users: users, orgs: orgs, mailer: mailer, appURL: appURL}
}

func (h *OnboardingHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Get)
	r.Patch("/", h.Update)
	r.Post("/invite", h.SendInvite)
	r.Post("/accept", h.AcceptInvite)
	r.Get("/status", h.Status)
	return r
}

// onboardingResponse is the shape returned to the frontend:
// matches the TypeScript OnboardingState interface.
type onboardingResponse struct {
	ID             string   `json:"id"`
	CompletedSteps []string `json:"completedSteps"`
	Completed      bool     `json:"completed"`
	Dismissed      bool     `json:"dismissed"`
	OrgName        string   `json:"orgName,omitempty"`
}

func toOnboardingResponse(state *domain.OrgOnboarding, orgName string) onboardingResponse {
	steps := state.CompletedSteps
	if steps == nil {
		steps = []string{}
	}
	return onboardingResponse{
		ID:             state.OrgID.String(),
		CompletedSteps: steps,
		Completed:      state.CompletedAt != nil,
		Dismissed:      state.Dismissed,
		OrgName:        orgName,
	}
}

// Get returns the current onboarding state for the caller's org.
// GET /api/v1/onboarding
func (h *OnboardingHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing org context")
		return
	}

	state, err := h.repo.GetOrCreate(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	orgName := ""
	if h.orgs != nil {
		if org, err := h.orgs.GetByID(r.Context(), orgID); err == nil {
			orgName = org.Name
		}
	}

	writeJSON(w, http.StatusOK, toOnboardingResponse(state, orgName))
}

// onboardingPatchRequest uses camelCase JSON tags to match the frontend payload.
type onboardingPatchRequest struct {
	CompletedSteps []string `json:"completedSteps"`
	Completed      bool     `json:"completed"`
	Dismissed      *bool    `json:"dismissed"`
	OrgName        string   `json:"orgName"`
}

// Update persists onboarding step progress and optionally updates the org name.
// PATCH /api/v1/onboarding
func (h *OnboardingHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing org context")
		return
	}

	var req onboardingPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Persist org name if provided.
	orgName := ""
	if orgName = strings.TrimSpace(req.OrgName); orgName != "" && h.orgs != nil {
		if err := h.orgs.UpdateName(r.Context(), orgID, orgName); err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	var completedAt *time.Time
	if req.Completed {
		now := time.Now().UTC()
		completedAt = &now
	}

	state, err := h.repo.UpdateSteps(r.Context(), orgID, req.CompletedSteps, completedAt, req.Dismissed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toOnboardingResponse(state, orgName))
}

type inviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// SendInvite emails a magic-link invite to a new team member.
// POST /api/v1/onboarding/invite
func (h *OnboardingHandler) SendInvite(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing org context")
		return
	}

	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusUnprocessableEntity, "validation error: email is invalid")
		return
	}
	if req.Role == "" {
		req.Role = "agent"
	}

	token, err := generateInviteToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	invite := &domain.OrgInvite{
		OrgID:     orgID,
		Email:     req.Email,
		Role:      req.Role,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(inviteTokenTTL),
	}

	saved, err := h.repo.CreateInvite(r.Context(), invite)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Best-effort email — log but don't fail the request.
	acceptURL := fmt.Sprintf("%s/onboarding/accept?token=%s", strings.TrimRight(h.appURL, "/"), token)
	_ = h.mailer.SendDirect(
		req.Email,
		"You've been invited to join Omnir",
		fmt.Sprintf("You've been invited to join the team. Accept your invite here:\n\n%s\n\nThis link expires in 72 hours.", acceptURL),
	)

	// Don't expose the raw token in the list response.
	saved.Token = ""
	writeJSON(w, http.StatusCreated, saved)
}

type acceptInviteRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// AcceptInvite accepts a magic-link invite and creates the new user account.
// POST /api/v1/onboarding/accept
func (h *OnboardingHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req acceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	req.Name = strings.TrimSpace(req.Name)
	if req.Token == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation error: token is required")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation error: name is required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusUnprocessableEntity, "validation error: password must be at least 8 characters")
		return
	}

	invite, err := h.repo.GetInviteByToken(r.Context(), req.Token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if invite == nil {
		writeError(w, http.StatusNotFound, "invite not found or already used")
		return
	}
	if invite.AcceptedAt != nil {
		writeError(w, http.StatusConflict, "invite has already been accepted")
		return
	}
	if time.Now().UTC().After(invite.ExpiresAt) {
		writeError(w, http.StatusGone, "invite has expired")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	ctx := domain.WithOrgID(r.Context(), invite.OrgID)
	user := &domain.User{
		OrgID: invite.OrgID,
		Email: invite.Email,
		Name:  req.Name,
		Role:  domain.UserRole(invite.Role),
	}
	created, err := h.users.Create(ctx, user, string(hash))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := h.repo.AcceptInvite(r.Context(), invite.ID); err != nil {
		// Non-fatal: user is created; mark accept best-effort.
		_ = err
	}

	writeJSON(w, http.StatusCreated, map[string]any{"user": created})
}

// Status returns a completion summary for the onboarding wizard.
// GET /api/v1/onboarding/status
func (h *OnboardingHandler) Status(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing org context")
		return
	}

	state, err := h.repo.GetOrCreate(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	invites, err := h.repo.ListInvites(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if invites == nil {
		invites = []*domain.OrgInvite{}
	}
	// Strip tokens from list output.
	for _, inv := range invites {
		inv.Token = ""
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"completed_steps": state.CompletedSteps,
		"completed":       state.CompletedAt != nil,
		"completed_at":    state.CompletedAt,
		"invites":         invites,
	})
}

func generateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
