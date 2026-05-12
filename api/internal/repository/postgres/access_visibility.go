package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

func addAccessVisibilityWhere(ctx context.Context, where *[]string, args *[]any, idx *int, module domain.ACLModule, access domain.SharingAccessLevel, ownerExprs ...string) {
	predicate, predicateArgs := accessVisibilityPredicate(ctx, *idx, module, access, ownerExprs...)
	if predicate == "" {
		return
	}
	*where = append(*where, predicate)
	*args = append(*args, predicateArgs...)
	*idx += len(predicateArgs)
}

func appendAccessVisibilitySQL(ctx context.Context, q *string, args *[]any, module domain.ACLModule, access domain.SharingAccessLevel, ownerExprs ...string) {
	predicate, predicateArgs := accessVisibilityPredicate(ctx, len(*args)+1, module, access, ownerExprs...)
	if predicate == "" {
		return
	}
	*q += " AND " + predicate
	*args = append(*args, predicateArgs...)
}

func accessVisibilityPredicate(ctx context.Context, startIdx int, module domain.ACLModule, access domain.SharingAccessLevel, ownerExprs ...string) (string, []any) {
	acl, ok := domain.AccessContextFromContext(ctx)
	if !ok || acl.CanAccessAllRecords(module, access) || len(ownerExprs) == 0 {
		return "", nil
	}

	userPlaceholder := fmt.Sprintf("$%d", startIdx)
	clauses := make([]string, 0, len(ownerExprs)*2)
	for _, ownerExpr := range ownerExprs {
		clauses = append(clauses, fmt.Sprintf("%s = %s", ownerExpr, userPlaceholder))
	}
	args := []any{acl.UserID}

	orgID := acl.OrgID
	if orgID == uuid.Nil {
		if ctxOrgID, ok := domain.OrgIDFromContext(ctx); ok {
			orgID = ctxOrgID
		}
	}
	if acl.RoleID != nil && orgID != uuid.Nil {
		rolePlaceholder := fmt.Sprintf("$%d", startIdx+len(args))
		args = append(args, *acl.RoleID)
		orgPlaceholder := fmt.Sprintf("$%d", startIdx+len(args))
		args = append(args, orgID)

		for _, ownerExpr := range ownerExprs {
			clauses = append(clauses, fmt.Sprintf(`EXISTS (
				SELECT 1
				FROM users owner_user
				LEFT JOIN crm_roles fallback_role
					ON fallback_role.org_id = owner_user.org_id
					AND fallback_role.system_key = owner_user.role
				JOIN crm_role_closure role_closure
					ON role_closure.org_id = owner_user.org_id
					AND role_closure.ancestor_id = %s
					AND role_closure.descendant_id = COALESCE(owner_user.role_id, fallback_role.id)
					AND role_closure.depth > 0
				WHERE owner_user.id = %s
				  AND owner_user.org_id = %s
				  AND owner_user.deleted_at IS NULL
			)`, rolePlaceholder, ownerExpr, orgPlaceholder))
		}
	}

	return "(" + strings.Join(clauses, " OR ") + ")", args
}
