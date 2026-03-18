package domain

import (
	"time"

	"github.com/google/uuid"
)

// SLAPolicy defines response and resolution time targets for tickets matching
// specific priorities within an org.
type SLAPolicy struct {
	ID                   uuid.UUID       `json:"id"`
	OrgID                uuid.UUID       `json:"org_id"`
	Name                 string          `json:"name"`
	ResponseTimeHours    float64         `json:"response_time_hours"`
	ResolutionTimeHours  float64         `json:"resolution_time_hours"`
	PriorityFilter       []TicketPriority `json:"priority_filter"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// SLAPolicyPatch holds optional fields for partial SLA policy updates.
type SLAPolicyPatch struct {
	Name                *string          `json:"name,omitempty"`
	ResponseTimeHours   *float64         `json:"response_time_hours,omitempty"`
	ResolutionTimeHours *float64         `json:"resolution_time_hours,omitempty"`
	PriorityFilter      []TicketPriority `json:"priority_filter,omitempty"`
}

// SLAStatus is the computed SLA state included in ticket responses.
type SLAStatus struct {
	PolicyID          *uuid.UUID `json:"policy_id,omitempty"`
	ResponseDueAt     *time.Time `json:"response_due_at,omitempty"`
	ResponseBreached  bool       `json:"response_breached"`
	FirstRespondedAt  *time.Time `json:"first_responded_at,omitempty"`
}
