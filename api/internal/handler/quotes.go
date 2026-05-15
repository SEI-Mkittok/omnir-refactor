package handler

import (
	"context"
	"encoding/json"
	"errors"
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
	repo     repository.QuoteRepository
	contacts repository.ContactRepository
	deals    repository.DealRepository
	cfDefs   repository.CustomFieldDefinitionRepository
	mailer   *email.Mailer
	from     string
}

func NewQuoteHandler(repo repository.QuoteRepository) *QuoteHandler {
	return &QuoteHandler{repo: repo}
}

func (h *QuoteHandler) WithRelations(contacts repository.ContactRepository, deals repository.DealRepository) *QuoteHandler {
	h.contacts = contacts
	h.deals = deals
	return h
}

func (h *QuoteHandler) WithCustomFields(r repository.CustomFieldDefinitionRepository) *QuoteHandler {
	h.cfDefs = r
	return h
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

func quoteOrgContext(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return uuid.Nil, false
	}
	return orgID, true
}

// DealQuotesRouter returns sub-routes mounted under /deals/{dealId}/quotes.
func (h *QuoteHandler) DealQuotesRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListByDeal)
	return r
}

func (h *QuoteHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, ok := quoteOrgContext(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()
	filter := domain.QuoteFilter{
		OrgID: orgID,
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
	if v := q.Get("account_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.AccountID = &id
		}
	}

	quotes, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	if err := h.expandQuoteCustomFields(r.Context(), quotes); err != nil {
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
	orgID, ok := quoteOrgContext(w, r)
	if !ok {
		return
	}
	filter := domain.QuoteFilter{OrgID: orgID, DealID: &dealID, Limit: 50}
	quotes, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	if err := h.expandQuoteCustomFields(r.Context(), quotes); err != nil {
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
	if _, ok := quoteOrgContext(w, r); !ok {
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
	if err := h.expandQuoteCustomFields(r.Context(), []*domain.Quote{quote}); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

func (h *QuoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, ok := quoteOrgContext(w, r)
	if !ok {
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
	if err := h.validateQuoteCustomFields(r.Context(), q.CustomFields, true); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}
	if err := validateContactDealPair(r.Context(), h.contacts, h.deals, q.ContactID, q.DealID); err != nil {
		handleDomainErr(w, err)
		return
	}

	created, err := h.repo.Create(r.Context(), &q)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	if err := h.expandQuoteCustomFields(r.Context(), []*domain.Quote{created}); err != nil {
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
	if _, ok := quoteOrgContext(w, r); !ok {
		return
	}
	var patch domain.QuotePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if err := h.validateQuoteCustomFields(r.Context(), patch.CustomFields, false); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}
	if h.contacts != nil && h.deals != nil && (patch.ContactID != nil || patch.DealID != nil) {
		current, err := h.repo.GetByID(r.Context(), id)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		contactID := current.ContactID
		dealID := current.DealID
		if patch.ContactID != nil {
			contactID = patch.ContactID
		}
		if patch.DealID != nil {
			dealID = patch.DealID
		}
		if err := validateContactDealPair(r.Context(), h.contacts, h.deals, contactID, dealID); err != nil {
			handleDomainErr(w, err)
			return
		}
	}
	updated, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "quote not found")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	if err := h.expandQuoteCustomFields(r.Context(), []*domain.Quote{updated}); err != nil {
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
	if _, ok := quoteOrgContext(w, r); !ok {
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
	if _, ok := quoteOrgContext(w, r); !ok {
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
	if err := h.expandQuoteCustomFields(r.Context(), []*domain.Quote{quote}); err != nil {
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
	if _, ok := quoteOrgContext(w, r); !ok {
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
	if err := h.expandQuoteCustomFields(r.Context(), []*domain.Quote{quote}); err != nil {
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
	if _, ok := quoteOrgContext(w, r); !ok {
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
	if err := h.expandQuoteCustomFields(r.Context(), []*domain.Quote{quote}); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

func (h *QuoteHandler) quoteCustomFieldDefinitions(ctx context.Context) ([]*domain.CustomFieldDefinition, error) {
	if h.cfDefs == nil {
		return nil, nil
	}
	et := domain.CustomFieldEntityQuote
	return h.cfDefs.List(ctx, domain.CustomFieldDefinitionFilter{EntityType: &et})
}

func (h *QuoteHandler) validateQuoteCustomFields(ctx context.Context, raw json.RawMessage, enforceRequired bool) error {
	if h.cfDefs == nil {
		return nil
	}
	if !enforceRequired && len(raw) == 0 {
		return nil
	}
	defs, err := h.quoteCustomFieldDefinitions(ctx)
	if err != nil {
		return err
	}
	return domain.ValidateCustomFields(raw, defs)
}

func (h *QuoteHandler) expandQuoteCustomFields(ctx context.Context, quotes []*domain.Quote) error {
	if h.cfDefs == nil || len(quotes) == 0 {
		return nil
	}
	defs, err := h.quoteCustomFieldDefinitions(ctx)
	if err != nil {
		return err
	}
	for _, quote := range quotes {
		if quote == nil {
			continue
		}
		quote.CustomFields = domain.ExpandCustomFields(quote.CustomFields, defs)
	}
	return nil
}
