package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SavedViewEntityType enumerates valid CRM entity types for saved views.
type SavedViewEntityType string

const (
	SavedViewEntityContacts SavedViewEntityType = "contacts"
	SavedViewEntityAccounts SavedViewEntityType = "accounts"
	SavedViewEntityDeals    SavedViewEntityType = "deals"
	SavedViewEntityLeads    SavedViewEntityType = "leads"
)

func (e SavedViewEntityType) IsValid() bool {
	switch e {
	case SavedViewEntityContacts, SavedViewEntityAccounts, SavedViewEntityDeals, SavedViewEntityLeads:
		return true
	}
	return false
}

// SavedView represents a named, saved filter set for a CRM entity list.
type SavedView struct {
	ID          uuid.UUID           `json:"id"`
	OrgID       uuid.UUID           `json:"org_id"`
	CreatedBy   uuid.UUID           `json:"created_by"`
	EntityType  SavedViewEntityType `json:"entity_type"`
	Name        string              `json:"name"`
	Filters     json.RawMessage     `json:"filters"`
	SortBy      *string             `json:"sort_by,omitempty"`
	SortDir     *string             `json:"sort_dir,omitempty"`
	IsShared    bool                `json:"is_shared"`
	IsPinned    bool                `json:"is_pinned"`
	PinnedOrder *int                `json:"pinned_order,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   *time.Time          `json:"deleted_at,omitempty"`
}

// SavedViewPatch holds optional fields for partial updates to a saved view.
type SavedViewPatch struct {
	Name        *string         `json:"name,omitempty"`
	Filters     json.RawMessage `json:"filters,omitempty"`
	SortBy      *string         `json:"sort_by,omitempty"`
	SortDir     *string         `json:"sort_dir,omitempty"`
	IsShared    *bool           `json:"is_shared,omitempty"`
	IsPinned    *bool           `json:"is_pinned,omitempty"`
	PinnedOrder *int            `json:"pinned_order,omitempty"`
}

// SavedViewFilter scopes a list query to the caller's visible views.
type SavedViewFilter struct {
	OrgID      uuid.UUID
	UserID     uuid.UUID
	EntityType SavedViewEntityType
}

// Validate checks required fields and value constraints on a SavedView.
func (v *SavedView) Validate() error {
	if v.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if !v.EntityType.IsValid() {
		return fmt.Errorf("%w: invalid entity_type %q", ErrValidation, v.EntityType)
	}
	if v.SortDir != nil && *v.SortDir != "asc" && *v.SortDir != "desc" {
		return fmt.Errorf("%w: sort_dir must be asc or desc", ErrValidation)
	}
	return nil
}
