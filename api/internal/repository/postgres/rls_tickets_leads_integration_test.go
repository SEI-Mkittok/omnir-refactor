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

// TestRLS_TicketTenantIsolation verifies that tickets created for org A are
// invisible when queried under org B's session variable.
func TestRLS_TicketTenantIsolation(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "ticket-org-a"},
		{orgB, "ticket-org-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, "Org "+org.slug, org.slug)
		require.NoError(t, err)
	}

	// Seed a user in org A for assignee reference.
	ownerA := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Ticket Owner A', 'admin')
	`, ownerA, orgA, "ticketowner+"+ownerA.String()+"@omnir.test")
	require.NoError(t, err)

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	repo := postgres.NewTicketRepo(pool)

	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	ticket, err := repo.Create(ctxA, &domain.Ticket{
		OrgID:   orgA,
		Subject: "RLS test ticket",
		Status:  domain.TicketStatusOpen,
	})
	require.NoError(t, err)
	require.NotNil(t, ticket)
	assert.Equal(t, orgA, ticket.OrgID)

	t.Run("org A can read its own ticket", func(t *testing.T) {
		got, err := repo.GetByID(ctxA, ticket.ID)
		require.NoError(t, err)
		assert.Equal(t, ticket.ID, got.ID)
	})

	t.Run("org B cannot read org A ticket", func(t *testing.T) {
		_, err := repo.GetByID(ctxB, ticket.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound,
			"RLS should make org A ticket invisible to org B")
	})

	t.Run("org B List returns empty", func(t *testing.T) {
		tickets, total, err := repo.List(ctxB, domain.TicketFilter{OrgID: orgB})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, tickets)
	})

	t.Run("org A List returns its ticket", func(t *testing.T) {
		tickets, total, err := repo.List(ctxA, domain.TicketFilter{OrgID: orgA})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, tickets, 1)
	})
}

func TestTicketDetailReturnsTicketAccountWhenDifferentFromContactAccount(t *testing.T) {
	pool, _ := setupDB(t)

	repo := postgres.NewTicketRepo(pool)
	accountRepo := postgres.NewAccountRepo(pool)
	contactRepo := postgres.NewContactRepo(pool)
	ctx := domain.WithOrgID(context.Background(), defaultOrgID)

	ownerID := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, $4, 'admin')
	`, ownerID, defaultOrgID, "account-owner+"+ownerID.String()+"@omnir.test", "Account Owner")
	require.NoError(t, err)

	contactAccount, err := accountRepo.Create(ctx, &domain.Account{
		OrgID:   defaultOrgID,
		Name:    "Contact Default Account",
		OwnerID: ownerID,
	})
	require.NoError(t, err)

	ticketAccount, err := accountRepo.Create(ctx, &domain.Account{
		OrgID:   defaultOrgID,
		Name:    "Escalation Account",
		OwnerID: ownerID,
	})
	require.NoError(t, err)

	contact, err := contactRepo.Create(ctx, &domain.Contact{
		OrgID:     defaultOrgID,
		FirstName: "Taylor",
		LastName:  "Contact",
		AccountID: &contactAccount.ID,
	})
	require.NoError(t, err)

	ticket, err := repo.Create(ctx, &domain.Ticket{
		OrgID:     defaultOrgID,
		Subject:   "Route ticket to escalation account",
		ContactID: &contact.ID,
		AccountID: &ticketAccount.ID,
	})
	require.NoError(t, err)

	detail, err := repo.GetDetailByID(ctx, ticket.ID)
	require.NoError(t, err)

	require.NotNil(t, detail.Contact)
	assert.Equal(t, contact.ID, detail.Contact.ID)

	require.NotNil(t, detail.Account)
	assert.Equal(t, ticketAccount.ID, detail.Account.ID)
	assert.Equal(t, ticketAccount.Name, detail.Account.Name)
	assert.NotEqual(t, contactAccount.ID, detail.Account.ID, "ticket detail account must come from tickets.account_id")
}

