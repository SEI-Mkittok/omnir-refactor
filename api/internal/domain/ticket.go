package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusPending    TicketStatus = "pending"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

// IsValid returns true if the status value is recognised.
func (s TicketStatus) IsValid() bool {
	switch s {
	case TicketStatusOpen, TicketStatusInProgress, TicketStatusPending, TicketStatusResolved, TicketStatusClosed:
		return true
	}
	return false
}

type TicketPriority string

const (
	TicketPriorityLow      TicketPriority = "low"
	TicketPriorityMedium   TicketPriority = "medium"
	TicketPriorityHigh     TicketPriority = "high"
	TicketPriorityCritical TicketPriority = "critical"
)

// IsValid returns true if the priority value is recognised.
func (p TicketPriority) IsValid() bool {
	switch p {
	case TicketPriorityLow, TicketPriorityMedium, TicketPriorityHigh, TicketPriorityCritical:
		return true
	}
	return false
}

// Ticket represents a help-desk support ticket.
type Ticket struct {
	ID           uuid.UUID      `json:"id"`
	OrgID        uuid.UUID      `json:"org_id"`
	Subject      string         `json:"subject"`
	Description  *string        `json:"description,omitempty"`
	Status       TicketStatus   `json:"status"`
	Priority     TicketPriority `json:"priority"`
	AssigneeID   *uuid.UUID     `json:"assignee_id,omitempty"`
	ContactID    *uuid.UUID     `json:"contact_id,omitempty"`
	AccountID    *uuid.UUID     `json:"account_id,omitempty"`
	Source       *string        `json:"source,omitempty"`
	Tags         []string       `json:"tags"`
	CustomFields []byte         `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    *time.Time     `json:"deleted_at,omitempty"`
}

// TicketPatch holds optional fields for partial ticket updates.
type TicketPatch struct {
	Subject      *string          `json:"subject,omitempty"`
	Description  *string          `json:"description,omitempty"`
	Status       *TicketStatus    `json:"status,omitempty"`
	Priority     *TicketPriority  `json:"priority,omitempty"`
	AssigneeID   *uuid.UUID       `json:"assignee_id,omitempty"`
	ContactID    *uuid.UUID       `json:"contact_id,omitempty"`
	AccountID    *uuid.UUID       `json:"account_id,omitempty"`
	Source       *string          `json:"source,omitempty"`
	CustomFields json.RawMessage  `json:"custom_fields,omitempty"`
}

// TicketFilter holds query parameters for listing tickets.
type TicketFilter struct {
	OrgID      uuid.UUID
	Status     *TicketStatus
	Priority   *TicketPriority
	AssigneeID *uuid.UUID
	ContactID  *uuid.UUID
	Q          string
	Page       int
	Limit      int
	Sort       string
	Order      string
}

// TicketComment is a reply or internal note attached to a ticket.
type TicketComment struct {
	ID         uuid.UUID  `json:"id"`
	TicketID   uuid.UUID  `json:"ticket_id"`
	OrgID      uuid.UUID  `json:"org_id"`
	AuthorID   *uuid.UUID `json:"author_id,omitempty"`
	Body       string     `json:"body"`
	IsInternal bool       `json:"is_internal"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

// TicketCommentFilter holds query parameters for listing ticket comments.
type TicketCommentFilter struct {
	TicketID   uuid.UUID
	OrgID      uuid.UUID
	IsInternal *bool // nil = all, true = internal only, false = public only
}

// TicketAttachment is a file attached to a ticket.
type TicketAttachment struct {
	ID          uuid.UUID  `json:"id"`
	TicketID    uuid.UUID  `json:"ticket_id"`
	OrgID       uuid.UUID  `json:"org_id"`
	UploadedBy  *uuid.UUID `json:"uploaded_by,omitempty"`
	Filename    string     `json:"filename"`
	ContentType string     `json:"content_type"`
	SizeBytes   *int64     `json:"size_bytes,omitempty"`
	StorageURL  string     `json:"url"`
	CreatedAt   time.Time  `json:"created_at"`
}
