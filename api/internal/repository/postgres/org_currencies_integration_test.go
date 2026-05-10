//go:build integration

package postgres_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

func TestCurrencyRepo_EnsureDefault_PreservesNonUSDDefault(t *testing.T) {
	pool, _ := setupDB(t)
	repo := postgres.NewCurrencyRepo(pool)

	_, err := repo.Replace(t.Context(), defaultOrgID, domain.OrgCurrencyUpdateRequest{
		DefaultCode: "EUR",
		Currencies: []domain.OrgCurrencyInput{
			{
				Code:          "USD",
				DisplayName:   "US Dollar",
				Symbol:        "$",
				DecimalPlaces: 2,
				IsActive:      true,
			},
			{
				Code:          "EUR",
				DisplayName:   "Euro",
				Symbol:        "EUR",
				DecimalPlaces: 2,
				IsActive:      true,
			},
		},
	})
	require.NoError(t, err)

	defaultCode, err := repo.GetDefaultCode(t.Context(), defaultOrgID)
	require.NoError(t, err)
	assert.Equal(t, "EUR", defaultCode)

	rows, err := repo.List(t.Context(), defaultOrgID)
	require.NoError(t, err)

	defaults := 0
	var foundEUR bool
	for _, row := range rows {
		if !row.IsDefault {
			continue
		}
		defaults++
		if row.Code == "EUR" {
			foundEUR = true
		}
	}

	assert.Equal(t, 1, defaults)
	assert.True(t, foundEUR)
}
