package domain

import (
	"encoding/json"
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
	ID                 uuid.UUID       `json:"id"`
	OrgID              uuid.UUID       `json:"org_id"`
	FirstName          string          `json:"first_name"`
	LastName           string          `json:"last_name"`
	Email              *string         `json:"email,omitempty"`
	Phone              *string         `json:"phone,omitempty"`
	Company            *string         `json:"company,omitempty"`
	LeadSource         *string         `json:"lead_source,omitempty"`
	Score              int             `json:"lead_score"`
	Status             LeadStatus      `json:"status"`
	OwnerID            *uuid.UUID      `json:"owner_id,omitempty"`
	ConvertedContactID *uuid.UUID      `json:"converted_contact_id,omitempty"`
	CustomFields       json.RawMessage `json:"custom_fields,omitempty"`
	VtigerLegacyID     *string         `json:"vtiger_legacy_id,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          *time.Time      `json:"deleted_at,omitempty"`
}

// LeadPatch holds optional fields for partial lead updates.
type LeadPatch struct {
	FirstName    *string         `json:"first_name,omitempty"`
	LastName     *string         `json:"last_name,omitempty"`
	Email        *string         `json:"email,omitempty"`
	Phone        *string         `json:"phone,omitempty"`
	Company      *string         `json:"company,omitempty"`
	LeadSource   *string         `json:"lead_source,omitempty"`
	Score        *int            `json:"lead_score,omitempty"`
	Status       *LeadStatus     `json:"status,omitempty"`
	OwnerID      *uuid.UUID      `json:"owner_id,omitempty"`
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
}

// ScoreForStatus returns the canonical lead score for a given status.
// Scores advance in 10% increments: new=0, contacted=20, qualified=40,
// converted=100. Unqualified resets to 0.
func ScoreForStatus(s LeadStatus) int {
	switch s {
	case LeadStatusContacted:
		return 20
	case LeadStatusQualified:
		return 40
	case LeadStatusConverted:
		return 100
	default: // new, unqualified, unknown
		return 0
	}
}

// SnapScoreToStep rounds score to the nearest 10% increment (0–100).
func SnapScoreToStep(score int) int {
	if score <= 0 {
		return 0
	}
	if score >= 100 {
		return 100
	}
	return ((score + 5) / 10) * 10
}

// LeadFilter holds query parameters for listing leads.
type LeadFilter struct {
	OrgID    uuid.UUID
	Status   *LeadStatus
	OwnerID  *uuid.UUID
	Source   *string
	ScoreMin *int
	ScoreMax *int
	Q        string
	Page     int
	Limit    int
	Sort     string
	Order    string
}
