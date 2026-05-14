package domain

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestModuleLayoutValidationRejectsHiddenRequiredSystemField(t *testing.T) {
	layout := DefaultModuleLayout(uuid.New(), CustomFieldEntityContact, nil)
	layout.Blocks[0].Fields[0].Visible = false

	err := ValidateModuleLayout(layout, nil)
	require.ErrorIs(t, err, ErrValidation)
	require.Contains(t, err.Error(), "first_name")
}

func TestModuleLayoutValidationRejectsRelaxedRequiredSystemField(t *testing.T) {
	layout := DefaultModuleLayout(uuid.New(), CustomFieldEntityDeal, nil)
	for bIdx := range layout.Blocks {
		for fIdx := range layout.Blocks[bIdx].Fields {
			if layout.Blocks[bIdx].Fields[fIdx].FieldKey == "pipeline_id" {
				layout.Blocks[bIdx].Fields[fIdx].Required = false
			}
		}
	}

	err := ValidateModuleLayout(layout, nil)
	require.ErrorIs(t, err, ErrValidation)
	require.Contains(t, err.Error(), "pipeline_id")
}

func TestModuleLayoutValidationRejectsRequiredCustomFieldRelaxation(t *testing.T) {
	defs := []*CustomFieldDefinition{{
		ID:         uuid.New(),
		EntityType: CustomFieldEntityDeal,
		Name:       "approval_code",
		Label:      "Approval code",
		FieldType:  CustomFieldTypeText,
		Required:   true,
	}}
	layout := DefaultModuleLayout(uuid.New(), CustomFieldEntityDeal, defs)
	for bIdx := range layout.Blocks {
		for fIdx := range layout.Blocks[bIdx].Fields {
			if layout.Blocks[bIdx].Fields[fIdx].FieldKey == "approval_code" {
				layout.Blocks[bIdx].Fields[fIdx].Required = false
			}
		}
	}

	err := ValidateModuleLayout(layout, defs)
	require.ErrorIs(t, err, ErrValidation)
	require.Contains(t, err.Error(), "approval_code")
}

func TestDefaultLayoutsKeepContactAndLeadEmailOptional(t *testing.T) {
	for _, entityType := range []CustomFieldEntityType{CustomFieldEntityContact, CustomFieldEntityLead} {
		layout := DefaultModuleLayout(uuid.New(), entityType, nil)
		var emailField *ModuleLayoutField
		for bIdx := range layout.Blocks {
			for fIdx := range layout.Blocks[bIdx].Fields {
				if layout.Blocks[bIdx].Fields[fIdx].FieldKey == "email" {
					emailField = &layout.Blocks[bIdx].Fields[fIdx]
				}
			}
		}
		require.NotNil(t, emailField)
		require.False(t, emailField.Required)
		require.True(t, emailField.Visible)
	}
}

func TestModuleLayoutValidationRejectsUnknownAndDuplicateFields(t *testing.T) {
	layout := DefaultModuleLayout(uuid.New(), CustomFieldEntityAccount, nil)
	layout.Blocks[0].Fields = append(layout.Blocks[0].Fields, ModuleLayoutField{
		Source:      ModuleLayoutFieldSourceStandard,
		FieldKey:    "not_a_field",
		Label:       "Nope",
		Visible:     true,
		QuickCreate: true,
	})
	err := ValidateModuleLayout(layout, nil)
	require.ErrorIs(t, err, ErrValidation)
	require.Contains(t, err.Error(), "not_a_field")

	layout = DefaultModuleLayout(uuid.New(), CustomFieldEntityAccount, nil)
	layout.Blocks[0].Fields = append(layout.Blocks[0].Fields, layout.Blocks[0].Fields[0])
	err = ValidateModuleLayout(layout, nil)
	require.ErrorIs(t, err, ErrValidation)
	require.Contains(t, err.Error(), "duplicate")
}

func TestModuleLayoutRequiredValuesIncludesCustomFields(t *testing.T) {
	defs := []*CustomFieldDefinition{{
		ID:         uuid.New(),
		EntityType: CustomFieldEntityContact,
		Name:       "territory",
		Label:      "Territory",
		FieldType:  CustomFieldTypeText,
		Required:   true,
	}}
	layout := DefaultModuleLayout(uuid.New(), CustomFieldEntityContact, defs)
	payload := map[string]json.RawMessage{
		"first_name": json.RawMessage(`"Ada"`),
		"last_name":  json.RawMessage(`"Lovelace"`),
		"email":      json.RawMessage(`"ada@example.com"`),
	}

	err := ValidateModuleLayoutRequiredValues(layout, payload)
	require.ErrorIs(t, err, ErrValidation)
	require.Contains(t, err.Error(), "territory")

	payload["custom_fields"] = json.RawMessage(`{"territory":"North"}`)
	require.NoError(t, ValidateModuleLayoutRequiredValues(layout, payload))
}

func TestModuleRelationshipValidationRejectsUnsafeSameModuleCardinality(t *testing.T) {
	def := &ModuleRelationshipDefinition{
		OrgID:           uuid.New(),
		RelationshipKey: "custom_account_parent",
		FromEntityType:  CustomFieldEntityAccount,
		ToEntityType:    CustomFieldEntityAccount,
		Label:           "Account parent",
		Cardinality:     RelationshipCardinalityManyToOne,
		StorageStrategy: RelationshipStorageCRMEntityLinks,
		IsEnabled:       true,
	}

	err := ValidateModuleRelationshipDefinition(def)
	require.ErrorIs(t, err, ErrValidation)
}

func TestMergeDefaultRelationshipDefinitionsKeepsSystemRows(t *testing.T) {
	orgID := uuid.New()
	stored := []*ModuleRelationshipDefinition{{
		ID:              uuid.New(),
		OrgID:           orgID,
		RelationshipKey: "deal_account",
		Label:           "Opportunity account",
		IsEnabled:       false,
		OrderIdx:        5,
	}}

	merged := MergeDefaultRelationshipDefinitions(orgID, stored)
	var found *ModuleRelationshipDefinition
	for _, row := range merged {
		if row.RelationshipKey == "deal_account" {
			found = row
			break
		}
	}
	require.NotNil(t, found)
	require.True(t, found.SystemLocked)
	require.Equal(t, RelationshipStorageNative, found.StorageStrategy)
	require.Equal(t, "Opportunity account", found.Label)
	require.False(t, found.IsEnabled)
}
