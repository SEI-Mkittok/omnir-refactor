package domain

import (
	"encoding/json"
	"fmt"
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
	OrgID        uuid.UUID       `json:"org_id"`
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

// IsValid returns true if the size is a known value.
func (s AccountSize) IsValid() bool {
	switch s {
	case AccountSize1_10, AccountSize11_50, AccountSize51_200, AccountSize201_500, AccountSize501Plus:
		return true
	}
	return false
}

// Validate checks required fields and value constraints on an Account.
func (a *Account) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if a.OwnerID == uuid.Nil {
		return fmt.Errorf("%w: owner_id is required", ErrValidation)
	}
	if a.Size != nil && !a.Size.IsValid() {
		return fmt.Errorf("%w: invalid size %q", ErrValidation, *a.Size)
	}
	return nil
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
	OrgID    uuid.UUID
	Q        string
	OwnerID  *uuid.UUID
	Industry *string
	Size     *AccountSize
	Page     int
	Limit    int
	Sort     string
	Order    string
}
