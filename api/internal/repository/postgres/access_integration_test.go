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

func TestAccessRepoResolveAccess_DoesNotInheritAncestorSharingGrants(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewAccessRepo(pool)

	var adminRoleID, agentRoleID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT admin.id, agent.id
		FROM crm_roles admin
		JOIN crm_roles agent ON agent.org_id = admin.org_id
		WHERE admin.org_id = $1
		  AND admin.system_key = 'admin'
		  AND agent.system_key = 'agent'
	`, defaultOrgID).Scan(&adminRoleID, &agentRoleID)
	require.NoError(t, err)

	agentID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Standard Agent', 'agent')
	`, agentID, defaultOrgID, "agent-"+agentID.String()+"@omnir.test")
	require.NoError(t, err)

	agentAccess, err := repo.ResolveAccess(ctx, agentID, defaultOrgID, string(domain.UserRoleAgent))
	require.NoError(t, err)
	require.NotNil(t, agentAccess.RoleID)
	assert.Equal(t, agentRoleID, *agentAccess.RoleID)
	assert.Contains(t, agentAccess.RoleLineageIDs, adminRoleID)
	assert.False(t, agentAccess.CanAccessAllRecords(domain.ACLModuleAccounts, domain.SharingAccessRead))
	assert.False(t, agentAccess.CanAccessAllRecords(domain.ACLModuleAccounts, domain.SharingAccessWrite))

	adminID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Workspace Admin', 'admin')
	`, adminID, defaultOrgID, "admin-"+adminID.String()+"@omnir.test")
	require.NoError(t, err)

	adminAccess, err := repo.ResolveAccess(ctx, adminID, defaultOrgID, string(domain.UserRoleAdmin))
	require.NoError(t, err)
	assert.True(t, adminAccess.CanAccessAllRecords(domain.ACLModuleAccounts, domain.SharingAccessRead))
	assert.True(t, adminAccess.CanAccessAllRecords(domain.ACLModuleAccounts, domain.SharingAccessWrite))
}

func TestAccessRepo_ACLObjectMutationsAreScopedToContextOrg(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewAccessRepo(pool)

	otherOrgID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO orgs (id, name, slug, plan)
		VALUES ($1, 'Other Org', $2, 'single')
	`, otherOrgID, "other-"+otherOrgID.String())
	require.NoError(t, err)

	otherRoleID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO crm_roles (id, org_id, name)
		VALUES ($1, $2, 'Other Role')
	`, otherRoleID, otherOrgID)
	require.NoError(t, err)

	otherProfileID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO crm_profiles (id, org_id, name)
		VALUES ($1, $2, 'Other Profile')
	`, otherProfileID, otherOrgID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO crm_profile_permissions (org_id, profile_id, module, action, allowed)
		VALUES ($1, $2, 'accounts', 'read', TRUE)
	`, otherOrgID, otherProfileID)
	require.NoError(t, err)

	otherUserID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Other User', 'agent')
	`, otherUserID, otherOrgID, "other-user-"+otherUserID.String()+"@omnir.test")
	require.NoError(t, err)
	otherGroupID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO crm_groups (id, org_id, name)
		VALUES ($1, $2, 'Other Group')
	`, otherGroupID, otherOrgID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO crm_group_members (org_id, group_id, user_id)
		VALUES ($1, $2, $3)
	`, otherOrgID, otherGroupID, otherUserID)
	require.NoError(t, err)

	newName := "Cross Tenant Write"
	_, err = repo.UpdateRole(ctx, otherRoleID, domain.ACLRolePatch{Name: &newName})
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = repo.MoveRole(ctx, otherRoleID, nil)
	require.ErrorIs(t, err, domain.ErrNotFound)
	err = repo.DeleteRole(ctx, otherRoleID)
	require.ErrorIs(t, err, domain.ErrNotFound)

	_, err = repo.UpdateProfile(ctx, otherProfileID, domain.ACLProfilePatch{Name: &newName})
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, _, err = repo.ListProfilePermissions(ctx, otherProfileID)
	require.ErrorIs(t, err, domain.ErrNotFound)
	err = repo.ReplaceProfilePermissions(ctx, otherProfileID, nil, nil)
	require.ErrorIs(t, err, domain.ErrNotFound)
	err = repo.DeleteProfile(ctx, otherProfileID)
	require.ErrorIs(t, err, domain.ErrNotFound)

	_, err = repo.UpdateGroup(ctx, otherGroupID, domain.ACLGroupPatch{Name: &newName})
	require.ErrorIs(t, err, domain.ErrNotFound)
	err = repo.ReplaceGroupMembers(ctx, otherGroupID, nil)
	require.ErrorIs(t, err, domain.ErrNotFound)
	err = repo.DeleteGroup(ctx, otherGroupID)
	require.ErrorIs(t, err, domain.ErrNotFound)

	var roleName, profileName, groupName string
	err = pool.QueryRow(ctx, `SELECT name FROM crm_roles WHERE id = $1 AND org_id = $2`, otherRoleID, otherOrgID).Scan(&roleName)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT name FROM crm_profiles WHERE id = $1 AND org_id = $2`, otherProfileID, otherOrgID).Scan(&profileName)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT name FROM crm_groups WHERE id = $1 AND org_id = $2`, otherGroupID, otherOrgID).Scan(&groupName)
	require.NoError(t, err)
	assert.Equal(t, "Other Role", roleName)
	assert.Equal(t, "Other Profile", profileName)
	assert.Equal(t, "Other Group", groupName)

	var permissionCount, memberCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM crm_profile_permissions WHERE org_id = $1 AND profile_id = $2
	`, otherOrgID, otherProfileID).Scan(&permissionCount)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM crm_group_members WHERE org_id = $1 AND group_id = $2
	`, otherOrgID, otherGroupID).Scan(&memberCount)
	require.NoError(t, err)
	assert.Equal(t, 1, permissionCount)
	assert.Equal(t, 1, memberCount)

	defaultGroup := &domain.ACLGroup{Name: "Default Org Group"}
	createdGroup, err := repo.CreateGroup(ctx, defaultGroup)
	require.NoError(t, err)
	err = repo.ReplaceGroupMembers(ctx, createdGroup.ID, []uuid.UUID{otherUserID})
	require.ErrorIs(t, err, domain.ErrNotFound)
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM crm_group_members WHERE org_id = $1 AND group_id = $2
	`, defaultOrgID, createdGroup.ID).Scan(&memberCount)
	require.NoError(t, err)
	assert.Equal(t, 0, memberCount)
}
