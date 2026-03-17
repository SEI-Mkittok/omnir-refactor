package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ActivityType represents the kind of activity.
type ActivityType string

const (
	ActivityTypeCall    ActivityType = "call"
	ActivityTypeEmail   ActivityType = "email"
	ActivityTypeMeeting ActivityType = "meeting"
	ActivityTypeTask    ActivityType = "task"
)

// IsValid returns true if the activity type is a known value.
func (t ActivityType) IsValid() bool {
	switch t {
	case ActivityTypeCall, ActivityTypeEmail, ActivityTypeMeeting, ActivityTypeTask:
		return true
	}
	return false
}

// Activity represents a logged interaction or scheduled task linked to CRM entities.
type Activity struct {
	ID          uuid.UUID    `json:"id"`
	Type        ActivityType `json:"type"`
	Subject     string       `json:"subject"`
	Description *string      `json:"description,omitempty"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	ContactID   *uuid.UUID   `json:"contact_id,omitempty"`
	AccountID   *uuid.UUID   `json:"account_id,omitempty"`
	DealID      *uuid.UUID   `json:"deal_id,omitempty"`
	OwnerID     uuid.UUID    `json:"owner_id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty"`
}

// Validate checks required fields and value constraints on an Activity.
func (a *Activity) Validate() error {
	if a.Subject == "" {
		return fmt.Errorf("%w: subject is required", ErrValidation)
	}
	if !a.Type.IsValid() {
		return fmt.Errorf("%w: invalid type %q; must be call, email, meeting, or task", ErrValidation, a.Type)
	}
	if a.OwnerID == uuid.Nil {
		return fmt.Errorf("%w: owner_id is required", ErrValidation)
	}
	return nil
}

// ActivityPatch holds optional fields for partial updates.
type ActivityPatch struct {
	Type        *ActivityType `json:"type,omitempty"`
	Subject     *string       `json:"subject,omitempty"`
	Description *string       `json:"description,omitempty"`
	DueDate     *time.Time    `json:"due_date,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	ContactID   *uuid.UUID    `json:"contact_id,omitempty"`
	AccountID   *uuid.UUID    `json:"account_id,omitempty"`
	DealID      *uuid.UUID    `json:"deal_id,omitempty"`
	OwnerID     *uuid.UUID    `json:"owner_id,omitempty"`
}

// ActivityFilter holds query parameters for listing activities.
type ActivityFilter struct {
	Q         string
	Type      *ActivityType
	OwnerID   *uuid.UUID
	ContactID *uuid.UUID
	AccountID *uuid.UUID
	DealID    *uuid.UUID
	Page      int
	Limit     int
	Sort      string
	Order     string
}
