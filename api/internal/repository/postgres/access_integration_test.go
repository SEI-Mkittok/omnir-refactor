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

func TestAccessRepoCanAccessRecordUsesSharingVisibility(t *testing.T) {
	pool, ctx := setupDB(t)
	accessRepo := postgres.NewAccessRepo(pool)
	contactRepo := postgres.NewContactRepo(pool)

	ownerID := uuid.New()
	otherID := uuid.New()
	adminID := uuid.New()
	for _, user := range []struct {
		id   uuid.UUID
		role string
		name string
	}{
		{ownerID, "agent", "Owner Agent"},
		{otherID, "agent", "Other Agent"},
		{adminID, "admin", "Admin User"},
	} {
		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, org_id, email, name, role)
			VALUES ($1, $2, $3, $4, $5)
		`, user.id, defaultOrgID, user.id.String()+"@omnir.test", user.name, user.role)
		require.NoError(t, err)
	}

	owned, err := contactRepo.Create(ctx, &domain.Contact{
		FirstName: "Owned",
		LastName:  "Contact",
		OwnerID:   ownerID,
		Stage:     domain.ContactStageLead,
	})
	require.NoError(t, err)
	other, err := contactRepo.Create(ctx, &domain.Contact{
		FirstName: "Private",
		LastName:  "Contact",
		OwnerID:   otherID,
		Stage:     domain.ContactStageLead,
	})
	require.NoError(t, err)

	ownerAccess, err := accessRepo.ResolveAccess(ctx, ownerID, defaultOrgID, string(domain.UserRoleAgent))
	require.NoError(t, err)
	ownerCtx := domain.WithAccessContext(ctx, ownerAccess)
	canAccess, err := accessRepo.CanAccessRecord(ownerCtx, domain.ACLModuleContacts, owned.ID, domain.SharingAccessRead)
	require.NoError(t, err)
	assert.True(t, canAccess)
	canAccess, err = accessRepo.CanAccessRecord(ownerCtx, domain.ACLModuleContacts, other.ID, domain.SharingAccessRead)
	require.NoError(t, err)
	assert.False(t, canAccess)

	adminAccess, err := accessRepo.ResolveAccess(ctx, adminID, defaultOrgID, string(domain.UserRoleAdmin))
	require.NoError(t, err)
	adminCtx := domain.WithAccessContext(ctx, adminAccess)
	canAccess, err = accessRepo.CanAccessRecord(adminCtx, domain.ACLModuleContacts, other.ID, domain.SharingAccessWrite)
	require.NoError(t, err)
	assert.True(t, canAccess)
}

func TestAccessRepoCanAccessRecordUsesAdvancedSharingRules(t *testing.T) {
	pool, ctx := setupDB(t)
	accessRepo := postgres.NewAccessRepo(pool)
	contactRepo := postgres.NewContactRepo(pool)

	sourceParent, err := accessRepo.CreateRole(ctx, &domain.ACLRole{Name: "Source Leadership"})
	require.NoError(t, err)
	sourceChild, err := accessRepo.CreateRole(ctx, &domain.ACLRole{Name: "Source Team", ParentID: &sourceParent.ID})
	require.NoError(t, err)
	targetRole, err := accessRepo.CreateRole(ctx, &domain.ACLRole{Name: "Target Auditor"})
	require.NoError(t, err)
	siblingRole, err := accessRepo.CreateRole(ctx, &domain.ACLRole{Name: "Sibling Auditor"})
	require.NoError(t, err)

	ownerID := uuid.New()
	targetID := uuid.New()
	siblingID := uuid.New()
	for _, user := range []struct {
		id     uuid.UUID
		name   string
		roleID uuid.UUID
	}{
		{ownerID, "Source Owner", sourceChild.ID},
		{targetID, "Target User", targetRole.ID},
		{siblingID, "Sibling User", siblingRole.ID},
	} {
		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, org_id, email, name, role, role_id)
			VALUES ($1, $2, $3, $4, 'agent', $5)
		`, user.id, defaultOrgID, user.id.String()+"@omnir.test", user.name, user.roleID)
		require.NoError(t, err)
	}

	owned, err := contactRepo.Create(ctx, &domain.Contact{
		FirstName: "Shared",
		LastName:  "Contact",
		OwnerID:   ownerID,
		Stage:     domain.ContactStageLead,
	})
	require.NoError(t, err)

	_, err = accessRepo.ReplaceSharingRules(ctx, defaultOrgID, &domain.ACLSharingRules{Rules: []domain.ACLSharingModuleRule{
		{
			Module: domain.ACLModuleContacts,
			Mode:   domain.SharingDefaultPrivate,
			AdvancedRules: []domain.ACLSharingRule{
				{
					SourceType:  domain.SharingPrincipalRoleSubordinates,
					SourceID:    &sourceParent.ID,
					TargetType:  domain.SharingPrincipalUser,
					TargetID:    targetID,
					AccessLevel: domain.SharingAccessRead,
				},
			},
		},
	}})
	require.NoError(t, err)

	targetAccess, err := accessRepo.ResolveAccess(ctx, targetID, defaultOrgID, string(domain.UserRoleAgent))
	require.NoError(t, err)
	targetCtx := domain.WithAccessContext(ctx, targetAccess)
	canAccess, err := accessRepo.CanAccessRecord(targetCtx, domain.ACLModuleContacts, owned.ID, domain.SharingAccessRead)
	require.NoError(t, err)
	assert.True(t, canAccess)
	canAccess, err = accessRepo.CanAccessRecord(targetCtx, domain.ACLModuleContacts, owned.ID, domain.SharingAccessWrite)
	require.NoError(t, err)
	assert.False(t, canAccess)

	siblingAccess, err := accessRepo.ResolveAccess(ctx, siblingID, defaultOrgID, string(domain.UserRoleAgent))
	require.NoError(t, err)
	siblingCtx := domain.WithAccessContext(ctx, siblingAccess)
	canAccess, err = accessRepo.CanAccessRecord(siblingCtx, domain.ACLModuleContacts, owned.ID, domain.SharingAccessRead)
	require.NoError(t, err)
	assert.False(t, canAccess)
}

