package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type productHelpSyncer interface {
	Sync(ctx context.Context) (*domain.ProductHelpSyncRun, error)
}

// ProductHelpHandler serves global Omnir product-help endpoints.
type ProductHelpHandler struct {
	repo   repository.ProductHelpRepository
	syncer productHelpSyncer
}

// NewProductHelpHandler constructs a ProductHelpHandler.
func NewProductHelpHandler(repo repository.ProductHelpRepository, syncer productHelpSyncer) *ProductHelpHandler {
	return &ProductHelpHandler{repo: repo, syncer: syncer}
}

// PublicRouter returns unauthenticated product-help routes.
func (h *ProductHelpHandler) PublicRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/categories", h.ListCategories)
	r.Get("/articles", h.ListArticles)
	r.Get("/articles/{slug}", h.GetArticle)
	r.Get("/search", h.Search)
	return r
}

// AdminRouter returns authenticated/admin product-help management routes.
func (h *ProductHelpHandler) AdminRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireRole(domain.UserRoleAdmin))
	r.Get("/sync", h.SyncStatus)
	r.Post("/sync", h.SyncNow)
	return r
}

func (h *ProductHelpHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.ListCategories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list product-help categories")
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *ProductHelpHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	filter := parseProductHelpArticleFilter(r)
	articles, _, err := h.repo.ListArticles(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list product-help articles")
		return
	}
	writeJSON(w, http.StatusOK, articles)
}

func (h *ProductHelpHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	article, err := h.repo.GetArticleBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	_ = h.repo.IncrementViewCount(r.Context(), article.ID)
	writeJSON(w, http.StatusOK, article)
}

func (h *ProductHelpHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q parameter is required")
		return
	}
	filter := parseProductHelpArticleFilter(r)
	filter.Query = q
	articles, _, err := h.repo.ListArticles(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "product-help search failed")
		return
	}
	writeJSON(w, http.StatusOK, articles)
}

func (h *ProductHelpHandler) SyncStatus(w http.ResponseWriter, r *http.Request) {
	run, err := h.repo.LatestSyncRun(r.Context())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"latest": nil})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load product-help sync status")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"latest": run})
}

func (h *ProductHelpHandler) SyncNow(w http.ResponseWriter, r *http.Request) {
	if h.syncer == nil {
		writeError(w, http.StatusServiceUnavailable, "product-help sync is not configured")
		return
	}
	run, err := h.syncer.Sync(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"latest": run, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"latest": run})
}

func parseProductHelpArticleFilter(r *http.Request) domain.ProductHelpArticleFilter {
	filter := domain.ProductHelpArticleFilter{
		Page:  1,
		Limit: 100,
	}
	q := r.URL.Query()
	if v := q.Get("q"); v != "" {
		filter.Query = v
	}
	if v := q.Get("category_slug"); v != "" {
		filter.CategorySlug = v
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
	return filter
}
