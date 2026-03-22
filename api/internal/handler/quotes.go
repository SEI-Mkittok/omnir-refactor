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

type QuoteHandler struct {
	repo   repository.QuoteRepository
	mailer *email.Mailer
	from   string
}

func NewQuoteHandler(repo repository.QuoteRepository) *QuoteHandler {
	return &QuoteHandler{repo: repo}
}

func (h *QuoteHandler) WithMailer(mailer *email.Mailer, from string) *QuoteHandler {
	h.mailer = mailer
	h.from = from
	return h
}

func (h *QuoteHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Create)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/send", h.Send)
		r.Post("/{id}/approve", h.Approve)
		r.Post("/{id}/reject", h.Reject)
	})
	return r
}

// DealQuotesRouter returns sub-routes mounted under /deals/{dealId}/quotes.
func (h *QuoteHandler) DealQuotesRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListByDeal)
	return r
}

func (h *QuoteHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	q := r.URL.Query()
	filter := domain.QuoteFilter{OrgID: orgID, Q: q.Get("q")}
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
	if v := q.Get("status"); v != "" {
		s := domain.QuoteStatus(v)
		filter.Status = &s
	}
	if v := q.Get("deal_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.DealID = &id
		}
	}
	if v := q.Get("contact_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.ContactID = &id
		}
	}

	quotes, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": quotes, "total": total})
}

func (h *QuoteHandler) ListByDeal(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "dealId"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid deal id")
		return
	}
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}
	filter := domain.QuoteFilter{OrgID: orgID, DealID: &dealID, Limit: 50}
	quotes, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": quotes, "total": total})
}

func (h *QuoteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid quote id")
		return
	}
	quote, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "quote not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

func (h *QuoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}
	var q domain.Quote
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	q.OrgID = orgID
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		q.CreatedBy = &claims.UserID
	}
	if err := q.Validate(); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), &q)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *QuoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid quote id")
		return
	}
	var patch domain.QuotePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	updated, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "quote not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *QuoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid quote id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "quote not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Send emails the quote to a contact and marks it as sent.
func (h *QuoteHandler) Send(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid quote id")
		return
	}

	var req domain.SendQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.To == "" {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "to is required")
		return
	}

	// Attempt email delivery (non-fatal if SMTP is disabled).
	if h.mailer != nil {
		subject := req.Subject
		if subject == "" {
			subject = "Your Quote"
		}
		_ = h.mailer.SendDirect(req.To, subject, req.Message)
	}

	quote, err := h.repo.MarkSent(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "quote not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

// Approve marks the quote as approved.
func (h *QuoteHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid quote id")
		return
	}
	quote, err := h.repo.MarkApproved(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "quote not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

// Reject marks the quote as rejected.
func (h *QuoteHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid quote id")
		return
	}
	quote, err := h.repo.MarkRejected(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "quote not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, quote)
}
