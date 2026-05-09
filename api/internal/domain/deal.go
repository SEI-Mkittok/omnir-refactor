package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DealStage string

const (
	DealStageLead        DealStage = "lead"
	DealStageQualified   DealStage = "qualified"
	DealStageProposal    DealStage = "proposal"
	DealStageNegotiation DealStage = "negotiation"
	DealStageClosedWon   DealStage = "closed_won"
	DealStageClosedLost  DealStage = "closed_lost"
)

// DealContact represents a contact linked to a deal with an optional role.
type DealContact struct {
	ContactID uuid.UUID `json:"contact_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Deal struct {
	ID                uuid.UUID       `json:"id"`
	OrgID             uuid.UUID       `json:"org_id"`
	Title             string          `json:"title"`
	ValueCents        int64           `json:"value_cents"`
	Currency          string          `json:"currency"`
	Stage             DealStage       `json:"stage"`
	Probability       int             `json:"probability"`
	ExpectedCloseDate *time.Time      `json:"expected_close_date,omitempty"`
	ContactID         *uuid.UUID      `json:"contact_id,omitempty"` // legacy; kept for backwards compat
	AccountID         *uuid.UUID      `json:"account_id,omitempty"`
	Account           *Account        `json:"account,omitempty"`
	Contact           *Contact        `json:"contact,omitempty"`
	OwnerID           uuid.UUID       `json:"owner_id"`
	PipelineID        uuid.UUID       `json:"pipeline_id"`
	CustomFields      json.RawMessage `json:"custom_fields,omitempty"`
	Contacts          []Contact       `json:"contacts,omitempty"` // populated on GetByID
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         *time.Time      `json:"deleted_at,omitempty"`
}

// IsValid returns true if the stage is a known value.
func (s DealStage) IsValid() bool {
	switch s {
	case DealStageLead, DealStageQualified, DealStageProposal,
		DealStageNegotiation, DealStageClosedWon, DealStageClosedLost:
		return true
	}
	return false
}

// Validate checks required fields and value constraints on a Deal.
func (d *Deal) Validate() error {
	if d.Title == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if d.OwnerID == uuid.Nil {
		return fmt.Errorf("%w: owner_id is required", ErrValidation)
	}
	if d.PipelineID == uuid.Nil {
		return fmt.Errorf("%w: pipeline_id is required", ErrValidation)
	}
	if d.Stage != "" && !d.Stage.IsValid() {
		return fmt.Errorf("%w: invalid stage %q", ErrValidation, d.Stage)
	}
	if d.Probability < 0 || d.Probability > 100 {
		return fmt.Errorf("%w: probability must be between 0 and 100", ErrValidation)
	}
	return nil
}

// DealPatch holds optional fields for partial updates.
type DealPatch struct {
	Title             *string         `json:"title,omitempty"`
	ValueCents        *int64          `json:"value_cents,omitempty"`
	Currency          *string         `json:"currency,omitempty"`
	Stage             *DealStage      `json:"stage,omitempty"`
	Probability       *int            `json:"probability,omitempty"`
	ExpectedCloseDate *time.Time      `json:"expected_close_date,omitempty"`
	ContactID         *uuid.UUID      `json:"contact_id,omitempty"`
	AccountID         *uuid.UUID      `json:"account_id,omitempty"`
	ClearContactID    bool            `json:"-"`
	ClearAccountID    bool            `json:"-"`
	OwnerID           *uuid.UUID      `json:"owner_id,omitempty"`
	PipelineID        *uuid.UUID      `json:"pipeline_id,omitempty"`
	CustomFields      json.RawMessage `json:"custom_fields,omitempty"`
}

// DealFilter holds query parameters for listing deals.
type DealFilter struct {
	OrgID      uuid.UUID
	Q          string
	OwnerID    *uuid.UUID
	Stage      *DealStage
	AccountID  *uuid.UUID
	ContactID  *uuid.UUID
	PipelineID *uuid.UUID
	Page       int
	Limit      int
	Sort       string
	Order      string
}
