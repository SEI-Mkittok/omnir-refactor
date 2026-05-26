package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/worker"
)

type SchedulerHandler struct {
	registry *worker.SchedulerRegistry
}

func NewSchedulerHandler(registry *worker.SchedulerRegistry) *SchedulerHandler {
	return &SchedulerHandler{registry: registry}
}

func (h *SchedulerHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/jobs", h.ListJobs)
	r.Get("/jobs/{jobKey}/runs", h.ListRuns)
	return r
}

func (h *SchedulerHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": h.registry.Jobs()})
}

func (h *SchedulerHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": h.registry.Runs(chi.URLParam(r, "jobKey"))})
}
