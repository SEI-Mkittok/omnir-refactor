//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

func TestReportsRepo_TicketMetrics(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewReportsRepo(pool)
	ownerID := seedUser(t, pool)

	// Seed tickets of different statuses.
	_, err := pool.Exec(ctx, `
		INSERT INTO tickets (id, org_id, subject, status, priority, assignee_id, created_at, updated_at)
		VALUES
			(gen_random_uuid(), $1, 'Open ticket',     'open',     'medium', $2, NOW(),               NOW()),
			(gen_random_uuid(), $1, 'Pending ticket',  'pending',  'low',    $2, NOW(),               NOW()),
			(gen_random_uuid(), $1, 'Resolved ticket', 'resolved', 'high',   $2, NOW() - INTERVAL '2h', NOW()),
			(gen_random_uuid(), $1, 'Old open ticket', 'open',     'medium', $2, NOW() - INTERVAL '72h', NOW())
	`, defaultOrgID, ownerID)
	require.NoError(t, err)

	t.Run("returns correct open/closed counts", func(t *testing.T) {
		report, err := repo.TicketMetrics(ctx, domain.ReportFilter{})
		require.NoError(t, err)
		assert.Equal(t, 3, report.TotalOpen)     // open + pending + old open
		assert.Equal(t, 1, report.TotalClosed)   // resolved
		assert.Equal(t, 3, len(report.ByStatus)) // 3 distinct statuses: open, pending, resolved
	})

	t.Run("breach rate reflects old open tickets", func(t *testing.T) {
		report, err := repo.TicketMetrics(ctx, domain.ReportFilter{})
		require.NoError(t, err)
		// 1 out of 3 open tickets is older than 48h → ~33%
		assert.Greater(t, report.BreachRate, 0.0)
		assert.Less(t, report.BreachRate, 100.0)
	})

	t.Run("date filter from restricts results", func(t *testing.T) {
		from := time.Now().Add(-time.Hour) // only tickets from last hour
		report, err := repo.TicketMetrics(ctx, domain.ReportFilter{From: &from})
		require.NoError(t, err)
		// The old open ticket (72h ago) and resolved (2h ago) should be excluded.
		assert.Equal(t, 2, report.TotalOpen) // open + pending (within last hour)
		assert.Equal(t, 0, report.TotalClosed)
	})
}

func TestReportsRepo_ContactMetrics(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewReportsRepo(pool)
	ownerID := seedUser(t, pool)

	_, err := pool.Exec(ctx, `
		INSERT INTO contacts (id, org_id, first_name, last_name, owner_id, created_at, updated_at)
		VALUES
			(gen_random_uuid(), $1, 'Alice', 'A', $2, NOW(),                NOW()),
			(gen_random_uuid(), $1, 'Bob',   'B', $2, NOW(),                NOW()),
			(gen_random_uuid(), $1, 'Old',   'C', $2, NOW() - INTERVAL '60 days', NOW())
	`, defaultOrgID, ownerID)
	require.NoError(t, err)

	t.Run("total counts all non-deleted contacts", func(t *testing.T) {
		report, err := repo.ContactMetrics(ctx, domain.ReportFilter{})
		require.NoError(t, err)
		assert.Equal(t, 3, report.TotalCount)
	})

	t.Run("date filter from restricts total", func(t *testing.T) {
		from := time.Now().Add(-time.Hour)
		report, err := repo.ContactMetrics(ctx, domain.ReportFilter{From: &from})
		require.NoError(t, err)
		assert.Equal(t, 2, report.NewCount)
	})

	t.Run("by_period returns monthly buckets", func(t *testing.T) {
		report, err := repo.ContactMetrics(ctx, domain.ReportFilter{})
		require.NoError(t, err)
		assert.NotEmpty(t, report.OverTime)
	})
}

func TestReportsRepo_DealMetrics(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewReportsRepo(pool)
	ownerID := seedUser(t, pool)

	_, err := pool.Exec(ctx, `
		INSERT INTO deals (id, org_id, title, value_cents, currency, stage, owner_id, pipeline_id, created_at, updated_at)
		VALUES
			(gen_random_uuid(), $1, 'Active deal',   100000, 'USD', 'qualified',  $2, $3, NOW(), NOW()),
			(gen_random_uuid(), $1, 'Won deal',      200000, 'USD', 'closed_won', $2, $3, NOW(), NOW()),
			(gen_random_uuid(), $1, 'Lost deal',      50000, 'USD', 'closed_lost',$2, $3, NOW(), NOW())
	`, defaultOrgID, ownerID, defaultPipelineID)
	require.NoError(t, err)

	t.Run("pipeline value excludes closed deals", func(t *testing.T) {
		report, err := repo.DealMetrics(ctx, domain.ReportFilter{})
		require.NoError(t, err)
		assert.Equal(t, int64(100000), report.PipelineValueCents)
		assert.Equal(t, 1, report.WonCount)
		assert.Equal(t, 1, report.LostCount)
		assert.Equal(t, 3, len(report.ByStage))
	})
}

