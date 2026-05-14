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

	if orgID != uuid.Nil {
		orgPlaceholder := fmt.Sprintf("$%d", startIdx+len(args))
		args = append(args, orgID)
		modulePlaceholder := fmt.Sprintf("$%d", startIdx+len(args))
		args = append(args, module)
		var roleID any
		if acl.RoleID != nil {
			roleID = *acl.RoleID
		}
		rolePlaceholder := fmt.Sprintf("$%d", startIdx+len(args))
		args = append(args, roleID)
		groupIDs := acl.GroupIDs
		if groupIDs == nil {
			groupIDs = []uuid.UUID{}
		}
		groupsPlaceholder := fmt.Sprintf("$%d", startIdx+len(args))
		args = append(args, groupIDs)

		accessPredicate := "sr.access_level = 'write'"
		if access == domain.SharingAccessRead {
			accessPredicate = "sr.access_level IN ('read', 'write')"
		}
		targetPredicate := fmt.Sprintf(`(
			(sr.target_type = 'user' AND sr.target_id = %s)
			OR (sr.target_type = 'role' AND %s::uuid IS NOT NULL AND sr.target_id = %s)
			OR (
				sr.target_type = 'role_subordinates'
				AND %s::uuid IS NOT NULL
				AND EXISTS (
					SELECT 1
					FROM crm_role_closure target_closure
					WHERE target_closure.org_id = sr.org_id
					  AND target_closure.ancestor_id = sr.target_id
					  AND target_closure.descendant_id = %s
				)
			)
			OR (sr.target_type = 'group' AND sr.target_id = ANY(%s))
		)`, userPlaceholder, rolePlaceholder, rolePlaceholder, rolePlaceholder, rolePlaceholder, groupsPlaceholder)

		for _, ownerExpr := range ownerExprs {
			sourcePredicate := fmt.Sprintf(`(
				sr.source_type = 'all'
				OR (sr.source_type = 'user' AND sr.source_id = %s)
				OR (
					sr.source_type = 'role'
					AND EXISTS (
						SELECT 1
						FROM users owner_user
						LEFT JOIN crm_roles fallback_role
							ON fallback_role.org_id = owner_user.org_id
							AND fallback_role.system_key = owner_user.role
						WHERE owner_user.id = %s
						  AND owner_user.org_id = %s
						  AND owner_user.deleted_at IS NULL
						  AND COALESCE(owner_user.role_id, fallback_role.id) = sr.source_id
					)
				)
				OR (
					sr.source_type = 'role_subordinates'
					AND EXISTS (
						SELECT 1
						FROM users owner_user
						LEFT JOIN crm_roles fallback_role
							ON fallback_role.org_id = owner_user.org_id
							AND fallback_role.system_key = owner_user.role
						JOIN crm_role_closure source_closure
							ON source_closure.org_id = owner_user.org_id
							AND source_closure.ancestor_id = sr.source_id
							AND source_closure.descendant_id = COALESCE(owner_user.role_id, fallback_role.id)
						WHERE owner_user.id = %s
						  AND owner_user.org_id = %s
						  AND owner_user.deleted_at IS NULL
					)
				)
				OR (
					sr.source_type = 'group'
					AND EXISTS (
						SELECT 1
						FROM crm_group_members source_group
						WHERE source_group.org_id = sr.org_id
						  AND source_group.group_id = sr.source_id
						  AND source_group.user_id = %s
					)
				)
			)`, ownerExpr, ownerExpr, orgPlaceholder, ownerExpr, orgPlaceholder, ownerExpr)
			clauses = append(clauses, fmt.Sprintf(`EXISTS (
				SELECT 1
				FROM crm_sharing_rules sr
				WHERE sr.org_id = %s
				  AND sr.module = %s
				  AND %s
				  AND %s
				  AND %s
			)`, orgPlaceholder, modulePlaceholder, accessPredicate, targetPredicate, sourcePredicate))
		}
	}

	return "(" + strings.Join(clauses, " OR ") + ")", args
}
