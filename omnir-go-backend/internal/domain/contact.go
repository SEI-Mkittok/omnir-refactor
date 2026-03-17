package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ContactStage string

const (
	ContactStageLead     ContactStage = "lead"
	ContactStageProspect ContactStage = "prospect"
	ContactStageCustomer ContactStage = "customer"
	ContactStageChurned  ContactStage = "churned"
)

type Contact struct {
	ID           uuid.UUID       `json:"id"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	Email        *string         `json:"email,omitempty"`
	Phone        *string         `json:"phone,omitempty"`
	AccountID    *uuid.UUID      `json:"account_id,omitempty"`
	OwnerID      uuid.UUID       `json:"owner_id"`
	LeadSource   *string         `json:"lead_source,omitempty"`
	Stage        ContactStage    `json:"stage"`
	Tags         []string        `json:"tags"`
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    *time.Time      `json:"deleted_at,omitempty"`
}

// ContactPatch holds optional fields for partial updates.
type ContactPatch struct {
	FirstName    *string         `json:"first_name,omitempty"`
	LastName     *string         `json:"last_name,omitempty"`
	Email        *string         `json:"email,omitempty"`
	Phone        *string         `json:"phone,omitempty"`
	AccountID    *uuid.UUID      `json:"account_id,omitempty"`
	OwnerID      *uuid.UUID      `json:"owner_id,omitempty"`
	LeadSource   *string         `json:"lead_source,omitempty"`
	Stage        *ContactStage   `json:"stage,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
}

// ContactFilter holds query parameters for listing contacts.
type ContactFilter struct {
	Q        string
	OwnerID  *uuid.UUID
	Stage    *ContactStage
	Page     int
	Limit    int
	Sort     string
	Order    string
}
