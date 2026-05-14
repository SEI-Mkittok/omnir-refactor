package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ProductHelpArticleStatus is the publication state for global Omnir product help.
type ProductHelpArticleStatus string

const (
	ProductHelpArticleStatusDraft     ProductHelpArticleStatus = "draft"
	ProductHelpArticleStatusPublished ProductHelpArticleStatus = "published"
)

func (s ProductHelpArticleStatus) IsValid() bool {
	return s == ProductHelpArticleStatusDraft || s == ProductHelpArticleStatusPublished
}

// ProductHelpSyncStatus is the outcome state for a product-help wiki sync run.
type ProductHelpSyncStatus string

const (
	ProductHelpSyncStatusSucceeded ProductHelpSyncStatus = "succeeded"
	ProductHelpSyncStatusFailed    ProductHelpSyncStatus = "failed"
)

// ProductHelpCategory groups global Omnir product-help articles.
type ProductHelpCategory struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	SortOrder    int        `json:"sort_order"`
	ArticleCount int        `json:"article_count,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

func (c *ProductHelpCategory) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if c.Slug == "" {
		return fmt.Errorf("%w: slug is required", ErrValidation)
	}
	return nil
}

// ProductHelpArticle is a global Omnir product-help article mirrored from GitHub Wiki.
type ProductHelpArticle struct {
	ID           uuid.UUID                `json:"id"`
	CategoryID   *uuid.UUID               `json:"category_id,omitempty"`
	CategorySlug string                   `json:"category_slug,omitempty"`
	CategoryName string                   `json:"category_name,omitempty"`
	Title        string                   `json:"title"`
	Slug         string                   `json:"slug"`
	Body         string                   `json:"body,omitempty"`
	Excerpt      string                   `json:"excerpt,omitempty"`
	Tags         []string                 `json:"tags"`
	Status       ProductHelpArticleStatus `json:"status"`
	SourcePath   string                   `json:"source_path"`
	WikiURL      string                   `json:"wiki_url"`
	EditURL      string                   `json:"edit_url"`
	SortOrder    int                      `json:"sort_order"`
	ViewCount    int                      `json:"view_count"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	DeletedAt    *time.Time               `json:"deleted_at,omitempty"`
}

func (a *ProductHelpArticle) Validate() error {
	if a.Title == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if a.Slug == "" {
		return fmt.Errorf("%w: slug is required", ErrValidation)
	}
	if a.SourcePath == "" {
		return fmt.Errorf("%w: source_path is required", ErrValidation)
	}
	if !a.Status.IsValid() {
		return fmt.Errorf("%w: invalid status %q", ErrValidation, a.Status)
	}
	return nil
}

// ProductHelpArticleFilter controls global product-help article listing.
type ProductHelpArticleFilter struct {
	Query        string
	CategorySlug string
	Status       *ProductHelpArticleStatus
	Page         int
	Limit        int
	IncludeBody  bool
}

// ProductHelpSyncRun records a GitHub Wiki sync attempt.
type ProductHelpSyncRun struct {
	ID              uuid.UUID             `json:"id"`
	Status          ProductHelpSyncStatus `json:"status"`
	Message         string                `json:"message,omitempty"`
	CategoriesCount int                   `json:"categories_count"`
	ArticlesCount   int                   `json:"articles_count"`
	StartedAt       time.Time             `json:"started_at"`
	FinishedAt      *time.Time            `json:"finished_at,omitempty"`
}
