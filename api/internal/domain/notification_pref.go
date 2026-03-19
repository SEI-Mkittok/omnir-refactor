package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserNotificationPref holds per-user email notification opt-in/out settings.
// A missing row means all notifications are enabled (default-on).
type UserNotificationPref struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	OrgID           uuid.UUID `json:"org_id"`
	EmailOnAssigned bool      `json:"email_on_assigned"`
	EmailOnResolved bool      `json:"email_on_resolved"`
	EmailOnClosed   bool      `json:"email_on_closed"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserNotificationPrefPatch holds optional fields for partial updates.
type UserNotificationPrefPatch struct {
	EmailOnAssigned *bool `json:"email_on_assigned,omitempty"`
	EmailOnResolved *bool `json:"email_on_resolved,omitempty"`
	EmailOnClosed   *bool `json:"email_on_closed,omitempty"`
}

// EmailEventKind describes which ticket state change triggered an email.
type EmailEventKind string

const (
	EmailEventAssigned EmailEventKind = "assigned"
	EmailEventResolved EmailEventKind = "resolved"
	EmailEventClosed   EmailEventKind = "closed"
)

// EmailJob is an item queued for async delivery.
type EmailJob struct {
	Kind          EmailEventKind
	ToEmail       string
	ToName        string
	TicketID      string // string form of UUID for URL building
	TicketSubject string
}
