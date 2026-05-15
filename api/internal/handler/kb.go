package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
	cfDefs     repository.CustomFieldDefinitionRepository
	orgs       repository.OrgRepository
}

// NewKBHandler constructs a KBHandler.
func NewKBHandler(articles repository.KBArticleRepository, categories repository.KBCategoryRepository) *KBHandler {
	return &KBHandler{articles: articles, categories: categories}
}

// WithOrgs enables public slug routes to resolve tenant slugs before querying KB data.
func (h *KBHandler) WithOrgs(orgs repository.OrgRepository) *KBHandler {
	h.orgs = orgs
	return h
}

func (h *KBHandler) WithCustomFields(r repository.CustomFieldDefinitionRepository) *KBHandler {
	h.cfDefs = r
	return h
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
	r.Get("/articles/suggest", h.SuggestByQuery)
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
	r.Get("/categories", h.PublicListCategories)
	r.Get("/articles", h.PublicListArticles)
	r.Get("/articles/{slug}", h.PublicGetArticle)
	r.Get("/search", h.PublicSearch)
	return r
}

// PublicCompatibilityRouter supports the help-center client paths under /api/public.
func (h *KBHandler) PublicCompatibilityRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/kb/{orgSlug}/categories", h.PublicListCategories)
	r.Get("/kb/{orgSlug}/articles", h.PublicListArticlesFlat)
	r.Get("/kb/{orgSlug}/articles/{slug}", h.PublicGetArticle)
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
	if err := h.expandKBArticleCustomFields(r.Context(), articles); err != nil {
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
	if err := h.expandKBArticleCustomFields(r.Context(), []*domain.KBArticle{article}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load article")
		return
	}
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
	if err := h.validateKBArticleCustomFields(r.Context(), a.CustomFields, true); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	created, err := h.articles.Create(r.Context(), &a)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create article")
		return
	}
	if err := h.expandKBArticleCustomFields(r.Context(), []*domain.KBArticle{created}); err != nil {
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
	if err := h.validateKBArticleCustomFields(r.Context(), patch.CustomFields, false); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	updated, err := h.articles.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if err := h.expandKBArticleCustomFields(r.Context(), []*domain.KBArticle{updated}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update article")
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
	if err := h.expandKBArticleCustomFields(r.Context(), articles); err != nil {
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
	results, err = h.hydrateKBSuggestSlugs(domain.WithOrgID(r.Context(), article.OrgID), results)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "suggest failed")
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *KBHandler) SuggestByQuery(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		q = r.URL.Query().Get("subject")
	}
	if q == "" {
		writeError(w, http.StatusBadRequest, "q parameter is required")
		return
	}

	results, err := h.articles.Suggest(r.Context(), orgID, q, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "suggest failed")
		return
	}
	results, err = h.hydrateKBSuggestSlugs(domain.WithOrgID(r.Context(), orgID), results)
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
	orgID, ok := h.publicOrgID(w, r)
	if !ok {
		return
	}

	published := domain.KBArticleStatusPublished
	filter := domain.KBArticleFilter{
		Status: &published,
		OrgID:  orgID,
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
	writeJSON(w, http.StatusOK, paginated(publicKBArticles(articles, true), total, filter.Page, filter.Limit))
}

func (h *KBHandler) PublicListArticlesFlat(w http.ResponseWriter, r *http.Request) {
	orgID, ok := h.publicOrgID(w, r)
	if !ok {
		return
	}

	published := domain.KBArticleStatusPublished
	filter := domain.KBArticleFilter{
		Status: &published,
		OrgID:  orgID,
		Page:   1,
		Limit:  50,
	}
	if q := r.URL.Query().Get("q"); q != "" {
		filter.Query = q
	}
	articles, _, err := h.articles.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list articles")
		return
	}
	writeJSON(w, http.StatusOK, publicKBArticleSummaries(articles))
}

func (h *KBHandler) PublicListCategories(w http.ResponseWriter, r *http.Request) {
	orgID, ok := h.publicOrgID(w, r)
	if !ok {
		return
	}
	ctx := domain.WithOrgID(r.Context(), orgID)

	cats, _, err := h.categories.List(ctx, domain.KBCategoryFilter{OrgID: orgID, Page: 1, Limit: 100})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}
	counts, err := h.publicKBArticleCountsByCategory(ctx, orgID, cats)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}
	resp := make([]map[string]any, 0, len(cats))
	for _, cat := range cats {
		resp = append(resp, map[string]any{
			"id":            cat.ID,
			"org_id":        cat.OrgID,
			"name":          cat.Name,
			"slug":          cat.Slug,
			"position":      cat.SortOrder,
			"sort_order":    cat.SortOrder,
			"article_count": counts[cat.ID],
			"created_at":    cat.CreatedAt,
			"updated_at":    cat.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *KBHandler) PublicGetArticle(w http.ResponseWriter, r *http.Request) {
	orgID, ok := h.publicOrgID(w, r)
	if !ok {
		return
	}
	ctx := domain.WithOrgID(r.Context(), orgID)

	slug := chi.URLParam(r, "slug")
	if id, err := uuid.Parse(slug); err == nil {
		article, err := h.articles.GetByID(ctx, id)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		if article.OrgID != orgID || article.Status != domain.KBArticleStatusPublished {
			handleDomainErr(w, domain.ErrNotFound)
			return
		}
		writeJSON(w, http.StatusOK, publicKBArticle(article, true))
		return
	}

	published := domain.KBArticleStatusPublished
	articles, _, err := h.articles.List(r.Context(), domain.KBArticleFilter{OrgID: orgID, Status: &published, Page: 1, Limit: 200})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load article")
		return
	}
	for _, article := range articles {
		if articleSlug(article) == slug {
			writeJSON(w, http.StatusOK, publicKBArticle(article, true))
			return
		}
	}
	handleDomainErr(w, domain.ErrNotFound)
}

func (h *KBHandler) PublicSearch(w http.ResponseWriter, r *http.Request) {
	orgID, ok := h.publicOrgID(w, r)
	if !ok {
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q parameter is required")
		return
	}
	published := domain.KBArticleStatusPublished
	filter := domain.KBArticleFilter{
		Status: &published,
		OrgID:  orgID,
		Query:  q,
		Page:   1,
		Limit:  50,
	}
	articles, total, err := h.articles.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}
	writeJSON(w, http.StatusOK, paginated(publicKBArticles(articles, true), total, filter.Page, filter.Limit))
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

func (h *KBHandler) publicOrgID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	slug := chi.URLParam(r, "orgSlug")
	if slug == "" {
		return domain.DefaultOrgID, true
	}
	if h.orgs == nil {
		writeError(w, http.StatusInternalServerError, "organization lookup unavailable")
		return uuid.Nil, false
	}

	org, err := h.orgs.GetBySlug(r.Context(), slug)
	if err != nil {
		handleDomainErr(w, err)
		return uuid.Nil, false
	}
	return org.ID, true
}

