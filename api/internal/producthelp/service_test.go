package producthelp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
)

func TestServiceSyncStoresPublishedWikiContent(t *testing.T) {
	manifest := Manifest{
		Version: 1,
		Categories: []ManifestCategory{
			{Name: "Getting Started", Slug: "getting-started", Position: 10},
		},
		Articles: []ManifestArticle{
			{
				Title:    "Welcome to Omnir",
				Slug:     "welcome-to-omnir",
				Category: "getting-started",
				Path:     "Welcome-to-Omnir.md",
				Status:   "published",
				Position: 10,
				Tags:     []string{"Start", "Admin"},
			},
			{
				Title:    "Draft Only",
				Slug:     "draft-only",
				Category: "getting-started",
				Path:     "Draft-Only.md",
				Status:   "draft",
			},
		},
	}
	srv := wikiFixture(t, manifest, map[string]string{
		"Welcome-to-Omnir.md": "# Welcome\n\nUse Omnir to manage CRM work.",
	})
	defer srv.Close()

	repo := newFakeProductHelpRepo()
	svc := NewService(repo, Config{
		RawBaseURL:   srv.URL,
		WikiBaseURL:  "https://github.com/acme/app/wiki",
		ManifestPath: "omnir-product-help-manifest.json",
	})

	run, err := svc.Sync(context.Background())

	require.NoError(t, err)
	require.Equal(t, domain.ProductHelpSyncStatusSucceeded, run.Status)
	require.Len(t, repo.categories, 1)
	require.Len(t, repo.articles, 1)
	require.Equal(t, "welcome-to-omnir", repo.articles["welcome-to-omnir"].Slug)
	require.Equal(t, "Use Omnir to manage CRM work.", repo.articles["welcome-to-omnir"].Excerpt)
	require.Equal(t, []string{"start", "admin"}, repo.articles["welcome-to-omnir"].Tags)
	require.Equal(t, "https://github.com/acme/app/wiki/Welcome-to-Omnir", repo.articles["welcome-to-omnir"].WikiURL)
}

func TestServiceSyncRejectsDuplicateArticleSlugs(t *testing.T) {
	manifest := Manifest{
		Version: 1,
		Categories: []ManifestCategory{
			{Name: "Getting Started", Slug: "getting-started"},
		},
		Articles: []ManifestArticle{
			{Title: "One", Slug: "duplicate", Category: "getting-started", Path: "One.md", Status: "published"},
			{Title: "Two", Slug: "duplicate", Category: "getting-started", Path: "Two.md", Status: "published"},
		},
	}
	srv := wikiFixture(t, manifest, nil)
	defer srv.Close()

	repo := newFakeProductHelpRepo()
	svc := NewService(repo, Config{RawBaseURL: srv.URL, WikiBaseURL: srv.URL, ManifestPath: "omnir-product-help-manifest.json"})

	run, err := svc.Sync(context.Background())

	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate article slug")
	require.Equal(t, domain.ProductHelpSyncStatusFailed, run.Status)
	require.Len(t, repo.runs, 1)
}

func TestServiceSyncSoftRemovesMissingArticlesThroughRepositoryContract(t *testing.T) {
	repo := newFakeProductHelpRepo()
	repo.articles["old-page"] = &domain.ProductHelpArticle{ID: uuid.New(), Slug: "old-page", Title: "Old"}

	manifest := Manifest{
		Version: 1,
		Categories: []ManifestCategory{
			{Name: "Getting Started", Slug: "getting-started"},
		},
		Articles: []ManifestArticle{
			{Title: "New Page", Slug: "new-page", Category: "getting-started", Path: "New-Page.md", Status: "published"},
		},
	}
	srv := wikiFixture(t, manifest, map[string]string{
		"New-Page.md": "# New Page\n\nFresh help.",
	})
	defer srv.Close()

	svc := NewService(repo, Config{RawBaseURL: srv.URL, WikiBaseURL: srv.URL, ManifestPath: "omnir-product-help-manifest.json"})

	_, err := svc.Sync(context.Background())

	require.NoError(t, err)
	require.Contains(t, repo.articles, "new-page")
	require.NotContains(t, repo.articles, "old-page")
}

func wikiFixture(t *testing.T, manifest Manifest, pages map[string]string) *httptest.Server {
	t.Helper()
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/omnir-product-help-manifest.json":
			_, _ = w.Write(manifestBytes)
			return
		default:
			path := r.URL.Path[1:]
			if body, ok := pages[path]; ok {
				_, _ = w.Write([]byte(body))
				return
			}
			http.NotFound(w, r)
		}
	}))
}

type fakeProductHelpRepo struct {
	categories map[string]*domain.ProductHelpCategory
	articles   map[string]*domain.ProductHelpArticle
	runs       []*domain.ProductHelpSyncRun
}

func newFakeProductHelpRepo() *fakeProductHelpRepo {
	return &fakeProductHelpRepo{
		categories: map[string]*domain.ProductHelpCategory{},
		articles:   map[string]*domain.ProductHelpArticle{},
	}
}

func (r *fakeProductHelpRepo) ListCategories(context.Context) ([]*domain.ProductHelpCategory, error) {
	return nil, nil
}

func (r *fakeProductHelpRepo) ListArticles(context.Context, domain.ProductHelpArticleFilter) ([]*domain.ProductHelpArticle, int, error) {
	return nil, 0, nil
}

func (r *fakeProductHelpRepo) GetArticleBySlug(context.Context, string) (*domain.ProductHelpArticle, error) {
	return nil, domain.ErrNotFound
}

func (r *fakeProductHelpRepo) IncrementViewCount(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeProductHelpRepo) LatestSyncRun(context.Context) (*domain.ProductHelpSyncRun, error) {
	if len(r.runs) == 0 {
		return nil, domain.ErrNotFound
	}
	return r.runs[len(r.runs)-1], nil
}

func (r *fakeProductHelpRepo) RecordSyncRun(_ context.Context, run *domain.ProductHelpSyncRun) (*domain.ProductHelpSyncRun, error) {
	r.runs = append(r.runs, run)
	return run, nil
}

func (r *fakeProductHelpRepo) ReplaceContent(_ context.Context, categories []*domain.ProductHelpCategory, articles []*domain.ProductHelpArticle, run *domain.ProductHelpSyncRun) (*domain.ProductHelpSyncRun, error) {
	nextCategories := map[string]*domain.ProductHelpCategory{}
	for _, c := range categories {
		nextCategories[c.Slug] = c
	}
	nextArticles := map[string]*domain.ProductHelpArticle{}
	for _, a := range articles {
		nextArticles[a.Slug] = a
	}
	r.categories = nextCategories
	r.articles = nextArticles
	r.runs = append(r.runs, run)
	return run, nil
}
