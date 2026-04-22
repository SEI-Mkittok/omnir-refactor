package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
	"github.com/omnir/crm-api/internal/worker"
)

var allowedRelationshipTypes = map[string]struct{}{
	"decision_maker": {},
	"billing":        {},
	"technical":      {},
	"executive":      {},
	"champion":       {},
}

func normalizeRelationshipType(v *string) {
	if v == nil {
		return
	}
	n := strings.ToLower(strings.TrimSpace(*v))
	*v = n
}

func isValidRelationshipType(v *string) bool {
	if v == nil || *v == "" {
		return true
	}
	normalized := strings.TrimSpace(strings.ToLower(*v))
	if normalized == "" {
		return false
	}
	_, ok := allowedRelationshipTypes[normalized]
	return ok
}

func normalizeRelationshipType(v *string) *string {
	if v == nil {
		return nil
	}
	normalized := strings.ToLower(strings.TrimSpace(*v))
	return &normalized
}

type ContactHandler struct {
	repo        repository.ContactRepository
	deals       repository.DealRepository
	cfDefs      repository.CustomFieldDefinitionRepository
	dispatcher  chan<- worker.WebhookEvent
	automations chan<- worker.AutomationEvent
}

func NewContactHandler(repo repository.ContactRepository) *ContactHandler {
	return &ContactHandler{repo: repo}
}

func (h *ContactHandler) WithCustomFields(r repository.CustomFieldDefinitionRepository) *ContactHandler {
	h.cfDefs = r
	return h
}

func (h *ContactHandler) WithDeals(d repository.DealRepository) *ContactHandler {
	h.deals = d
	return h
}

func (h *ContactHandler) WithDispatcher(d chan<- worker.WebhookEvent) *ContactHandler {
	h.dispatcher = d
	return h
}

func (h *ContactHandler) WithAutomationEvents(ch chan<- worker.AutomationEvent) *ContactHandler {
	h.automations = ch
	return h
}

func (h *ContactHandler) emitWebhook(r *http.Request, event domain.WebhookEvent, entityID uuid.UUID, data any) {
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

func (h *ContactHandler) emitAutomation(r *http.Request, trigger domain.TriggerType, entityID uuid.UUID, entityType string, data map[string]interface{}) {
	if h.automations == nil {
		return
	}
	orgID, _ := domain.OrgIDFromContext(r.Context())
	evt := worker.AutomationEvent{OrgID: orgID, TriggerType: trigger, EntityID: entityID, EntityType: entityType, Data: data}
	select {
	case h.automations <- evt:
	default:
	}
}

func contactToData(c *domain.Contact) map[string]interface{} {
	d := map[string]interface{}{
		"id":         c.ID.String(),
		"org_id":     c.OrgID.String(),
		"stage":      string(c.Stage),
		"owner_id":   c.OwnerID.String(),
		"first_name": c.FirstName,
		"last_name":  c.LastName,
	}
	if c.Email != nil {
		d["email"] = *c.Email
	}
	return d
}

func (h *ContactHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/lead-sources", h.LeadSources)
	r.Get("/{id}", h.GetByID)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Create)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Patch("/{id}/score", h.UpdateLeadScore)
		r.Post("/{id}/convert", h.ConvertLead)
	})
	return r
}

func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.ContactFilter{Q: q.Get("q"), Sort: q.Get("sort"), Order: q.Get("order")}
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n <= 200 {
			filter.Limit = n
		}
	}
	if v := q.Get("owner_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.OwnerID = &id
		}
	}
	if v := q.Get("stage"); v != "" {
		s := domain.ContactStage(v)
		filter.Stage = &s
	}
	if v := q.Get("source"); v != "" {
		filter.Source = &v
	}
	if v := q.Get("score_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.ScoreMin = &n
		}
	}
	if v := q.Get("score_max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.ScoreMax = &n
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}
	contacts, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(contacts, total, filter.Page, filter.Limit))
}

func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c domain.Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if c.Stage == "" {
		c.Stage = domain.ContactStageLead
	}
	normalizeRelationshipType(c.RelationshipType)
	if !isValidRelationshipType(c.RelationshipType) {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid relationship_type")
		return
	}
	if err := c.Validate(); err != nil {
		handleDomainErr(w, err)
		return
	}
	created, err := h.repo.Create(r.Context(), &c)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	h.emitWebhook(r, domain.WebhookEventContactCreated, created.ID, created)
	h.emitAutomation(r, domain.TriggerContactCreated, created.ID, "contact", contactToData(created))
	writeJSON(w, http.StatusCreated, created)
}

func (h *ContactHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	c, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if h.cfDefs != nil {
		et := domain.CustomFieldEntityContact
		defs, err := h.cfDefs.List(r.Context(), domain.CustomFieldDefinitionFilter{EntityType: &et})
		if err == nil && len(defs) > 0 {
			c.CustomFields = domain.ExpandCustomFields(c.CustomFields, defs)
		}
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.ContactPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	normalizeRelationshipType(patch.RelationshipType)
	if !isValidRelationshipType(patch.RelationshipType) {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid relationship_type")
		return
	}
	if h.cfDefs != nil && len(patch.CustomFields) > 0 {
		et := domain.CustomFieldEntityContact
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
	c, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	h.emitWebhook(r, domain.WebhookEventContactUpdated, c.ID, c)
	writeJSON(w, http.StatusOK, c)
}

func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// LeadSources returns distinct lead_source values for contacts with stage='lead'.
// GET /api/v1/contacts/lead-sources
func (h *ContactHandler) LeadSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.repo.ListLeadSources(r.Context())
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sources": sources})
}

// UpdateLeadScore updates the lead_score on a contact (absolute or delta).
// PATCH /api/v1/contacts/{id}/score
func (h *ContactHandler) UpdateLeadScore(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.LeadScorePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if patch.Score == nil && patch.Delta == nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "score or delta is required")
		return
	}
	contact, err := h.repo.UpdateLeadScore(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, contact)
}

// ConvertLead promotes a contact from stage='lead' to stage='prospect',
// optionally creating a deal linked to the contact.
// POST /api/v1/contacts/{id}/convert
func (h *ContactHandler) ConvertLead(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}

	var req domain.LeadConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	// Resolve converting user from JWT claims.
	byUserID := uuid.Nil
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		byUserID = claims.UserID
	}

	var dealID *uuid.UUID
	if req.CreateDeal && h.deals != nil {
		contact, err := h.repo.GetByID(r.Context(), id)
		if err != nil {
			handleDomainErr(w, err)
			return
		}

		title := req.DealTitle
		if title == "" {
			title = contact.FirstName + " " + contact.LastName + " — Deal"
		}

		pipelineID := uuid.Nil
		if req.PipelineID != nil {
			pipelineID = *req.PipelineID
		}

		deal := &domain.Deal{
			Title:      title,
			Stage:      domain.DealStageLead,
			OwnerID:    contact.OwnerID,
			ContactID:  &contact.ID,
			PipelineID: pipelineID,
		}
		created, err := h.deals.Create(r.Context(), deal)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		dealID = &created.ID
	}

	contact, err := h.repo.ConvertLead(r.Context(), id, byUserID, dealID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	resp := map[string]any{"contact": contact}
	if dealID != nil {
		resp["deal_id"] = dealID
	}
	writeJSON(w, http.StatusOK, resp)
}
