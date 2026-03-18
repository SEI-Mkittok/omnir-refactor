package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

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

	results, total, err := h.repo.Search(r.Context(), q, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"results": results,
		"total":   total,
	})
}