func articleSlug(article *domain.KBArticle) string {
	if article.Number != nil {
		return strings.ToLower(article.NumberPrefix) + "-" + strconv.FormatInt(*article.Number, 10)
	}
	slug := strings.ToLower(article.Title)
	var b strings.Builder
	lastDash := false
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func publicKBArticleSummaries(articles []*domain.KBArticle) []map[string]any {
	resp := make([]map[string]any, 0, len(articles))
	for _, article := range articles {
		resp = append(resp, map[string]any{
			"id":          article.ID,
			"category_id": article.CategoryID,
			"title":       article.Title,
			"slug":        articleSlug(article),
			"status":      article.Status,
			"view_count":  article.ViewCount,
			"created_at":  article.CreatedAt,
			"updated_at":  article.UpdatedAt,
		})
	}
	return resp
}

func publicKBArticles(articles []*domain.KBArticle, includeBody bool) []map[string]any {
	resp := make([]map[string]any, 0, len(articles))
	for _, article := range articles {
		resp = append(resp, publicKBArticle(article, includeBody))
	}
	return resp
}

func publicKBArticle(article *domain.KBArticle, includeBody bool) map[string]any {
	resp := map[string]any{
		"id":            article.ID,
		"org_id":        article.OrgID,
		"category_id":   article.CategoryID,
		"title":         article.Title,
		"slug":          articleSlug(article),
		"tags":          article.Tags,
		"status":        article.Status,
		"view_count":    article.ViewCount,
		"number":        article.Number,
		"number_prefix": article.NumberPrefix,
		"created_at":    article.CreatedAt,
		"updated_at":    article.UpdatedAt,
	}
	if includeBody {
		resp["body"] = article.Body
	}
	return resp
}

func (h *KBHandler) publicKBArticleCountsByCategory(ctx context.Context, orgID uuid.UUID, cats []*domain.KBCategory) (map[uuid.UUID]int, error) {
	counts := make(map[uuid.UUID]int, len(cats))
	published := domain.KBArticleStatusPublished
	for _, cat := range cats {
		categoryID := cat.ID
		_, total, err := h.articles.List(ctx, domain.KBArticleFilter{
			OrgID:      orgID,
			CategoryID: &categoryID,
			Status:     &published,
			Page:       1,
			Limit:      1,
		})
		if err != nil {
			return nil, err
		}
		counts[cat.ID] = total
	}
	return counts, nil
}

func (h *KBHandler) hydrateKBSuggestSlugs(ctx context.Context, results []*domain.KBSuggestResult) ([]*domain.KBSuggestResult, error) {
	for _, result := range results {
		if result.Slug != "" {
			continue
		}
		article, err := h.articles.GetByID(ctx, result.ID)
		if err != nil {
			return nil, err
		}
		result.Slug = articleSlug(article)
	}
	return results, nil
}

func (h *KBHandler) kbArticleCustomFieldDefinitions(ctx context.Context) ([]*domain.CustomFieldDefinition, error) {
	if h.cfDefs == nil {
		return nil, nil
	}
	et := domain.CustomFieldEntityKBArticle
	return h.cfDefs.List(ctx, domain.CustomFieldDefinitionFilter{EntityType: &et})
}

func (h *KBHandler) validateKBArticleCustomFields(ctx context.Context, raw json.RawMessage, enforceRequired bool) error {
	if h.cfDefs == nil {
		return nil
	}
	if !enforceRequired && len(raw) == 0 {
		return nil
	}
	defs, err := h.kbArticleCustomFieldDefinitions(ctx)
	if err != nil {
		return err
	}
	return domain.ValidateCustomFields(raw, defs)
}

func (h *KBHandler) expandKBArticleCustomFields(ctx context.Context, articles []*domain.KBArticle) error {
	if h.cfDefs == nil || len(articles) == 0 {
		return nil
	}
	defs, err := h.kbArticleCustomFieldDefinitions(ctx)
	if err != nil {
		return err
	}
	for _, article := range articles {
		if article == nil {
			continue
		}
		article.CustomFields = domain.ExpandCustomFields(article.CustomFields, defs)
	}
	return nil
}
