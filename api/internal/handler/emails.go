package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// EmailHandler handles HTTP requests for the emails resource.
type EmailHandler struct {
	repo   repository.EmailRepository
	mailer *email.Mailer
	from   string
}

func NewEmailHandler(repo repository.EmailRepository, mailer *email.Mailer, from string) *EmailHandler {
	return &EmailHandler{repo: repo, mailer: mailer, from: from}
}

// Router returns the top-level /emails routes.
func (h *EmailHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Send)
	})
	return r
}

// ContactEmailRouter returns sub-routes mounted under /contacts/{id}/emails.
func (h *EmailHandler) ContactEmailRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListByContact)
	return r
}

// Send composes and sends an outbound email, then persists it.
func (h *EmailHandler) Send(w http.ResponseWriter, r *http.Request) {
	var req domain.SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if err := req.Validate(); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}

	// Attempt SMTP delivery (non-fatal when SMTP is disabled).
	_ = h.mailer.SendDirect(req.To, req.Subject, req.Body)

	record := &domain.ContactEmail{
		ContactID: req.ContactID,
		DealID:    req.DealID,
		Direction: domain.EmailDirectionOutbound,
		FromAddr:  h.from,
		ToAddr:    req.To,
		Subject:   req.Subject,
		Body:      req.Body,
		ThreadID:  req.ThreadID,
	}

	created, err := h.repo.Create(r.Context(), record)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// ListByContact returns emails for a specific contact.
func (h *EmailHandler) ListByContact(w http.ResponseWriter, r *http.Request) {
	contactID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid contact id")
		return
	}

	q := r.URL.Query()
	filter := domain.EmailFilter{
		ContactID: &contactID,
	}
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
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	emails, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(emails, total, filter.Page, filter.Limit))
}
