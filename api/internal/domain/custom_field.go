package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CustomFieldEntityType identifies which entity type a field definition belongs to.
type CustomFieldEntityType string

const (
	CustomFieldEntityTicket  CustomFieldEntityType = "ticket"
	CustomFieldEntityContact CustomFieldEntityType = "contact"
	CustomFieldEntityLead    CustomFieldEntityType = "lead"
	CustomFieldEntityDeal    CustomFieldEntityType = "deal"
	CustomFieldEntityAccount CustomFieldEntityType = "account"
)

// IsValid returns true if the entity type is recognised.
func (e CustomFieldEntityType) IsValid() bool {
	switch e {
	case CustomFieldEntityTicket, CustomFieldEntityContact, CustomFieldEntityLead,
		CustomFieldEntityDeal, CustomFieldEntityAccount:
		return true
	}
	return false
}

// CustomFieldType is the data type of a custom field value.
type CustomFieldType string

const (
	CustomFieldTypeText        CustomFieldType = "text"
	CustomFieldTypeNumber      CustomFieldType = "number"
	CustomFieldTypeDate        CustomFieldType = "date"
	CustomFieldTypeURL         CustomFieldType = "url"
	CustomFieldTypeCheckbox    CustomFieldType = "checkbox"
	CustomFieldTypeSelect      CustomFieldType = "select"
	CustomFieldTypeMultiSelect CustomFieldType = "multiselect"
)

// IsValid returns true if the field type is recognised.
func (t CustomFieldType) IsValid() bool {
	switch t {
	case CustomFieldTypeText, CustomFieldTypeNumber, CustomFieldTypeDate,
		CustomFieldTypeURL, CustomFieldTypeCheckbox, CustomFieldTypeSelect, CustomFieldTypeMultiSelect:
		return true
	}
	return false
}

// CustomFieldDefinition is the admin-defined schema for a custom field on an entity.
type CustomFieldDefinition struct {
	ID         uuid.UUID             `json:"id"`
	OrgID      uuid.UUID             `json:"org_id"`
	EntityType CustomFieldEntityType `json:"entity_type"`
	Name       string                `json:"name"`   // snake_case key used in JSON payloads
	Label      string                `json:"label"`  // display name
	FieldType  CustomFieldType       `json:"field_type"`
	Options    []string              `json:"options,omitempty"` // for select/multi_select
	Required   bool                  `json:"required"`
	OrderIdx   int                   `json:"order_idx"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
	DeletedAt  *time.Time            `json:"deleted_at,omitempty"`
}

// CustomFieldDefinitionPatch holds optional fields for partial updates.
type CustomFieldDefinitionPatch struct {
	Label    *string   `json:"label,omitempty"`
	Options  []string  `json:"options,omitempty"`
	Required *bool     `json:"required,omitempty"`
	OrderIdx *int      `json:"order_idx,omitempty"`
}

// CustomFieldDefinitionFilter holds query parameters for listing definitions.
type CustomFieldDefinitionFilter struct {
	OrgID      uuid.UUID
	EntityType *CustomFieldEntityType
}

// ValidateCustomFields checks that the provided JSONB values conform to the given definitions.
// It returns a ValidationError listing all violations, or nil if all fields are valid.
func ValidateCustomFields(rawFields []byte, defs []*CustomFieldDefinition) error {
	if len(rawFields) == 0 || string(rawFields) == "null" {
		// Validate required fields even when custom_fields is absent.
		for _, d := range defs {
			if d.Required {
				return &ValidationError{Field: d.Name, Message: fmt.Sprintf("required custom field %q is missing", d.Name)}
			}
		}
		return nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rawFields, &fields); err != nil {
		return &ValidationError{Field: "custom_fields", Message: "must be a JSON object"}
	}

	defsByName := make(map[string]*CustomFieldDefinition, len(defs))
	for _, d := range defs {
		defsByName[d.Name] = d
	}

	// Check required fields are present.
	for _, d := range defs {
		if d.Required {
			if _, ok := fields[d.Name]; !ok {
				return &ValidationError{Field: d.Name, Message: fmt.Sprintf("required custom field %q is missing", d.Name)}
			}
		}
	}

	// Type-check provided values.
	for key, raw := range fields {
		def, ok := defsByName[key]
		if !ok {
			// Unknown fields are allowed (forward compat); skip.
			continue
		}
		if err := validateFieldValue(def, raw); err != nil {
			return err
		}
	}
	return nil
}

func validateFieldValue(def *CustomFieldDefinition, raw json.RawMessage) error {
	makeErr := func(msg string) *ValidationError {
		return &ValidationError{Field: def.Name, Message: fmt.Sprintf("custom field %q: %s", def.Name, msg)}
	}

	switch def.FieldType {
	case CustomFieldTypeText:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return makeErr("must be a string")
		}
	case CustomFieldTypeNumber:
		var n float64
		if err := json.Unmarshal(raw, &n); err != nil {
			return makeErr("must be a number")
		}
	case CustomFieldTypeDate:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return makeErr("must be a date string (YYYY-MM-DD)")
		}
	case CustomFieldTypeCheckbox:
		var b bool
		if err := json.Unmarshal(raw, &b); err != nil {
			return makeErr("must be a boolean")
		}
	case CustomFieldTypeURL:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return makeErr("must be a string")
		}
	case CustomFieldTypeSelect:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return makeErr("must be a string")
		}
		if len(def.Options) > 0 && !contains(def.Options, s) {
			return makeErr(fmt.Sprintf("must be one of: %v", def.Options))
		}
	case CustomFieldTypeMultiSelect:
		var vals []string
		if err := json.Unmarshal(raw, &vals); err != nil {
			return makeErr("must be an array of strings")
		}
		if len(def.Options) > 0 {
			for _, v := range vals {
				if !contains(def.Options, v) {
					return makeErr(fmt.Sprintf("%q is not a valid option; must be one of: %v", v, def.Options))
				}
			}
		}
	}
	return nil
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// ValidationError carries a single field-level validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %q: %s", e.Field, e.Message)
}
