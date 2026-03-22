package domain

import (
	"time"

	"github.com/google/uuid"
)

// BillingPlan represents the subscription tier for an org.
type BillingPlan string

const (
	BillingPlanFree       BillingPlan = "free"
	BillingPlanPro        BillingPlan = "pro"
	BillingPlanEnterprise BillingPlan = "enterprise"
)

// SubscriptionStatus mirrors Stripe subscription statuses.
type SubscriptionStatus string

const (
	SubscriptionStatusActive     SubscriptionStatus = "active"
	SubscriptionStatusTrialing   SubscriptionStatus = "trialing"
	SubscriptionStatusPastDue    SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled   SubscriptionStatus = "canceled"
	SubscriptionStatusIncomplete SubscriptionStatus = "incomplete"
)

// OrgPlanRecord holds the billing state for an org.
type OrgPlanRecord struct {
	ID                   uuid.UUID          `json:"id"`
	OrgID                uuid.UUID          `json:"org_id"`
	Plan                 BillingPlan        `json:"plan"`
	StripeCustomerID     *string            `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string            `json:"stripe_subscription_id,omitempty"`
	Status               SubscriptionStatus `json:"status"`
	CurrentPeriodEnd     *time.Time         `json:"current_period_end,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

// InvoiceStatus mirrors Stripe invoice statuses.
type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "draft"
	InvoiceStatusOpen          InvoiceStatus = "open"
	InvoiceStatusPaid          InvoiceStatus = "paid"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
	InvoiceStatusVoid          InvoiceStatus = "void"
)

// Invoice represents a Stripe invoice record stored for an org.
type Invoice struct {
	ID              uuid.UUID     `json:"id"`
	OrgID           uuid.UUID     `json:"org_id"`
	StripeInvoiceID string        `json:"stripe_invoice_id"`
	AmountCents     int64         `json:"amount_cents"`
	Currency        string        `json:"currency"`
	Status          InvoiceStatus `json:"status"`
	PDFURL          *string       `json:"pdf_url,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}

// OrgPlanPatch is used to update billing state (internal use — webhook-driven).
type OrgPlanPatch struct {
	Plan                 *BillingPlan        `json:"plan,omitempty"`
	StripeCustomerID     *string             `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string             `json:"stripe_subscription_id,omitempty"`
	Status               *SubscriptionStatus `json:"status,omitempty"`
	CurrentPeriodEnd     *time.Time          `json:"current_period_end,omitempty"`
}