func TestReportsRepo_LeadMetrics(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewReportsRepo(pool)
	ownerID := seedUser(t, pool)

	_, err := pool.Exec(ctx, `
		INSERT INTO leads (id, org_id, first_name, last_name, status, owner_id, created_at, updated_at)
		VALUES
			(gen_random_uuid(), $1, 'Lead1', 'A', 'new',       $2, NOW(), NOW()),
			(gen_random_uuid(), $1, 'Lead2', 'B', 'converted', $2, NOW(), NOW()),
			(gen_random_uuid(), $1, 'Lead3', 'C', 'converted', $2, NOW(), NOW())
	`, defaultOrgID, ownerID)
	require.NoError(t, err)

	t.Run("conversion rate is calculated correctly", func(t *testing.T) {
		report, err := repo.LeadMetrics(ctx, domain.ReportFilter{})
		require.NoError(t, err)
		assert.Equal(t, 3, report.NewCount)
		assert.Equal(t, 2, report.ConvertedCount)
		assert.InDelta(t, 0.6667, report.ConversionRate, 0.001)
	})
}

func TestReportsRepo_ManagerDashboard_RangeAndOrgIsolation(t *testing.T) {
	pool, ctx := setupDB(t)
	repo := postgres.NewReportsRepo(pool)

	ownerID := seedUser(t, pool)
	otherOrgID := uuid.New()
	otherPipelineID := uuid.New()
	otherOwnerID := uuid.New()

	_, err := pool.Exec(context.Background(), `
		INSERT INTO orgs (id, name, slug, plan) VALUES ($1, 'Other Org', 'other-org', 'pro')
	`, otherOrgID)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), `
		INSERT INTO pipelines (id, org_id, name, stages) VALUES ($1, $2, 'Other Pipeline', '[]'::jsonb)
	`, otherPipelineID, otherOrgID)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role) VALUES ($1, $2, $3, 'Other User', 'agent')
	`, otherOwnerID, otherOrgID, "other+"+otherOwnerID.String()+"@omnir.test")
	require.NoError(t, err)

	now := time.Now().UTC()
	recent := now.Add(-2 * time.Hour)
	old := now.Add(-72 * time.Hour)

	_, err = pool.Exec(context.Background(), `
		INSERT INTO deals (id, org_id, title, value_cents, currency, stage, owner_id, pipeline_id, created_at, updated_at)
		VALUES
			(gen_random_uuid(), $1, 'Recent Qualified', 100000, 'USD', 'qualified',  $2, $3, $4, $4),
			(gen_random_uuid(), $1, 'Old Lost',          50000, 'USD', 'closed_lost', $2, $3, $5, $5),
			(gen_random_uuid(), $6, 'Other Won',        999999, 'USD', 'closed_won',  $7, $8, $4, $4)
	`, defaultOrgID, ownerID, defaultPipelineID, recent, old, otherOrgID, otherOwnerID, otherPipelineID)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), `
		INSERT INTO tickets (id, org_id, subject, status, priority, assignee_id, created_at, updated_at)
		VALUES
			(gen_random_uuid(), $1, 'Recent Open',     'open',     'medium', $2, $4, $4),
			(gen_random_uuid(), $1, 'Old Resolved',    'resolved', 'medium', $2, $5, $4),
			(gen_random_uuid(), $6, 'Other Open',      'open',     'low',    $7, $4, $4)
	`, defaultOrgID, ownerID, recent, old, otherOrgID, otherOwnerID)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), `
		INSERT INTO activities (id, org_id, type, subject, owner_id, created_at, updated_at, completed_at)
		VALUES
			(gen_random_uuid(), $1, 'call', 'Recent Activity', $2, $4, $4, $4),
			(gen_random_uuid(), $1, 'task', 'Old Activity',    $2, $5, $5, NULL),
			(gen_random_uuid(), $6, 'email','Other Activity',  $6, $3, $3, $3)
	`, defaultOrgID, ownerID, recent, old, otherOrgID, otherOwnerID)
	require.NoError(t, err)

	from := now.Add(-24 * time.Hour)
	report, err := repo.ManagerDashboard(ctx, domain.ReportFilter{From: &from})
	require.NoError(t, err)

	assert.Equal(t, int64(100000), report.CRM.PipelineValueCents)
	assert.Equal(t, 0, report.CRM.WonCount)
	assert.Equal(t, 0, report.CRM.LostCount)
	assert.Equal(t, 1, report.HelpDesk.OpenCount)
	assert.Equal(t, 0, report.HelpDesk.BacklogCount)
	require.Len(t, report.TeamActivity.ByUser, 1)
	assert.Equal(t, ownerID, report.TeamActivity.ByUser[0].OwnerID)
	assert.Equal(t, 1, report.TeamActivity.ByUser[0].CreatedCount)
	assert.Equal(t, 1, report.TeamActivity.ByUser[0].CompletedCount)
}
