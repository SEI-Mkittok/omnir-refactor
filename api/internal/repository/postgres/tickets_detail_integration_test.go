//go:build integration

package postgres_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

func TestTicketDetailReturnsTicketAccountWhenDifferentFromContactAccount(t *testing.T) {
	pool, ctx := setupDB(t)

	repo := postgres.NewTicketRepo(pool)
	accountRepo := postgres.NewAccountRepo(pool)
	contactRepo := postgres.NewContactRepo(pool)

	ownerID := seedUser(t, pool)

	contactAccount, err := accountRepo.Create(ctx, &domain.Account{
		Name:    "Contact Default Account",
		OwnerID: ownerID,
	})
	require.NoError(t, err)

	ticketAccount, err := accountRepo.Create(ctx, &domain.Account{
		Name:    "Escalation Account",
		OwnerID: ownerID,
	})
	require.NoError(t, err)

	contact, err := contactRepo.Create(ctx, &domain.Contact{
		FirstName: "Ari",
		LastName:  "Contact",
		OwnerID:   ownerID,
		AccountID: &contactAccount.ID,
	})
	require.NoError(t, err)

	ticket, err := repo.Create(ctx, &domain.Ticket{
		Subject:   "Account mismatch detail test",
		ContactID: &contact.ID,
		AccountID: &ticketAccount.ID,
	})
	require.NoError(t, err)

	detail, err := repo.GetDetailByID(ctx, ticket.ID)
	require.NoError(t, err)
	require.NotNil(t, detail.Account)

	assert.Equal(t, ticketAccount.ID, detail.Account.ID)
	assert.Equal(t, ticketAccount.Name, detail.Account.Name)
}