func TestAccessRepoReplaceSharingRulesRejectsUnknownRuleID(t *testing.T) {
	pool, ctx := setupDB(t)
	accessRepo := postgres.NewAccessRepo(pool)

	targetRole, err := accessRepo.CreateRole(ctx, &domain.ACLRole{Name: "Target Auditor"})
	require.NoError(t, err)

	_, err = accessRepo.ReplaceSharingRules(ctx, defaultOrgID, &domain.ACLSharingRules{Rules: []domain.ACLSharingModuleRule{
		{
			Module: domain.ACLModuleContacts,
			Mode:   domain.SharingDefaultPrivate,
			AdvancedRules: []domain.ACLSharingRule{
				{
					ID:          uuid.New(),
					SourceType:  domain.SharingPrincipalAll,
					TargetType:  domain.SharingPrincipalRole,
					TargetID:    targetRole.ID,
					AccessLevel: domain.SharingAccessRead,
				},
			},
		},
	}})
	require.ErrorIs(t, err, domain.ErrValidation)
}

func TestAccessRepoCanAccessAccountRelationshipRequiresLinkedAccountVisibility(t *testing.T) {
	pool, ctx := setupDB(t)
	accessRepo := postgres.NewAccessRepo(pool)
	accountRepo := postgres.NewAccountRepo(pool)

	ownerID := uuid.New()
	otherID := uuid.New()
	adminID := uuid.New()
	for _, user := range []struct {
		id   uuid.UUID
		role string
		name string
	}{
		{ownerID, "agent", "Owner Agent"},
		{otherID, "agent", "Other Agent"},
		{adminID, "admin", "Admin User"},
	} {
		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, org_id, email, name, role)
			VALUES ($1, $2, $3, $4, $5)
		`, user.id, defaultOrgID, user.id.String()+"@omnir.test", user.name, user.role)
		require.NoError(t, err)
	}

	parent, err := accountRepo.Create(ctx, &domain.Account{
		Name:    "Owned Parent",
		OwnerID: ownerID,
	})
	require.NoError(t, err)
	child, err := accountRepo.Create(ctx, &domain.Account{
		Name:    "Private Child",
		OwnerID: otherID,
	})
	require.NoError(t, err)
	rel, err := accountRepo.CreateRelationship(ctx, &domain.AccountRelationship{
		ParentAccountID:  parent.ID,
		ChildAccountID:   child.ID,
		RelationshipType: domain.AccountRelationshipTypePartner,
	})
	require.NoError(t, err)

	ownerAccess, err := accessRepo.ResolveAccess(ctx, ownerID, defaultOrgID, string(domain.UserRoleAgent))
	require.NoError(t, err)
	ownerCtx := domain.WithAccessContext(ctx, ownerAccess)
	canAccess, err := accessRepo.CanAccessAccountRelationship(ownerCtx, rel.ID, domain.SharingAccessWrite)
	require.NoError(t, err)
	assert.False(t, canAccess)

	adminAccess, err := accessRepo.ResolveAccess(ctx, adminID, defaultOrgID, string(domain.UserRoleAdmin))
	require.NoError(t, err)
	adminCtx := domain.WithAccessContext(ctx, adminAccess)
	canAccess, err = accessRepo.CanAccessAccountRelationship(adminCtx, rel.ID, domain.SharingAccessWrite)
	require.NoError(t, err)
	assert.True(t, canAccess)
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

func TestAccessRepo_MoveRoleKeepsSystemRolesImmutable(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewAccessRepo(pool)

	var agentRoleID, originalParentID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT id, parent_id
		FROM crm_roles
		WHERE org_id = $1 AND system_key = 'agent'
	`, defaultOrgID).Scan(&agentRoleID, &originalParentID)
	require.NoError(t, err)

	customParent, err := repo.CreateRole(ctx, &domain.ACLRole{Name: "Custom Parent"})
	require.NoError(t, err)

	_, err = repo.MoveRole(ctx, agentRoleID, &customParent.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)

	var parentID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT parent_id
		FROM crm_roles
		WHERE id = $1 AND org_id = $2
	`, agentRoleID, defaultOrgID).Scan(&parentID)
	require.NoError(t, err)
	assert.Equal(t, originalParentID, parentID)
}

func TestAccessRepo_CreateGroupWithInvalidMemberRollsBack(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewAccessRepo(pool)

	otherOrgID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO orgs (id, name, slug, plan)
		VALUES ($1, 'Other Org', $2, 'single')
	`, otherOrgID, "other-"+otherOrgID.String())
	require.NoError(t, err)

	otherUserID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Other User', 'agent')
	`, otherUserID, otherOrgID, "other-"+otherUserID.String()+"@omnir.test")
	require.NoError(t, err)

	groupID := uuid.New()
	_, err = repo.CreateGroup(ctx, &domain.ACLGroup{
		ID:      groupID,
		Name:    "Invalid Member Group",
		UserIDs: []uuid.UUID{otherUserID},
	})
	require.ErrorIs(t, err, domain.ErrNotFound)

	var groupCount, memberCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM crm_groups WHERE id = $1`, groupID).Scan(&groupCount)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM crm_group_members WHERE group_id = $1`, groupID).Scan(&memberCount)
	require.NoError(t, err)
	assert.Equal(t, 0, groupCount)
	assert.Equal(t, 0, memberCount)
}
