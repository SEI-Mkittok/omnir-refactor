package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
	"github.com/omnir/crm-api/internal/worker"
)

type DealHandler struct {
	repo          repository.DealRepository
	cfDefs        repository.CustomFieldDefinitionRepository
	dispatcher    chan<- worker.WebhookEvent
	automations   chan<- worker.AutomationEvent
	notifications repository.NotificationRepository
	slaInstances  repository.SLAInstanceRepository
	slaPolicies   repository.SLAPolicyRepository
	teamsNotifier *worker.TeamsNotifier
	pushNotifier  *worker.PushNotifier
}

func NewDealHandler(repo repository.DealRepository) *DealHandler {
	return &DealHandler{repo: repo}
}

func (h *DealHandler) WithCustomFields(r repository.CustomFieldDefinitionRepository) *DealHandler {
	h.cfDefs = r
	return h
}

func (h *DealHandler) WithDispatcher(d chan<- worker.WebhookEvent) *DealHandler {
	h.dispatcher = d
	return h
}

func (h *DealHandler) WithAutomationEvents(ch chan<- worker.AutomationEvent) *DealHandler {
	h.automations = ch
	return h
}

func (h *DealHandler) WithNotifications(r repository.NotificationRepository) *DealHandler {
	h.notifications = r
	return h
}

func (h *DealHandler) WithSLA(policies repository.SLAPolicyRepository, instances repository.SLAInstanceRepository) *DealHandler {
	h.slaPolicies = policies
	h.slaInstances = instances
	return h
}

func (h *DealHandler) WithTeamsNotifier(n *worker.TeamsNotifier) *DealHandler {
	h.teamsNotifier = n
	return h
}

func (h *DealHandler) WithPushNotifier(n *worker.PushNotifier) *DealHandler {
	h.pushNotifier = n
	return h
}

func (h *DealHandler) emitWebhook(r *http.Request, event domain.WebhookEvent, entityID uuid.UUID, data any) {
	if h.dispatcher == nil {
		return
	}
	orgID, _ := domain.OrgIDFromContext(r.Context())
	evt := worker.WebhookEvent{OrgID: orgID, EntityID: entityID, Event: event, Data: data}
	select {
	case h.dispatcher <- evt:
	default:
	}
}

func (h *DealHandler) emitAutomation(r *http.Request, trigger domain.TriggerType, entityID uuid.UUID, data map[string]interface{}) {
	if h.automations == nil {
		return
	}
	orgID, _ := domain.OrgIDFromContext(r.Context())
	evt := worker.AutomationEvent{OrgID: orgID, TriggerType: trigger, EntityID: entityID, EntityType: "deal", Data: data}
	select {
	case h.automations <- evt:
	default:
	}
}

func dealToData(d *domain.Deal) map[string]interface{} {
	return map[string]interface{}{
		"id":       d.ID.String(),
		"org_id":   d.OrgID.String(),
		"stage":    string(d.Stage),
		"owner_id": d.OwnerID.String(),
		"title":    d.Title,
	}
}

func (h *DealHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Create)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/contacts", h.AddContact)
	})
	return r
}

func (h *DealHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.DealFilter{Q: q.Get("q"), Sort: q.Get("sort"), Order: q.Get("order")}
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	// Accept ?per_page (frontend) or ?limit (canonical).
	if limitVal := q.Get("per_page"); limitVal == "" {
		if v := q.Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n <= 200 {
				filter.Limit = n
			}
		}
	} else if n, err := strconv.Atoi(limitVal); err == nil && n <= 200 {
		filter.Limit = n
	}
	if v := q.Get("owner_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.OwnerID = &id
		}
	}
	if v := q.Get("stage"); v != "" {
		s := domain.DealStage(v)
		filter.Stage = &s
	}
	if v := q.Get("account_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.AccountID = &id
		}
	}
	if v := q.Get("contact_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.ContactID = &id
		}
	}
	if v := q.Get("pipeline_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.PipelineID = &id
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}
	deals, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(deals, total, filter.Page, filter.Limit))
}

