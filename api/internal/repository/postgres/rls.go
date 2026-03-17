package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// orgScopedTables are all tables that carry an org_id column and have RLS
// policies defined by migration 20240101000014_add_rls_policies.sql.
var orgScopedTables = []string{
	"users",
	"accounts",
	"contacts",
	"pipelines",
	"deals",
	"activities",
	"notes",
	"notifications",
}

// EnableRLS activates FORCE ROW LEVEL SECURITY on every org-scoped table so
// that the database enforces tenant isolation even for the application DB role.
// Call this at startup when ORG_MODE is 'multitenant' or 'enterprise'.
func EnableRLS(ctx context.Context, db *pgxpool.Pool) error {
	for _, table := range orgScopedTables {
		sql := fmt.Sprintf("ALTER TABLE %s FORCE ROW LEVEL SECURITY", table)
		if _, err := db.Exec(ctx, sql); err != nil {
			return fmt.Errorf("EnableRLS: table %s: %w", table, err)
		}
	}
	return nil
}

// DisableRLS removes FORCE ROW LEVEL SECURITY from every org-scoped table.
// Intended for use in tests or when switching back to single-tenant mode.
func DisableRLS(ctx context.Context, db *pgxpool.Pool) error {
	for _, table := range orgScopedTables {
		sql := fmt.Sprintf("ALTER TABLE %s NO FORCE ROW LEVEL SECURITY", table)
		if _, err := db.Exec(ctx, sql); err != nil {
			return fmt.Errorf("DisableRLS: table %s: %w", table, err)
		}
	}
	return nil
}
