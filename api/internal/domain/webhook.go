package domain

import (
	"time"

	"github.com/google/uuid"
)

// WebhookEvent is the name of a CRM event that can trigger a webhook delivery.
type WebhookEvent string

const (
	WebhookEventDealCreated      WebhookEvent = "deal.created"
	WebhookEventDealUpdated      WebhookEvent = "deal.updated"
	WebhookEventDealStageChanged WebhookEvent = "deal.stage_changed"
	WebhookEventDealDeleted      WebhookEvent = "deal.deleted"
	WebhookEventContactCreated   WebhookEvent = "contact.created"
	WebhookEventContactUpdated   WebhookEvent = "contact.updated"
	WebhookEventActivityCreated  WebhookEvent = "activity.created"
	WebhookEventTicketCreated    WebhookEvent = "ticket.created"
)

// IsValid reports whether e is a known WebhookEvent.
func (e WebhookEvent) IsValid() bool {
	switch e {
	case WebhookEventDealCreated, WebhookEventDealUpdated, WebhookEventDealStageChanged,
		WebhookEventDealDeleted, WebhookEventContactCreated, WebhookEventContactUpdated,
		WebhookEventActivityCreated, WebhookEventTicketCreated:
		return true
	}
	return false
}

// DeliveryStatus represents the delivery state of a webhook attempt.
type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
)

// Webhook is a registered outbound webhook endpoint for an org.
type Webhook struct {
	ID        uuid.UUID      `json:"id"`
	OrgID     uuid.UUID      `json:"org_id"`
	URL       string         `json:"url"`
	Events    []WebhookEvent `json:"events"`
	Secret    string         `json:"secret,omitempty"`
	Active    bool           `json:"active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// WebhookPatch contains the fields that can be updated on a webhook.
type WebhookPatch struct {
	URL    *string        `json:"url,omitempty"`
	Events []WebhookEvent `json:"events,omitempty"`
	Active *bool          `json:"active,omitempty"`
}

// WebhookDelivery records one delivery attempt for a webhook event.
type WebhookDelivery struct {
	ID          uuid.UUID      `json:"id"`
	WebhookID   uuid.UUID      `json:"webhook_id"`
	Event       WebhookEvent   `json:"event"`
	Payload     []byte         `json:"payload"`
	Status      DeliveryStatus `json:"status"`
	Attempts    int            `json:"attempts"`
	NextRetryAt *time.Time     `json:"next_retry_at,omitempty"`
	DeliveredAt *time.Time     `json:"delivered_at,omitempty"`
	LastError   *string        `json:"last_error,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// WebhookEventPayload is the JSON body sent to a webhook endpoint.
type WebhookEventPayload struct {
	Event    WebhookEvent `json:"event"`
	OrgID    uuid.UUID    `json:"org_id"`
	EntityID uuid.UUID    `json:"entity_id"`
	Data     any          `json:"data"`
}
