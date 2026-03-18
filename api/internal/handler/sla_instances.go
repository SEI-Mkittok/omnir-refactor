package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// SLAInstanceHandler serves /sla-instances and /sla-dashboard.
type SLAInstanceHandler struct {
	instances repository.SLAInstanceRepository
}

func NewSLAInstanceHandler(instances repository.SLAInstanceRepository) *SLAInstanceHandler {
	return &SLAInstanceHandler{instances: instances}
}

func (h *SLAInstanceHandler) Router() chi.Router {
	r := chi.NewRouter()
	agentOnly := middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent)
	r.With(agentOnly).Get("/", h.List)
	r.With(agentOnly).Get("/{id}", h.GetByID)
	return r
}

func (h *SLAInstanceHandler) DashboardRouter() chi.Router {
	r := chi.NewRouter()
	agentOnly := middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent)
	r.With(agentOnly).Get("/", h.Dashboard)
	return r
}

func (h *SLAInstanceHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.SLAInstanceFilter{Limit: 50}

	if v := q.Get("entity_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.EntityID = &id
		}
	}
	if v := q.Get("entity_type"); v != "" {
		filter.EntityType = domain.SLAEntityType(v)
	}
	if v := q.Get("breached"); v == "true" {
		t := true
		filter.Breached = &t
	} else if v == "false" {
		f := false
		filter.Breached = &f
	}

	instances, err := h.instances.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, instances)
}

func (h *SLAInstanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	inst, err := h.instances.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inst)
}

func (h *SLAInstanceHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	summary, err := h.instances.Dashboard(r.Context())
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
