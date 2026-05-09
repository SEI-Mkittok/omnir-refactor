package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// SearchHandler handles global full-text search across CRM entities.
type SearchHandler struct {
	repo repository.SearchRepository
}

func NewSearchHandler(repo repository.SearchRepository) *SearchHandler {
	return &SearchHandler{repo: repo}
}

func (h *SearchHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Search)
	return r
}

// Search handles GET /api/v1/search?q=...&limit=20
// Returns results grouped by entity type: { contacts, accounts, deals, tickets }.
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "q is required")
		return
	}

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	filter := domain.SearchFilter{
		Query:            q,
		Limit:            limit,
		RelationshipType: r.URL.Query().Get("relationship_type"),
	}
	if entityType := domain.SearchEntityType(r.URL.Query().Get("entity_type")); entityType != "" {
		if !entityType.IsValid() {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "entity_type is invalid")
			return
		}
		filter.EntityType = entityType
	}
	if raw := r.URL.Query().Get("account_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "account_id is invalid")
			return
		}
		filter.AccountID = &id
	}
	if raw := r.URL.Query().Get("contact_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "contact_id is invalid")
			return
		}
		filter.ContactID = &id
	}

	results, err := h.repo.Search(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}