// TestRLS_TicketCommentTenantIsolation verifies that ticket comments created
// for org A are invisible when listed under org B's session variable.
func TestRLS_TicketCommentTenantIsolation(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "comment-org-a"},
		{orgB, "comment-org-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, "Org "+org.slug, org.slug)
		require.NoError(t, err)
	}

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	ticketRepo := postgres.NewTicketRepo(pool)
	commentRepo := postgres.NewTicketCommentRepo(pool)

	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	ticket, err := ticketRepo.Create(ctxA, &domain.Ticket{
		OrgID:   orgA,
		Subject: "Comment RLS test",
		Status:  domain.TicketStatusOpen,
	})
	require.NoError(t, err)

	comment, err := commentRepo.Create(ctxA, &domain.TicketComment{
		TicketID: ticket.ID,
		OrgID:    orgA,
		Body:     "private note",
	})
	require.NoError(t, err)
	require.NotNil(t, comment)

	t.Run("org A can list its comment", func(t *testing.T) {
		comments, err := commentRepo.List(ctxA, domain.TicketCommentFilter{
			TicketID: ticket.ID,
			OrgID:    orgA,
		})
		require.NoError(t, err)
		assert.Len(t, comments, 1)
	})

	t.Run("org B sees no comments for org A ticket", func(t *testing.T) {
		comments, err := commentRepo.List(ctxB, domain.TicketCommentFilter{
			TicketID: ticket.ID,
			OrgID:    orgB,
		})
		require.NoError(t, err)
		assert.Empty(t, comments, "RLS should hide org A comments from org B")
	})
}

// TestRLS_LeadTenantIsolation verifies that leads created for org A are
// invisible when queried under org B's session variable.
func TestRLS_LeadTenantIsolation(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "lead-org-a"},
		{orgB, "lead-org-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, "Org "+org.slug, org.slug)
		require.NoError(t, err)
	}

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	repo := postgres.NewLeadRepo(pool)

	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	lead, err := repo.Create(ctxA, &domain.Lead{
		OrgID:     orgA,
		FirstName: "RLS",
		LastName:  "TestLead",
		Status:    domain.LeadStatusNew,
	})
	require.NoError(t, err)
	require.NotNil(t, lead)
	assert.Equal(t, orgA, lead.OrgID)

	t.Run("org A can read its own lead", func(t *testing.T) {
		got, err := repo.GetByID(ctxA, lead.ID)
		require.NoError(t, err)
		assert.Equal(t, lead.ID, got.ID)
	})

	t.Run("org B cannot read org A lead", func(t *testing.T) {
		_, err := repo.GetByID(ctxB, lead.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound,
			"RLS should make org A lead invisible to org B")
	})

	t.Run("org B List returns empty", func(t *testing.T) {
		leads, total, err := repo.List(ctxB, domain.LeadFilter{OrgID: orgB})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, leads)
	})

	t.Run("org A List returns its lead", func(t *testing.T) {
		leads, total, err := repo.List(ctxA, domain.LeadFilter{OrgID: orgA})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, leads, 1)
	})
}

// TestRLS_OrgModeTransactionScope verifies that SET LOCAL for app.current_org_id
// is scoped to the transaction and does not leak across connections.
func TestRLS_OrgModeTransactionScope(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO orgs (id, name, slug, plan)
		VALUES ($1, 'Scope Org A', 'scope-org-a', 'starter')
	`, orgA)
	require.NoError(t, err)

	conn, err := pool.Acquire(context.Background())
	require.NoError(t, err)
	defer conn.Release()

	// Set the session variable using SET LOCAL inside a transaction.
	tx, err := conn.Begin(context.Background())
	require.NoError(t, err)

	_, err = tx.Exec(context.Background(),
		`SELECT set_config('app.current_org_id', $1, true)`, orgA.String())
	require.NoError(t, err)

	// Inside the transaction the variable is set.
	var inside string
	err = tx.QueryRow(context.Background(),
		`SELECT current_setting('app.current_org_id', true)`).Scan(&inside)
	require.NoError(t, err)
	assert.Equal(t, orgA.String(), inside)

	require.NoError(t, tx.Rollback(context.Background()))

	// After rollback (SET LOCAL reverts) the variable is cleared.
	var outside string
	err = conn.QueryRow(context.Background(),
		`SELECT current_setting('app.current_org_id', true)`).Scan(&outside)
	require.NoError(t, err)
	assert.Empty(t, outside, "SET LOCAL should not persist after transaction ends")
}
