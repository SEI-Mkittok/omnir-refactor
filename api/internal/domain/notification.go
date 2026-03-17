package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the kind of reminder notification.
type NotificationType string

const (
	NotificationTypeUpcoming15m NotificationType = "upcoming_15m"
	NotificationTypeUpcoming1h  NotificationType = "upcoming_1h"
	NotificationTypeUpcoming1d  NotificationType = "upcoming_1d"
	NotificationTypeOverdue     NotificationType = "overdue"
)

// Notification is a reminder linked to a user and an activity.
type Notification struct {
	ID         uuid.UUID        `json:"id"`
	OrgID      uuid.UUID        `json:"org_id"`
	UserID     uuid.UUID        `json:"user_id"`
	ActivityID uuid.UUID        `json:"activity_id"`
	Type       NotificationType `json:"type"`
	ReadAt     *time.Time       `json:"read_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
}

// NotificationFilter holds query parameters for listing notifications.
type NotificationFilter struct {
	OrgID  uuid.UUID
	UserID uuid.UUID
	Unread bool
	Page   int
	Limit  int
}
