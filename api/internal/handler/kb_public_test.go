package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
)

func TestPublicKBCompatibilityRoutesResolveOrgSlug(t *testing.T) {
	orgID := uuid.New()
	number := int64(42)
	articles := &fakeKBArticleRepo{articles: []*domain.KBArticle{
		{
			ID:           uuid.New(),
			OrgID:        orgID,
			Title:        "Reset Password",
			Status:       domain.KBArticleStatusPublished,
			Number:       &number,
			NumberPrefix: "KB",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}}
	h := NewKBHandler(articles, &fakeKBCategoryRepo{}).WithOrgs(&fakeOrgRepo{
		bySlug: map[string]*domain.Organization{
			"acme": {ID: orgID, Slug: "acme", Name: "Acme"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/kb/acme/articles", nil)
	rr := httptest.NewRecorder()

	h.PublicCompatibilityRouter().ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, orgID, articles.lastFilter.OrgID)

	var body []map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	require.Len(t, body, 1)
	require.Equal(t, "kb-42", body[0]["slug"])
}

func TestPublicKBCompatibilityRoutesRejectUnknownOrgSlug(t *testing.T) {
	articles := &fakeKBArticleRepo{}
	h := NewKBHandler(articles, &fakeKBCategoryRepo{}).WithOrgs(&fakeOrgRepo{
		bySlug: map[string]*domain.Organization{},
	})

	req := httptest.NewRequest(http.MethodGet, "/kb/missing/articles", nil)
	rr := httptest.NewRecorder()

	h.PublicCompatibilityRouter().ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	require.Equal(t, uuid.Nil, articles.lastFilter.OrgID)
}

func TestKBSuggestByQueryPopulatesGeneratedSlugs(t *testing.T) {
	orgID := uuid.New()
	articleID := uuid.New()
	number := int64(42)
	articles := &fakeKBArticleRepo{
		byID: map[uuid.UUID]*domain.KBArticle{
			articleID: {
				ID:           articleID,
				OrgID:        orgID,
				Title:        "Reset Password",
				Status:       domain.KBArticleStatusPublished,
				Number:       &number,
				NumberPrefix: "KB",
			},
		},
		suggestResults: []*domain.KBSuggestResult{
			{ID: articleID, Title: "Reset Password"},
		},
	}
	h := NewKBHandler(articles, &fakeKBCategoryRepo{})

	req := httptest.NewRequest(http.MethodGet, "/articles/suggest?q=password", nil)
	req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
	rr := httptest.NewRecorder()

	h.Router().ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, orgID, articles.suggestOrgID)

	var body []domain.KBSuggestResult
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	require.Len(t, body, 1)
	require.Equal(t, "kb-42", body[0].Slug)
}

func TestPublicKBUUIDArticleLookupRejectsCrossOrgArticle(t *testing.T) {
	orgID := uuid.New()
	otherOrgID := uuid.New()
	articleID := uuid.New()
	articles := &fakeKBArticleRepo{
		byID: map[uuid.UUID]*domain.KBArticle{
			articleID: {
				ID:     articleID,
				OrgID:  otherOrgID,
				Title:  "Other Tenant Article",
				Status: domain.KBArticleStatusPublished,
			},
		},
	}
	h := NewKBHandler(articles, &fakeKBCategoryRepo{}).WithOrgs(&fakeOrgRepo{
		bySlug: map[string]*domain.Organization{
			"acme": {ID: orgID, Slug: "acme", Name: "Acme"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/kb/acme/articles/"+articleID.String(), nil)
	rr := httptest.NewRecorder()

	h.PublicCompatibilityRouter().ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestPublicKBArticleResponsesOmitCustomFields(t *testing.T) {
	orgID := uuid.New()
	articleID := uuid.New()
	articles := &fakeKBArticleRepo{
		byID: map[uuid.UUID]*domain.KBArticle{
			articleID: {
				ID:           articleID,
				OrgID:        orgID,
				Title:        "Deployment Guide",
				Body:         "Keep this public.",
				Status:       domain.KBArticleStatusPublished,
				CustomFields: json.RawMessage(`{"internal_owner":"support-ops"}`),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}
	h := NewKBHandler(articles, &fakeKBCategoryRepo{}).WithOrgs(&fakeOrgRepo{
		bySlug: map[string]*domain.Organization{
			"acme": {ID: orgID, Slug: "acme", Name: "Acme"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/kb/acme/articles/"+articleID.String(), nil)
	rr := httptest.NewRecorder()

	h.PublicCompatibilityRouter().ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	require.Equal(t, "Deployment Guide", body["title"])
	require.Equal(t, "Keep this public.", body["body"])
	_, leaked := body["custom_fields"]
	require.False(t, leaked)
}

func TestPublicKBCategoriesIncludePublishedArticleCounts(t *testing.T) {
	orgID := uuid.New()
	categoryID := uuid.New()
	otherCategoryID := uuid.New()
	otherOrgID := uuid.New()
	draft := domain.KBArticleStatusDraft
	published := domain.KBArticleStatusPublished
	articles := &fakeKBArticleRepo{articles: []*domain.KBArticle{
		{ID: uuid.New(), OrgID: orgID, CategoryID: &categoryID, Status: published, Title: "One"},
		{ID: uuid.New(), OrgID: orgID, CategoryID: &categoryID, Status: published, Title: "Two"},
		{ID: uuid.New(), OrgID: orgID, CategoryID: &categoryID, Status: draft, Title: "Draft"},
		{ID: uuid.New(), OrgID: orgID, CategoryID: &otherCategoryID, Status: published, Title: "Other Category"},
		{ID: uuid.New(), OrgID: otherOrgID, CategoryID: &categoryID, Status: published, Title: "Other Org"},
	}}
	h := NewKBHandler(articles, &fakeKBCategoryRepo{categories: []*domain.KBCategory{
		{ID: categoryID, OrgID: orgID, Name: "Account Help", Slug: "account-help"},
	}}).WithOrgs(&fakeOrgRepo{
		bySlug: map[string]*domain.Organization{
			"acme": {ID: orgID, Slug: "acme", Name: "Acme"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/kb/acme/categories", nil)
	rr := httptest.NewRecorder()

	h.PublicCompatibilityRouter().ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var body []map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	require.Len(t, body, 1)
	require.Equal(t, float64(2), body[0]["article_count"])
}

type fakeOrgRepo struct {
	bySlug map[string]*domain.Organization
}

func (r *fakeOrgRepo) Create(context.Context, *domain.Organization) (*domain.Organization, error) {
	return nil, nil
}

func (r *fakeOrgRepo) GetByID(context.Context, uuid.UUID) (*domain.Organization, error) {
	return nil, nil
}

func (r *fakeOrgRepo) GetBySlug(_ context.Context, slug string) (*domain.Organization, error) {
	org, ok := r.bySlug[slug]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return org, nil
}

func (r *fakeOrgRepo) SlugExists(context.Context, string) (bool, error) {
	return false, nil
}

func (r *fakeOrgRepo) HasAny(context.Context) (bool, error) {
	return false, nil
}

func (r *fakeOrgRepo) List(context.Context) ([]*domain.Organization, error) {
	return nil, nil
}

func (r *fakeOrgRepo) UpdateName(context.Context, uuid.UUID, string) error {
	return nil
}

type fakeKBArticleRepo struct {
	lastFilter     domain.KBArticleFilter
	suggestOrgID   uuid.UUID
	articles       []*domain.KBArticle
	byID           map[uuid.UUID]*domain.KBArticle
	suggestResults []*domain.KBSuggestResult
}

func (r *fakeKBArticleRepo) Create(context.Context, *domain.KBArticle) (*domain.KBArticle, error) {
	return nil, nil
}

func (r *fakeKBArticleRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.KBArticle, error) {
	article, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return article, nil
}

func (r *fakeKBArticleRepo) Update(context.Context, uuid.UUID, domain.KBArticlePatch) (*domain.KBArticle, error) {
	return nil, nil
}

func (r *fakeKBArticleRepo) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeKBArticleRepo) List(_ context.Context, filter domain.KBArticleFilter) ([]*domain.KBArticle, int, error) {
	r.lastFilter = filter
	out := make([]*domain.KBArticle, 0, len(r.articles))
	for _, article := range r.articles {
		if filter.OrgID != uuid.Nil && article.OrgID != filter.OrgID {
			continue
		}
		if filter.Status != nil && article.Status != *filter.Status {
			continue
		}
		if filter.CategoryID != nil {
			if article.CategoryID == nil || *article.CategoryID != *filter.CategoryID {
				continue
			}
		}
		out = append(out, article)
	}
	return out, len(out), nil
}

func (r *fakeKBArticleRepo) IncrementViewCount(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeKBArticleRepo) Suggest(_ context.Context, orgID uuid.UUID, _ string, _ int) ([]*domain.KBSuggestResult, error) {
	r.suggestOrgID = orgID
	return r.suggestResults, nil
}

type fakeKBCategoryRepo struct {
	categories []*domain.KBCategory
}

func (r *fakeKBCategoryRepo) Create(context.Context, *domain.KBCategory) (*domain.KBCategory, error) {
	return nil, nil
}

func (r *fakeKBCategoryRepo) GetByID(context.Context, uuid.UUID) (*domain.KBCategory, error) {
	return nil, nil
}

func (r *fakeKBCategoryRepo) Update(context.Context, uuid.UUID, domain.KBCategoryPatch) (*domain.KBCategory, error) {
	return nil, nil
}

func (r *fakeKBCategoryRepo) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeKBCategoryRepo) List(_ context.Context, filter domain.KBCategoryFilter) ([]*domain.KBCategory, int, error) {
	out := make([]*domain.KBCategory, 0, len(r.categories))
	for _, category := range r.categories {
		if filter.OrgID != uuid.Nil && category.OrgID != filter.OrgID {
			continue
		}
		out = append(out, category)
	}
	return out, len(out), nil
}
