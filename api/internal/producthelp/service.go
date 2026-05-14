package producthelp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

const (
	defaultMaxManifestBytes = 128 * 1024
	defaultMaxArticleBytes  = 256 * 1024
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Config controls GitHub Wiki product-help sync behavior.
type Config struct {
	RawBaseURL       string
	WikiBaseURL      string
	ManifestPath     string
	MaxManifestBytes int64
	MaxArticleBytes  int64
}

// Service syncs global Omnir product-help content from GitHub Wiki.
type Service struct {
	repo   repository.ProductHelpRepository
	cfg    Config
	client *http.Client
}

// Manifest is the JSON source of truth checked into the GitHub Wiki.
type Manifest struct {
	Version    int                `json:"version"`
	Categories []ManifestCategory `json:"categories"`
	Articles   []ManifestArticle  `json:"articles"`
}

// ManifestCategory describes one product-help category in the wiki manifest.
type ManifestCategory struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Position int    `json:"position"`
}

// ManifestArticle describes one product-help article in the wiki manifest.
type ManifestArticle struct {
	Title    string   `json:"title"`
	Slug     string   `json:"slug"`
	Category string   `json:"category"`
	Path     string   `json:"path"`
	Status   string   `json:"status"`
	Position int      `json:"position"`
	Tags     []string `json:"tags"`
	Excerpt  string   `json:"excerpt"`
}

// NewService constructs a product-help sync service.
func NewService(repo repository.ProductHelpRepository, cfg Config) *Service {
	if cfg.MaxManifestBytes <= 0 {
		cfg.MaxManifestBytes = defaultMaxManifestBytes
	}
	if cfg.MaxArticleBytes <= 0 {
		cfg.MaxArticleBytes = defaultMaxArticleBytes
	}
	return &Service{
		repo: repo,
		cfg:  cfg,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// Start launches periodic background sync. It runs one sync immediately, then on interval.
func (s *Service) Start(ctx context.Context, interval time.Duration, logger *slog.Logger) {
	if interval <= 0 {
		return
	}
	go func() {
		if _, err := s.Sync(ctx); err != nil && logger != nil {
			logger.Warn("product help sync failed", "err", err)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := s.Sync(ctx); err != nil && logger != nil {
					logger.Warn("product help sync failed", "err", err)
				}
			}
		}
	}()
}

// Sync fetches the GitHub Wiki manifest and published markdown articles into local storage.
func (s *Service) Sync(ctx context.Context) (*domain.ProductHelpSyncRun, error) {
	started := time.Now().UTC()
	manifestURL, err := joinURL(s.cfg.RawBaseURL, s.cfg.ManifestPath)
	if err != nil {
		return s.recordFailure(ctx, started, err)
	}

	body, err := s.fetch(ctx, manifestURL, s.cfg.MaxManifestBytes)
	if err != nil {
		return s.recordFailure(ctx, started, err)
	}

	var manifest Manifest
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&manifest); err != nil {
		return s.recordFailure(ctx, started, fmt.Errorf("decode manifest: %w", err))
	}

	categories, articles, err := s.contentFromManifest(ctx, manifest)
	if err != nil {
		return s.recordFailure(ctx, started, err)
	}

	finished := time.Now().UTC()
	run := &domain.ProductHelpSyncRun{
		ID:              uuid.New(),
		Status:          domain.ProductHelpSyncStatusSucceeded,
		Message:         "synced product help from GitHub Wiki",
		CategoriesCount: len(categories),
		ArticlesCount:   len(articles),
		StartedAt:       started,
		FinishedAt:      &finished,
	}
	return s.repo.ReplaceContent(ctx, categories, articles, run)
}

func (s *Service) recordFailure(ctx context.Context, started time.Time, err error) (*domain.ProductHelpSyncRun, error) {
	finished := time.Now().UTC()
	run := &domain.ProductHelpSyncRun{
		ID:         uuid.New(),
		Status:     domain.ProductHelpSyncStatusFailed,
		Message:    err.Error(),
		StartedAt:  started,
		FinishedAt: &finished,
	}
	recorded, recordErr := s.repo.RecordSyncRun(ctx, run)
	if recordErr != nil {
		return nil, err
	}
	return recorded, err
}

