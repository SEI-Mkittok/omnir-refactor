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
	if access.RoleID != nil {
		roleGrantIDs = append(roleGrantIDs, *access.RoleID)
	}
	rows, err = r.db.Query(ctx, `
		SELECT module, access_level
		FROM crm_sharing_grants
		WHERE org_id = $1
		  AND (
			(grantee_type = 'role' AND grantee_id = ANY($2))
			OR (grantee_type = 'group' AND grantee_id = ANY($3))
		  )
	`, orgID, roleGrantIDs, access.GroupIDs)
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
	table, ownerPredicate, ok := recordAccessTarget(module)
	if !ok {
		return true, nil
	}

	access, ok := domain.AccessContextFromContext(ctx)
	if !ok || access.CanAccessAllRecords(module, accessLevel) {
		return r.recordExists(ctx, table, id, "")
	}

	return r.recordExists(ctx, table, id, ownerPredicate)
}

func recordAccessTarget(module domain.ACLModule) (table string, ownerPredicate string, ok bool) {
	switch module {
	case domain.ACLModuleAccounts:
		return "accounts", "owner_id = %s", true
	case domain.ACLModuleContacts:
		return "contacts", "owner_id = %s", true
	case domain.ACLModuleDeals:
		return "deals", "owner_id = %s", true
	case domain.ACLModuleLeads:
		return "leads", "owner_id = %s", true
	case domain.ACLModuleTickets:
		return "tickets", "(assignee_id = %s OR submitted_by_user_id = %s)", true
	default:
		return "", "", false
	}
}

func (r *AccessRepo) recordExists(ctx context.Context, table string, id uuid.UUID, ownerPredicate string) (bool, error) {
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
	if ownerPredicate != "" {
		placeholder := "$3"
		if strings.Count(ownerPredicate, "%s") > 1 {
			q += " AND " + fmt.Sprintf(ownerPredicate, placeholder, placeholder)
		} else {
			q += " AND " + fmt.Sprintf(ownerPredicate, placeholder)
		}
		access, ok := domain.AccessContextFromContext(ctx)
		if !ok {
			return false, nil
		}
		args = append(args, access.UserID)
	}
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
	var out []*domain.ACLRole
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
		WHERE id = $2 AND org_id = $3
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
	var out []*domain.ACLProfile
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
	var out []*domain.ACLGroup
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
	created, err := scanACLGroup(r.db.QueryRow(ctx, `
		INSERT INTO crm_groups (id, org_id, name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, org_id, name, description, ARRAY[]::uuid[]
	`, group.ID, group.OrgID, group.Name, group.Description))
	if err != nil {
		return nil, err
	}
	if len(group.UserIDs) > 0 {
		if err := r.ReplaceGroupMembers(ctx, created.ID, group.UserIDs); err != nil {
			return nil, err
		}
		created.UserIDs = group.UserIDs
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
	if err := tx.QueryRow(ctx, `SELECT org_id FROM crm_groups WHERE id = $1 AND org_id = $2`, groupID, orgID).Scan(&orgID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM crm_group_members WHERE org_id = $1 AND group_id = $2`, orgID, groupID); err != nil {
		return err
	}
	seen := map[uuid.UUID]struct{}{}
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
			return err
		}
		if result.RowsAffected() == 0 {
			return domain.ErrNotFound
		}
	}
	return tx.Commit(ctx)
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
		rulesByModule[module] = &domain.ACLSharingModuleRule{Module: module, Mode: mode, Grants: []domain.ACLSharingGrant{}}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = r.db.Query(ctx, `
		SELECT id, org_id, module, grantee_type, grantee_id, access_level
		FROM crm_sharing_grants
		WHERE org_id = $1
		ORDER BY module, grantee_type, grantee_id
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var grant domain.ACLSharingGrant
		if err := rows.Scan(&grant.ID, &grant.OrgID, &grant.Module, &grant.GranteeType, &grant.GranteeID, &grant.AccessLevel); err != nil {
			return nil, err
		}
		rule := rulesByModule[grant.Module]
		if rule == nil {
			rule = &domain.ACLSharingModuleRule{Module: grant.Module, Mode: domain.SharingDefaultPrivate, Grants: []domain.ACLSharingGrant{}}
			rulesByModule[grant.Module] = rule
		}
		rule.Grants = append(rule.Grants, grant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := &domain.ACLSharingRules{Rules: []domain.ACLSharingModuleRule{}}
	for _, module := range domain.SharingModules {
		if rule := rulesByModule[module]; rule != nil {
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
	if _, err := tx.Exec(ctx, `DELETE FROM crm_sharing_grants WHERE org_id = $1`, orgID); err != nil {
		return nil, err
	}
	for _, rule := range rules.Rules {
		_, err := tx.Exec(ctx, `
			INSERT INTO crm_sharing_defaults (org_id, module, mode, updated_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (org_id, module) DO UPDATE
			SET mode = EXCLUDED.mode, updated_at = NOW()
		`, orgID, rule.Module, rule.Mode)
		if err != nil {
			return nil, err
		}
		for _, grant := range rule.Grants {
			_, err := tx.Exec(ctx, `
				INSERT INTO crm_sharing_grants (org_id, module, grantee_type, grantee_id, access_level)
				VALUES ($1, $2, $3, $4, $5)
			`, orgID, rule.Module, grant.GranteeType, grant.GranteeID, grant.AccessLevel)
			if err != nil {
				return nil, err
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetSharingRules(ctx, orgID)
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
