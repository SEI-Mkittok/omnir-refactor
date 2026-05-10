package domain

import (
	"time"

	"github.com/google/uuid"
)

type PicklistValue struct {
	ID            uuid.UUID `json:"id"`
	OrgID         uuid.UUID `json:"org_id"`
	CustomFieldID uuid.UUID `json:"custom_field_id"`
	Value         string    `json:"value"`
	DisplayLabel  string    `json:"display_label"`
	OrderIdx      int       `json:"order_idx"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PicklistValueInput struct {
	Value        string `json:"value"`
	DisplayLabel string `json:"display_label"`
	OrderIdx     int    `json:"order_idx"`
	IsActive     bool   `json:"is_active"`
}

type PicklistDependency struct {
	ID            uuid.UUID             `json:"id"`
	OrgID         uuid.UUID             `json:"org_id"`
	EntityType    CustomFieldEntityType `json:"entity_type"`
	SourceFieldID uuid.UUID             `json:"source_field_id"`
	TargetFieldID uuid.UUID             `json:"target_field_id"`
	Mapping       map[string][]string   `json:"mapping"`
	IsActive      bool                  `json:"is_active"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type PicklistDependencyInput struct {
	EntityType    CustomFieldEntityType `json:"entity_type"`
	SourceFieldID uuid.UUID             `json:"source_field_id"`
	TargetFieldID uuid.UUID             `json:"target_field_id"`
	Mapping       map[string][]string   `json:"mapping"`
	IsActive      bool                  `json:"is_active"`
}
