package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type CurrencySettingsHandler struct {
	repo    repository.CurrencyRepository
	auditor Auditor
}

func NewCurrencySettingsHandler(repo repository.CurrencyRepository) *CurrencySettingsHandler {
	return &CurrencySettingsHandler{repo: repo}
}

func (h *CurrencySettingsHandler) WithAuditLog(r repository.AuditLogRepository) *CurrencySettingsHandler {
	h.auditor = newAuditor(r)
	return h
}

func (h *CurrencySettingsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Get)
	r.With(requireAdminRole).Patch("/", h.Update)
	return r
}

func requireAdminRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *CurrencySettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rows, err := h.repo.List(r.Context(), claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defaultCode, err := h.repo.GetDefaultCode(r.Context(), claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"default_code": defaultCode,
		"currencies":   rows,
	})
}

func (h *CurrencySettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req domain.OrgCurrencyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rows, err := h.repo.Replace(r.Context(), claims.OrgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defaultCode, err := h.repo.GetDefaultCode(r.Context(), claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if h.auditor.repo != nil {
		name := "settings.currencies"
		entry := domain.AuditEntry{
			OrgID:      claims.OrgID,
			Action:     domain.AuditActionUpdated,
			EntityType: domain.AuditEntityUser,
			EntityName: &name,
			Changes: domain.AuditChanges{
				"default_code": domain.FieldChange{To: defaultCode},
				"currencies":   domain.FieldChange{To: rows},
			},
		}
		uid := claims.UserID
		entry.UserID = &uid
		_ = h.auditor.repo.Append(r.Context(), entry)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"default_code": defaultCode,
		"currencies":   rows,
	})
}
