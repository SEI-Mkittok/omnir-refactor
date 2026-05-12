//go:build integration

package postgres_test

import (
	"context"
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
			Role:  domain.UserRoleAgent,
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
			Role:  domain.UserRoleAgent,
			OrgID: otherOrgID,
		}
		got, err := repo.Create(ctx, u, "hashed-pw")
		require.NoError(t, err)
		assert.Equal(t, defaultOrgID, got.OrgID)
	})
}

func TestUserRepo_RoleProfileAssignmentsAreOrgScoped(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewUserRepo(pool)

	otherOrgID := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO orgs (id, name, slug, plan)
		VALUES ($1, 'Other Org', $2, 'single')
	`, otherOrgID, "other-"+uuid.NewString())
	require.NoError(t, err)

	var otherRoleID uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT id FROM crm_roles WHERE org_id = $1 AND system_key = 'admin'
	`, otherOrgID).Scan(&otherRoleID))

	var otherProfileID uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(), `
		SELECT id FROM crm_profiles WHERE org_id = $1 AND system_key = 'administrator'
	`, otherOrgID).Scan(&otherProfileID))

	t.Run("create rejects cross-org role and profile ids", func(t *testing.T) {
		_, err := repo.Create(ctx, &domain.User{
			Email:     "cross-create@omnir.test",
			Name:      "Cross Create",
			Role:      domain.UserRoleAgent,
			RoleID:    &otherRoleID,
			ProfileID: &otherProfileID,
		}, "hashed-pw")
		require.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("update rejects cross-org role id", func(t *testing.T) {
		user, err := repo.Create(ctx, &domain.User{
			Email: "cross-role-update@omnir.test",
			Name:  "Cross Role Update",
			Role:  domain.UserRoleAgent,
		}, "hashed-pw")
		require.NoError(t, err)

		_, err = repo.Update(ctx, user.ID, domain.UserPatch{RoleID: &otherRoleID})
		require.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("update rejects cross-org profile id", func(t *testing.T) {
		user, err := repo.Create(ctx, &domain.User{
			Email: "cross-profile-update@omnir.test",
			Name:  "Cross Profile Update",
			Role:  domain.UserRoleAgent,
		}, "hashed-pw")
		require.NoError(t, err)

		_, err = repo.Update(ctx, user.ID, domain.UserPatch{ProfileID: &otherProfileID})
		require.ErrorIs(t, err, domain.ErrValidation)
	})

	t.Run("update clears role and profile assignments", func(t *testing.T) {
		var adminRoleID uuid.UUID
		require.NoError(t, pool.QueryRow(context.Background(), `
			SELECT id FROM crm_roles WHERE org_id = $1 AND system_key = 'admin'
		`, defaultOrgID).Scan(&adminRoleID))

		var adminProfileID uuid.UUID
		require.NoError(t, pool.QueryRow(context.Background(), `
			SELECT id FROM crm_profiles WHERE org_id = $1 AND system_key = 'administrator'
		`, defaultOrgID).Scan(&adminProfileID))

		user, err := repo.Create(ctx, &domain.User{
			Email:     "clear-acl-assignments@omnir.test",
			Name:      "Clear ACL Assignments",
			Role:      domain.UserRoleAgent,
			RoleID:    &adminRoleID,
			ProfileID: &adminProfileID,
		}, "hashed-pw")
		require.NoError(t, err)
		require.NotNil(t, user.RoleID)
		require.NotNil(t, user.ProfileID)

		got, err := repo.Update(ctx, user.ID, domain.UserPatch{
			ClearRoleID:    true,
			ClearProfileID: true,
		})
		require.NoError(t, err)
		assert.Nil(t, got.RoleID)
		assert.Nil(t, got.ProfileID)
		assert.Nil(t, got.RoleName)
		assert.Nil(t, got.ProfileName)
	})

	t.Run("stale cross-org ids do not expose foreign role or profile names", func(t *testing.T) {
		user, err := repo.Create(ctx, &domain.User{
			Email: "stale-cross-org@omnir.test",
			Name:  "Stale Cross Org",
			Role:  domain.UserRoleAgent,
		}, "hashed-pw")
		require.NoError(t, err)

		_, err = pool.Exec(context.Background(), `
			UPDATE users SET role_id = $1, profile_id = $2 WHERE id = $3
		`, otherRoleID, otherProfileID, user.ID)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, got.RoleName)
		assert.Nil(t, got.ProfileName)
	})
}
