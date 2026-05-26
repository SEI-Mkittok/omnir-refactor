package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// orgScopedTables are all tables that carry an org_id column and have RLS
// policies defined. When ORG_MODE is 'saas', 'multitenant', or 'enterprise',
// EnableRLS applies FORCE ROW LEVEL SECURITY to every table in this list so
// that the database enforces tenant isolation even for the application DB role.
var orgScopedTables = []string{
	"users",
	"accounts",
	"contacts",
	"pipelines",
	"deals",
	"activities",
	"notes",
	"notifications",
	"tickets",
	"ticket_comments",
	"ticket_attachments",
	"leads",
	"crm_entity_links",
	"crm_roles",
	"crm_role_closure",
	"crm_profiles",
	"crm_profile_permissions",
	"crm_profile_field_permissions",
	"crm_groups",
	"crm_group_members",
	"crm_sharing_defaults",
	"crm_sharing_grants",
	"crm_sharing_rules",
	"module_layouts",
	"module_relationship_definitions",
	"webforms",
	"webform_submissions",
	"mail_converter_rules",
	"mail_converter_runs",
	"mail_converter_logs",
	"campaigns",
	"campaign_members",
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