func (h *DealHandler) Create(w http.ResponseWriter, r *http.Request) {
	var d domain.Deal
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if d.Stage == "" {
		d.Stage = domain.DealStageLead
	}
	if d.Currency == "" {
		d.Currency = "USD"
	}
	if err := d.Validate(); err != nil {
		handleDomainErr(w, err)
		return
	}
	created, err := h.repo.Create(r.Context(), &d)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	h.emitWebhook(r, domain.WebhookEventDealCreated, created.ID, created)
	h.emitAutomation(r, domain.TriggerDealCreated, created.ID, dealToData(created))
	h.attachSLAInstances(r.Context(), created.ID, domain.SLAEntityTypeDeal, created.CreatedAt)
	writeJSON(w, http.StatusCreated, created)
}

func (h *DealHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	d, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if h.cfDefs != nil {
		et := domain.CustomFieldEntityDeal
		defs, err := h.cfDefs.List(r.Context(), domain.CustomFieldDefinitionFilter{EntityType: &et})
		if err == nil && len(defs) > 0 {
			d.CustomFields = domain.ExpandCustomFields(d.CustomFields, defs)
		}
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *DealHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.DealPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if h.cfDefs != nil && len(patch.CustomFields) > 0 {
		et := domain.CustomFieldEntityDeal
		defs, err := h.cfDefs.List(r.Context(), domain.CustomFieldDefinitionFilter{EntityType: &et})
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		if err := domain.ValidateCustomFields(patch.CustomFields, defs); err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	}
	d, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if patch.Stage != nil && h.notifications != nil {
		title := "Deal stage changed"
		entityType := "deal"
		_, _ = h.notifications.Create(r.Context(), &domain.Notification{
			OrgID:      d.OrgID,
			UserID:     d.OwnerID,
			Kind:       domain.NotificationKindDealStageChanged,
			EntityType: &entityType,
			EntityID:   &d.ID,
			Title:      title,
		})
	}
	h.emitWebhook(r, domain.WebhookEventDealUpdated, d.ID, d)
	if patch.Stage != nil {
		h.emitWebhook(r, domain.WebhookEventDealStageChanged, d.ID, d)
		h.emitAutomation(r, domain.TriggerDealStageChanged, d.ID, dealToData(d))
		if h.teamsNotifier != nil {
			h.teamsNotifier.NotifyDealStageChanged(d.OrgID, d.ID, d.Title, string(d.Stage))
		}
		if h.pushNotifier != nil && h.pushNotifier.Enabled() {
			h.pushNotifier.NotifyDealStageChanged(d.OrgID, d.OwnerID, d.ID, d.Title, string(d.Stage))
		}
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *DealHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	h.emitWebhook(r, domain.WebhookEventDealDeleted, id, map[string]any{"id": id})
	w.WriteHeader(http.StatusNoContent)
}

func (h *DealHandler) AddContact(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid deal id")
		return
	}
	var body struct {
		ContactID uuid.UUID `json:"contact_id"`
		Role      string    `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if body.ContactID == uuid.Nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "contact_id is required")
		return
	}
	if err := h.repo.AddContact(r.Context(), dealID, body.ContactID, body.Role); err != nil {
		handleDomainErr(w, err)
		return
	}
	deal, err := h.repo.GetByID(r.Context(), dealID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, deal)
}

// attachSLAInstances creates SLA instance rows for all matching deal policies.
// Errors are logged and silently ignored so they don't fail the primary request.
func (h *DealHandler) attachSLAInstances(ctx context.Context, entityID uuid.UUID, entityType domain.SLAEntityType, startedAt time.Time) {
	if h.slaPolicies == nil || h.slaInstances == nil {
		return
	}
	policies, err := h.slaPolicies.MatchForEntity(ctx, entityType)
	if err != nil {
		slog.Default().Error("sla policy match failed", "entity_type", entityType, "err", err)
		return
	}
	for _, policy := range policies {
		responseWindow := time.Duration(policy.ResponseTimeHours * float64(time.Hour))
		resolutionWindow := time.Duration(policy.ResolutionTimeHours * float64(time.Hour))
		inst := &domain.SLAInstance{
			PolicyID:        policy.ID,
			EntityID:        entityID,
			EntityType:      entityType,
			ResponseDueAt:   startedAt.Add(responseWindow),
			ResolutionDueAt: startedAt.Add(resolutionWindow),
		}
		if _, err := h.slaInstances.Create(ctx, inst); err != nil {
			slog.Default().Error("sla instance create failed", "policy_id", policy.ID, "entity_id", entityID, "err", err)
		}
	}
}
