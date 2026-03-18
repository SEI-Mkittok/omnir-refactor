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

// TestPortal_ClientTicketIsolation verifies that client users can only
// retrieve tickets they submitted (via submitted_by_user_id).
func TestPortal_ClientTicketIsolation(t *testing.T) {
	pool, ctx := setupDB(t)

	clientA := uuid.New()
	clientB := uuid.New()
	for _, u := range []struct {
		id    uuid.UUID
		email string
	}{
		{clientA, "clienta+" + clientA.String() + "@omnir.test"},
		{clientB, "clientb+" + clientB.String() + "@omnir.test"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO users (id, org_id, email, name, role)
			VALUES ($1, $2, $3, 'Portal Client', 'client')
		`, u.id, defaultOrgID, u.email)
		require.NoError(t, err)
	}

	repo := postgres.NewTicketRepo(pool)

	// Client A submits a ticket.
	ticketA, err := repo.Create(ctx, &domain.Ticket{
		Subject:           "Client A ticket",
		SubmittedByUserID: &clientA,
	})
	require.NoError(t, err)
	require.NotNil(t, ticketA)
	assert.Equal(t, &clientA, ticketA.SubmittedByUserID)

	// Client B submits a ticket.
	ticketB, err := repo.Create(ctx, &domain.Ticket{
		Subject:           "Client B ticket",
		SubmittedByUserID: &clientB,
	})
	require.NoError(t, err)
	require.NotNil(t, ticketB)

	t.Run("GetByID returns client A ticket", func(t *testing.T) {
		got, err := repo.GetByID(ctx, ticketA.ID)
		require.NoError(t, err)
		assert.Equal(t, ticketA.ID, got.ID)
		assert.Equal(t, &clientA, got.SubmittedByUserID)
	})

	t.Run("List filtered by clientA returns only client A tickets", func(t *testing.T) {
		tickets, total, err := repo.List(ctx, domain.TicketFilter{
			SubmittedByUserID: &clientA,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		require.Len(t, tickets, 1)
		assert.Equal(t, ticketA.ID, tickets[0].ID)
	})

	t.Run("List filtered by clientB returns only client B tickets", func(t *testing.T) {
		tickets, total, err := repo.List(ctx, domain.TicketFilter{
			SubmittedByUserID: &clientB,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		require.Len(t, tickets, 1)
		assert.Equal(t, ticketB.ID, tickets[0].ID)
	})

	t.Run("submitted_by_user_id persists through scan", func(t *testing.T) {
		tickets, _, err := repo.List(ctx, domain.TicketFilter{
			SubmittedByUserID: &clientA,
		})
		require.NoError(t, err)
		require.Len(t, tickets, 1)
		assert.NotNil(t, tickets[0].SubmittedByUserID)
		assert.Equal(t, clientA, *tickets[0].SubmittedByUserID)
	})
}

// TestPortal_ClientTicketComment verifies that comments can be created and
// listed on client-submitted tickets.
func TestPortal_ClientTicketComment(t *testing.T) {
	pool, ctx := setupDB(t)

	clientID := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Portal Commenter', 'client')
	`, clientID, defaultOrgID, "comment+"+clientID.String()+"@omnir.test")
	require.NoError(t, err)

	ticketRepo := postgres.NewTicketRepo(pool)
	commentRepo := postgres.NewTicketCommentRepo(pool)

	ticket, err := ticketRepo.Create(ctx, &domain.Ticket{
		Subject:           "Portal comment test",
		SubmittedByUserID: &clientID,
	})
	require.NoError(t, err)

	comment, err := commentRepo.Create(ctx, &domain.TicketComment{
		TicketID:   ticket.ID,
		AuthorID:   &clientID,
		Body:       "this is my comment",
		IsInternal: false,
	})
	require.NoError(t, err)
	require.NotNil(t, comment)
	assert.Equal(t, "this is my comment", comment.Body)
	assert.False(t, comment.IsInternal)
	assert.Equal(t, &clientID, comment.AuthorID)

	t.Run("comment is listed on ticket", func(t *testing.T) {
		f := false
		comments, err := commentRepo.List(ctx, domain.TicketCommentFilter{
			TicketID:   ticket.ID,
			IsInternal: &f,
		})
		require.NoError(t, err)
		require.Len(t, comments, 1)
		assert.Equal(t, comment.ID, comments[0].ID)
	})
}

// TestPortal_RLSClientTicketIsolation verifies that with RLS enabled,
// client tickets from org A are invisible to org B queries.
func TestPortal_RLSClientTicketIsolation(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "portal-org-a-" + orgA.String()[:8]},
		{orgB, "portal-org-b-" + orgB.String()[:8]},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, "Portal "+org.slug, org.slug)
		require.NoError(t, err)
	}

	clientA := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Portal RLS Client', 'client')
	`, clientA, orgA, "portalrls+"+clientA.String()+"@omnir.test")
	require.NoError(t, err)

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	repo := postgres.NewTicketRepo(pool)
	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	ticket, err := repo.Create(ctxA, &domain.Ticket{
		OrgID:             orgA,
		Subject:           "Portal RLS test",
		SubmittedByUserID: &clientA,
	})
	require.NoError(t, err)
	require.NotNil(t, ticket)

	t.Run("org A client can see own ticket", func(t *testing.T) {
		got, err := repo.GetByID(ctxA, ticket.ID)
		require.NoError(t, err)
		assert.Equal(t, ticket.ID, got.ID)
	})

	t.Run("org B cannot see org A client ticket via RLS", func(t *testing.T) {
		_, err := repo.GetByID(ctxB, ticket.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("org B list returns no tickets for clientA filter", func(t *testing.T) {
		tickets, total, err := repo.List(ctxB, domain.TicketFilter{
			OrgID:             orgB,
			SubmittedByUserID: &clientA,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, tickets)
	})
}
