package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// TimelineHandler serves normalized timeline events across CRM modules.
type TimelineHandler struct {
	repo repository.TimelineRepository
}

func NewTimelineHandler(repo repository.TimelineRepository) *TimelineHandler {
	return &TimelineHandler{repo: repo}
}

func (h *TimelineHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	return r
}

func (h *TimelineHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.TimelineFilter{
		Page:  1,
		Limit: 50,
	}

	if v := q.Get("account_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid account_id")
			return
		}
		filter.AccountID = &id
	}
	if v := q.Get("contact_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid contact_id")
			return
		}
		filter.ContactID = &id
	}
	if v := q.Get("start_at"); v != "" {
		ts, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid start_at: use RFC3339")
			return
		}
		filter.OccurredAtGTE = &ts
	}
	if v := q.Get("end_at"); v != "" {
		ts, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid end_at: use RFC3339")
			return
		}
		filter.OccurredAtLTE = &ts
	}
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

	events, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, paginated(events, total, filter.Page, filter.Limit))
}
