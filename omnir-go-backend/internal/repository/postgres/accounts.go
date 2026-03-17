package postgres

import (
	"context"
	"errors"
	"fmt"
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
	id, name, domain, industry, size,
	owner_id, tags, custom_fields, created_at, updated_at, deleted_at
`

func scanAccount(row pgx.Row) (*domain.Account, error) {
	var a domain.Account
	err := row.Scan(
		&a.ID, &a.Name, &a.Domain, &a.Industry, &a.Size,
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

func (r *AccountRepo) Create(ctx context.Context, a *domain.Account) (*domain.Account, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO accounts
			(id, name, domain, industry, size,
			 owner_id, tags, custom_fields, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+accountCols,
		a.ID, a.Name, a.Domain, a.Industry, a.Size,
		a.OwnerID, a.Tags, a.CustomFields, a.CreatedAt, a.UpdatedAt,
	)
	return scanAccount(row)
}

func (r *AccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+accountCols+` FROM accounts WHERE id=$1 AND deleted_at IS NULL`, id)
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

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE accounts SET %s WHERE id=$%d AND deleted_at IS NULL RETURNING %s`,
		strings.Join(sets, ", "), i, accountCols,
	)
	row := r.db.QueryRow(ctx, query, args...)
	return scanAccount(row)
}

func (r *AccountRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE accounts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
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
			`(name ILIKE $%d OR domain ILIKE $%d)`, i, i,
		))
		args = append(args, "%"+f.Q+"%")
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
			&a.ID, &a.Name, &a.Domain, &a.Industry, &a.Size,
			&a.OwnerID, &a.Tags, &a.CustomFields, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		accounts = append(accounts, &a)
	}
	return accounts, total, rows.Err()
}
