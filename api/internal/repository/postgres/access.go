package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type AccessRepo struct {
	db *pgxpool.Pool
}

func NewAccessRepo(db *pgxpool.Pool) *AccessRepo {
	return &AccessRepo{db: db}
}

func (r *AccessRepo) ResolveAccess(ctx context.Context, userID, orgID uuid.UUID, platformRole string) (*domain.AccessContext, error) {
	access := &domain.AccessContext{
		UserID:           userID,
		OrgID:            orgID,
		PlatformRole:     platformRole,
		Permissions:      map[domain.ACLModule]map[domain.ACLAction]bool{},
		FieldWrite:       map[domain.ACLModule]map[string]bool{},
		Sharing:          map[domain.ACLModule]domain.ACLSharingAccess{},
		SuperAdminBypass: platformRole == string(domain.UserRoleSuperAdmin),
	}
	if userID == uuid.Nil || orgID == uuid.Nil {
		return access, nil
	}

	var roleID, profileID *uuid.UUID
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(u.role_id, fallback_role.id), COALESCE(u.profile_id, fallback_profile.id)
		FROM users u
		LEFT JOIN crm_roles fallback_role
			ON fallback_role.org_id = u.org_id AND fallback_role.system_key = u.role
		LEFT JOIN crm_profiles fallback_profile
			ON fallback_profile.org_id = u.org_id
			AND fallback_profile.system_key = CASE
				WHEN u.role IN ('super_admin', 'admin') THEN 'administrator'
				WHEN u.role = 'agent' THEN 'sales'
				ELSE 'guest'
			END
		WHERE u.id = $1 AND u.org_id = $2 AND u.deleted_at IS NULL
	`, userID, orgID).Scan(&roleID, &profileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return access, nil
		}
		return nil, err
	}
	access.RoleID = roleID
	access.ProfileID = profileID

	if roleID != nil {
		rows, err := r.db.Query(ctx, `
			SELECT ancestor_id
			FROM crm_role_closure
			WHERE org_id = $1 AND descendant_id = $2
		`, orgID, *roleID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			access.RoleLineageIDs = append(access.RoleLineageIDs, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
		if len(access.RoleLineageIDs) == 0 {
			access.RoleLineageIDs = append(access.RoleLineageIDs, *roleID)
		}
	}

	rows, err := r.db.Query(ctx, `
		SELECT group_id
		FROM crm_group_members
		WHERE org_id = $1 AND user_id = $2
	`, orgID, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		access.GroupIDs = append(access.GroupIDs, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if profileID != nil {
		rows, err = r.db.Query(ctx, `
			SELECT module, action, allowed
			FROM crm_profile_permissions
			WHERE org_id = $1 AND profile_id = $2
		`, orgID, *profileID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var module domain.ACLModule
			var action domain.ACLAction
			var allowed bool
			if err := rows.Scan(&module, &action, &allowed); err != nil {
				rows.Close()
				return nil, err
			}
			if _, ok := access.Permissions[module]; !ok {
				access.Permissions[module] = map[domain.ACLAction]bool{}
			}
			access.Permissions[module][action] = allowed
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}

		rows, err = r.db.Query(ctx, `
			SELECT module, field_name, can_write
			FROM crm_profile_field_permissions
			WHERE org_id = $1 AND profile_id = $2
		`, orgID, *profileID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var module domain.ACLModule
			var field string
			var canWrite bool
			if err := rows.Scan(&module, &field, &canWrite); err != nil {
				rows.Close()
				return nil, err
			}
			if _, ok := access.FieldWrite[module]; !ok {
				access.FieldWrite[module] = map[string]bool{}
			}
			access.FieldWrite[module][field] = canWrite
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	rows, err = r.db.Query(ctx, `
		SELECT module, mode
		FROM crm_sharing_defaults
		WHERE org_id = $1
	`, orgID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var module domain.ACLModule
		var mode domain.SharingDefaultMode
		if err := rows.Scan(&module, &mode); err != nil {
			rows.Close()
			return nil, err
		}
		access.Sharing[module] = domain.ACLSharingAccess{Mode: mode}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	roleGrantIDs := []uuid.UUID{}
	var roleParam any
	if access.RoleID != nil {
		roleGrantIDs = append(roleGrantIDs, *access.RoleID)
		roleParam = *access.RoleID
	}
	rows, err = r.db.Query(ctx, `
		SELECT module, access_level
		FROM crm_sharing_rules sr
		WHERE org_id = $1
		  AND source_type = 'all'
		  AND (
			(target_type = 'user' AND target_id = $2)
			OR (target_type = 'role' AND target_id = ANY($3))
			OR (
				target_type = 'role_subordinates'
				AND $4::uuid IS NOT NULL
				AND EXISTS (
					SELECT 1
					FROM crm_role_closure target_closure
					WHERE target_closure.org_id = sr.org_id
					  AND target_closure.ancestor_id = sr.target_id
					  AND target_closure.descendant_id = $4
				)
			)
			OR (target_type = 'group' AND target_id = ANY($5))
		  )
	`, orgID, access.UserID, roleGrantIDs, roleParam, access.GroupIDs)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var module domain.ACLModule
		var level domain.SharingAccessLevel
		if err := rows.Scan(&module, &level); err != nil {
			rows.Close()
			return nil, err
		}
		rule := access.Sharing[module]
		if level == domain.SharingAccessWrite {
			rule.WriteAll = true
			rule.ReadAll = true
		} else {
			rule.ReadAll = true
		}
		access.Sharing[module] = rule
	}
	rows.Close()
	return access, rows.Err()
}

func (r *AccessRepo) CanAccessRecord(ctx context.Context, module domain.ACLModule, id uuid.UUID, accessLevel domain.SharingAccessLevel) (bool, error) {
	table, ownerExprs, ok := recordAccessTarget(module)
	if !ok {
		return true, nil
	}

	access, ok := domain.AccessContextFromContext(ctx)
	if !ok || access.CanAccessAllRecords(module, accessLevel) {
		return r.recordExists(ctx, table, id, module, accessLevel)
	}

	return r.recordExists(ctx, table, id, module, accessLevel, ownerExprs...)
}

func (r *AccessRepo) CanAccessAccountRelationship(ctx context.Context, relationshipID uuid.UUID, accessLevel domain.SharingAccessLevel) (bool, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		if access, hasAccess := domain.AccessContextFromContext(ctx); hasAccess {
			orgID = access.OrgID
		}
	}
	if orgID == uuid.Nil {
		return false, nil
	}

	var parentID, childID uuid.UUID
	err := r.db.QueryRow(ctx, `
		SELECT parent_account_id, child_account_id
		FROM account_relationships
		WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL
	`, relationshipID, orgID).Scan(&parentID, &childID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	access, ok := domain.AccessContextFromContext(ctx)
	if !ok || access.CanAccessAllRecords(domain.ACLModuleAccounts, accessLevel) {
		return true, nil
	}

	canAccessParent, err := r.CanAccessRecord(ctx, domain.ACLModuleAccounts, parentID, accessLevel)
	if err != nil || !canAccessParent {
		return canAccessParent, err
	}
	return r.CanAccessRecord(ctx, domain.ACLModuleAccounts, childID, accessLevel)
}

func recordAccessTarget(module domain.ACLModule) (table string, ownerExprs []string, ok bool) {
	switch module {
	case domain.ACLModuleAccounts:
		return "accounts", []string{"owner_id"}, true
	case domain.ACLModuleContacts:
		return "contacts", []string{"owner_id"}, true
	case domain.ACLModuleDeals:
		return "deals", []string{"owner_id"}, true
	case domain.ACLModuleLeads:
		return "leads", []string{"owner_id"}, true
	case domain.ACLModuleTickets:
		return "tickets", []string{"assignee_id", "submitted_by_user_id"}, true
	default:
		return "", nil, false
	}
}

func (r *AccessRepo) recordExists(ctx context.Context, table string, id uuid.UUID, module domain.ACLModule, accessLevel domain.SharingAccessLevel, ownerExprs ...string) (bool, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		if access, hasAccess := domain.AccessContextFromContext(ctx); hasAccess {
			orgID = access.OrgID
		}
	}
	if orgID == uuid.Nil {
		return false, nil
	}

	q := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL`, table)
	args := []any{id, orgID}
	appendAccessVisibilitySQL(ctx, &q, &args, module, accessLevel, ownerExprs...)
	q += ")"

	var exists bool
	err := r.db.QueryRow(ctx, q, args...).Scan(&exists)
	return exists, err
}

