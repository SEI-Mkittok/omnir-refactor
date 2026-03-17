package domain

import (
	"encoding/json"
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

type Deal struct {
	ID                uuid.UUID       `json:"id"`
	Title             string          `json:"title"`
	ValueCents        int64           `json:"value_cents"`
	Currency          string          `json:"currency"`
	Stage             DealStage       `json:"stage"`
	Probability       int             `json:"probability"`
	ExpectedCloseDate *time.Time      `json:"expected_close_date,omitempty"`
	ContactID         *uuid.UUID      `json:"contact_id,omitempty"`
	AccountID         *uuid.UUID      `json:"account_id,omitempty"`
	OwnerID           uuid.UUID       `json:"owner_id"`
	PipelineID        uuid.UUID       `json:"pipeline_id"`
	CustomFields      json.RawMessage `json:"custom_fields,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         *time.Time      `json:"deleted_at,omitempty"`
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
	OwnerID           *uuid.UUID      `json:"owner_id,omitempty"`
	PipelineID        *uuid.UUID      `json:"pipeline_id,omitempty"`
	CustomFields      json.RawMessage `json:"custom_fields,omitempty"`
}

// DealFilter holds query parameters for listing deals.
type DealFilter struct {
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
