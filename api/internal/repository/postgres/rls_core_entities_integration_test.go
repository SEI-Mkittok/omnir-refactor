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

// TestRLS_AccountTenantIsolation verifies RLS on the accounts table.
func TestRLS_AccountTenantIsolation(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "acct-rls-org-a"},
		{orgB, "acct-rls-org-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, "Org "+org.slug, org.slug)
		require.NoError(t, err)
	}

	// Each org needs an owner user.
	ownerA := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Account Owner A', 'admin')
	`, ownerA, orgA, "acctowner+"+ownerA.String()+"@omnir.test")
	require.NoError(t, err)

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	repo := postgres.NewAccountRepo(pool)
	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	account, err := repo.Create(ctxA, &domain.Account{
		OrgID:   orgA,
		Name:    "Acme Corp",
		OwnerID: ownerA,
	})
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.Equal(t, orgA, account.OrgID)

	t.Run("org A can read its own account", func(t *testing.T) {
		got, err := repo.GetByID(ctxA, account.ID)
		require.NoError(t, err)
		assert.Equal(t, account.ID, got.ID)
	})

	t.Run("org B cannot read org A account", func(t *testing.T) {
		_, err := repo.GetByID(ctxB, account.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound, "RLS should hide org A account from org B")
	})

	t.Run("org B List returns empty", func(t *testing.T) {
		accounts, total, err := repo.List(ctxB, domain.AccountFilter{OrgID: orgB})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, accounts)
	})

	t.Run("org A List returns its account", func(t *testing.T) {
		accounts, total, err := repo.List(ctxA, domain.AccountFilter{OrgID: orgA})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, accounts, 1)
	})
}

// TestRLS_DealTenantIsolation verifies RLS on the deals table.
func TestRLS_DealTenantIsolation(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "deal-rls-org-a"},
		{orgB, "deal-rls-org-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, "Org "+org.slug, org.slug)
		require.NoError(t, err)
	}

	ownerA := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Deal Owner A', 'admin')
	`, ownerA, orgA, "dealowner+"+ownerA.String()+"@omnir.test")
	require.NoError(t, err)

	// Seed a pipeline for orgA (pipeline_id is NOT NULL on deals).
	pipelineA := uuid.New()
	_, err = pool.Exec(context.Background(), `
		INSERT INTO pipelines (id, org_id, name, stages)
		VALUES ($1, $2, 'Default', '[]'::jsonb)
	`, pipelineA, orgA)
	require.NoError(t, err)

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	repo := postgres.NewDealRepo(pool)
	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	deal, err := repo.Create(ctxA, &domain.Deal{
		OrgID:      orgA,
		Title:      "RLS Test Deal",
		Stage:      domain.DealStageLead,
		Currency:   "USD",
		OwnerID:    ownerA,
		PipelineID: pipelineA,
	})
	require.NoError(t, err)
	require.NotNil(t, deal)
	assert.Equal(t, orgA, deal.OrgID)

	t.Run("org A can read its own deal", func(t *testing.T) {
		got, err := repo.GetByID(ctxA, deal.ID)
		require.NoError(t, err)
		assert.Equal(t, deal.ID, got.ID)
	})

	t.Run("org B cannot read org A deal", func(t *testing.T) {
		_, err := repo.GetByID(ctxB, deal.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound, "RLS should hide org A deal from org B")
	})

	t.Run("org B List returns empty", func(t *testing.T) {
		deals, total, err := repo.List(ctxB, domain.DealFilter{OrgID: orgB})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, deals)
	})

	t.Run("org A List returns its deal", func(t *testing.T) {
		deals, total, err := repo.List(ctxA, domain.DealFilter{OrgID: orgA})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, deals, 1)
	})
}

// TestRLS_ActivityTenantIsolation verifies RLS on the activities table.
func TestRLS_ActivityTenantIsolation(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "activity-rls-org-a"},
		{orgB, "activity-rls-org-b"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO orgs (id, name, slug, plan)
			VALUES ($1, $2, $3, 'starter')
		`, org.id, "Org "+org.slug, org.slug)
		require.NoError(t, err)
	}

	ownerA := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Activity Owner A', 'admin')
	`, ownerA, orgA, "activityowner+"+ownerA.String()+"@omnir.test")
	require.NoError(t, err)

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	repo := postgres.NewActivityRepo(pool)
	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	activity, err := repo.Create(ctxA, &domain.Activity{
		OrgID:   orgA,
		Type:    domain.ActivityTypeTask,
		Subject: "RLS Test Activity",
		OwnerID: ownerA,
	})
	require.NoError(t, err)
	require.NotNil(t, activity)
	assert.Equal(t, orgA, activity.OrgID)

	t.Run("org A can read its own activity", func(t *testing.T) {
		got, err := repo.GetByID(ctxA, activity.ID)
		require.NoError(t, err)
		assert.Equal(t, activity.ID, got.ID)
	})

	t.Run("org B cannot read org A activity", func(t *testing.T) {
		_, err := repo.GetByID(ctxB, activity.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound, "RLS should hide org A activity from org B")
	})

	t.Run("org B List returns empty", func(t *testing.T) {
		activities, total, err := repo.List(ctxB, domain.ActivityFilter{OrgID: orgB})
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, activities)
	})

	t.Run("org A List returns its activity", func(t *testing.T) {
		activities, total, err := repo.List(ctxA, domain.ActivityFilter{OrgID: orgA})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, activities, 1)
	})
}

