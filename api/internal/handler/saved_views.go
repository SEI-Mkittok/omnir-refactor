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

type SavedViewHandler struct {
	repo repository.SavedViewRepository
}

func NewSavedViewHandler(repo repository.SavedViewRepository) *SavedViewHandler {
	return &SavedViewHandler{repo: repo}
}

func (h *SavedViewHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/pin", h.Pin)
	return r
}

func (h *SavedViewHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entityType := domain.SavedViewEntityType(r.URL.Query().Get("entityType"))
	if !entityType.IsValid() {
		writeError(w, http.StatusBadRequest, "entityType must be one of: contacts, accounts, deals, leads")
		return
	}

	orgID, _ := domain.OrgIDFromContext(r.Context())
	filter := domain.SavedViewFilter{
		OrgID:      orgID,
		UserID:     claims.UserID,
		EntityType: entityType,
	}

	views, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if views == nil {
		views = []*domain.SavedView{}
	}
	writeJSON(w, http.StatusOK, views)
}

func (h *SavedViewHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var v domain.SavedView
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	v.CreatedBy = claims.UserID
	if err := v.Validate(); err != nil {
		handleDomainErr(w, err)
		return
	}

	created, err := h.repo.Create(r.Context(), &v)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *SavedViewHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	v, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	// Access: own view, shared view, or admin
	claims, _ := middleware.ClaimsFromContext(r)
	if !v.IsShared && v.CreatedBy != claims.UserID && claims.Role != string(domain.UserRoleAdmin) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	writeJSON(w, http.StatusOK, v)
}

func (h *SavedViewHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if existing.CreatedBy != claims.UserID && claims.Role != string(domain.UserRoleAdmin) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var patch domain.SavedViewPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if patch.SortDir != nil && *patch.SortDir != "asc" && *patch.SortDir != "desc" {
		writeError(w, http.StatusUnprocessableEntity, "sort_dir must be asc or desc")
		return
	}

	updated, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *SavedViewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if existing.CreatedBy != claims.UserID && claims.Role != string(domain.UserRoleAdmin) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SavedViewHandler) Pin(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if existing.CreatedBy != claims.UserID && claims.Role != string(domain.UserRoleAdmin) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Toggle the pinned state.
	updated, err := h.repo.Pin(r.Context(), id, !existing.IsPinned)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
