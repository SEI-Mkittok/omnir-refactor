//go:build integration

package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

func TestContactRepo_IsRelatedToAccount_PrimaryAndMembership(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	accountRepo := postgres.NewAccountRepo(pool)
	contactRepo := postgres.NewContactRepo(pool)

	account, err := accountRepo.Create(ctx, &domain.Account{Name: "Acme", OwnerID: ownerID})
	require.NoError(t, err)

	primaryContact, err := contactRepo.Create(ctx, &domain.Contact{
		FirstName: "Primary", LastName: "Contact", OwnerID: ownerID, Stage: domain.ContactStageLead, AccountID: &account.ID,
	})
	require.NoError(t, err)
	related, err := contactRepo.IsRelatedToAccount(ctx, primaryContact.ID, account.ID)
	require.NoError(t, err)
	require.True(t, related)

	memberContact, err := contactRepo.Create(ctx, &domain.Contact{
		FirstName: "Member", LastName: "Contact", OwnerID: ownerID, Stage: domain.ContactStageLead,
	})
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO account_contacts (account_id, contact_id) VALUES ($1,$2)`, account.ID, memberContact.ID)
	require.NoError(t, err)
	related, err = contactRepo.IsRelatedToAccount(ctx, memberContact.ID, account.ID)
	require.NoError(t, err)
	require.True(t, related)

	unrelatedContact, err := contactRepo.Create(ctx, &domain.Contact{
		FirstName: "Other", LastName: "Contact", OwnerID: ownerID, Stage: domain.ContactStageLead,
	})
	require.NoError(t, err)
	related, err = contactRepo.IsRelatedToAccount(ctx, unrelatedContact.ID, account.ID)
	require.NoError(t, err)
	require.False(t, related)
}

func TestContactAccountPairingTriggers_BlockInvalidPairs(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	accountRepo := postgres.NewAccountRepo(pool)
	contactRepo := postgres.NewContactRepo(pool)
	dealRepo := postgres.NewDealRepo(pool)

	accountA, err := accountRepo.Create(ctx, &domain.Account{Name: "A", OwnerID: ownerID})
	require.NoError(t, err)
	accountB, err := accountRepo.Create(ctx, &domain.Account{Name: "B", OwnerID: ownerID})
	require.NoError(t, err)
	contactA, err := contactRepo.Create(ctx, &domain.Contact{
		FirstName: "Ada", LastName: "Lovelace", OwnerID: ownerID, Stage: domain.ContactStageLead, AccountID: &accountA.ID,
	})
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO tickets (id, org_id, subject, status, priority, contact_id, account_id)
		VALUES ($1,$2,'Invalid ticket','open','medium',$3,$4)
	`, uuid.New(), defaultOrgID, contactA.ID, accountB.ID)
	require.Error(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO deals (id, org_id, title, stage, owner_id, pipeline_id, contact_id, account_id)
		VALUES ($1,$2,'Invalid deal','lead',$3,$4,$5,$6)
	`, uuid.New(), defaultOrgID, ownerID, defaultPipelineID, contactA.ID, accountB.ID)
	require.Error(t, err)

	validDeal, err := dealRepo.Create(ctx, &domain.Deal{
		Title: "Quote Deal", Stage: domain.DealStageLead, OwnerID: ownerID, PipelineID: defaultPipelineID, AccountID: &accountB.ID,
	})
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO quotes (id, org_id, deal_id, contact_id, title, status, currency)
		VALUES ($1,$2,$3,$4,'Invalid quote','draft','USD')
	`, uuid.New(), defaultOrgID, validDeal.ID, contactA.ID)
	require.Error(t, err)
}
