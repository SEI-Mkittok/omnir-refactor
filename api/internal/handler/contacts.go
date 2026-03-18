package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
	"github.com/omnir/crm-api/internal/worker"
)

type ContactHandler struct {
	repo       repository.ContactRepository
	dispatcher chan<- worker.WebhookEvent
	cfDefs     repository.CustomFieldDefinitionRepository
}

func NewContactHandler(repo repository.ContactRepository) *ContactHandler {
	return &ContactHandler{repo: repo}
}

func (h *ContactHandler) WithCustomFields(r repository.CustomFieldDefinitionRepository) *ContactHandler {
	h.cfDefs = r
	return h
}

func (h *ContactHandler) WithDispatcher(d chan<- worker.WebhookEvent) *ContactHandler {
	h.dispatcher = d
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

func (h *ContactHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Create)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
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
