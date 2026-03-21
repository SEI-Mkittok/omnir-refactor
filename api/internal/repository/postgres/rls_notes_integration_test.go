//go:build integration

package postgres_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

// TestRLS_NoteTenantIsolation verifies that notes created for org A are
// invisible when queried under org B's session variable.
func TestRLS_NoteTenantIsolation(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "note-org-a"},
		{orgB, "note-org-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, "Org "+org.slug, org.slug)
		require.NoError(t, err)
	}

	// Seed a user in org A for the author field.
	authorA := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Note Author A', 'admin')
	`, authorA, orgA, "noteauthor+"+authorA.String()+"@omnir.test")
	require.NoError(t, err)

	// Seed a contact in org A to attach the note to.
	contactA := uuid.New()
	_, err = pool.Exec(context.Background(), `
		INSERT INTO contacts (id, org_id, first_name, last_name, stage, owner_id)
		VALUES ($1, $2, 'NoteRLS', 'OrgA', 'lead', $3)
	`, contactA, orgA, authorA)
	require.NoError(t, err)

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	repo := postgres.NewNoteRepo(pool)
	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	note, err := repo.Create(ctxA, &domain.Note{
		OrgID:      orgA,
		Content:    "private note for org A",
		EntityType: domain.NoteEntityContact,
		EntityID:   contactA,
		AuthorID:   authorA,
	})
	require.NoError(t, err)
	require.NotNil(t, note)
	assert.Equal(t, orgA, note.OrgID)

	t.Run("org A can read its own note", func(t *testing.T) {
		got, err := repo.GetByID(ctxA, note.ID)
		require.NoError(t, err)
		assert.Equal(t, note.ID, got.ID)
	})

	t.Run("org B cannot read org A note", func(t *testing.T) {
		_, err := repo.GetByID(ctxB, note.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound,
			"RLS should make org A note invisible to org B")
	})

	t.Run("org B ListByEntity returns empty for org A contact", func(t *testing.T) {
		notes, total, err := repo.ListByEntity(ctxB, domain.NoteFilter{
			EntityType: domain.NoteEntityContact,
			EntityID:   contactA,
			OrgID:      orgB,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, notes)
	})

	t.Run("org A ListByEntity returns its note", func(t *testing.T) {
		notes, total, err := repo.ListByEntity(ctxA, domain.NoteFilter{
			EntityType: domain.NoteEntityContact,
			EntityID:   contactA,
			OrgID:      orgA,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, notes, 1)
	})
}

// TestRLS_NullOrgContextBlocksAll verifies that FORCE ROW LEVEL SECURITY blocks
// all rows when no org context is set. This test requires a non-superuser DB role
// (omnir_app_test) because PostgreSQL superusers bypass FORCE RLS unconditionally.
// The CI workflow creates this role; local runs skip if the role is unavailable.
func TestRLS_NullOrgContextBlocksAll(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO orgs (id, name, slug, plan)
		VALUES ($1, 'Null Ctx Org', 'null-ctx-org', 'starter')
	`, orgA)
	require.NoError(t, err)

	authorA := uuid.New()
	_, err = pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Author A', 'admin')
	`, authorA, orgA, "nullctx+"+authorA.String()+"@omnir.test")
	require.NoError(t, err)

	contactA := uuid.New()
	_, err = pool.Exec(context.Background(), `
		INSERT INTO contacts (id, org_id, first_name, last_name, stage, owner_id)
		VALUES ($1, $2, 'NullCtx', 'OrgA', 'lead', $3)
	`, contactA, orgA, authorA)
	require.NoError(t, err)

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	// Connect as the non-superuser app role. FORCE RLS applies only to
	// non-superusers — the omnir_app_test role is created by the CI setup step.
	// Skip the test if the role is not available (local dev without the role).
	testDSN := os.Getenv("TEST_DATABASE_URL")
	if testDSN == "" {
		testDSN = "postgres://localhost/omnir_crm_test?sslmode=disable"
	}
	appDSN := strings.Replace(testDSN, "omnir:omnir_dev@", "omnir_app_test:test@", 1)
	appPool, connErr := pgxpool.New(context.Background(), appDSN)
	if connErr != nil {
		t.Skipf("skipping: omnir_app_test role unavailable (%v)", connErr)
	}
	defer appPool.Close()

	appConn, err := appPool.Acquire(context.Background())
	require.NoError(t, err)
	defer appConn.Release()

	// Explicitly clear any existing org context on this connection.
	_, err = appConn.Exec(context.Background(),
		`SELECT set_config('app.current_org_id', '', false)`)
	require.NoError(t, err)

	// With current_org_id = '' → current_org_id() returns NULL →
	// policy org_id = NULL evaluates to NULL (never true) → no rows visible.
	var count int
	err = appConn.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM contacts WHERE id=$1`, contactA).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count,
		"FORCE RLS with null org context must block all rows")
}

// TestRLS_SuperAdminTenantSwitch verifies that a super_admin can switch org
// context and access exactly the switched-to org's data via RLS, with no
// cross-org leakage in either direction.
func TestRLS_SuperAdminTenantSwitch(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "sa-switch-org-a"},
		{orgB, "sa-switch-org-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, "Org "+org.slug, org.slug)
		require.NoError(t, err)
	}

	ownerA, ownerB := uuid.New(), uuid.New()
	for _, u := range []struct {
		id    uuid.UUID
		orgID uuid.UUID
		name  string
	}{
		{ownerA, orgA, "sa-owner-a"},
		{ownerB, orgB, "sa-owner-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO users (id, org_id, email, name, role)
			VALUES ($1, $2, $3, $4, 'admin')
		`, u.id, u.orgID, u.name+"+"+u.id.String()+"@omnir.test", u.name)
		require.NoError(t, err)
	}

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	repo := postgres.NewContactRepo(pool)

	// Seed contacts for each org.
	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	contactA, err := repo.Create(ctxA, &domain.Contact{
		FirstName: "Alice",
		LastName:  "OrgA",
		OwnerID:   ownerA,
		Stage:     domain.ContactStageLead,
	})
	require.NoError(t, err)

	contactB, err := repo.Create(ctxB, &domain.Contact{
		FirstName: "Bob",
		LastName:  "OrgB",
		OwnerID:   ownerB,
		Stage:     domain.ContactStageLead,
	})
	require.NoError(t, err)

	// Super-admin "before switch" — operating in org A context.
	t.Run("super_admin in org A context sees only org A data", func(t *testing.T) {
		got, err := repo.GetByID(ctxA, contactA.ID)
		require.NoError(t, err)
		assert.Equal(t, contactA.ID, got.ID)

		_, err = repo.GetByID(ctxA, contactB.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound,
			"org A context must not see org B contact")
	})

	// Super-admin "after switch" — tenant switcher issues JWT with org B's ID,
	// which becomes the new context. Verify they now see org B and not org A.
	t.Run("super_admin after switch to org B sees only org B data", func(t *testing.T) {
		got, err := repo.GetByID(ctxB, contactB.ID)
		require.NoError(t, err)
		assert.Equal(t, contactB.ID, got.ID)

		_, err = repo.GetByID(ctxB, contactA.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound,
			"org B context must not see org A contact")
	})
}