func (s *Service) contentFromManifest(ctx context.Context, manifest Manifest) ([]*domain.ProductHelpCategory, []*domain.ProductHelpArticle, error) {
	if err := validateManifest(manifest); err != nil {
		return nil, nil, err
	}

	categories := make([]*domain.ProductHelpCategory, 0, len(manifest.Categories))
	categorySlugs := make(map[string]struct{}, len(manifest.Categories))
	for _, c := range manifest.Categories {
		categorySlugs[c.Slug] = struct{}{}
		categories = append(categories, &domain.ProductHelpCategory{
			Name:      strings.TrimSpace(c.Name),
			Slug:      c.Slug,
			SortOrder: c.Position,
		})
	}

	articles := make([]*domain.ProductHelpArticle, 0, len(manifest.Articles))
	for _, item := range manifest.Articles {
		status := domain.ProductHelpArticleStatus(item.Status)
		if status == "" {
			status = domain.ProductHelpArticleStatusPublished
		}
		if status != domain.ProductHelpArticleStatusPublished {
			continue
		}
		if _, ok := categorySlugs[item.Category]; !ok {
			return nil, nil, fmt.Errorf("article %q references unknown category %q", item.Slug, item.Category)
		}

		articleURL, err := joinURL(s.cfg.RawBaseURL, item.Path)
		if err != nil {
			return nil, nil, err
		}
		body, err := s.fetch(ctx, articleURL, s.cfg.MaxArticleBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("fetch article %q: %w", item.Path, err)
		}

		bodyText := strings.TrimSpace(string(body))
		excerpt := strings.TrimSpace(item.Excerpt)
		if excerpt == "" {
			excerpt = excerptFromMarkdown(bodyText)
		}
		articles = append(articles, &domain.ProductHelpArticle{
			Title:        strings.TrimSpace(item.Title),
			Slug:         item.Slug,
			Body:         bodyText,
			Excerpt:      excerpt,
			Tags:         cleanTags(item.Tags),
			Status:       status,
			SourcePath:   item.Path,
			CategorySlug: item.Category,
			WikiURL:      wikiPageURL(s.cfg.WikiBaseURL, item.Path),
			EditURL:      wikiPageURL(s.cfg.WikiBaseURL, item.Path) + "/_edit",
			SortOrder:    item.Position,
		})
	}

	return categories, articles, nil
}

func validateManifest(manifest Manifest) error {
	if manifest.Version <= 0 {
		return fmt.Errorf("manifest version is required")
	}
	if len(manifest.Categories) == 0 {
		return fmt.Errorf("manifest must include at least one category")
	}
	categorySlugs := map[string]struct{}{}
	for _, c := range manifest.Categories {
		if strings.TrimSpace(c.Name) == "" {
			return fmt.Errorf("category name is required")
		}
		if !validSlug(c.Slug) {
			return fmt.Errorf("invalid category slug %q", c.Slug)
		}
		if _, ok := categorySlugs[c.Slug]; ok {
			return fmt.Errorf("duplicate category slug %q", c.Slug)
		}
		categorySlugs[c.Slug] = struct{}{}
	}

	articleSlugs := map[string]struct{}{}
	for _, a := range manifest.Articles {
		if strings.TrimSpace(a.Title) == "" {
			return fmt.Errorf("article title is required")
		}
		if !validSlug(a.Slug) {
			return fmt.Errorf("invalid article slug %q", a.Slug)
		}
		if _, ok := articleSlugs[a.Slug]; ok {
			return fmt.Errorf("duplicate article slug %q", a.Slug)
		}
		articleSlugs[a.Slug] = struct{}{}
		status := a.Status
		if status == "" {
			status = string(domain.ProductHelpArticleStatusPublished)
		}
		if status != string(domain.ProductHelpArticleStatusPublished) && status != string(domain.ProductHelpArticleStatusDraft) {
			return fmt.Errorf("article %q has invalid status %q", a.Slug, a.Status)
		}
		if _, ok := categorySlugs[a.Category]; !ok {
			return fmt.Errorf("article %q references unknown category %q", a.Slug, a.Category)
		}
		if !validWikiPath(a.Path) {
			return fmt.Errorf("article %q has invalid path %q", a.Slug, a.Path)
		}
	}
	return nil
}

func validSlug(slug string) bool {
	return slugPattern.MatchString(slug)
}

func validWikiPath(path string) bool {
	if path == "" || strings.Contains(path, "/") || strings.Contains(path, `\`) || strings.Contains(path, "..") {
		return false
	}
	return strings.HasSuffix(strings.ToLower(path), ".md")
}

func cleanTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, tag := range tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func excerptFromMarkdown(body string) string {
	for _, block := range strings.Split(body, "\n\n") {
		line := strings.TrimSpace(block)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.Trim(line, "`*_>")
		if len(line) > 240 {
			return strings.TrimSpace(line[:240]) + "..."
		}
		return line
	}
	return ""
}

func wikiPageURL(base, sourcePath string) string {
	page := strings.TrimSuffix(sourcePath, ".md")
	page = strings.TrimSuffix(page, ".markdown")
	u, err := joinURL(base, page)
	if err != nil {
		return ""
	}
	return u
}

func joinURL(base, path string) (string, error) {
	if strings.TrimSpace(base) == "" {
		return "", fmt.Errorf("base URL is required")
	}
	u, err := url.JoinPath(strings.TrimRight(base, "/"), path)
	if err != nil {
		return "", err
	}
	return u, nil
}

func (s *Service) fetch(ctx context.Context, target string, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s returned %d", target, resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, maxBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("GET %s exceeded %d bytes", target, maxBytes)
	}
	return body, nil
}
