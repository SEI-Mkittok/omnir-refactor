//go:build integration

package postgres_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

func TestPicklistRepo_RemapAndDeleteValue_SelectScalarValue(t *testing.T) {
	pool, ctx := setupDB(t)

	customFieldRepo := postgres.NewCustomFieldDefinitionRepo(pool)
	picklistRepo := postgres.NewPicklistRepo(pool)
	leadRepo := postgres.NewLeadRepo(pool)

	fieldName := fmt.Sprintf("region_%s", uuid.NewString()[:8])
	def, err := customFieldRepo.Create(ctx, &domain.CustomFieldDefinition{
		EntityType: domain.CustomFieldEntityLead,
		Name:       fieldName,
		Label:      "Region",
		FieldType:  domain.CustomFieldTypeSelect,
		Options:    []string{"North", "South"},
		OrderIdx:   1,
	})
	require.NoError(t, err)

	_, err = picklistRepo.UpsertValues(ctx, defaultOrgID, def.ID, []domain.PicklistValueInput{
		{Value: "North", DisplayLabel: "North", IsActive: true},
		{Value: "South", DisplayLabel: "South", IsActive: true},
	})
	require.NoError(t, err)

	email := fmt.Sprintf("lead+%s@omnir.test", uuid.NewString())
	customFields := []byte(fmt.Sprintf(`{"%s":"North"}`, fieldName))
	lead, err := leadRepo.Create(ctx, &domain.Lead{
		FirstName:    "Scalar",
		LastName:     "Select",
		Email:        &email,
		Status:       domain.LeadStatusNew,
		CustomFields: &customFields,
	})
	require.NoError(t, err)

	toValue := "South"
	err = picklistRepo.RemapAndDeleteValue(ctx, defaultOrgID, def.ID, "North", &toValue)
	require.NoError(t, err)

	updatedLead, err := leadRepo.GetByID(ctx, lead.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedLead.CustomFields)

	var got map[string]any
	require.NoError(t, json.Unmarshal(*updatedLead.CustomFields, &got))
	assert.Equal(t, "South", got[fieldName])
}