func (r *AccessRepo) ListRoles(ctx context.Context, orgID uuid.UUID) ([]*domain.ACLRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, org_id, name, description, system_key, parent_id, created_at, updated_at
		FROM crm_roles
		WHERE org_id = $1
		ORDER BY COALESCE(system_key, name), name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.ACLRole{}
	for rows.Next() {
		role, err := scanACLRole(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, role)
	}
	return out, rows.Err()
}

func (r *AccessRepo) CreateRole(ctx context.Context, role *domain.ACLRole) (*domain.ACLRole, error) {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	if role.OrgID == uuid.Nil {
		if orgID, ok := domain.OrgIDFromContext(ctx); ok {
			role.OrgID = orgID
		}
	}
	if err := r.ensureRoleParentSafe(ctx, role.OrgID, role.ID, role.ParentID); err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO crm_roles (id, org_id, name, description, parent_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, org_id, name, description, system_key, parent_id, created_at, updated_at
	`, role.ID, role.OrgID, role.Name, role.Description, role.ParentID)
	created, err := scanACLRole(row)
	if err != nil {
		return nil, err
	}
	return created, r.rebuildRoleClosure(ctx, role.OrgID)
}

func (r *AccessRepo) UpdateRole(ctx context.Context, id uuid.UUID, patch domain.ACLRolePatch) (*domain.ACLRole, error) {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return nil, err
	}
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1
	add := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	if patch.Name != nil {
		add("name", *patch.Name)
	}
	if patch.Description != nil {
		add("description", *patch.Description)
	}
	if patch.ParentID != nil {
		if err := r.ensureRoleParentSafe(ctx, orgID, id, patch.ParentID); err != nil {
			return nil, err
		}
		add("parent_id", *patch.ParentID)
	}
	args = append(args, id, orgID)
	query := fmt.Sprintf(`
		UPDATE crm_roles
		SET %s
		WHERE id = $%d AND org_id = $%d AND system_key IS NULL
		RETURNING id, org_id, name, description, system_key, parent_id, created_at, updated_at
	`, strings.Join(sets, ", "), i, i+1)
	role, err := scanACLRole(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, err
	}
	return role, r.rebuildRoleClosure(ctx, orgID)
}

func (r *AccessRepo) DeleteRole(ctx context.Context, id uuid.UUID) error {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `DELETE FROM crm_roles WHERE id = $1 AND org_id = $2 AND system_key IS NULL`, id, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM crm_sharing_grants WHERE org_id = $1 AND grantee_type = 'role' AND grantee_id = $2`, orgID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM crm_sharing_rules
		WHERE org_id = $1
		  AND (
			(source_type IN ('role', 'role_subordinates') AND source_id = $2)
			OR (target_type IN ('role', 'role_subordinates') AND target_id = $2)
		  )
	`, orgID, id); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return r.rebuildRoleClosure(ctx, orgID)
}

func (r *AccessRepo) MoveRole(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) (*domain.ACLRole, error) {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return nil, err
	}
	if err := r.ensureRoleParentSafe(ctx, orgID, id, parentID); err != nil {
		return nil, err
	}
	role, err := scanACLRole(r.db.QueryRow(ctx, `
		UPDATE crm_roles
		SET parent_id = $1, updated_at = NOW()
		WHERE id = $2 AND org_id = $3 AND system_key IS NULL
		RETURNING id, org_id, name, description, system_key, parent_id, created_at, updated_at
	`, parentID, id, orgID))
	if err != nil {
		return nil, err
	}
	return role, r.rebuildRoleClosure(ctx, orgID)
}

func (r *AccessRepo) ListProfiles(ctx context.Context, orgID uuid.UUID) ([]*domain.ACLProfile, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, org_id, name, description, system_key
		FROM crm_profiles
		WHERE org_id = $1
		ORDER BY COALESCE(system_key, name), name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.ACLProfile{}
	for rows.Next() {
		profile, err := scanACLProfile(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, profile)
	}
	return out, rows.Err()
}

func (r *AccessRepo) CreateProfile(ctx context.Context, profile *domain.ACLProfile) (*domain.ACLProfile, error) {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	if profile.OrgID == uuid.Nil {
		if orgID, ok := domain.OrgIDFromContext(ctx); ok {
			profile.OrgID = orgID
		}
	}
	return scanACLProfile(r.db.QueryRow(ctx, `
		INSERT INTO crm_profiles (id, org_id, name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, org_id, name, description, system_key
	`, profile.ID, profile.OrgID, profile.Name, profile.Description))
}

func (r *AccessRepo) UpdateProfile(ctx context.Context, id uuid.UUID, patch domain.ACLProfilePatch) (*domain.ACLProfile, error) {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return nil, err
	}
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1
	if patch.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", i))
		args = append(args, *patch.Name)
		i++
	}
	if patch.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", i))
		args = append(args, *patch.Description)
		i++
	}
	args = append(args, id, orgID)
	return scanACLProfile(r.db.QueryRow(ctx, fmt.Sprintf(`
		UPDATE crm_profiles
		SET %s
		WHERE id = $%d AND org_id = $%d AND system_key IS NULL
		RETURNING id, org_id, name, description, system_key
	`, strings.Join(sets, ", "), i, i+1), args...))
}

func (r *AccessRepo) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return err
	}
	result, err := r.db.Exec(ctx, `DELETE FROM crm_profiles WHERE id = $1 AND org_id = $2 AND system_key IS NULL`, id, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AccessRepo) ListProfilePermissions(ctx context.Context, profileID uuid.UUID) ([]domain.ACLProfilePermission, []domain.ACLProfileFieldPermission, error) {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := r.ensureProfileInOrg(ctx, orgID, profileID); err != nil {
		return nil, nil, err
	}
	perms := []domain.ACLProfilePermission{}
	rows, err := r.db.Query(ctx, `
		SELECT profile_id, module, action, allowed
		FROM crm_profile_permissions
		WHERE org_id = $1 AND profile_id = $2
		ORDER BY module, action
	`, orgID, profileID)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var p domain.ACLProfilePermission
		if err := rows.Scan(&p.ProfileID, &p.Module, &p.Action, &p.Allowed); err != nil {
			rows.Close()
			return nil, nil, err
		}
		perms = append(perms, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	fields := []domain.ACLProfileFieldPermission{}
	rows, err = r.db.Query(ctx, `
		SELECT profile_id, module, field_name, can_write
		FROM crm_profile_field_permissions
		WHERE org_id = $1 AND profile_id = $2
		ORDER BY module, field_name
	`, orgID, profileID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var f domain.ACLProfileFieldPermission
		if err := rows.Scan(&f.ProfileID, &f.Module, &f.FieldName, &f.CanWrite); err != nil {
			return nil, nil, err
		}
		fields = append(fields, f)
	}
	return perms, fields, rows.Err()
}

func (r *AccessRepo) ReplaceProfilePermissions(ctx context.Context, profileID uuid.UUID, permissions []domain.ACLProfilePermission, fieldPermissions []domain.ACLProfileFieldPermission) error {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, `SELECT org_id FROM crm_profiles WHERE id = $1 AND org_id = $2`, profileID, orgID).Scan(&orgID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM crm_profile_permissions WHERE org_id = $1 AND profile_id = $2`, orgID, profileID); err != nil {
		return err
	}
	for _, p := range permissions {
		_, err := tx.Exec(ctx, `
			INSERT INTO crm_profile_permissions (org_id, profile_id, module, action, allowed)
			VALUES ($1, $2, $3, $4, $5)
		`, orgID, profileID, p.Module, p.Action, p.Allowed)
		if err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM crm_profile_field_permissions WHERE org_id = $1 AND profile_id = $2`, orgID, profileID); err != nil {
		return err
	}
	for _, f := range fieldPermissions {
		_, err := tx.Exec(ctx, `
			INSERT INTO crm_profile_field_permissions (org_id, profile_id, module, field_name, can_write)
			VALUES ($1, $2, $3, $4, $5)
		`, orgID, profileID, f.Module, f.FieldName, f.CanWrite)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *AccessRepo) ListGroups(ctx context.Context, orgID uuid.UUID) ([]*domain.ACLGroup, error) {
	rows, err := r.db.Query(ctx, `
		SELECT g.id, g.org_id, g.name, g.description,
			COALESCE(array_agg(gm.user_id) FILTER (WHERE gm.user_id IS NOT NULL), ARRAY[]::uuid[])
		FROM crm_groups g
		LEFT JOIN crm_group_members gm ON gm.group_id = g.id
		WHERE g.org_id = $1
		GROUP BY g.id
		ORDER BY g.name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.ACLGroup{}
	for rows.Next() {
		group, err := scanACLGroup(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, group)
	}
	return out, rows.Err()
}

func (r *AccessRepo) CreateGroup(ctx context.Context, group *domain.ACLGroup) (*domain.ACLGroup, error) {
	if group.ID == uuid.Nil {
		group.ID = uuid.New()
	}
	if group.OrgID == uuid.Nil {
		if orgID, ok := domain.OrgIDFromContext(ctx); ok {
			group.OrgID = orgID
		}
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	created, err := scanACLGroup(tx.QueryRow(ctx, `
		INSERT INTO crm_groups (id, org_id, name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, org_id, name, description, ARRAY[]::uuid[]
	`, group.ID, group.OrgID, group.Name, group.Description))
	if err != nil {
		return nil, err
	}
	if len(group.UserIDs) > 0 {
		userIDs, err := replaceGroupMembersTx(ctx, tx, group.OrgID, created.ID, group.UserIDs)
		if err != nil {
			return nil, err
		}
		created.UserIDs = userIDs
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

func (r *AccessRepo) UpdateGroup(ctx context.Context, id uuid.UUID, patch domain.ACLGroupPatch) (*domain.ACLGroup, error) {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return nil, err
	}
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1
	if patch.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", i))
		args = append(args, *patch.Name)
		i++
	}
	if patch.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", i))
		args = append(args, *patch.Description)
		i++
	}
	args = append(args, id, orgID)
	return scanACLGroup(r.db.QueryRow(ctx, fmt.Sprintf(`
		WITH updated AS (
			UPDATE crm_groups SET %s WHERE id = $%d AND org_id = $%d
			RETURNING id, org_id, name, description
		)
		SELECT updated.id, updated.org_id, updated.name, updated.description,
			COALESCE(array_agg(gm.user_id) FILTER (WHERE gm.user_id IS NOT NULL), ARRAY[]::uuid[])
		FROM updated
		LEFT JOIN crm_group_members gm ON gm.group_id = updated.id
		GROUP BY updated.id, updated.org_id, updated.name, updated.description
	`, strings.Join(sets, ", "), i, i+1), args...))
}

func (r *AccessRepo) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `DELETE FROM crm_groups WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM crm_sharing_grants WHERE org_id = $1 AND grantee_type = 'group' AND grantee_id = $2`, orgID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM crm_sharing_rules
		WHERE org_id = $1
		  AND (
			(source_type = 'group' AND source_id = $2)
			OR (target_type = 'group' AND target_id = $2)
		  )
	`, orgID, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *AccessRepo) ReplaceGroupMembers(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error {
	orgID, err := accessOrgID(ctx)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := replaceGroupMembersTx(ctx, tx, orgID, groupID, userIDs); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func replaceGroupMembersTx(ctx context.Context, tx pgx.Tx, orgID, groupID uuid.UUID, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	if err := tx.QueryRow(ctx, `SELECT org_id FROM crm_groups WHERE id = $1 AND org_id = $2`, groupID, orgID).Scan(&orgID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM crm_group_members WHERE org_id = $1 AND group_id = $2`, orgID, groupID); err != nil {
		return nil, err
	}
	seen := map[uuid.UUID]struct{}{}
	applied := []uuid.UUID{}
	for _, userID := range userIDs {
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		result, err := tx.Exec(ctx, `
			INSERT INTO crm_group_members (org_id, group_id, user_id)
			SELECT $1, $2, u.id
			FROM users u
			WHERE u.id = $3 AND u.org_id = $1 AND u.deleted_at IS NULL
		`, orgID, groupID, userID)
		if err != nil {
			return nil, err
		}
		if result.RowsAffected() == 0 {
			return nil, domain.ErrNotFound
		}
		applied = append(applied, userID)
	}
	return applied, nil
}

