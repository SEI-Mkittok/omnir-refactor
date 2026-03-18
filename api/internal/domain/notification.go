package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationKind is the type of CRM notification.
type NotificationKind string

const (
	NotificationKindActivityReminder NotificationKind = "activity_reminder"
	NotificationKindDealStageChanged NotificationKind = "deal_stage_changed"
	NotificationKindMention          NotificationKind = "mention"
	NotificationKindAssignment       NotificationKind = "assignment"
	NotificationKindSLAWarning       NotificationKind = "sla_warning"
	NotificationKindSLABreached      NotificationKind = "sla_breached"
)

// Notification is a general CRM notification for a user.
type Notification struct {
	ID         uuid.UUID        `json:"id"`
	OrgID      uuid.UUID        `json:"org_id,omitempty"`
	UserID     uuid.UUID        `json:"user_id,omitempty"`
	ActorID    *uuid.UUID       `json:"actor_id,omitempty"`
	Kind       NotificationKind `json:"kind"`
	EntityType *string          `json:"entity_type,omitempty"`
	EntityID   *uuid.UUID       `json:"entity_id,omitempty"`
	Title      string           `json:"title"`
	Body       *string          `json:"body,omitempty"`
	ReadAt     *time.Time       `json:"read_at,omitempty"`
	EmailedAt  *time.Time       `json:"emailed_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
}

// NotificationFilter holds query parameters for listing notifications.
type NotificationFilter struct {
	OrgID      uuid.UUID
	UserID     uuid.UUID
	UnreadOnly bool
	Limit      int
	Before     *time.Time // cursor-based pagination
}
