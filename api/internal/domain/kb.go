package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// KBArticleStatus is the publication state of a knowledge base article.
type KBArticleStatus string

const (
	KBArticleStatusDraft     KBArticleStatus = "draft"
	KBArticleStatusPublished KBArticleStatus = "published"
)

func (s KBArticleStatus) IsValid() bool {
	return s == KBArticleStatusDraft || s == KBArticleStatusPublished
}

// KBCategory is a grouping bucket for KB articles.
type KBCategory struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *KBCategory) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if c.Slug == "" {
		return fmt.Errorf("%w: slug is required", ErrValidation)
	}
	return nil
}

// KBCategoryPatch holds optional update fields for a category.
type KBCategoryPatch struct {
	Name      *string `json:"name"`
	Slug      *string `json:"slug"`
	SortOrder *int    `json:"sort_order"`
}

// KBCategoryFilter holds query parameters for listing categories.
type KBCategoryFilter struct {
	OrgID uuid.UUID
	Page  int
	Limit int
}

// KBArticle is a knowledge base article.
type KBArticle struct {
	ID         uuid.UUID       `json:"id"`
	OrgID      uuid.UUID       `json:"org_id"`
	Title      string          `json:"title"`
	Body       string          `json:"body"`
	CategoryID *uuid.UUID      `json:"category_id,omitempty"`
	Tags       []string        `json:"tags"`
	Status     KBArticleStatus `json:"status"`
	AuthorID   uuid.UUID       `json:"author_id"`
	ViewCount  int             `json:"view_count"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	DeletedAt  *time.Time      `json:"deleted_at,omitempty"`
}

func (a *KBArticle) Validate() error {
	if a.Title == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if !a.Status.IsValid() {
		return fmt.Errorf("%w: invalid status %q", ErrValidation, a.Status)
	}
	if a.AuthorID == uuid.Nil {
		return fmt.Errorf("%w: author_id is required", ErrValidation)
	}
	return nil
}

// KBArticlePatch holds optional update fields for an article.
type KBArticlePatch struct {
	Title      *string          `json:"title"`
	Body       *string          `json:"body"`
	CategoryID *uuid.UUID       `json:"category_id"`
	Tags       []string         `json:"tags"`
	Status     *KBArticleStatus `json:"status"`
}

// KBArticleFilter holds query parameters for listing/searching articles.
type KBArticleFilter struct {
	OrgID      uuid.UUID
	Query      string
	CategoryID *uuid.UUID
	Status     *KBArticleStatus // nil = all
	Page       int
	Limit      int
}

// KBSuggestResult is a slim article summary for ticket-deflection suggestions.
type KBSuggestResult struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Slug  string    `json:"slug,omitempty"`
}
