package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AccountRelationshipType string

const (
	AccountRelationshipTypeParent     AccountRelationshipType = "parent"
	AccountRelationshipTypeSubsidiary AccountRelationshipType = "subsidiary"
	AccountRelationshipTypePartner    AccountRelationshipType = "partner"
	AccountRelationshipTypeReseller   AccountRelationshipType = "reseller"
	AccountRelationshipTypeVendor     AccountRelationshipType = "vendor"
)

func (t AccountRelationshipType) IsValid() bool {
	switch t {
	case AccountRelationshipTypeParent,
		AccountRelationshipTypeSubsidiary,
		AccountRelationshipTypePartner,
		AccountRelationshipTypeReseller,
		AccountRelationshipTypeVendor:
		return true
	default:
		return false
	}
}

type AccountRelationship struct {
	ID               uuid.UUID               `json:"id"`
	OrgID            uuid.UUID               `json:"org_id"`
	ParentAccountID  uuid.UUID               `json:"parent_account_id"`
	ChildAccountID   uuid.UUID               `json:"child_account_id"`
	RelationshipType AccountRelationshipType `json:"relationship_type"`
	OwnershipPercent *float64                `json:"ownership_percent,omitempty"`
	EffectiveFrom    time.Time               `json:"effective_from"`
	EffectiveTo      *time.Time              `json:"effective_to,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
	DeletedAt        *time.Time              `json:"deleted_at,omitempty"`
	DeletedBy        *uuid.UUID              `json:"deleted_by,omitempty"`
}

func (r *AccountRelationship) Validate() error {
	if r.ParentAccountID == uuid.Nil {
		return fmt.Errorf("%w: parent_account_id is required", ErrValidation)
	}
	if r.ChildAccountID == uuid.Nil {
		return fmt.Errorf("%w: child_account_id is required", ErrValidation)
	}
	if r.ParentAccountID == r.ChildAccountID {
		return fmt.Errorf("%w: parent_account_id must differ from child_account_id", ErrValidation)
	}
	if !r.RelationshipType.IsValid() {
		return fmt.Errorf("%w: invalid relationship_type %q", ErrValidation, r.RelationshipType)
	}
	if r.OwnershipPercent != nil && (*r.OwnershipPercent < 0 || *r.OwnershipPercent > 100) {
		return fmt.Errorf("%w: ownership_percent must be between 0 and 100", ErrValidation)
	}
	if r.EffectiveTo != nil && r.EffectiveTo.Before(r.EffectiveFrom) {
		return fmt.Errorf("%w: effective_to must be on or after effective_from", ErrValidation)
	}
	return nil
}
