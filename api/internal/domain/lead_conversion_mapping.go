package domain

import (
	"time"

	"github.com/google/uuid"
)

type LeadConversionTarget string

const (
	LeadConversionTargetContact LeadConversionTarget = "contact"
	LeadConversionTargetAccount LeadConversionTarget = "account"
	LeadConversionTargetDeal    LeadConversionTarget = "deal"
)

func (t LeadConversionTarget) IsValid() bool {
	return t == LeadConversionTargetContact || t == LeadConversionTargetAccount || t == LeadConversionTargetDeal
}

type LeadConversionMapping struct {
	ID           uuid.UUID            `json:"id"`
	OrgID        uuid.UUID            `json:"org_id"`
	LeadField    string               `json:"lead_field"`
	TargetEntity LeadConversionTarget `json:"target_entity"`
	TargetField  string               `json:"target_field"`
	IsActive     bool                 `json:"is_active"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

