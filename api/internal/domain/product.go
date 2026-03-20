package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the org's catalog.
type Product struct {
	ID             uuid.UUID  `json:"id"`
	OrgID          uuid.UUID  `json:"org_id"`
	Name           string     `json:"name"`
	SKU            string     `json:"sku,omitempty"`
	Description    string     `json:"description,omitempty"`
	UnitPriceCents int64      `json:"unit_price_cents"`
	Currency       string     `json:"currency"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (p *Product) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if p.UnitPriceCents < 0 {
		return fmt.Errorf("%w: unit_price_cents must be >= 0", ErrValidation)
	}
	return nil
}

// ProductPatch holds optional fields for partial product updates.
type ProductPatch struct {
	Name           *string `json:"name,omitempty"`
	SKU            *string `json:"sku,omitempty"`
	Description    *string `json:"description,omitempty"`
	UnitPriceCents *int64  `json:"unit_price_cents,omitempty"`
	Currency       *string `json:"currency,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
}

// ProductFilter holds query params for listing products.
type ProductFilter struct {
	OrgID    uuid.UUID
	Q        string
	IsActive *bool
	Page     int
	Limit    int
}
