//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/testutil"
)

var (
	defaultOrgID      = domain.DefaultOrgID
	defaultPipelineID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
)

// setupDB connects to the test DB, truncates all tables, and re-seeds required
// fixtures (default org + default pipeline). Returns a pool and a context
// pre-loaded with the default org ID.
func setupDB(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	pool := testutil.NewTestDB(t)
	testutil.TruncateAll(t, pool)

	ctx := context.Background()

	// Re-seed default organization (truncated above).
	_, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, slug, plan)
		VALUES ($1, 'Default', 'default', 'self_hosted')
	`, defaultOrgID)
	require.NoError(t, err)

	// Re-seed default pipeline.
	_, err = pool.Exec(ctx, `
		INSERT INTO pipelines (id, org_id, name, stages)
		VALUES ($1, $2, 'Default', '[]'::jsonb)
	`, defaultPipelineID, defaultOrgID)
	require.NoError(t, err)

	return pool, domain.WithOrgID(ctx, defaultOrgID)
}

// seedUser inserts a minimal user row and returns its ID.
func seedUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, org_id, email, name, role)
		VALUES ($1, $2, $3, 'Test User', 'admin')
	`, id, defaultOrgID, "test+"+id.String()+"@omnir.test")
	require.NoError(t, err)
	return id
}
