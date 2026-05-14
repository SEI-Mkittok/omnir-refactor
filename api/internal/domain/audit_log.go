package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditActionCreated   AuditAction = "created"
	AuditActionUpdated   AuditAction = "updated"
	AuditActionDeleted   AuditAction = "deleted"
	AuditActionConverted AuditAction = "converted"
	AuditActionLogin     AuditAction = "login"
	AuditActionExport    AuditAction = "export"
)

type AuditEntityType string

const (
	AuditEntityContact     AuditEntityType = "contact"
	AuditEntityAccount     AuditEntityType = "account"
	AuditEntityDeal        AuditEntityType = "deal"
	AuditEntityLead        AuditEntityType = "lead"
	AuditEntitySharingRule AuditEntityType = "sharing_rule"
	AuditEntityUser        AuditEntityType = "user"
	AuditEntityView        AuditEntityType = "view"
)

// FieldChange captures the before/after value for a single field.
type FieldChange struct {
	From interface{} `json:"from"`
	To   interface{} `json:"to"`
}

// AuditChanges is a map of field name to before/after values.
type AuditChanges map[string]FieldChange

type AuditLog struct {
	ID           uuid.UUID       `json:"id"`
	OrgID        uuid.UUID       `json:"org_id"`
	UserID       *uuid.UUID      `json:"user_id,omitempty"`
	AgentID      *string         `json:"agent_id,omitempty"`
	ActorType    string          `json:"actor_type"`
	ActorName    *string         `json:"actor_name,omitempty"`
	ActorEmail   *string         `json:"actor_email,omitempty"`
	ActorDisplay string          `json:"actor_display"`
	Action       AuditAction     `json:"action"`
	EntityType   AuditEntityType `json:"entity_type"`
	EntityID     *uuid.UUID      `json:"entity_id,omitempty"`
	EntityName   *string         `json:"entity_name,omitempty"`
	Changes      AuditChanges    `json:"changes,omitempty"`
	IPAddress    *string         `json:"ip_address,omitempty"`
	UserAgent    *string         `json:"user_agent,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type AuditLogFilter struct {
	OrgID      uuid.UUID
	Q          string
	EntityType *AuditEntityType
	EntityID   *uuid.UUID
	UserID     *uuid.UUID
	Action     *AuditAction
	From       *time.Time
	To         *time.Time
	Page       int
	Limit      int
}

// AuditEntry is the input for writing a new audit log entry.
type AuditEntry struct {
	OrgID      uuid.UUID
	UserID     *uuid.UUID
	AgentID    *string
	Action     AuditAction
	EntityType AuditEntityType
	EntityID   *uuid.UUID
	EntityName *string
	Changes    AuditChanges
	IPAddress  *string
	UserAgent  *string
}

// MarshalChanges serialises AuditChanges to JSON bytes for storage.
func (c AuditChanges) MarshalJSON() ([]byte, error) {
	type plain map[string]FieldChange
	return json.Marshal(plain(c))
}
