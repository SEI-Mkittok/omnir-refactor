package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// SearchResult is the unified search response.
type SearchResult struct {
	Contacts []*domain.Contact `json:"contacts"`
	Accounts []*domain.Account `json:"accounts"`
	Deals    []*domain.Deal    `json:"deals"`
	Total    int               `json:"total"`
}

// SearchHandler handles global search across CRM entities.
type SearchHandler struct {
	contacts repository.ContactRepository
	accounts repository.AccountRepository
	deals    repository.DealRepository
}

func NewSearchHandler(
	contacts repository.ContactRepository,
	accounts repository.AccountRepository,
	deals repository.DealRepository,
) *SearchHandler {
	return &SearchHandler{contacts: contacts, accounts: accounts, deals: deals}
}

func (h *SearchHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Search)
	return r
}

// Search handles GET /api/v1/search?q=...&limit=5
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "q is required")
		return
	}

	limit := 5
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	contacts, _, err := h.contacts.List(r.Context(), domain.ContactFilter{Q: q, Page: 1, Limit: limit})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	accounts, _, err := h.accounts.List(r.Context(), domain.AccountFilter{Q: q, Page: 1, Limit: limit})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	deals, _, err := h.deals.List(r.Context(), domain.DealFilter{Q: q, Page: 1, Limit: limit})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if contacts == nil {
		contacts = []*domain.Contact{}
	}
	if accounts == nil {
		accounts = []*domain.Account{}
	}
	if deals == nil {
		deals = []*domain.Deal{}
	}

	result := SearchResult{
		Contacts: contacts,
		Accounts: accounts,
		Deals:    deals,
		Total:    len(contacts) + len(accounts) + len(deals),
	}
	writeJSON(w, http.StatusOK, result)
}
