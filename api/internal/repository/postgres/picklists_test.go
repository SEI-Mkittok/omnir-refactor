package postgres

import (
	"strings"
	"testing"

	"github.com/omnir/crm-api/internal/domain"
)

func TestPicklistSQLForQuoteOmitsSoftDeleteFilter(t *testing.T) {
	storage, ok := picklistStorageFor(domain.CustomFieldEntityQuote)
	if !ok {
		t.Fatal("expected quote picklist storage")
	}

	for name, sql := range map[string]string{
		"usage count":  picklistUsageCountSQL(storage),
		"remap select": picklistRemapSelectSQL(storage),
	} {
		if !strings.Contains(sql, "FROM quotes") {
			t.Fatalf("%s SQL should query quotes: %s", name, sql)
		}
		if strings.Contains(sql, "deleted_at IS NULL") {
			t.Fatalf("%s SQL should not filter deleted_at for quotes: %s", name, sql)
		}
	}
	if clause := softDeleteClause(storage); clause != "" {
		t.Fatalf("quote soft delete clause should be empty, got %q", clause)
	}
}

func TestPicklistSQLForKBArticleKeepsSoftDeleteFilter(t *testing.T) {
	storage, ok := picklistStorageFor(domain.CustomFieldEntityKBArticle)
	if !ok {
		t.Fatal("expected KB article picklist storage")
	}

	for name, sql := range map[string]string{
		"usage count":  picklistUsageCountSQL(storage),
		"remap select": picklistRemapSelectSQL(storage),
	} {
		if !strings.Contains(sql, "FROM articles") {
			t.Fatalf("%s SQL should query articles: %s", name, sql)
		}
		if !strings.Contains(sql, "deleted_at IS NULL") {
			t.Fatalf("%s SQL should keep deleted_at filter for articles: %s", name, sql)
		}
	}
}
