package domain

import (
	"time"

	"github.com/google/uuid"
)

// TimelineEntityRef is a typed entity pointer attached to a timeline event.
type TimelineEntityRef struct {
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
}

// TimelineEvent is the normalized timeline event envelope used across CRM modules.
type TimelineEvent struct {
	EventType  string              `json:"event_type"`
	EventID    uuid.UUID           `json:"event_id"`
	OccurredAt time.Time           `json:"occurred_at"`
	ActorID    *uuid.UUID          `json:"actor_id,omitempty"`
	EntityRefs []TimelineEntityRef `json:"entity_refs"`
	Preview    string              `json:"preview"`
}

// TimelineFilter holds query options for timeline listing.
type TimelineFilter struct {
	OrgID         uuid.UUID
	AccountID     *uuid.UUID
	ContactID     *uuid.UUID
	OccurredAtGTE *time.Time
	OccurredAtLTE *time.Time
	Page          int
	Limit         int
}
