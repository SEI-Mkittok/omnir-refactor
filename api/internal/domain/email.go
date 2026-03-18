package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EmailDirection indicates whether an email was sent or received.
type EmailDirection string

const (
	EmailDirectionInbound  EmailDirection = "inbound"
	EmailDirectionOutbound EmailDirection = "outbound"
)

// ContactEmail represents an email linked to a CRM contact or deal.
type ContactEmail struct {
	ID        uuid.UUID      `json:"id"`
	OrgID     uuid.UUID      `json:"org_id"`
	ContactID *uuid.UUID     `json:"contact_id,omitempty"`
	DealID    *uuid.UUID     `json:"deal_id,omitempty"`
	Direction EmailDirection `json:"direction"`
	FromAddr  string         `json:"from_addr"`
	ToAddr    string         `json:"to_addr"`
	Subject   string         `json:"subject"`
	Body      string         `json:"body"`
	ThreadID  string         `json:"thread_id"`
	MessageID *string        `json:"message_id,omitempty"`
	SentAt    time.Time      `json:"sent_at"`
	CreatedAt time.Time      `json:"created_at"`
}

// SendEmailRequest is the payload for composing and sending a new outbound email.
type SendEmailRequest struct {
	ContactID *uuid.UUID `json:"contact_id"`
	DealID    *uuid.UUID `json:"deal_id"`
	To        string     `json:"to"`
	Subject   string     `json:"subject"`
	Body      string     `json:"body"`
	// ThreadID groups replies into a conversation. Leave empty to start a new thread.
	ThreadID string `json:"thread_id"`
}

// Validate checks required fields.
func (r *SendEmailRequest) Validate() error {
	if r.To == "" {
		return fmt.Errorf("%w: to is required", ErrValidation)
	}
	if r.Subject == "" {
		return fmt.Errorf("%w: subject is required", ErrValidation)
	}
	if r.Body == "" {
		return fmt.Errorf("%w: body is required", ErrValidation)
	}
	return nil
}

// EmailFilter holds query parameters for listing contact emails.
type EmailFilter struct {
	OrgID     uuid.UUID
	ContactID *uuid.UUID
	DealID    *uuid.UUID
	Page      int
	Limit     int
}
