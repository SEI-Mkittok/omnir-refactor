package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SLAEntityType identifies which entity type an SLA policy/instance applies to.
type SLAEntityType string

const (
	SLAEntityTypeDeal     SLAEntityType = "deal"
	SLAEntityTypeContact  SLAEntityType = "contact"
	SLAEntityTypeActivity SLAEntityType = "activity"
	SLAEntityTypeTicket   SLAEntityType = "ticket"
)

// SLABreachType describes which deadline was missed.
type SLABreachType string

const (
	SLABreachTypeNone       SLABreachType = "none"
	SLABreachTypeResponse   SLABreachType = "response"
	SLABreachTypeResolution SLABreachType = "resolution"
)

// SLAPolicy defines response and resolution time targets for tickets matching
// specific priorities within an org.
type SLAPolicy struct {
	ID                  uuid.UUID        `json:"id"`
	OrgID               uuid.UUID        `json:"org_id"`
	Name                string           `json:"name"`
	EntityType          SLAEntityType    `json:"entity_type"`
	Conditions          json.RawMessage  `json:"conditions"`
	ResponseTimeHours   float64          `json:"response_time_hours"`
	ResolutionTimeHours float64          `json:"resolution_time_hours"`
	PriorityFilter      []TicketPriority `json:"priority_filter"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

// SLAPolicyPatch holds optional fields for partial SLA policy updates.
type SLAPolicyPatch struct {
	Name                *string          `json:"name,omitempty"`
	EntityType          *SLAEntityType   `json:"entity_type,omitempty"`
	Conditions          json.RawMessage  `json:"conditions,omitempty"`
	ResponseTimeHours   *float64         `json:"response_time_hours,omitempty"`
	ResolutionTimeHours *float64         `json:"resolution_time_hours,omitempty"`
	PriorityFilter      []TicketPriority `json:"priority_filter,omitempty"`
}

// SLAStatus is the computed SLA state included in ticket responses.
type SLAStatus struct {
	PolicyID           *uuid.UUID `json:"policy_id,omitempty"`
	PolicyName         string     `json:"policy_name,omitempty"`
	ResponseDueAt      *time.Time `json:"response_due_at,omitempty"`
	ResolutionDueAt    *time.Time `json:"resolution_due_at,omitempty"`
	ResponseBreached   bool       `json:"response_breached"`
	ResolutionBreached bool       `json:"resolution_breached"`
	FirstRespondedAt   *time.Time `json:"first_responded_at,omitempty"`
	Status             string     `json:"status"` // on_track | at_risk | breached
}

// SLAInstance tracks SLA state for a specific entity (deal, contact, activity, ticket).
type SLAInstance struct {
	ID              uuid.UUID     `json:"id"`
	OrgID           uuid.UUID     `json:"org_id"`
	PolicyID        uuid.UUID     `json:"policy_id"`
	EntityID        uuid.UUID     `json:"entity_id"`
	EntityType      SLAEntityType `json:"entity_type"`
	StartedAt       time.Time     `json:"started_at"`
	ResponseDueAt   time.Time     `json:"response_due_at"`
	ResolutionDueAt time.Time     `json:"resolution_due_at"`
	RespondedAt     *time.Time    `json:"responded_at,omitempty"`
	ResolvedAt      *time.Time    `json:"resolved_at,omitempty"`
	Breached        bool          `json:"breached"`
	BreachType      SLABreachType `json:"breach_type"`
	WarnedAt        *time.Time    `json:"warned_at,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	// Computed status for API responses
	Status string `json:"status,omitempty"` // on_track | at_risk | breached
}

// ComputeStatus fills the Status field based on current time and deadlines.
func (s *SLAInstance) ComputeStatus() {
	if s.Breached {
		s.Status = "breached"
		return
	}
	now := time.Now().UTC()
	// at_risk: within 20% of the resolution window remaining
	total := s.ResolutionDueAt.Sub(s.StartedAt)
	remaining := s.ResolutionDueAt.Sub(now)
	if remaining <= 0 || remaining <= total/5 {
		s.Status = "at_risk"
	} else {
		s.Status = "on_track"
	}
}

// SLAInstanceFilter holds query parameters for listing SLA instances.
type SLAInstanceFilter struct {
	OrgID      uuid.UUID
	EntityID   *uuid.UUID
	EntityType SLAEntityType
	Breached   *bool
	Limit      int
	Offset     int
}

// SLADashboard summarises SLA health across an org.
type SLADashboard struct {
	OnTrack  int `json:"on_track"`
	AtRisk   int `json:"at_risk"`
	Breached int `json:"breached"`
}
