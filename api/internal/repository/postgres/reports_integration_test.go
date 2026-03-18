//go:build integration

package postgres_test

import (
	"testing"
	"time"

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
		assert.Equal(t, 3, report.TotalOpen)  // open + pending + old open
		assert.Equal(t, 1, report.TotalClosed) // resolved
		assert.Equal(t, 4, len(report.ByStatus))
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
		assert.Equal(t, 2, report.TotalOpen)  // open + pending (within last hour)
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
		assert.Equal(t, 2, report.TotalCount)
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
		assert.InDelta(t, 66.67, report.ConversionRate, 0.1)
	})
}
