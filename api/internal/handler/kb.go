package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// KBHandler serves knowledge base article and category endpoints.
type KBHandler struct {
	articles   repository.KBArticleRepository
	categories repository.KBCategoryRepository
}

// NewKBHandler constructs a KBHandler.
func NewKBHandler(articles repository.KBArticleRepository, categories repository.KBCategoryRepository) *KBHandler {
	return &KBHandler{articles: articles, categories: categories}
}

// Router returns the authenticated sub-router (mounted under /api/v1/kb).
func (h *KBHandler) Router() chi.Router {
	r := chi.NewRouter()

	// Categories
	r.Get("/categories", h.ListCategories)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/categories", h.CreateCategory)
		r.Patch("/categories/{catID}", h.UpdateCategory)
		r.Delete("/categories/{catID}", h.DeleteCategory)
	})

	// Articles
	r.Get("/articles", h.ListArticles)
	r.Get("/articles/{articleID}", h.GetArticle)
	r.Get("/search", h.Search)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/articles", h.CreateArticle)
		r.Patch("/articles/{articleID}", h.UpdateArticle)
		r.Delete("/articles/{articleID}", h.DeleteArticle)
	})
	r.Get("/articles/{articleID}/suggest", h.Suggest)

	return r
}

// PublicRouter returns the unauthenticated help-portal router (mounted under /api/portal/help).
func (h *KBHandler) PublicRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/articles", h.PublicListArticles)
	r.Get("/search", h.PublicSearch)
	return r
}

// ---------------------------------------------------------------------------
// Categories
// ---------------------------------------------------------------------------

func (h *KBHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	filter := domain.KBCategoryFilter{Page: 1, Limit: 100}
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	cats, total, err := h.categories.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}
	writeJSON(w, http.StatusOK, paginated(cats, total, filter.Page, filter.Limit))
}

func (h *KBHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var c domain.KBCategory
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := c.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	created, err := h.categories.Create(r.Context(), &c)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create category")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *KBHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	catID, err := uuid.Parse(chi.URLParam(r, "catID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}
	var patch domain.KBCategoryPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	updated, err := h.categories.Update(r.Context(), catID, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *KBHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	catID, err := uuid.Parse(chi.URLParam(r, "catID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}
	if err := h.categories.Delete(r.Context(), catID); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Articles (authenticated)
// ---------------------------------------------------------------------------

func (h *KBHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	filter := h.parseArticleFilter(r, nil)
	articles, total, err := h.articles.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list articles")
		return
	}
	writeJSON(w, http.StatusOK, paginated(articles, total, filter.Page, filter.Limit))
}

func (h *KBHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "articleID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid article id")
		return
	}
	article, err := h.articles.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	// Best-effort view count increment (fire-and-forget, ignore error).
	_ = h.articles.IncrementViewCount(r.Context(), id)
	writeJSON(w, http.StatusOK, article)
}

func (h *KBHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	var a domain.KBArticle
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	a.AuthorID = claims.UserID
	if a.Status == "" {
		a.Status = domain.KBArticleStatusDraft
	}
	if err := a.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	created, err := h.articles.Create(r.Context(), &a)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create article")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *KBHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "articleID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid article id")
		return
	}
	var patch domain.KBArticlePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	updated, err := h.articles.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *KBHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "articleID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid article id")
		return
	}
	if err := h.articles.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *KBHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q parameter is required")
		return
	}
	filter := h.parseArticleFilter(r, nil)
	filter.Query = q
	articles, total, err := h.articles.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}
	writeJSON(w, http.StatusOK, paginated(articles, total, filter.Page, filter.Limit))
}

func (h *KBHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "articleID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid article id")
		return
	}
	// Verify the article exists and belongs to the org.
	article, err := h.articles.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	subject := r.URL.Query().Get("subject")
	if subject == "" {
		subject = article.Title
	}
	results, err := h.articles.Suggest(r.Context(), article.OrgID, subject, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "suggest failed")
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// ---------------------------------------------------------------------------
// Public portal endpoints (no auth)
// ---------------------------------------------------------------------------

func (h *KBHandler) PublicListArticles(w http.ResponseWriter, r *http.Request) {
	// Public always returns only published articles; no org context from JWT.
	// Org must be derivable from host header or query param in multi-tenant setups;
	// for now we use a "published only" filter with no org restriction (single-org mode).
	published := domain.KBArticleStatusPublished
	filter := domain.KBArticleFilter{
		Status: &published,
		Page:   1,
		Limit:  50,
	}
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	articles, total, err := h.articles.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list articles")
		return
	}
	writeJSON(w, http.StatusOK, paginated(articles, total, filter.Page, filter.Limit))
}

func (h *KBHandler) PublicSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q parameter is required")
		return
	}
	published := domain.KBArticleStatusPublished
	filter := domain.KBArticleFilter{
		Status: &published,
		Query:  q,
		Page:   1,
		Limit:  50,
	}
	articles, total, err := h.articles.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}
	writeJSON(w, http.StatusOK, paginated(articles, total, filter.Page, filter.Limit))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (h *KBHandler) parseArticleFilter(r *http.Request, defaultStatus *domain.KBArticleStatus) domain.KBArticleFilter {
	f := domain.KBArticleFilter{
		Status: defaultStatus,
		Page:   1,
		Limit:  50,
	}
	q := r.URL.Query()
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Page = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n <= 200 {
			f.Limit = n
		}
	}
	if v := q.Get("category_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.CategoryID = &id
		}
	}
	if v := q.Get("status"); v != "" {
		s := domain.KBArticleStatus(v)
		if s.IsValid() {
			f.Status = &s
		}
	}
	return f
}

