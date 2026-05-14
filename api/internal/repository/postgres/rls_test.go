package postgres

import "testing"

func TestOrgScopedTablesIncludesBundle5ConfigurationTables(t *testing.T) {
	tables := make(map[string]bool, len(orgScopedTables))
	for _, table := range orgScopedTables {
		tables[table] = true
	}

	for _, table := range []string{"module_layouts", "module_relationship_definitions"} {
		if !tables[table] {
			t.Fatalf("orgScopedTables missing %s; org-scoped tables must receive FORCE ROW LEVEL SECURITY in multi-tenant modes", table)
		}
	}
}