func (r *AccessRepo) GetSharingRules(ctx context.Context, orgID uuid.UUID) (*domain.ACLSharingRules, error) {
	rows, err := r.db.Query(ctx, `
		SELECT module, mode
		FROM crm_sharing_defaults
		WHERE org_id = $1
		ORDER BY module
	`, orgID)
	if err != nil {
		return nil, err
	}
	rulesByModule := map[domain.ACLModule]*domain.ACLSharingModuleRule{}
	for rows.Next() {
		var module domain.ACLModule
		var mode domain.SharingDefaultMode
		if err := rows.Scan(&module, &mode); err != nil {
			rows.Close()
			return nil, err
		}
		rulesByModule[module] = &domain.ACLSharingModuleRule{Module: module, Mode: mode, AdvancedRules: []domain.ACLSharingRule{}}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = r.db.Query(ctx, `
		SELECT id, org_id, module, source_type, source_id, target_type, target_id, access_level, created_at, updated_at
		FROM crm_sharing_rules
		WHERE org_id = $1
		ORDER BY module, source_type, source_id, target_type, target_id, access_level
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var advancedRule domain.ACLSharingRule
		if err := rows.Scan(
			&advancedRule.ID,
			&advancedRule.OrgID,
			&advancedRule.Module,
			&advancedRule.SourceType,
			&advancedRule.SourceID,
			&advancedRule.TargetType,
			&advancedRule.TargetID,
			&advancedRule.AccessLevel,
			&advancedRule.CreatedAt,
			&advancedRule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rule := rulesByModule[advancedRule.Module]
		if rule == nil {
			rule = &domain.ACLSharingModuleRule{Module: advancedRule.Module, Mode: domain.SharingDefaultPrivate, AdvancedRules: []domain.ACLSharingRule{}}
			rulesByModule[advancedRule.Module] = rule
		}
		rule.AdvancedRules = append(rule.AdvancedRules, advancedRule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := &domain.ACLSharingRules{Rules: []domain.ACLSharingModuleRule{}}
	for _, module := range domain.SharingModules {
		if rule := rulesByModule[module]; rule != nil {
			if rule.AdvancedRules == nil {
				rule.AdvancedRules = []domain.ACLSharingRule{}
			}
			out.Rules = append(out.Rules, *rule)
		}
	}
	return out, nil
}

func (r *AccessRepo) ReplaceSharingRules(ctx context.Context, orgID uuid.UUID, rules *domain.ACLSharingRules) (*domain.ACLSharingRules, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	existingRules, err := existingSharingRules(ctx, tx, orgID)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM crm_sharing_grants WHERE org_id = $1`, orgID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM crm_sharing_rules WHERE org_id = $1`, orgID); err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	seenIDs := map[uuid.UUID]struct{}{}
	for _, rule := range rules.Rules {
		if !validSharingModule(rule.Module) {
			return nil, fmt.Errorf("%w: invalid sharing module", domain.ErrValidation)
		}
		if !validSharingDefaultMode(rule.Mode) {
			return nil, fmt.Errorf("%w: invalid sharing default mode", domain.ErrValidation)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO crm_sharing_defaults (org_id, module, mode, updated_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (org_id, module) DO UPDATE
			SET mode = EXCLUDED.mode, updated_at = NOW()
		`, orgID, rule.Module, rule.Mode)
		if err != nil {
			return nil, err
		}
		for _, advancedRule := range sharingRulesForReplace(rule) {
			advancedRule.Module = rule.Module
			advancedRule.OrgID = orgID
			key := sharingRuleKey(advancedRule)
			if _, exists := seen[key]; exists {
				continue
			}
			if advancedRule.ID == uuid.Nil {
				if existingID, ok := existingRules.idsByKey[key]; ok {
					advancedRule.ID = existingID
				} else {
					advancedRule.ID = uuid.New()
				}
			} else {
				if _, ok := existingRules.ids[advancedRule.ID]; !ok {
					return nil, fmt.Errorf("%w: sharing rule id not found in org", domain.ErrValidation)
				}
			}
			if _, ok := seenIDs[advancedRule.ID]; ok {
				return nil, fmt.Errorf("%w: duplicate sharing rule id", domain.ErrValidation)
			}
			seenIDs[advancedRule.ID] = struct{}{}
			if err := validateSharingRule(ctx, tx, orgID, advancedRule); err != nil {
				return nil, err
			}
			seen[key] = struct{}{}
			_, err := tx.Exec(ctx, `
				INSERT INTO crm_sharing_rules
					(id, org_id, module, source_type, source_id, target_type, target_id, access_level, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
			`, advancedRule.ID, orgID, advancedRule.Module, advancedRule.SourceType, advancedRule.SourceID, advancedRule.TargetType, advancedRule.TargetID, advancedRule.AccessLevel)
			if err != nil {
				return nil, err
			}
			if advancedRule.SourceType == domain.SharingPrincipalAll &&
				(advancedRule.TargetType == domain.SharingPrincipalRole || advancedRule.TargetType == domain.SharingPrincipalGroup) {
				_, err := tx.Exec(ctx, `
					INSERT INTO crm_sharing_grants (org_id, module, grantee_type, grantee_id, access_level)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT DO NOTHING
				`, orgID, rule.Module, advancedRule.TargetType, advancedRule.TargetID, advancedRule.AccessLevel)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetSharingRules(ctx, orgID)
}

func sharingRulesForReplace(rule domain.ACLSharingModuleRule) []domain.ACLSharingRule {
	if len(rule.Grants) == 0 {
		return rule.AdvancedRules
	}

	rules := make([]domain.ACLSharingRule, 0, len(rule.AdvancedRules)+len(rule.Grants))
	rules = append(rules, rule.AdvancedRules...)
	for _, grant := range rule.Grants {
		rules = append(rules, domain.ACLSharingRule{
			OrgID:       grant.OrgID,
			Module:      rule.Module,
			SourceType:  domain.SharingPrincipalAll,
			SourceID:    nil,
			TargetType:  domain.SharingPrincipalType(grant.GranteeType),
			TargetID:    grant.GranteeID,
			AccessLevel: grant.AccessLevel,
		})
	}

	return rules
}

type existingSharingRuleLookup struct {
	ids      map[uuid.UUID]struct{}
	idsByKey map[string]uuid.UUID
}

func existingSharingRules(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) (existingSharingRuleLookup, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, module, source_type, source_id, target_type, target_id, access_level
		FROM crm_sharing_rules
		WHERE org_id = $1
	`, orgID)
	if err != nil {
		return existingSharingRuleLookup{}, err
	}
	defer rows.Close()

	lookup := existingSharingRuleLookup{
		ids:      map[uuid.UUID]struct{}{},
		idsByKey: map[string]uuid.UUID{},
	}
	for rows.Next() {
		var rule domain.ACLSharingRule
		if err := rows.Scan(
			&rule.ID,
			&rule.Module,
			&rule.SourceType,
			&rule.SourceID,
			&rule.TargetType,
			&rule.TargetID,
			&rule.AccessLevel,
		); err != nil {
			return existingSharingRuleLookup{}, err
		}
		lookup.ids[rule.ID] = struct{}{}
		lookup.idsByKey[sharingRuleKey(rule)] = rule.ID
	}
	return lookup, rows.Err()
}

type sharingRowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func validSharingModule(module domain.ACLModule) bool {
	for _, candidate := range domain.SharingModules {
		if candidate == module {
			return true
		}
	}
	return false
}

func validSharingDefaultMode(mode domain.SharingDefaultMode) bool {
	switch mode {
	case domain.SharingDefaultPrivate, domain.SharingDefaultPublicRO, domain.SharingDefaultPublicRW:
		return true
	default:
		return false
	}
}

func validSharingAccessLevel(level domain.SharingAccessLevel) bool {
	switch level {
	case domain.SharingAccessRead, domain.SharingAccessWrite:
		return true
	default:
		return false
	}
}

func sharingRuleKey(rule domain.ACLSharingRule) string {
	sourceID := ""
	if rule.SourceID != nil {
		sourceID = rule.SourceID.String()
	}
	return strings.Join([]string{
		string(rule.Module),
		string(rule.SourceType),
		sourceID,
		string(rule.TargetType),
		rule.TargetID.String(),
		string(rule.AccessLevel),
	}, "|")
}

func validateSharingRule(ctx context.Context, q sharingRowQuerier, orgID uuid.UUID, rule domain.ACLSharingRule) error {
	if !validSharingModule(rule.Module) {
		return fmt.Errorf("%w: invalid sharing module", domain.ErrValidation)
	}
	if !validSharingAccessLevel(rule.AccessLevel) {
		return fmt.Errorf("%w: invalid sharing access level", domain.ErrValidation)
	}
	if rule.SourceType == domain.SharingPrincipalAll {
		if rule.SourceID != nil {
			return fmt.Errorf("%w: all-source sharing rules cannot include source_id", domain.ErrValidation)
		}
	} else {
		if rule.SourceID == nil || *rule.SourceID == uuid.Nil {
			return fmt.Errorf("%w: source_id is required", domain.ErrValidation)
		}
		if err := ensureSharingPrincipalInOrg(ctx, q, orgID, rule.SourceType, *rule.SourceID, true); err != nil {
			return err
		}
	}
	if rule.TargetID == uuid.Nil {
		return fmt.Errorf("%w: target_id is required", domain.ErrValidation)
	}
	return ensureSharingPrincipalInOrg(ctx, q, orgID, rule.TargetType, rule.TargetID, false)
}

func ensureSharingPrincipalInOrg(ctx context.Context, q sharingRowQuerier, orgID uuid.UUID, principalType domain.SharingPrincipalType, id uuid.UUID, allowAll bool) error {
	var query string
	switch principalType {
	case domain.SharingPrincipalAll:
		if allowAll {
			return nil
		}
		return fmt.Errorf("%w: all is not a valid target", domain.ErrValidation)
	case domain.SharingPrincipalUser:
		query = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL)`
	case domain.SharingPrincipalRole, domain.SharingPrincipalRoleSubordinates:
		query = `SELECT EXISTS(SELECT 1 FROM crm_roles WHERE id = $1 AND org_id = $2)`
	case domain.SharingPrincipalGroup:
		query = `SELECT EXISTS(SELECT 1 FROM crm_groups WHERE id = $1 AND org_id = $2)`
	default:
		return fmt.Errorf("%w: invalid sharing principal type", domain.ErrValidation)
	}
	var exists bool
	if err := q.QueryRow(ctx, query, id, orgID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}
	return nil
}

func scanACLRole(row pgx.Row) (*domain.ACLRole, error) {
	var role domain.ACLRole
	if err := row.Scan(&role.ID, &role.OrgID, &role.Name, &role.Description, &role.SystemKey, &role.ParentID, &role.CreatedAt, &role.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func scanACLProfile(row pgx.Row) (*domain.ACLProfile, error) {
	var profile domain.ACLProfile
	if err := row.Scan(&profile.ID, &profile.OrgID, &profile.Name, &profile.Description, &profile.SystemKey); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func scanACLGroup(row pgx.Row) (*domain.ACLGroup, error) {
	var group domain.ACLGroup
	if err := row.Scan(&group.ID, &group.OrgID, &group.Name, &group.Description, &group.UserIDs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &group, nil
}

func accessOrgID(ctx context.Context) (uuid.UUID, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok || orgID == uuid.Nil {
		return uuid.Nil, domain.ErrNotFound
	}
	return orgID, nil
}

func (r *AccessRepo) ensureProfileInOrg(ctx context.Context, orgID, profileID uuid.UUID) error {
	var exists bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM crm_profiles WHERE id = $1 AND org_id = $2)
	`, profileID, orgID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AccessRepo) ensureRoleParentSafe(ctx context.Context, orgID, roleID uuid.UUID, parentID *uuid.UUID) error {
	if parentID == nil {
		return nil
	}
	if *parentID == roleID {
		return fmt.Errorf("%w: role cannot be its own parent", domain.ErrValidation)
	}
	var exists bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM crm_roles WHERE id = $1 AND org_id = $2)
	`, *parentID, orgID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}
	if roleID == uuid.Nil {
		return nil
	}
	var cycle bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM crm_role_closure
			WHERE ancestor_id = $1 AND descendant_id = $2
		)
	`, roleID, *parentID).Scan(&cycle); err != nil {
		return err
	}
	if cycle {
		return fmt.Errorf("%w: role hierarchy cycle", domain.ErrValidation)
	}
	return nil
}

func (r *AccessRepo) rebuildRoleClosure(ctx context.Context, orgID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM crm_role_closure WHERE org_id = $1`, orgID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		WITH RECURSIVE walk AS (
			SELECT org_id, id AS ancestor_id, id AS descendant_id, 0 AS depth
			FROM crm_roles
			WHERE org_id = $1
			UNION ALL
			SELECT p.org_id, w.ancestor_id, c.id AS descendant_id, w.depth + 1
			FROM walk w
			JOIN crm_roles p ON p.id = w.descendant_id
			JOIN crm_roles c ON c.parent_id = p.id
			WHERE c.org_id = $1
		)
		INSERT INTO crm_role_closure (org_id, ancestor_id, descendant_id, depth)
		SELECT org_id, ancestor_id, descendant_id, depth
		FROM walk
	`, orgID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
