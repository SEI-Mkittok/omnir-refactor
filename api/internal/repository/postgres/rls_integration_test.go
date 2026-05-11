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

// TestRLS_TenantIsolation proves that with RLS FORCE enabled, a contact
// created for org A is invisible when the session variable is set to org B.
func TestRLS_TenantIsolation(t *testing.T) {
	pool, _ := setupDB(t)

	// Seed two separate orgs.
	orgA := uuid.New()
	orgB := uuid.New()

	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "org-a"},
		{orgB, "org-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, org.slug, org.slug)
		require.NoError(t, err)
	}

	// Seed a user in org A.
	ownerID := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Org A User', 'admin')
	`, ownerID, orgA, "orga+"+ownerID.String()+"@omnir.test")
	require.NoError(t, err)

	// Enable FORCE RLS so policies are evaluated even for the app DB role.
	err = postgres.EnableRLS(context.Background(), pool)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = postgres.DisableRLS(context.Background(), pool)
	})

	repo := postgres.NewContactRepo(pool)

	// Create a contact scoped to org A.
	ctxA := domain.WithOrgID(context.Background(), orgA)
	contact, err := repo.Create(ctxA, &domain.Contact{
		FirstName: "Alice",
		LastName:  "OrgA",
		OwnerID:   ownerID,
		Stage:     domain.ContactStageLead,
	})
	require.NoError(t, err)
	require.NotNil(t, contact)
	assert.Equal(t, orgA, contact.OrgID)

	t.Run("org A can see its own contact", func(t *testing.T) {
		got, err := repo.GetByID(ctxA, contact.ID)
		require.NoError(t, err)
		assert.Equal(t, contact.ID, got.ID)
	})

	t.Run("org B cannot see org A contact", func(t *testing.T) {
		ctxB := domain.WithOrgID(context.Background(), orgB)
		_, err := repo.GetByID(ctxB, contact.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound,
			"RLS should make the contact invisible to org B")
	})

	t.Run("org B List returns empty", func(t *testing.T) {
		ctxB := domain.WithOrgID(context.Background(), orgB)
		contacts, total, err := repo.List(ctxB, domain.ContactFilter{})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, contacts)
	})

	t.Run("org A List returns its contact", func(t *testing.T) {
		contacts, total, err := repo.List(ctxA, domain.ContactFilter{})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, contacts, 1)
	})
}

func TestRLS_Bundle4OrgSeedTriggerUsesNewOrgContext(t *testing.T) {
	pool, _ := setupDB(t)
	err := postgres.EnableRLS(context.Background(), pool)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = postgres.DisableRLS(context.Background(), pool)
	})

	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	_, err = conn.Exec(ctx, `SELECT set_config('app.current_org_id', $1, false)`, defaultOrgID.String())
	require.NoError(t, err)

	orgID := uuid.New()
	_, err = conn.Exec(ctx, `
		INSERT INTO orgs (id, name, slug, plan)
		VALUES ($1, 'Seeded Org', $2, 'starter')
	`, orgID, "seeded-"+orgID.String())
	require.NoError(t, err)

	var currentOrg string
	err = conn.QueryRow(ctx, `SELECT current_setting('app.current_org_id', true)`).Scan(&currentOrg)
	require.NoError(t, err)
	assert.Equal(t, defaultOrgID.String(), currentOrg)

	_, err = conn.Exec(ctx, `SELECT set_config('app.current_org_id', $1, false)`, orgID.String())
	require.NoError(t, err)

	var roleCount, profileCount, sharingDefaultCount, sharingGrantCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM crm_roles WHERE org_id = $1`, orgID).Scan(&roleCount)
	require.NoError(t, err)
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM crm_profiles WHERE org_id = $1`, orgID).Scan(&profileCount)
	require.NoError(t, err)
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM crm_sharing_defaults WHERE org_id = $1`, orgID).Scan(&sharingDefaultCount)
	require.NoError(t, err)
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM crm_sharing_grants WHERE org_id = $1`, orgID).Scan(&sharingGrantCount)
	require.NoError(t, err)

	assert.Equal(t, 4, roleCount)
	assert.Equal(t, 6, profileCount)
	assert.Equal(t, 5, sharingDefaultCount)
	assert.Equal(t, 10, sharingGrantCount)
}

// TestRLS_DisabledInSingleMode verifies that when RLS is not forced, contacts
// are accessible without an org_id in the session (superuser / single-tenant).
func TestRLS_DisabledInSingleMode(t *testing.T) {
	pool, ctxDefault := setupDB(t)

	// Ensure RLS is not forced.
	err := postgres.DisableRLS(context.Background(), pool)
	require.NoError(t, err)

	ownerID := seedUser(t, pool)
	repo := postgres.NewContactRepo(pool)

	contact, err := repo.Create(ctxDefault, &domain.Contact{
		FirstName: "Bob",
		LastName:  "Single",
		OwnerID:   ownerID,
		Stage:     domain.ContactStageLead,
	})
	require.NoError(t, err)

	// Without RLS forced, a context with no org_id can still read the row
	// (application-level org_id filter will be skipped when org not in context).
	got, err := repo.GetByID(ctxDefault, contact.ID)
	require.NoError(t, err)
	assert.Equal(t, contact.ID, got.ID)
}
