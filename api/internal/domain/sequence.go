package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SequenceStatus represents the lifecycle state of an email sequence.
type SequenceStatus string

const (
	SequenceStatusDraft    SequenceStatus = "draft"
	SequenceStatusActive   SequenceStatus = "active"
	SequenceStatusPaused   SequenceStatus = "paused"
	SequenceStatusArchived SequenceStatus = "archived"
)

// StepKind represents the type of a sequence step.
type StepKind string

const (
	StepKindEmail StepKind = "email"
	StepKindWait  StepKind = "wait"
)

// EnrollmentStatus represents the state of a contact's enrollment in a sequence.
type EnrollmentStatus string

const (
	EnrollmentStatusActive       EnrollmentStatus = "active"
	EnrollmentStatusCompleted    EnrollmentStatus = "completed"
	EnrollmentStatusUnsubscribed EnrollmentStatus = "unsubscribed"
	EnrollmentStatusBounced      EnrollmentStatus = "bounced"
	EnrollmentStatusPaused       EnrollmentStatus = "paused"
)

// SequenceEventKind represents the type of a sequence event.
type SequenceEventKind string

const (
	SequenceEventSent         SequenceEventKind = "sent"
	SequenceEventOpened       SequenceEventKind = "opened"
	SequenceEventClicked      SequenceEventKind = "clicked"
	SequenceEventCompleted    SequenceEventKind = "completed"
	SequenceEventBounced      SequenceEventKind = "bounced"
	SequenceEventUnsubscribed SequenceEventKind = "unsubscribed"
)

// EmailSequence is a named drip-email campaign.
type EmailSequence struct {
	ID          uuid.UUID      `json:"id"`
	OrgID       uuid.UUID      `json:"org_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      SequenceStatus `json:"status"`
	CreatedBy   *uuid.UUID     `json:"created_by,omitempty"`
	Steps       []SequenceStep `json:"steps,omitempty"`
	// Analytics fields (populated on list)
	EnrolledCount int       `json:"enrolled_count"`
	OpenRate      float64   `json:"open_rate"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SequenceStep is one step within an email sequence.
type SequenceStep struct {
	ID                uuid.UUID  `json:"id"`
	SequenceID        uuid.UUID  `json:"sequence_id"`
	OrgID             uuid.UUID  `json:"org_id"`
	Position          int        `json:"position"`
	Kind              StepKind   `json:"kind"`
	Subject           string     `json:"subject,omitempty"`
	Body              string     `json:"body,omitempty"`
	TemplateID        *uuid.UUID `json:"template_id,omitempty"`
	WaitDurationHours *int       `json:"wait_duration_hours,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// SequenceEnrollment represents a contact enrolled in a sequence.
type SequenceEnrollment struct {
	ID          uuid.UUID        `json:"id"`
	SequenceID  uuid.UUID        `json:"sequence_id"`
	ContactID   uuid.UUID        `json:"contact_id"`
	OrgID       uuid.UUID        `json:"org_id"`
	Status      EnrollmentStatus `json:"status"`
	CurrentStep int              `json:"current_step"`
	EnrolledAt  time.Time        `json:"enrolled_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	// Joined fields
	ContactName  string `json:"contact_name,omitempty"`
	ContactEmail string `json:"contact_email,omitempty"`
}

// SequenceEvent records an analytics event for a sequence.
type SequenceEvent struct {
	ID           uuid.UUID         `json:"id"`
	SequenceID   uuid.UUID         `json:"sequence_id"`
	StepID       *uuid.UUID        `json:"step_id,omitempty"`
	EnrollmentID uuid.UUID         `json:"enrollment_id"`
	ContactID    uuid.UUID         `json:"contact_id"`
	OrgID        uuid.UUID         `json:"org_id"`
	Kind         SequenceEventKind `json:"kind"`
	OccurredAt   time.Time         `json:"occurred_at"`
}

// SequenceAnalytics holds per-sequence aggregate metrics.
type SequenceAnalytics struct {
	SequenceID   uuid.UUID       `json:"sequence_id"`
	Sent         int             `json:"sent"`
	Opened       int             `json:"opened"`
	Clicked      int             `json:"clicked"`
	Completed    int             `json:"completed"`
	Bounced      int             `json:"bounced"`
	Unsubscribed int             `json:"unsubscribed"`
	OpenRate     float64         `json:"open_rate"`
	ClickRate    float64         `json:"click_rate"`
	Steps        []StepAnalytics `json:"steps"`
}

// StepAnalytics holds per-step metrics.
type StepAnalytics struct {
	StepID   uuid.UUID `json:"step_id"`
	Position int       `json:"position"`
	Sent     int       `json:"sent"`
	Opened   int       `json:"opened"`
	Clicked  int       `json:"clicked"`
}

// --- Request / Filter types ---

// CreateSequenceRequest is the payload to create a new sequence.
type CreateSequenceRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Steps       []CreateStepRequest `json:"steps"`
}

func (r *CreateSequenceRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	return nil
}

// UpdateSequenceRequest is the payload to update a sequence's metadata or status.
type UpdateSequenceRequest struct {
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	Status      *SequenceStatus     `json:"status,omitempty"`
	Steps       []CreateStepRequest `json:"steps,omitempty"` // full replacement of steps when provided
}

// CreateStepRequest describes one step in a sequence.
type CreateStepRequest struct {
	Kind              StepKind   `json:"kind"`
	Position          int        `json:"position"`
	Subject           string     `json:"subject,omitempty"`
	Body              string     `json:"body,omitempty"`
	TemplateID        *uuid.UUID `json:"template_id,omitempty"`
	WaitDurationHours *int       `json:"wait_duration_hours,omitempty"`
}

// EnrollRequest is the payload to enroll contacts into a sequence.
type EnrollRequest struct {
	ContactIDs []uuid.UUID `json:"contact_ids"`
}

func (r *EnrollRequest) Validate() error {
	if len(r.ContactIDs) == 0 {
		return fmt.Errorf("%w: at least one contact_id is required", ErrValidation)
	}
	return nil
}

// SequenceFilter holds query parameters for listing sequences.
type SequenceFilter struct {
	Status *SequenceStatus
	Page   int
	Limit  int
}
