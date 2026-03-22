package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// NotificationHandler handles HTTP requests for the notifications resource.
type NotificationHandler struct {
	repo repository.NotificationRepository
}

func NewNotificationHandler(repo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

func (h *NotificationHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/unread-count", h.UnreadCount)
	r.Post("/{id}/read", h.MarkRead)
	r.Post("/read-all", h.MarkAllRead)
	return r
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	q := r.URL.Query()
	filter := domain.NotificationFilter{
		OrgID:      claims.OrgID,
		UserID:     claims.UserID,
		UnreadOnly: q.Get("unread_only") == "true",
		Limit:      50,
	}
	// Accept ?per_page (frontend) or ?limit (canonical).
	if limitVal := q.Get("per_page"); limitVal == "" {
		if v := q.Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
				filter.Limit = n
			}
		}
	} else if n, err := strconv.Atoi(limitVal); err == nil && n > 0 && n <= 200 {
		filter.Limit = n
	}
	if v := q.Get("before"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.Before = &t
		}
	}

	notifications, err := h.repo.ListByUser(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if notifications == nil {
		notifications = []*domain.Notification{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": notifications})
}

func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	count, err := h.repo.UnreadCount(r.Context(), claims.UserID, claims.OrgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	if err := h.repo.MarkRead(r.Context(), id, claims.UserID); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err := h.repo.MarkAllRead(r.Context(), claims.UserID, claims.OrgID); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
