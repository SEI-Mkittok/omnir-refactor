package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EmailProvider identifies which email service a connection uses.
type EmailProvider string

const (
	EmailProviderGmail   EmailProvider = "gmail"
	EmailProviderOutlook EmailProvider = "outlook"
)

// IsValid reports whether p is a known EmailProvider.
func (p EmailProvider) IsValid() bool {
	return p == EmailProviderGmail || p == EmailProviderOutlook
}

// EmailConnection represents a user's OAuth connection to Gmail or Outlook.
type EmailConnection struct {
	ID           uuid.UUID     `json:"id"`
	OrgID        uuid.UUID     `json:"org_id"`
	UserID       uuid.UUID     `json:"user_id"`
	Provider     EmailProvider `json:"provider"`
	EmailAddress string        `json:"email_address"`
	// AccessToken is stored encrypted at rest; omitted from API responses.
	AccessToken  string     `json:"-"`
	RefreshToken *string    `json:"-"`
	TokenExpiry  time.Time  `json:"token_expiry"`
	SyncCursor   *string    `json:"-"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// EmailConnectionPatch holds mutable fields for updating a connection.
type EmailConnectionPatch struct {
	AccessToken  *string
	RefreshToken *string
	TokenExpiry  *time.Time
	SyncCursor   *string
	LastSyncedAt *time.Time
	EmailAddress *string
}

// EmailConnectionFilter is used when listing connections.
type EmailConnectionFilter struct {
	OrgID    uuid.UUID
	UserID   *uuid.UUID
	Provider *EmailProvider
}

// Validate checks required fields on a new EmailConnection.
func (c *EmailConnection) Validate() error {
	if c.OrgID == uuid.Nil {
		return fmt.Errorf("%w: org_id is required", ErrValidation)
	}
	if c.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrValidation)
	}
	if !c.Provider.IsValid() {
		return fmt.Errorf("%w: invalid provider %q", ErrValidation, c.Provider)
	}
	return nil
}

// EmailInboxMessage is a single email fetched from Gmail or Outlook.
type EmailInboxMessage struct {
	ID           uuid.UUID      `json:"id"`
	OrgID        uuid.UUID      `json:"org_id"`
	ConnectionID uuid.UUID      `json:"connection_id"`
	MessageID    string         `json:"message_id"`
	ThreadID     string         `json:"thread_id"`
	FromAddr     string         `json:"from_addr"`
	ToAddrs      []string       `json:"to_addrs"`
	Subject      string         `json:"subject"`
	BodyText     *string        `json:"body_text,omitempty"`
	BodyHTML     *string        `json:"body_html,omitempty"`
	ContactID    *uuid.UUID     `json:"contact_id,omitempty"`
	Direction    EmailDirection `json:"direction"`
	SentAt       time.Time      `json:"sent_at"`
	ReadAt       *time.Time     `json:"read_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

// EmailInboxThreadSummary is the thread-first view used by the inbox UI.
type EmailInboxThreadSummary struct {
	ThreadID      string     `json:"thread_id"`
	OrgID         uuid.UUID  `json:"org_id"`
	ConnectionID  uuid.UUID  `json:"connection_id"`
	Subject       string     `json:"subject"`
	Participants  []string   `json:"participants"`
	Snippet       string     `json:"snippet"`
	Unread        bool       `json:"unread"`
	MessageCount  int        `json:"message_count"`
	LastMessageAt time.Time  `json:"last_message_at"`
	ContactID     *uuid.UUID `json:"contact_id,omitempty"`
}

// EmailInboxThreadDetail is a thread summary plus ordered messages.
type EmailInboxThreadDetail struct {
	ThreadSummary EmailInboxThreadSummary `json:"thread"`
	Messages      []*EmailInboxMessage    `json:"messages"`
}

// EmailInboxFilter holds query parameters for listing inbox messages.
type EmailInboxFilter struct {
	OrgID        uuid.UUID
	ConnectionID *uuid.UUID
	ContactID    *uuid.UUID
	ThreadID     *string
	UnreadOnly   bool
	Page         int
	Limit        int
}

// SendInboxEmailRequest is the payload for sending via a connected account.
type SendInboxEmailRequest struct {
	ConnectionID uuid.UUID `json:"connection_id"`
	To           []string  `json:"to"`
	CC           []string  `json:"cc,omitempty"`
	BCC          []string  `json:"bcc,omitempty"`
	Subject      string    `json:"subject"`
	BodyHTML     string    `json:"body_html"`
	// ThreadID links a reply to an existing conversation. Optional.
	ThreadID  *string    `json:"thread_id,omitempty"`
	ContactID *uuid.UUID `json:"contact_id,omitempty"`
}

// Validate checks required fields.
func (r *SendInboxEmailRequest) Validate() error {
	if r.ConnectionID == uuid.Nil {
		return fmt.Errorf("%w: connection_id is required", ErrValidation)
	}
	if len(r.To) == 0 {
		return fmt.Errorf("%w: to is required", ErrValidation)
	}
	if r.Subject == "" {
		return fmt.Errorf("%w: subject is required", ErrValidation)
	}
	if r.BodyHTML == "" {
		return fmt.Errorf("%w: body_html is required", ErrValidation)
	}
	return nil
}
