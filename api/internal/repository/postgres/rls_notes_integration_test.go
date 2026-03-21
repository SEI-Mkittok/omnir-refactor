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
