package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EntityType represents the CRM entity an attachment belongs to.
type EntityType string

const (
	EntityTypeContact EntityType = "contact"
	EntityTypeAccount EntityType = "account"
	EntityTypeDeal    EntityType = "deal"
)

// IsValid returns true if the entity type is known.
func (t EntityType) IsValid() bool {
	switch t {
	case EntityTypeContact, EntityTypeAccount, EntityTypeDeal:
		return true
	}
	return false
}

// EntityAttachment is a file attached to a contact, account, or deal.
type EntityAttachment struct {
	ID          uuid.UUID  `json:"id"`
	EntityType  EntityType `json:"entity_type"`
	EntityID    uuid.UUID  `json:"entity_id"`
	OrgID       uuid.UUID  `json:"org_id"`
	UploadedBy  *uuid.UUID `json:"uploaded_by,omitempty"`
	Filename    string     `json:"filename"`
	ContentType string     `json:"content_type"`
	SizeBytes   *int64     `json:"size_bytes,omitempty"`
	StoragePath string     `json:"-"` // Not exposed in API
	URL         string     `json:"url,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Validate checks required fields on an EntityAttachment.
func (a *EntityAttachment) Validate() error {
	if !a.EntityType.IsValid() {
		return fmt.Errorf("%w: invalid entity_type %q", ErrValidation, a.EntityType)
	}
	if a.EntityID == uuid.Nil {
		return fmt.Errorf("%w: entity_id is required", ErrValidation)
	}
	if a.Filename == "" {
		return fmt.Errorf("%w: filename is required", ErrValidation)
	}
	if a.StoragePath == "" {
		return fmt.Errorf("%w: storage_path is required", ErrValidation)
	}
	return nil
}
