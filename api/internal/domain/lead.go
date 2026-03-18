package domain

import (
	"time"

	"github.com/google/uuid"
)

type LeadStatus string

const (
	LeadStatusNew         LeadStatus = "new"
	LeadStatusContacted   LeadStatus = "contacted"
	LeadStatusQualified   LeadStatus = "qualified"
	LeadStatusUnqualified LeadStatus = "unqualified"
	LeadStatusConverted   LeadStatus = "converted"
)

// Lead represents a prospective customer before they are converted to a Contact.
type Lead struct {
	ID                 uuid.UUID  `json:"id"`
	OrgID              uuid.UUID  `json:"org_id"`
	FirstName          string     `json:"first_name"`
	LastName           string     `json:"last_name"`
	Email              *string    `json:"email,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	Company            *string    `json:"company,omitempty"`
	LeadSource         *string    `json:"lead_source,omitempty"`
	Status             LeadStatus `json:"status"`
	OwnerID            *uuid.UUID `json:"owner_id,omitempty"`
	ConvertedContactID *uuid.UUID `json:"converted_contact_id,omitempty"`
	CustomFields       *[]byte    `json:"custom_fields,omitempty"`
	VtigerLegacyID     *string    `json:"vtiger_legacy_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

// LeadPatch holds optional fields for partial lead updates.
type LeadPatch struct {
	FirstName  *string     `json:"first_name,omitempty"`
	LastName   *string     `json:"last_name,omitempty"`
	Email      *string     `json:"email,omitempty"`
	Phone      *string     `json:"phone,omitempty"`
	Company    *string     `json:"company,omitempty"`
	LeadSource *string     `json:"lead_source,omitempty"`
	Status     *LeadStatus `json:"status,omitempty"`
	OwnerID    *uuid.UUID  `json:"owner_id,omitempty"`
}

// LeadFilter holds query parameters for listing leads.
type LeadFilter struct {
	OrgID   uuid.UUID
	Status  *LeadStatus
	OwnerID *uuid.UUID
	Q       string
	Page    int
	Limit   int
	Sort    string
	Order   string
}
