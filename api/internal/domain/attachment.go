package domain

import (
	"time"

	"github.com/google/uuid"
)

// EntityType identifies which CRM entity an attachment belongs to.
type EntityType string

const (
	EntityTypeContact EntityType = "contact"
	EntityTypeAccount EntityType = "account"
	EntityTypeDeal    EntityType = "deal"
	EntityTypeTicket  EntityType = "ticket"
)

// EntityAttachment represents a file attached to a CRM entity.
type EntityAttachment struct {
	ID         uuid.UUID  `json:"id"`
	OrgID      uuid.UUID  `json:"org_id"`
	EntityType EntityType `json:"entity_type"`
	EntityID   uuid.UUID  `json:"entity_id"`
	Filename   string     `json:"filename"`
	MimeType   string     `json:"mime_type"`
	SizeBytes  int64      `json:"size_bytes"`
	StorageKey string     `json:"storage_key"`
	UploadedBy uuid.UUID  `json:"uploaded_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
