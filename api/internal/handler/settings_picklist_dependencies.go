package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type PicklistDependencySettingsHandler struct {
	repo repository.PicklistRepository
}

func NewPicklistDependencySettingsHandler(repo repository.PicklistRepository) *PicklistDependencySettingsHandler {
	return &PicklistDependencySettingsHandler{repo: repo}
}

func (h *PicklistDependencySettingsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.With(requireAdminRole).Get("/", h.List)
	r.With(requireAdminRole).Put("/", h.Upsert)
	r.With(requireAdminRole).Delete("/{dependencyID}", h.Delete)
	return r
}

func (h *PicklistDependencySettingsHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	entityTypeParam := r.URL.Query().Get("entity_type")
	var entityType *domain.CustomFieldEntityType
	if entityTypeParam != "" {
		v := domain.CustomFieldEntityType(entityTypeParam)
		if !v.IsValid() {
			writeError(w, http.StatusUnprocessableEntity, "invalid entity_type")
			return
		}
		entityType = &v
	}
	rows, err := h.repo.ListDependencies(r.Context(), claims.OrgID, entityType)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": rows})
}

func (h *PicklistDependencySettingsHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req domain.PicklistDependencyInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.repo.UpsertDependency(r.Context(), claims.OrgID, req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PicklistDependencySettingsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "dependencyID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dependency id")
		return
	}
	if err := h.repo.DeleteDependency(r.Context(), claims.OrgID, id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