// TestRLS_CrossOrgDataLeakage is a broad sanity check ensuring no cross-org
// contamination occurs when two orgs have data in accounts, deals, and activities.
func TestRLS_CrossOrgDataLeakage(t *testing.T) {
	pool, _ := setupDB(t)

	orgA := uuid.New()
	orgB := uuid.New()
	for _, org := range []struct {
		id   uuid.UUID
		slug string
	}{
		{orgA, "cross-leak-org-a"},
		{orgB, "cross-leak-org-b"},
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
		email string
	}{
		{ownerA, orgA, "cross+a+" + ownerA.String() + "@omnir.test"},
		{ownerB, orgB, "cross+b+" + ownerB.String() + "@omnir.test"},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO users (id, org_id, email, name, role)
			VALUES ($1, $2, $3, 'User', 'admin')
		`, u.id, u.orgID, u.email)
		require.NoError(t, err)
	}

	// Seed pipelines for each org (pipeline_id is NOT NULL on deals).
	pipelineA, pipelineB := uuid.New(), uuid.New()
	for _, p := range []struct {
		id    uuid.UUID
		orgID uuid.UUID
	}{
		{pipelineA, orgA},
		{pipelineB, orgB},
	} {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO pipelines (id, org_id, name, stages)
			VALUES ($1, $2, 'Default', '[]'::jsonb)
		`, p.id, p.orgID)
		require.NoError(t, err)
	}

	require.NoError(t, postgres.EnableRLS(context.Background(), pool))
	t.Cleanup(func() { _ = postgres.DisableRLS(context.Background(), pool) })

	ctxA := domain.WithOrgID(context.Background(), orgA)
	ctxB := domain.WithOrgID(context.Background(), orgB)

	accountRepo := postgres.NewAccountRepo(pool)
	dealRepo := postgres.NewDealRepo(pool)
	activityRepo := postgres.NewActivityRepo(pool)

	// Seed data for both orgs.
	_, err := accountRepo.Create(ctxA, &domain.Account{OrgID: orgA, Name: "OrgA Account", OwnerID: ownerA})
	require.NoError(t, err)
	_, err = accountRepo.Create(ctxB, &domain.Account{OrgID: orgB, Name: "OrgB Account", OwnerID: ownerB})
	require.NoError(t, err)

	_, err = dealRepo.Create(ctxA, &domain.Deal{OrgID: orgA, Title: "OrgA Deal", Stage: domain.DealStageLead, Currency: "USD", OwnerID: ownerA, PipelineID: pipelineA})
	require.NoError(t, err)
	_, err = dealRepo.Create(ctxB, &domain.Deal{OrgID: orgB, Title: "OrgB Deal", Stage: domain.DealStageLead, Currency: "USD", OwnerID: ownerB, PipelineID: pipelineB})
	require.NoError(t, err)

	_, err = activityRepo.Create(ctxA, &domain.Activity{OrgID: orgA, Type: domain.ActivityTypeCall, Subject: "OrgA Call", OwnerID: ownerA})
	require.NoError(t, err)
	_, err = activityRepo.Create(ctxB, &domain.Activity{OrgID: orgB, Type: domain.ActivityTypeCall, Subject: "OrgB Call", OwnerID: ownerB})
	require.NoError(t, err)

	// Each org should see exactly 1 of each entity type.
	t.Run("each org sees only its own accounts", func(t *testing.T) {
		accsA, totalA, err := accountRepo.List(ctxA, domain.AccountFilter{OrgID: orgA})
		require.NoError(t, err)
		assert.Equal(t, 1, totalA)
		assert.Equal(t, "OrgA Account", accsA[0].Name)

		accsB, totalB, err := accountRepo.List(ctxB, domain.AccountFilter{OrgID: orgB})
		require.NoError(t, err)
		assert.Equal(t, 1, totalB)
		assert.Equal(t, "OrgB Account", accsB[0].Name)
	})

	t.Run("each org sees only its own deals", func(t *testing.T) {
		dealsA, totalA, err := dealRepo.List(ctxA, domain.DealFilter{OrgID: orgA})
		require.NoError(t, err)
		assert.Equal(t, 1, totalA)
		assert.Equal(t, "OrgA Deal", dealsA[0].Title)

		dealsB, totalB, err := dealRepo.List(ctxB, domain.DealFilter{OrgID: orgB})
		require.NoError(t, err)
		assert.Equal(t, 1, totalB)
		assert.Equal(t, "OrgB Deal", dealsB[0].Title)
	})

	t.Run("each org sees only its own activities", func(t *testing.T) {
		actsA, totalA, err := activityRepo.List(ctxA, domain.ActivityFilter{OrgID: orgA})
		require.NoError(t, err)
		assert.Equal(t, 1, totalA)
		assert.Equal(t, "OrgA Call", actsA[0].Subject)

		actsB, totalB, err := activityRepo.List(ctxB, domain.ActivityFilter{OrgID: orgB})
		require.NoError(t, err)
		assert.Equal(t, 1, totalB)
		assert.Equal(t, "OrgB Call", actsB[0].Subject)
	})
}
