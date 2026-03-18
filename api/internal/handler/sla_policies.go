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

// SLAPolicyHandler serves the /sla-policies resource.
type SLAPolicyHandler struct {
	policies repository.SLAPolicyRepository
}

func NewSLAPolicyHandler(policies repository.SLAPolicyRepository) *SLAPolicyHandler {
	return &SLAPolicyHandler{policies: policies}
}

func (h *SLAPolicyHandler) Router() chi.Router {
	r := chi.NewRouter()

	agentOnly := middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent)

	r.With(agentOnly).Get("/", h.List)
	r.With(agentOnly).Post("/", h.Create)
	r.With(agentOnly).Get("/{id}", h.GetByID)
	r.With(agentOnly).Put("/{id}", h.Update)
	r.With(agentOnly).Delete("/{id}", h.Delete)

	return r
}

func (h *SLAPolicyHandler) List(w http.ResponseWriter, r *http.Request) {
	policies, err := h.policies.List(r.Context(), uuid.Nil)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

func (h *SLAPolicyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p domain.SLAPolicy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if p.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	if p.ResponseTimeHours <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "response_time_hours must be greater than 0")
		return
	}
	if p.ResolutionTimeHours <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "resolution_time_hours must be greater than 0")
		return
	}

	created, err := h.policies.Create(r.Context(), &p)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *SLAPolicyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	p, err := h.policies.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *SLAPolicyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.SLAPolicyPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if patch.ResponseTimeHours != nil && *patch.ResponseTimeHours <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "response_time_hours must be greater than 0")
		return
	}
	if patch.ResolutionTimeHours != nil && *patch.ResolutionTimeHours <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "resolution_time_hours must be greater than 0")
		return
	}

	p, err := h.policies.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *SLAPolicyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	if err := h.policies.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
