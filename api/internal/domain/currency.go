package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrgCurrency struct {
	ID            uuid.UUID `json:"id"`
	OrgID         uuid.UUID `json:"org_id"`
	Code          string    `json:"code"`
	DisplayName   string    `json:"display_name"`
	Symbol        string    `json:"symbol"`
	DecimalPlaces int       `json:"decimal_places"`
	IsActive      bool      `json:"is_active"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type OrgCurrencyInput struct {
	Code          string `json:"code"`
	DisplayName   string `json:"display_name"`
	Symbol        string `json:"symbol"`
	DecimalPlaces int    `json:"decimal_places"`
	IsActive      bool   `json:"is_active"`
}

type OrgCurrencyUpdateRequest struct {
	DefaultCode string             `json:"default_code"`
	Currencies  []OrgCurrencyInput `json:"currencies"`
}

