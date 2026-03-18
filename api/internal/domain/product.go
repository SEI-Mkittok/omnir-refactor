package domain

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a line-item product in the catalog.
type Product struct {
	ID          uuid.UUID `json:"id"`
	OrgID       uuid.UUID `json:"org_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	SKU         *string   `json:"sku,omitempty"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductPatch holds optional fields for partial product updates.
type ProductPatch struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	SKU         *string  `json:"sku,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Currency    *string  `json:"currency,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

// PriceBook is a named set of product price overrides.
type PriceBook struct {
	ID        uuid.UUID         `json:"id"`
	OrgID     uuid.UUID         `json:"org_id"`
	Name      string            `json:"name"`
	IsDefault bool              `json:"is_default"`
	Entries   []*PriceBookEntry `json:"entries,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// PriceBookEntry links a product to a price book with an optional override.
type PriceBookEntry struct {
	ID            uuid.UUID `json:"id"`
	PriceBookID   uuid.UUID `json:"price_book_id"`
	ProductID     uuid.UUID `json:"product_id"`
	Product       *Product  `json:"product,omitempty"`
	PriceOverride *float64  `json:"price_override,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// DealLineItem is a line on a deal quote.
type DealLineItem struct {
	ID          uuid.UUID  `json:"id"`
	DealID      uuid.UUID  `json:"deal_id"`
	ProductID   *uuid.UUID `json:"product_id,omitempty"`
	Name        string     `json:"name"`
	Quantity    float64    `json:"quantity"`
	UnitPrice   float64    `json:"unit_price"`
	DiscountPct float64    `json:"discount_pct"`
	Subtotal    float64    `json:"subtotal"`
	Position    int        `json:"position"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
