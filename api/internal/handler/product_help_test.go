package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
)

func TestProductHelpPublicRoutesAreSeparateFromTenantKB(t *testing.T) {
	categoryID := uuid.New()
	articleID := uuid.New()
	repo := &fakeProductHelpHandlerRepo{
		categories: []*domain.ProductHelpCategory{
			{ID: categoryID, Name: "Getting Started", Slug: "getting-started", ArticleCount: 1},
		},
		articles: []*domain.ProductHelpArticle{
			{
				ID:           articleID,
				CategoryID:   &categoryID,
				CategorySlug: "getting-started",
				Title:        "Welcome to Omnir",
				Slug:         "welcome-to-omnir",
				Excerpt:      "Start here.",
				Body:         "# Welcome",
				Status:       domain.ProductHelpArticleStatusPublished,
				UpdatedAt:    time.Now(),
			},
		},
	}
	h := NewProductHelpHandler(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/articles?category_slug=getting-started", nil)
	rr := httptest.NewRecorder()
	h.PublicRouter().ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, domain.ProductHelpArticleFilter{CategorySlug: "getting-started", Page: 1, Limit: 100}, repo.lastFilter)

	var list []domain.ProductHelpArticle
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &list))
	require.Len(t, list, 1)
	require.Equal(t, "welcome-to-omnir", list[0].Slug)

	req = httptest.NewRequest(http.MethodGet, "/articles/welcome-to-omnir", nil)
	rr = httptest.NewRecorder()
	h.PublicRouter().ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, articleID, repo.incrementedID)
}

func TestProductHelpSearchRequiresQuery(t *testing.T) {
	h := NewProductHelpHandler(&fakeProductHelpHandlerRepo{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	rr := httptest.NewRecorder()
	h.PublicRouter().ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProductHelpAdminSyncNow(t *testing.T) {
	runID := uuid.New()
	repo := &fakeProductHelpHandlerRepo{}
	syncer := fakeProductHelpSyncer{
		run: &domain.ProductHelpSyncRun{
			ID:     runID,
			Status: domain.ProductHelpSyncStatusSucceeded,
		},
	}
	h := NewProductHelpHandler(repo, syncer)

	req := httptest.NewRequest(http.MethodPost, "/sync", nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{Role: string(domain.UserRoleAdmin)}))
	rr := httptest.NewRecorder()
	h.AdminRouter().ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), runID.String())
}

type fakeProductHelpHandlerRepo struct {
	categories    []*domain.ProductHelpCategory
	articles      []*domain.ProductHelpArticle
	lastFilter    domain.ProductHelpArticleFilter
	incrementedID uuid.UUID
	latest        *domain.ProductHelpSyncRun
}

func (r *fakeProductHelpHandlerRepo) ListCategories(context.Context) ([]*domain.ProductHelpCategory, error) {
	return r.categories, nil
}

func (r *fakeProductHelpHandlerRepo) ListArticles(_ context.Context, filter domain.ProductHelpArticleFilter) ([]*domain.ProductHelpArticle, int, error) {
	r.lastFilter = filter
	out := make([]*domain.ProductHelpArticle, 0, len(r.articles))
	for _, article := range r.articles {
		if filter.CategorySlug != "" && article.CategorySlug != filter.CategorySlug {
			continue
		}
		if filter.Query != "" && !strings.Contains(strings.ToLower(article.Title+" "+article.Excerpt), strings.ToLower(filter.Query)) {
			continue
		}
		out = append(out, article)
	}
	return out, len(out), nil
}

func (r *fakeProductHelpHandlerRepo) GetArticleBySlug(_ context.Context, slug string) (*domain.ProductHelpArticle, error) {
	for _, article := range r.articles {
		if article.Slug == slug {
			return article, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeProductHelpHandlerRepo) IncrementViewCount(_ context.Context, id uuid.UUID) error {
	r.incrementedID = id
	return nil
}

func (r *fakeProductHelpHandlerRepo) LatestSyncRun(context.Context) (*domain.ProductHelpSyncRun, error) {
	if r.latest == nil {
		return nil, domain.ErrNotFound
	}
	return r.latest, nil
}

func (r *fakeProductHelpHandlerRepo) RecordSyncRun(_ context.Context, run *domain.ProductHelpSyncRun) (*domain.ProductHelpSyncRun, error) {
	r.latest = run
	return run, nil
}

func (r *fakeProductHelpHandlerRepo) ReplaceContent(_ context.Context, _ []*domain.ProductHelpCategory, _ []*domain.ProductHelpArticle, run *domain.ProductHelpSyncRun) (*domain.ProductHelpSyncRun, error) {
	r.latest = run
	return run, nil
}

type fakeProductHelpSyncer struct {
	run *domain.ProductHelpSyncRun
	err error
}

func (s fakeProductHelpSyncer) Sync(context.Context) (*domain.ProductHelpSyncRun, error) {
	return s.run, s.err
}
