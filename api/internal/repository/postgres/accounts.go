package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type AccountRepo struct {
	db *pgxpool.Pool
}

func NewAccountRepo(db *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{db: db}
}

const accountCols = `
	id, org_id, name, domain, industry, size,
	owner_id, tags, custom_fields, created_at, updated_at, deleted_at
`

const accountRelationshipCols = `
	id, org_id, parent_account_id, child_account_id, relationship_type,
	ownership_percent, effective_from, effective_to, created_at, updated_at, deleted_at, deleted_by
`

func scanAccount(row pgx.Row) (*domain.Account, error) {
	var a domain.Account
	err := row.Scan(
		&a.ID, &a.OrgID, &a.Name, &a.Domain, &a.Industry, &a.Size,
		&a.OwnerID, &a.Tags, &a.CustomFields, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func scanAccountRelationship(row pgx.Row) (*domain.AccountRelationship, error) {
	var rel domain.AccountRelationship
	err := row.Scan(
		&rel.ID, &rel.OrgID, &rel.ParentAccountID, &rel.ChildAccountID, &rel.RelationshipType,
		&rel.OwnershipPercent, &rel.EffectiveFrom, &rel.EffectiveTo, &rel.CreatedAt, &rel.UpdatedAt, &rel.DeletedAt, &rel.DeletedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &rel, nil
}

func (r *AccountRepo) Create(ctx context.Context, a *domain.Account) (*domain.Account, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		a.OrgID = orgID
	}
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	if a.Tags == nil {
		a.Tags = []string{}
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO accounts
			(id, org_id, name, domain, industry, size,
			 owner_id, tags, custom_fields, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+accountCols,
		a.ID, a.OrgID, a.Name, a.Domain, a.Industry, a.Size,
		a.OwnerID, a.Tags, a.CustomFields, a.CreatedAt, a.UpdatedAt,
	)
	return scanAccount(row)
}

func (r *AccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	q := `SELECT ` + accountCols + ` FROM accounts WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}
	i := 2

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
		i++
	}
	if acl, ok := domain.AccessContextFromContext(ctx); ok && !acl.CanAccessAllRecords(domain.ACLModuleAccounts, domain.SharingAccessRead) {
		q += fmt.Sprintf(` AND owner_id=$%d`, i)
		args = append(args, acl.UserID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	return scanAccount(row)
}

func (r *AccountRepo) GetByName(ctx context.Context, name string) (*domain.Account, error) {
	q := `SELECT ` + accountCols + ` FROM accounts WHERE name=$1 AND deleted_at IS NULL`
	args := []any{name}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	return scanAccount(row)
}

func (r *AccountRepo) Update(ctx context.Context, id uuid.UUID, patch domain.AccountPatch) (*domain.Account, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.Name != nil {
		addArg("name", *patch.Name)
	}
	if patch.Domain != nil {
		addArg("domain", *patch.Domain)
	}
	if patch.Industry != nil {
		addArg("industry", *patch.Industry)
	}
	if patch.Size != nil {
		addArg("size", *patch.Size)
	}
	if patch.OwnerID != nil {
		addArg("owner_id", *patch.OwnerID)
	}
	if patch.Tags != nil {
		addArg("tags", patch.Tags)
	}
	if patch.CustomFields != nil {
		addArg("custom_fields", patch.CustomFields)
	}

	whereClause := fmt.Sprintf(`id=$%d AND deleted_at IS NULL`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
		i++
	}
	if acl, ok := domain.AccessContextFromContext(ctx); ok && !acl.CanAccessAllRecords(domain.ACLModuleAccounts, domain.SharingAccessWrite) {
		whereClause += fmt.Sprintf(` AND owner_id=$%d`, i)
		args = append(args, acl.UserID)
	}

	query := fmt.Sprintf(
		`UPDATE accounts SET %s WHERE %s RETURNING %s`,
		strings.Join(sets, ", "), whereClause, accountCols,
	)
	row := r.db.QueryRow(ctx, query, args...)
	return scanAccount(row)
}

func (r *AccountRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE accounts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}
	i := 2

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
		i++
	}
	if acl, ok := domain.AccessContextFromContext(ctx); ok && !acl.CanAccessAllRecords(domain.ACLModuleAccounts, domain.SharingAccessWrite) {
		q += fmt.Sprintf(` AND owner_id=$%d`, i)
		args = append(args, acl.UserID)
	}

	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AccountRepo) List(ctx context.Context, f domain.AccountFilter) ([]*domain.Account, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"deleted_at IS NULL"}
	args := []any{}
	i := 1

	addWhere := func(expr string, val any) {
		where = append(where, fmt.Sprintf("%s = $%d", expr, i))
		args = append(args, val)
		i++
	}

	// Always scope by org_id: prefer context, fall back to filter field.
	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if !hasCtxOrg {
		orgID = f.OrgID
	}
	if orgID != uuid.Nil {
		addWhere("org_id", orgID)
	}
	addAccessVisibilityWhere(ctx, &where, &args, &i, domain.ACLModuleAccounts, domain.SharingAccessRead, "owner_id = %s")

	if f.OwnerID != nil {
		addWhere("owner_id", *f.OwnerID)
	}
	if f.Industry != nil {
		addWhere("industry", *f.Industry)
	}
	if f.Size != nil {
		addWhere("size", *f.Size)
	}
	if f.Q != "" {
		where = append(where, fmt.Sprintf(
			`to_tsvector('english', name || ' ' || coalesce(domain, '') || ' ' || coalesce(industry, '')) @@ plainto_tsquery('english', $%d)`, i,
		))
		args = append(args, f.Q)
		i++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM accounts WHERE `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sortCol := "created_at"
	allowedSorts := map[string]bool{
		"created_at": true, "updated_at": true, "name": true,
	}
	if allowedSorts[f.Sort] {
		sortCol = f.Sort
	}
	order := "DESC"
	if strings.ToUpper(f.Order) == "ASC" {
		order = "ASC"
	}

	rows, err := r.db.Query(ctx,
		fmt.Sprintf(
			`SELECT %s FROM accounts WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
			accountCols, whereClause, sortCol, order, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		var a domain.Account
		if err := rows.Scan(
			&a.ID, &a.OrgID, &a.Name, &a.Domain, &a.Industry, &a.Size,
			&a.OwnerID, &a.Tags, &a.CustomFields, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		accounts = append(accounts, &a)
	}
	return accounts, total, rows.Err()
}

func (r *AccountRepo) ListLinkedEntities(ctx context.Context, id uuid.UUID, f domain.LinkedEntityFilter) ([]domain.LinkedEntity, int, error) {
	if f.Limit <= 0 {
		f.Limit = 25
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	args := []any{id}
	i := 2
	where := []string{}

	if f.Type != nil && *f.Type != "" {
		where = append(where, fmt.Sprintf("entity_type = $%d", i))
		args = append(args, *f.Type)
		i++
	}
	if f.Role != nil && *f.Role != "" {
		where = append(where, fmt.Sprintf("entity_role = $%d", i))
		args = append(args, *f.Role)
		i++
	}
	if f.Since != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", i))
		args = append(args, *f.Since)
		i++
	}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		where = append(where, fmt.Sprintf("org_id = $%d", i))
		args = append(args, orgID)
		i++
	}

	filterClause := ""
	if len(where) > 0 {
		filterClause = " WHERE " + strings.Join(where, " AND ")
	}

	query := `
		WITH linked AS (
			SELECT c.org_id AS org_id, 'contact'::text AS entity_type, c.id::text AS entity_id,
				'member'::text AS entity_role, 'contacts.account_id'::text AS source, c.created_at
			FROM contacts c
			WHERE c.account_id = $1 AND c.deleted_at IS NULL
			UNION ALL
			SELECT d.org_id AS org_id, 'deal'::text AS entity_type, d.id::text AS entity_id,
				'linked'::text AS entity_role, 'deals.account_id'::text AS source, d.created_at
			FROM deals d
			WHERE d.account_id = $1 AND d.deleted_at IS NULL
		), filtered AS (
			SELECT * FROM linked` + filterClause + `
		)
		SELECT entity_type, entity_id, entity_role, source, created_at,
			COUNT(*) OVER() AS total
		FROM filtered
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(i) + ` OFFSET $` + strconv.Itoa(i+1)

	args = append(args, f.Limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.LinkedEntity, 0)
	total := 0
	for rows.Next() {
		var e domain.LinkedEntity
		if err := rows.Scan(&e.Type, &e.ID, &e.Role, &e.Source, &e.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *AccountRepo) CreateRelationship(ctx context.Context, rel *domain.AccountRelationship) (*domain.AccountRelationship, error) {
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		rel.OrgID = orgID
	}
	if rel.EffectiveFrom.IsZero() {
		rel.EffectiveFrom = time.Now().UTC()
	}
	if err := rel.Validate(); err != nil {
		return nil, err
	}

	var cycleExists bool
	err := r.db.QueryRow(ctx, `
		WITH RECURSIVE descendants(id) AS (
			SELECT $1::uuid
			UNION
			SELECT ar.child_account_id
			FROM account_relationships ar
			JOIN descendants d ON d.id = ar.parent_account_id
			WHERE ar.org_id = $3
			  AND ar.deleted_at IS NULL
		)
		SELECT EXISTS(SELECT 1 FROM descendants WHERE id = $2::uuid)`,
		rel.ChildAccountID, rel.ParentAccountID, rel.OrgID,
	).Scan(&cycleExists)
	if err != nil {
		return nil, err
	}
	if cycleExists {
		return nil, fmt.Errorf("%w: relationship creates cycle", domain.ErrValidation)
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO account_relationships
			(org_id, parent_account_id, child_account_id, relationship_type, ownership_percent, effective_from, effective_to)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING `+accountRelationshipCols,
		rel.OrgID, rel.ParentAccountID, rel.ChildAccountID, rel.RelationshipType, rel.OwnershipPercent, rel.EffectiveFrom, rel.EffectiveTo,
	)
	return scanAccountRelationship(row)
}

func (r *AccountRepo) DeleteRelationship(ctx context.Context, relationshipID uuid.UUID, deletedBy *uuid.UUID) error {
	q := `UPDATE account_relationships SET deleted_at=NOW(), deleted_by=$2 WHERE id=$1 AND deleted_at IS NULL`
	args := []any{relationshipID, deletedBy}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$3`
		args = append(args, orgID)
	}
	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AccountRepo) ListDescendants(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("%w: org context required", domain.ErrValidation)
	}

	rows, err := r.db.Query(ctx, `
		WITH RECURSIVE descendants(id) AS (
			SELECT child_account_id
			FROM account_relationships
			WHERE org_id = $1 AND parent_account_id = $2 AND deleted_at IS NULL
			UNION
			SELECT ar.child_account_id
			FROM account_relationships ar
			JOIN descendants d ON d.id = ar.parent_account_id
			WHERE ar.org_id = $1 AND ar.deleted_at IS NULL
		)
		SELECT id FROM descendants`, orgID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *AccountRepo) ListAncestors(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("%w: org context required", domain.ErrValidation)
	}

	rows, err := r.db.Query(ctx, `
		WITH RECURSIVE ancestors(id) AS (
			SELECT parent_account_id
			FROM account_relationships
			WHERE org_id = $1 AND child_account_id = $2 AND deleted_at IS NULL
			UNION
			SELECT ar.parent_account_id
			FROM account_relationships ar
			JOIN ancestors a ON a.id = ar.child_account_id
			WHERE ar.org_id = $1 AND ar.deleted_at IS NULL
		)
		SELECT id FROM ancestors`, orgID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
