package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CRMEntityType represents any entity that can participate in a reusable CRM link.
type CRMEntityType string

const (
	CRMEntityTypeAccount          CRMEntityType = "account"
	CRMEntityTypeContact          CRMEntityType = "contact"
	CRMEntityTypeDeal             CRMEntityType = "deal"
	CRMEntityTypeLead             CRMEntityType = "lead"
	CRMEntityTypeTicket           CRMEntityType = "ticket"
	CRMEntityTypeQuote            CRMEntityType = "quote"
	CRMEntityTypeActivity         CRMEntityType = "activity"
	CRMEntityTypeSequence         CRMEntityType = "sequence"
	CRMEntityTypeNote             CRMEntityType = "note"
	CRMEntityTypeEntityAttachment CRMEntityType = "entity_attachment"
)

// IsValid reports whether the CRM entity type is supported by crm_entity_links.
func (t CRMEntityType) IsValid() bool {
	switch t {
	case CRMEntityTypeAccount,
		CRMEntityTypeContact,
		CRMEntityTypeDeal,
		CRMEntityTypeLead,
		CRMEntityTypeTicket,
		CRMEntityTypeQuote,
		CRMEntityTypeActivity,
		CRMEntityTypeSequence,
		CRMEntityTypeNote,
		CRMEntityTypeEntityAttachment:
		return true
	}
	return false
}

// CRMEntityLink captures a reusable typed relationship between two CRM entities.
type CRMEntityLink struct {
	ID                       uuid.UUID       `json:"id"`
	OrgID                    uuid.UUID       `json:"org_id"`
	RelationshipDefinitionID *uuid.UUID      `json:"relationship_definition_id,omitempty"`
	FromEntityType           CRMEntityType   `json:"from_entity_type"`
	FromEntityID             uuid.UUID       `json:"from_entity_id"`
	ToEntityType             CRMEntityType   `json:"to_entity_type"`
	ToEntityID               uuid.UUID       `json:"to_entity_id"`
	LinkType                 string          `json:"link_type"`
	Metadata                 json.RawMessage `json:"metadata,omitempty"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
}

// CRMEntityLinkFilter selects links associated with a specific source or target entity.
type CRMEntityLinkFilter struct {
	OrgID                    uuid.UUID
	EntityType               CRMEntityType
	EntityID                 uuid.UUID
	RelationshipDefinitionID *uuid.UUID
	LinkType                 string
	IncludeFrom              bool
	IncludeTo                bool
}

// Validate checks required fields on a CRMEntityLink.
func (l *CRMEntityLink) Validate() error {
	if !l.FromEntityType.IsValid() {
		return fmt.Errorf("%w: invalid from_entity_type %q", ErrValidation, l.FromEntityType)
	}
	if l.FromEntityID == uuid.Nil {
		return fmt.Errorf("%w: from_entity_id is required", ErrValidation)
	}
	if !l.ToEntityType.IsValid() {
		return fmt.Errorf("%w: invalid to_entity_type %q", ErrValidation, l.ToEntityType)
	}
	if l.ToEntityID == uuid.Nil {
		return fmt.Errorf("%w: to_entity_id is required", ErrValidation)
	}
	if l.LinkType == "" {
		return fmt.Errorf("%w: link_type is required", ErrValidation)
	}
	if len(l.Metadata) == 0 {
		l.Metadata = json.RawMessage(`{}`)
	}
	return nil
}

// Normalize configures default filter behavior.
func (f *CRMEntityLinkFilter) Normalize() {
	if !f.IncludeFrom && !f.IncludeTo {
		f.IncludeFrom = true
		f.IncludeTo = true
	}
}
