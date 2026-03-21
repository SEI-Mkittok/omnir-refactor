package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type QuoteStatus string

const (
	QuoteStatusDraft    QuoteStatus = "draft"
	QuoteStatusSent     QuoteStatus = "sent"
	QuoteStatusApproved QuoteStatus = "approved"
	QuoteStatusRejected QuoteStatus = "rejected"
	QuoteStatusExpired  QuoteStatus = "expired"
)

func (s QuoteStatus) IsValid() bool {
	switch s {
	case QuoteStatusDraft, QuoteStatusSent, QuoteStatusApproved, QuoteStatusRejected, QuoteStatusExpired:
		return true
	}
	return false
}

// QuoteLineItem represents a single line on a quote.
type QuoteLineItem struct {
	ID             uuid.UUID  `json:"id"`
	QuoteID        uuid.UUID  `json:"quote_id"`
	ProductID      *uuid.UUID `json:"product_id,omitempty"`
	ProductName    string     `json:"product_name"`
	Description    string     `json:"description,omitempty"`
	Quantity       float64    `json:"quantity"`
	UnitPriceCents int64      `json:"unit_price_cents"`
	DiscountPct    float64    `json:"discount_pct"`
	TotalCents     int64      `json:"total_cents"`
	SortOrder      int        `json:"sort_order"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ComputeTotal recalculates TotalCents from quantity, unit_price, and discount.
func (li *QuoteLineItem) ComputeTotal() {
	gross := float64(li.UnitPriceCents) * li.Quantity
	discount := gross * (li.DiscountPct / 100.0)
	li.TotalCents = int64(gross - discount)
}

// Quote represents a sales quote linked to a deal.
type Quote struct {
	ID         uuid.UUID       `json:"id"`
	OrgID      uuid.UUID       `json:"org_id"`
	DealID     *uuid.UUID      `json:"deal_id,omitempty"`
	ContactID  *uuid.UUID      `json:"contact_id,omitempty"`
	Title      string          `json:"title"`
	Status     QuoteStatus     `json:"status"`
	Currency   string          `json:"currency"`
	ValidUntil *time.Time      `json:"valid_until,omitempty"`
	Notes      string          `json:"notes,omitempty"`
	SentAt     *time.Time      `json:"sent_at,omitempty"`
	ApprovedAt *time.Time      `json:"approved_at,omitempty"`
	RejectedAt *time.Time      `json:"rejected_at,omitempty"`
	CreatedBy  *uuid.UUID      `json:"created_by,omitempty"`
	LineItems  []QuoteLineItem `json:"line_items,omitempty"`
	TotalCents int64           `json:"total_cents"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`

	// Joined
	Contact *Contact `json:"contact,omitempty"`
	Deal    *Deal    `json:"deal,omitempty"`
}

func (q *Quote) Validate() error {
	if q.Title == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if q.Status != "" && !q.Status.IsValid() {
		return fmt.Errorf("%w: invalid status %q", ErrValidation, q.Status)
	}
	return nil
}

// ComputeTotal sums all line item totals.
func (q *Quote) ComputeTotal() {
	var total int64
	for i := range q.LineItems {
		q.LineItems[i].ComputeTotal()
		total += q.LineItems[i].TotalCents
	}
	q.TotalCents = total
}

// QuoteLineItemInput is used when creating/updating line items.
type QuoteLineItemInput struct {
	ProductID      *uuid.UUID `json:"product_id,omitempty"`
	ProductName    string     `json:"product_name"`
	Description    string     `json:"description,omitempty"`
	Quantity       float64    `json:"quantity"`
	UnitPriceCents int64      `json:"unit_price_cents"`
	DiscountPct    float64    `json:"discount_pct"`
	SortOrder      int        `json:"sort_order"`
}

// QuotePatch holds optional fields for partial quote updates.
type QuotePatch struct {
	Title      *string              `json:"title,omitempty"`
	Status     *QuoteStatus         `json:"status,omitempty"`
	Currency   *string              `json:"currency,omitempty"`
	ValidUntil *time.Time           `json:"valid_until,omitempty"`
	Notes      *string              `json:"notes,omitempty"`
	ContactID  *uuid.UUID           `json:"contact_id,omitempty"`
	DealID     *uuid.UUID           `json:"deal_id,omitempty"`
	LineItems  []QuoteLineItemInput `json:"line_items,omitempty"`
}

// QuoteFilter holds query params for listing quotes.
type QuoteFilter struct {
	OrgID     uuid.UUID
	DealID    *uuid.UUID
	ContactID *uuid.UUID
	Status    *QuoteStatus
	Q         string
	Page      int
	Limit     int
}

// SendQuoteRequest is used for the send-via-email action.
type SendQuoteRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject,omitempty"`
	Message string `json:"message,omitempty"`
}
