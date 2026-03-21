package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CalendarProvider identifies which calendar service a connection uses.
type CalendarProvider string

const (
	CalendarProviderGoogle    CalendarProvider = "google"
	CalendarProviderMicrosoft CalendarProvider = "microsoft"
)

// IsValid returns true if the provider is a known value.
func (p CalendarProvider) IsValid() bool {
	return p == CalendarProviderGoogle || p == CalendarProviderMicrosoft
}

// CalendarConnection represents a user's OAuth connection to an external
// calendar provider (Google Calendar or Microsoft Outlook).
type CalendarConnection struct {
	ID           uuid.UUID        `json:"id"`
	OrgID        uuid.UUID        `json:"org_id"`
	UserID       uuid.UUID        `json:"user_id"`
	Provider     CalendarProvider `json:"provider"`
	AccessToken  string           `json:"-"` // never serialised to clients
	RefreshToken *string          `json:"-"`
	TokenExpiry  *time.Time       `json:"token_expiry,omitempty"`
	SyncCursor   *string          `json:"-"` // provider-specific page/sync token
	CalendarID   *string          `json:"calendar_id,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// CalendarConnectionPatch holds mutable fields for updating a connection.
type CalendarConnectionPatch struct {
	AccessToken  *string
	RefreshToken *string
	TokenExpiry  *time.Time
	SyncCursor   *string
	CalendarID   *string
}

// CalendarConnectionFilter is used when listing connections.
type CalendarConnectionFilter struct {
	OrgID    uuid.UUID
	UserID   *uuid.UUID
	Provider *CalendarProvider
}

// CalendarEvent is a normalised calendar event received from a provider.
// It is used internally by the sync worker — not exposed via API.
type CalendarEvent struct {
	ExternalID  string
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	IsAllDay    bool
}

// Validate checks required fields on a new CalendarConnection.
func (c *CalendarConnection) Validate() error {
	if !c.Provider.IsValid() {
		return fmt.Errorf("%w: invalid calendar provider %q", ErrValidation, c.Provider)
	}
	if c.AccessToken == "" {
		return fmt.Errorf("%w: access_token is required", ErrValidation)
	}
	if c.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrValidation)
	}
	return nil
}
