package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AccountSize string

const (
	AccountSize1_10    AccountSize = "1-10"
	AccountSize11_50   AccountSize = "11-50"
	AccountSize51_200  AccountSize = "51-200"
	AccountSize201_500 AccountSize = "201-500"
	AccountSize501Plus AccountSize = "501+"
)

type Account struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	Domain       *string         `json:"domain,omitempty"`
	Industry     *string         `json:"industry,omitempty"`
	Size         *AccountSize    `json:"size,omitempty"`
	OwnerID      uuid.UUID       `json:"owner_id"`
	Tags         []string        `json:"tags"`
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    *time.Time      `json:"deleted_at,omitempty"`
}

// AccountPatch holds optional fields for partial updates.
type AccountPatch struct {
	Name         *string         `json:"name,omitempty"`
	Domain       *string         `json:"domain,omitempty"`
	Industry     *string         `json:"industry,omitempty"`
	Size         *AccountSize    `json:"size,omitempty"`
	OwnerID      *uuid.UUID      `json:"owner_id,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
	CustomFields json.RawMessage `json:"custom_fields,omitempty"`
}

// AccountFilter holds query parameters for listing accounts.
type AccountFilter struct {
	Q       string
	OwnerID *uuid.UUID
	Page    int
	Limit   int
	Sort    string
	Order   string
}
