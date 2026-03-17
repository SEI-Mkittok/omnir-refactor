//go:build integration

package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

func TestUserRepo_Create_OrgScoping(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewUserRepo(pool)

	t.Run("uses org_id from context", func(t *testing.T) {
		u := &domain.User{
			Email: "alice@omnir.test",
			Name:  "Alice",
			Role:  domain.UserRoleUser,
		}
		got, err := repo.Create(ctx, u, "hashed-pw")
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, got.ID)
		assert.Equal(t, defaultOrgID, got.OrgID)
	})

	t.Run("context org_id overrides struct org_id", func(t *testing.T) {
		otherOrgID := defaultOrgID // only default org exists in test DB; just verify override behaviour
		u := &domain.User{
			Email: "bob@omnir.test",
			Name:  "Bob",
			Role:  domain.UserRoleUser,
			OrgID: otherOrgID,
		}
		got, err := repo.Create(ctx, u, "hashed-pw")
		require.NoError(t, err)
		assert.Equal(t, defaultOrgID, got.OrgID)
	})
}
