package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditEntity identifies the entity type for an audit log entry.
type AuditEntity string

const (
	AuditEntityContact AuditEntity = "contact"
	AuditEntityAccount AuditEntity = "account"
	AuditEntityDeal    AuditEntity = "deal"
	AuditEntityView    AuditEntity = "view"
)

// AuditAction describes the action recorded.
type AuditAction string

const (
	AuditActionExport AuditAction = "export"
)

// AuditLogEntry represents a single entry in the audit log.
type AuditLogEntry struct {
	ID         uuid.UUID   `json:"id"`
	OrgID      uuid.UUID   `json:"org_id"`
	ActorID    uuid.UUID   `json:"actor_id"`
	Entity     AuditEntity `json:"entity"`
	EntityID   *uuid.UUID  `json:"entity_id,omitempty"`
	Action     AuditAction `json:"action"`
	Changes    interface{} `json:"changes,omitempty"`
	RemoteAddr string      `json:"remote_addr,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}
