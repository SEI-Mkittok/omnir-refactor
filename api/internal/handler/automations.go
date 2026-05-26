package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
	"github.com/omnir/crm-api/internal/worker"
)

// AutomationHandler serves /automations.
type AutomationHandler struct {
	repo   repository.AutomationRepository
	worker *worker.AutomationWorker
}

func NewAutomationHandler(repo repository.AutomationRepository) *AutomationHandler {
	return &AutomationHandler{repo: repo}
}

func (h *AutomationHandler) WithWorker(w *worker.AutomationWorker) *AutomationHandler {
	h.worker = w
	return h
}

func (h *AutomationHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/metadata", h.Metadata)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}/runs", h.ListRuns)
	r.Post("/{id}/execute", h.Execute)
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
		handleDomainErr(w, err)
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
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		a.CreatedBy = &claims.UserID
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
	current, err := h.repo.GetAutomation(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if req.Status != nil && !req.Status.IsValid() {
		handleDomainErr(w, fmt.Errorf("%w: status is invalid", domain.ErrValidation))
		return
	}
	trigger := current.Trigger
	if req.Trigger != nil {
		trigger = *req.Trigger
	}
	conditions := current.Conditions
	if req.Conditions != nil {
		conditions = req.Conditions
	}
	actions := current.Actions
	if req.Actions != nil {
		actions = req.Actions
	}
	if len(actions) == 0 {
		handleDomainErr(w, fmt.Errorf("%w: at least one action is required", domain.ErrValidation))
		return
	}
	if err := domain.ValidateAutomationConfig(&trigger, conditions, actions); err != nil {
		handleDomainErr(w, err)
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

func (h *AutomationHandler) Metadata(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"triggers": []map[string]string{
			{"type": string(domain.TriggerContactCreated), "label": "Contact created"},
			{"type": string(domain.TriggerContactUpdated), "label": "Contact updated"},
			{"type": string(domain.TriggerDealCreated), "label": "Deal created"},
			{"type": string(domain.TriggerDealStageChanged), "label": "Deal stage changed"},
			{"type": string(domain.TriggerActivityOverdue), "label": "Activity overdue"},
			{"type": string(domain.TriggerTicketCreated), "label": "Ticket created"},
			{"type": string(domain.TriggerManual), "label": "Manual"},
		},
		"operators": []string{
			string(domain.ConditionOpEquals),
			string(domain.ConditionOpNotEquals),
			string(domain.ConditionOpContains),
			string(domain.ConditionOpNotContains),
			string(domain.ConditionOpGreaterThan),
			string(domain.ConditionOpLessThan),
			string(domain.ConditionOpIsSet),
			string(domain.ConditionOpIsNotSet),
		},
		"actions": []map[string]string{
			{"type": string(domain.ActionAssignOwner), "label": "Assign owner"},
			{"type": string(domain.ActionSendEmail), "label": "Send email"},
			{"type": string(domain.ActionEnrollInSequence), "label": "Enroll in sequence"},
			{"type": string(domain.ActionCreateActivity), "label": "Create activity"},
			{"type": string(domain.ActionWebhook), "label": "Webhook"},
		},
		"fields": map[string][]string{
			"contact":  {"first_name", "last_name", "email", "owner_id", "stage", "lead_source"},
			"deal":     {"title", "stage_id", "owner_id", "value_cents"},
			"ticket":   {"subject", "status", "priority", "assignee_id", "source"},
			"activity": {"subject", "owner_id", "due_date"},
		},
	})
}

func (h *AutomationHandler) Execute(w http.ResponseWriter, r *http.Request) {
	if h.worker == nil {
		writeError(w, http.StatusServiceUnavailable, "automation worker unavailable")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		EntityType string                 `json:"entity_type"`
		EntityID   *uuid.UUID             `json:"entity_id,omitempty"`
		Data       map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	a, err := h.repo.GetAutomation(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if req.Data == nil {
		req.Data = map[string]interface{}{}
	}
	entityID := uuid.Nil
	if req.EntityID != nil {
		entityID = *req.EntityID
	}
	orgID, _ := domain.OrgIDFromContext(r.Context())
	run, execErr := h.worker.ExecuteAutomation(r.Context(), a, worker.AutomationEvent{
		OrgID:       orgID,
		TriggerType: domain.TriggerManual,
		EntityID:    entityID,
		EntityType:  req.EntityType,
		Data:        req.Data,
	})
	if run == nil && execErr != nil {
		handleDomainErr(w, execErr)
		return
	}
	writeJSON(w, http.StatusAccepted, run)
}
