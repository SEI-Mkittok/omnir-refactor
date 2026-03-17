package domain

import (
	"encoding/json"
	"fmt"
	"strings"
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

// IsValid returns true if the stage is a known value.
func (s ContactStage) IsValid() bool {
	switch s {
	case ContactStageLead, ContactStageProspect, ContactStageCustomer, ContactStageChurned:
		return true
	}
	return false
}

// Validate checks required fields and value constraints on a Contact.
func (c *Contact) Validate() error {
	if c.FirstName == "" {
		return fmt.Errorf("%w: first_name is required", ErrValidation)
	}
	if c.Email != nil && *c.Email != "" {
		if !isValidEmail(*c.Email) {
			return fmt.Errorf("%w: invalid email format", ErrValidation)
		}
	}
	if c.Stage != "" && !c.Stage.IsValid() {
		return fmt.Errorf("%w: invalid stage %q", ErrValidation, c.Stage)
	}
	return nil
}

func isValidEmail(email string) bool {
	atIdx := strings.Index(email, "@")
	if atIdx < 1 {
		return false
	}
	domain := email[atIdx+1:]
	return strings.Contains(domain, ".")
}

// ContactFilter holds query parameters for listing contacts.
type ContactFilter struct {
	Q       string
	OwnerID *uuid.UUID
	Stage   *ContactStage
	Page    int
	Limit   int
	Sort    string
	Order   string
}
