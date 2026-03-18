package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NoteEntityType is the type of CRM record a note belongs to.
type NoteEntityType string

const (
	NoteEntityContact NoteEntityType = "contact"
	NoteEntityAccount NoteEntityType = "account"
	NoteEntityDeal    NoteEntityType = "deal"
	NoteEntityLead    NoteEntityType = "lead"
)

// IsValid returns true if the entity type is known.
func (t NoteEntityType) IsValid() bool {
	switch t {
	case NoteEntityContact, NoteEntityAccount, NoteEntityDeal, NoteEntityLead:
		return true
	}
	return false
}

// Note is a free-text note attached to a CRM record.
type Note struct {
	ID         uuid.UUID      `json:"id"`
	OrgID      uuid.UUID      `json:"org_id"`
	Content    string         `json:"content"`
	EntityType NoteEntityType `json:"entity_type"`
	EntityID   uuid.UUID      `json:"entity_id"`
	AuthorID   uuid.UUID      `json:"author_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  *time.Time     `json:"deleted_at,omitempty"`
}

// Validate checks required fields on a Note.
func (n *Note) Validate() error {
	if n.Content == "" {
		return fmt.Errorf("%w: content is required", ErrValidation)
	}
	if !n.EntityType.IsValid() {
		return fmt.Errorf("%w: invalid entity_type %q", ErrValidation, n.EntityType)
	}
	if n.EntityID == uuid.Nil {
		return fmt.Errorf("%w: entity_id is required", ErrValidation)
	}
	if n.AuthorID == uuid.Nil {
		return fmt.Errorf("%w: author_id is required", ErrValidation)
	}
	return nil
}

// NoteFilter holds query parameters for listing notes on an entity.
type NoteFilter struct {
	OrgID      uuid.UUID
	EntityType NoteEntityType
	EntityID   uuid.UUID
	Page       int
	Limit      int
}
