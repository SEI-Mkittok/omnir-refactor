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
