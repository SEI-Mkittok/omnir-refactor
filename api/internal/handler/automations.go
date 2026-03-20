package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// AutomationHandler serves /automations.
type AutomationHandler struct {
	repo repository.AutomationRepository
}

func NewAutomationHandler(repo repository.AutomationRepository) *AutomationHandler {
	return &AutomationHandler{repo: repo}
}

func (h *AutomationHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}/runs", h.ListRuns)
	return r
}

// List handles GET /automations
func (h *AutomationHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := domain.AutomationFilter{Page: 1, Limit: 50}
	q := r.URL.Query()
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Page = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			filter.Limit = n
		}
	}
	if v := q.Get("status"); v != "" {
		st := domain.AutomationStatus(v)
		filter.Status = &st
	}

	list, total, err := h.repo.ListAutomations(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list automations")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": list, "total": total})
}

// Create handles POST /automations
func (h *AutomationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAutomationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a := &domain.Automation{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Trigger:     req.Trigger,
		Conditions:  req.Conditions,
		Actions:     req.Actions,
	}
	if a.Conditions == nil {
		a.Conditions = []domain.AutomationCondition{}
	}

	created, err := h.repo.CreateAutomation(r.Context(), a)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create automation")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Get handles GET /automations/{id}
func (h *AutomationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	a, err := h.repo.GetAutomation(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "automation not found")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

// Update handles PATCH /automations/{id}
func (h *AutomationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req domain.UpdateAutomationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.UpdateAutomation(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update automation")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /automations/{id}
func (h *AutomationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.DeleteAutomation(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete automation")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListRuns handles GET /automations/{id}/runs
func (h *AutomationHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	filter := domain.AutomationRunFilter{AutomationID: id, Page: 1, Limit: 50}
	q := r.URL.Query()
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Page = n
		}
	}

	runs, total, err := h.repo.ListRuns(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list runs")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": runs, "total": total})
}
