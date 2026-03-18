package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
)

func def(name string, ft domain.CustomFieldType, required bool, opts ...string) *domain.CustomFieldDefinition {
	return &domain.CustomFieldDefinition{
		ID:        uuid.New(),
		Name:      name,
		FieldType: ft,
		Required:  required,
		Options:   opts,
	}
}

func raw(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func fields(kv ...any) []byte {
	m := make(map[string]any)
	for i := 0; i < len(kv)-1; i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	b, _ := json.Marshal(m)
	return b
}

func TestValidateCustomFields_NilFields(t *testing.T) {
	defs := []*domain.CustomFieldDefinition{
		def("notes", domain.CustomFieldTypeText, false),
	}
	err := domain.ValidateCustomFields(nil, defs)
	require.NoError(t, err)
}

func TestValidateCustomFields_RequiredFieldMissing(t *testing.T) {
	defs := []*domain.CustomFieldDefinition{
		def("customer_id", domain.CustomFieldTypeText, true),
	}
	err := domain.ValidateCustomFields(fields(), defs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "customer_id")
}

func TestValidateCustomFields_RequiredFieldPresent(t *testing.T) {
	defs := []*domain.CustomFieldDefinition{
		def("customer_id", domain.CustomFieldTypeText, true),
	}
	err := domain.ValidateCustomFields(fields("customer_id", "ABC-123"), defs)
	require.NoError(t, err)
}

func TestValidateCustomFields_TextType(t *testing.T) {
	defs := []*domain.CustomFieldDefinition{def("notes", domain.CustomFieldTypeText, false)}

	// valid
	require.NoError(t, domain.ValidateCustomFields(fields("notes", "hello"), defs))
	// invalid — number
	require.Error(t, domain.ValidateCustomFields(fields("notes", 42), defs))
}

func TestValidateCustomFields_NumberType(t *testing.T) {
	defs := []*domain.CustomFieldDefinition{def("amount", domain.CustomFieldTypeNumber, false)}

	require.NoError(t, domain.ValidateCustomFields(fields("amount", 3.14), defs))
	require.NoError(t, domain.ValidateCustomFields(fields("amount", 0), defs))
	require.Error(t, domain.ValidateCustomFields(fields("amount", "not-a-number"), defs))
}

func TestValidateCustomFields_BooleanType(t *testing.T) {
	defs := []*domain.CustomFieldDefinition{def("flagged", domain.CustomFieldTypeCheckbox, false)}

	require.NoError(t, domain.ValidateCustomFields(fields("flagged", true), defs))
	require.Error(t, domain.ValidateCustomFields(fields("flagged", "yes"), defs))
}

func TestValidateCustomFields_SelectType(t *testing.T) {
	defs := []*domain.CustomFieldDefinition{def("priority", domain.CustomFieldTypeSelect, false, "low", "medium", "high")}

	require.NoError(t, domain.ValidateCustomFields(fields("priority", "low"), defs))
	require.Error(t, domain.ValidateCustomFields(fields("priority", "critical"), defs))
}

func TestValidateCustomFields_MultiSelectType(t *testing.T) {
	defs := []*domain.CustomFieldDefinition{def("tags", domain.CustomFieldTypeMultiSelect, false, "billing", "tech", "hr")}

	require.NoError(t, domain.ValidateCustomFields(fields("tags", []string{"billing", "tech"}), defs))
	require.Error(t, domain.ValidateCustomFields(fields("tags", []string{"billing", "unknown"}), defs))
}

func TestValidateCustomFields_UnknownFieldsAllowed(t *testing.T) {
	// Unknown keys should be silently accepted (forward compat).
	defs := []*domain.CustomFieldDefinition{def("notes", domain.CustomFieldTypeText, false)}
	err := domain.ValidateCustomFields(fields("notes", "hi", "unknown_key", "value"), defs)
	require.NoError(t, err)
}

func TestValidateCustomFields_NullFields(t *testing.T) {
	defs := []*domain.CustomFieldDefinition{def("notes", domain.CustomFieldTypeText, false)}
	err := domain.ValidateCustomFields(raw(nil), defs)
	require.NoError(t, err)
}

func TestCustomFieldEntityType_IsValid(t *testing.T) {
	assert.True(t, domain.CustomFieldEntityTicket.IsValid())
	assert.True(t, domain.CustomFieldEntityContact.IsValid())
	assert.True(t, domain.CustomFieldEntityLead.IsValid())
	assert.True(t, domain.CustomFieldEntityType("deal").IsValid())
	assert.True(t, domain.CustomFieldEntityType("account").IsValid())
	assert.False(t, domain.CustomFieldEntityType("invoice").IsValid())
}

func TestCustomFieldType_IsValid(t *testing.T) {
	for _, ft := range []domain.CustomFieldType{
		domain.CustomFieldTypeText, domain.CustomFieldTypeNumber, domain.CustomFieldTypeDate,
		domain.CustomFieldTypeURL, domain.CustomFieldTypeCheckbox, domain.CustomFieldTypeSelect, domain.CustomFieldTypeMultiSelect,
	} {
		assert.True(t, ft.IsValid(), "expected %s to be valid", ft)
	}
	assert.False(t, domain.CustomFieldType("json").IsValid())
}
