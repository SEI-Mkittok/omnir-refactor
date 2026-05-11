package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/omnir/crm-api/internal/domain"
)

func addAccessVisibilityWhere(ctx context.Context, where *[]string, args *[]any, idx *int, module domain.ACLModule, access domain.SharingAccessLevel, ownerPredicate string) {
	acl, ok := domain.AccessContextFromContext(ctx)
	if !ok || acl.CanAccessAllRecords(module, access) {
		return
	}
	placeholder := fmt.Sprintf("$%d", *idx)
	if strings.Count(ownerPredicate, "%s") > 1 {
		*where = append(*where, fmt.Sprintf(ownerPredicate, placeholder, placeholder))
	} else {
		*where = append(*where, fmt.Sprintf(ownerPredicate, placeholder))
	}
	*args = append(*args, acl.UserID)
	*idx = *idx + 1
}
