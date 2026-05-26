package postgres

import "testing"

func TestOrgScopedTablesIncludesNewOrgScopedTables(t *testing.T) {
	tables := make(map[string]bool, len(orgScopedTables))
	for _, table := range orgScopedTables {
		tables[table] = true
	}

	for _, table := range []string{
		"module_layouts",
		"module_relationship_definitions",
		"crm_sharing_rules",
		"webforms",
		"webform_submissions",
		"mail_converter_rules",
		"mail_converter_runs",
		"mail_converter_logs",
		"campaigns",
		"campaign_members",
	} {
		if !tables[table] {
			t.Fatalf("orgScopedTables missing %s; org-scoped tables must receive FORCE ROW LEVEL SECURITY in multi-tenant modes", table)
		}
	}
}
