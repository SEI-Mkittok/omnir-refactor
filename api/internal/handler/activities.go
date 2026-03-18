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

// ActivityHandler handles HTTP requests for the activities resource.
type ActivityHandler struct {
	repo       repository.ActivityRepository
	dispatcher chan<- worker.WebhookEvent
}

func NewActivityHandler(repo repository.ActivityRepository) *ActivityHandler {
	return &ActivityHandler{repo: repo}
}

// WithDispatcher attaches the webhook event dispatcher to the handler.
func (h *ActivityHandler) WithDispatcher(d chan<- worker.WebhookEvent) *ActivityHandler {
	h.dispatcher = d
	return h
}

func (h *ActivityHandler) emitWebhook(r *http.Request, event domain.WebhookEvent, entityID uuid.UUID, data any) {
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

func (h *ActivityHandler) Router() chi.Router {
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

func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.ActivityFilter{
		Q:     q.Get("q"),
		Sort:  q.Get("sort"),
		Order: q.Get("order"),
	}
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
	if v := q.Get("type"); v != "" {
		t := domain.ActivityType(v)
		filter.Type = &t
	}
	if v := q.Get("owner_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.OwnerID = &id
		}
	}
	if v := q.Get("contact_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.ContactID = &id
		}
	}
	if v := q.Get("account_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.AccountID = &id
		}
	}
	if v := q.Get("deal_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.DealID = &id
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}
	activities, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(activities, total, filter.Page, filter.Limit))
}

func (h *ActivityHandler) Create(w http.ResponseWriter, r *http.Request) {
	var a domain.Activity
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if err := a.Validate(); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}
	created, err := h.repo.Create(r.Context(), &a)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	h.emitWebhook(r, domain.WebhookEventActivityCreated, created.ID, created)
	writeJSON(w, http.StatusCreated, created)
}

func (h *ActivityHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	a, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *ActivityHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.ActivityPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	a, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *ActivityHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
