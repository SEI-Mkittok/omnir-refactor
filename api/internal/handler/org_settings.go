package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// OrgSettingsHandler handles org-level settings such as document numbering configuration.
type OrgSettingsHandler struct {
	settings repository.OrgSettingsRepository
}

func NewOrgSettingsHandler(settings repository.OrgSettingsRepository) *OrgSettingsHandler {
	return &OrgSettingsHandler{settings: settings}
}

func (h *OrgSettingsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAdmin)
	r.Get("/", h.Get)
	r.Patch("/", h.Update)
	return r
}

func (h *OrgSettingsHandler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.ClaimsFromContext(r)
		if !ok || !domain.IsAdminRole(claims.Role) {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Get returns current org settings (including numbering start values).
// GET /api/v1/settings/numbering
func (h *OrgSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r)
	s, err := h.settings.GetOrCreate(r.Context(), claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// Update patches org settings fields.
// PATCH /api/v1/settings/numbering
func (h *OrgSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r)

	var patch domain.OrgSettingsPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate: starting numbers must be >= 1.
	for _, v := range []*int64{
		patch.QuoteNumberStart, patch.TicketNumberStart,
		patch.KBArticleNumberStart, patch.InvoiceNumberStart,
	} {
		if v != nil && *v < 1 {
			writeError(w, http.StatusUnprocessableEntity, "starting numbers must be >= 1")
			return
		}
	}

	s, err := h.settings.Update(r.Context(), claims.OrgID, patch)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, s)
}
