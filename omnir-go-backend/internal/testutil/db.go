package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const defaultTestDSN = "postgres://omnir:omnir_dev@localhost:5433/omnir_crm_test?sslmode=disable"

// NewTestDB returns a pgxpool connected to the test database.
// It registers cleanup to close the pool when the test finishes.
func NewTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err, "failed to connect to test DB")
	t.Cleanup(pool.Close)
	return pool
}

// TruncateAll clears all core tables between tests for isolation.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		TRUNCATE
			notes,
			activities,
			deals,
			contacts,
			accounts,
			pipelines,
			users
		RESTART IDENTITY CASCADE
	`)
	require.NoError(t, err, "failed to truncate tables")
}
