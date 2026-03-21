package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// NewOrgScopedPool creates a pgxpool that automatically sets the PostgreSQL
// session variable app.current_org_id on every acquired connection using the
// org_id stored in the request context. This variable is read by Row-Level
// Security policies to enforce tenant isolation.
//
// If no org_id is present in the context, domain.DefaultOrgID is used so that
// single-tenant deployments work without JWT org claims.
//
// Implementation note — session-level vs SET LOCAL:
// set_config is called with is_local=false (session-scoped). SET LOCAL
// (is_local=true) would only persist until the end of the current transaction.
// Because BeforeAcquire runs outside any active transaction, using is_local=true
// would clear the variable after the first implicit transaction (i.e. the first
// query), leaving subsequent queries in the same request without an org context.
// Session-scoped setting combined with AfterRelease cleanup provides equivalent
// per-request isolation: the variable is always set before the first query and
// always cleared before the connection is returned to the pool.
func NewOrgScopedPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		orgID, ok := domain.OrgIDFromContext(ctx)
		if !ok {
			orgID = domain.DefaultOrgID
		}
		_, err := conn.Exec(ctx,
			`SELECT set_config('app.current_org_id', $1, false)`,
			orgID.String(),
		)
		return err == nil
	}

	cfg.AfterRelease = func(conn *pgx.Conn) bool {
		_, err := conn.Exec(context.Background(),
			`SELECT set_config('app.current_org_id', '', false)`,
		)
		return err == nil
	}

	return pgxpool.NewWithConfig(ctx, cfg)
}
