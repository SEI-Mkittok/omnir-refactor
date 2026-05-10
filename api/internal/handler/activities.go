package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
	"github.com/omnir/crm-api/internal/worker"
)

type activityCreateRequest struct {
	Type        domain.ActivityType `json:"type"`
	Subject     string              `json:"subject"`
	Description *string             `json:"description,omitempty"`
	DueDate     *string             `json:"due_date,omitempty"`
	StartAt     *string             `json:"start_at,omitempty"`
	EndAt       *string             `json:"end_at,omitempty"`
	Completed   *bool               `json:"completed,omitempty"`
	ContactID   *uuid.UUID          `json:"contact_id,omitempty"`
	AccountID   *uuid.UUID          `json:"account_id,omitempty"`
	DealID      *uuid.UUID          `json:"deal_id,omitempty"`
	OwnerID     *uuid.UUID          `json:"owner_id,omitempty"`
}

type activityUpdateRequest struct {
	Type        *domain.ActivityType `json:"type,omitempty"`
	Subject     *string              `json:"subject,omitempty"`
	Description *string              `json:"description,omitempty"`
	DueDate     *string              `json:"due_date,omitempty"`
	StartAt     *string              `json:"start_at,omitempty"`
	EndAt       *string              `json:"end_at,omitempty"`
	Completed   *bool                `json:"completed,omitempty"`
	ContactID   *uuid.UUID           `json:"contact_id,omitempty"`
	AccountID   *uuid.UUID           `json:"account_id,omitempty"`
	DealID      *uuid.UUID           `json:"deal_id,omitempty"`
	OwnerID     *uuid.UUID           `json:"owner_id,omitempty"`
}

type ActivityHandler struct {
	repo       repository.ActivityRepository
	contacts   repository.ContactRepository
	dispatcher chan<- worker.WebhookEvent
}

func parseActivityTime(field string, raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	value := strings.TrimSpace(*raw)
	if value == "" {
		return nil, nil
	}

	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}
	for _, layout := range formats {
		if parsed, err := time.Parse(layout, value); err == nil {
			utc := parsed.UTC()
			return &utc, nil
		}
	}
	return nil, fmt.Errorf("%w: %s must be ISO 8601 date or datetime", domain.ErrValidation, field)
}

func NewActivityHandler(repo repository.ActivityRepository) *ActivityHandler {
	return &ActivityHandler{repo: repo}
}

func (h *ActivityHandler) WithContacts(r repository.ContactRepository) *ActivityHandler {
	h.contacts = r
	return h
}

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
	filter := domain.ActivityFilter{Q: q.Get("q"), Sort: q.Get("sort"), Order: q.Get("order")}
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
	var req activityCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	dueDate, err := parseActivityTime("due_date", req.DueDate)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	startAt, err := parseActivityTime("start_at", req.StartAt)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	endAt, err := parseActivityTime("end_at", req.EndAt)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	ownerID := uuid.Nil
	if req.OwnerID != nil {
		ownerID = *req.OwnerID
	} else if claims, ok := middleware.ClaimsFromContext(r); ok {
		ownerID = claims.UserID
	}

	a := domain.Activity{
		Type:        req.Type,
		Subject:     strings.TrimSpace(req.Subject),
		Description: req.Description,
		DueDate:     dueDate,
		StartAt:     startAt,
		EndAt:       endAt,
		ContactID:   req.ContactID,
		AccountID:   req.AccountID,
		DealID:      req.DealID,
		OwnerID:     ownerID,
	}
	if a.DueDate == nil && a.StartAt != nil {
		a.DueDate = a.StartAt
	}
	if req.Completed != nil && *req.Completed {
		now := time.Now().UTC()
		a.CompletedAt = &now
	}

	if err := a.Validate(); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}
	if err := validateContactAccountPair(r.Context(), h.contacts, a.ContactID, a.AccountID); err != nil {
		handleDomainErr(w, err)
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
	var req activityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	dueDate, err := parseActivityTime("due_date", req.DueDate)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	startAt, err := parseActivityTime("start_at", req.StartAt)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	endAt, err := parseActivityTime("end_at", req.EndAt)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	patch := domain.ActivityPatch{
		Type:        req.Type,
		Subject:     req.Subject,
		Description: req.Description,
		ContactID:   req.ContactID,
		AccountID:   req.AccountID,
		DealID:      req.DealID,
		OwnerID:     req.OwnerID,
		DueDate:     dueDate,
		StartAt:     startAt,
		EndAt:       endAt,
	}
	if patch.DueDate == nil && patch.StartAt != nil {
		patch.DueDate = patch.StartAt
	}
	if req.Completed != nil {
		if *req.Completed {
			now := time.Now().UTC()
			patch.CompletedAt = &now
		} else {
			zero := time.Time{}
			patch.CompletedAt = &zero
		}
	}

	needsCurrent := (h.contacts != nil && (patch.ContactID != nil || patch.AccountID != nil)) ||
		patch.StartAt != nil || patch.EndAt != nil || patch.DueDate != nil
	var current *domain.Activity
	if needsCurrent {
		current, err = h.repo.GetByID(r.Context(), id)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
	}

	if patch.StartAt != nil || patch.EndAt != nil || patch.DueDate != nil {
		effectiveStart := current.StartAt
		if effectiveStart == nil {
			effectiveStart = current.DueDate
		}
		if patch.DueDate != nil {
			effectiveStart = patch.DueDate
		}
		if patch.StartAt != nil {
			effectiveStart = patch.StartAt
		}

		effectiveEnd := current.EndAt
		if patch.EndAt != nil {
			effectiveEnd = patch.EndAt
		}

		if effectiveEnd != nil {
			if effectiveStart == nil {
				handleDomainErr(w, fmt.Errorf("%w: end_at requires start_at or due_date", domain.ErrValidation))
				return
			}
			if effectiveEnd.Before(*effectiveStart) {
				handleDomainErr(w, fmt.Errorf("%w: end_at must be on or after start_at", domain.ErrValidation))
				return
			}
		}
	}

	if h.contacts != nil && (patch.ContactID != nil || patch.AccountID != nil) {
		contactID := current.ContactID
		accountID := current.AccountID
		if patch.ContactID != nil {
			contactID = patch.ContactID
		}
		if patch.AccountID != nil {
			accountID = patch.AccountID
		}
		if err := validateContactAccountPair(r.Context(), h.contacts, contactID, accountID); err != nil {
			handleDomainErr(w, err)
			return
		}
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
