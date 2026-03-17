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

// ActivityRepo is the Postgres implementation of repository.ActivityRepository.
type ActivityRepo struct {
	db *pgxpool.Pool
}

func NewActivityRepo(db *pgxpool.Pool) *ActivityRepo {
	return &ActivityRepo{db: db}
}

const activityCols = `
	id, org_id, type, subject, description, due_date, completed_at,
	contact_id, account_id, deal_id, owner_id,
	created_at, updated_at, deleted_at
`

func scanActivity(row pgx.Row) (*domain.Activity, error) {
	var a domain.Activity
	err := row.Scan(
		&a.ID, &a.OrgID, &a.Type, &a.Subject, &a.Description, &a.DueDate, &a.CompletedAt,
		&a.ContactID, &a.AccountID, &a.DealID, &a.OwnerID,
		&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *ActivityRepo) Create(ctx context.Context, a *domain.Activity) (*domain.Activity, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		a.OrgID = orgID
	}
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO activities
			(id, org_id, type, subject, description, due_date, completed_at,
			 contact_id, account_id, deal_id, owner_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING `+activityCols,
		a.ID, a.OrgID, a.Type, a.Subject, a.Description, a.DueDate, a.CompletedAt,
		a.ContactID, a.AccountID, a.DealID, a.OwnerID, a.CreatedAt, a.UpdatedAt,
	)
	return scanActivity(row)
}

func (r *ActivityRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Activity, error) {
	q := `SELECT ` + activityCols + ` FROM activities WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	return scanActivity(row)
}

func (r *ActivityRepo) Update(ctx context.Context, id uuid.UUID, patch domain.ActivityPatch) (*domain.Activity, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.Type != nil {
		addArg("type", *patch.Type)
	}
	if patch.Subject != nil {
		addArg("subject", *patch.Subject)
	}
	if patch.Description != nil {
		addArg("description", *patch.Description)
	}
	if patch.DueDate != nil {
		addArg("due_date", *patch.DueDate)
	}
	if patch.CompletedAt != nil {
		addArg("completed_at", *patch.CompletedAt)
	}
	if patch.ContactID != nil {
		addArg("contact_id", *patch.ContactID)
	}
	if patch.AccountID != nil {
		addArg("account_id", *patch.AccountID)
	}
	if patch.DealID != nil {
		addArg("deal_id", *patch.DealID)
	}
	if patch.OwnerID != nil {
		addArg("owner_id", *patch.OwnerID)
	}

	whereClause := fmt.Sprintf(`id=$%d AND deleted_at IS NULL`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
	}

	query := fmt.Sprintf(
		`UPDATE activities SET %s WHERE %s RETURNING %s`,
		strings.Join(sets, ", "), whereClause, activityCols,
	)
	row := r.db.QueryRow(ctx, query, args...)
	return scanActivity(row)
}

func (r *ActivityRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE activities SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
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

func (r *ActivityRepo) List(ctx context.Context, f domain.ActivityFilter) ([]*domain.Activity, int, error) {
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

	if f.Type != nil {
		addWhere("type", *f.Type)
	}
	if f.OwnerID != nil {
		addWhere("owner_id", *f.OwnerID)
	}
	if f.ContactID != nil {
		addWhere("contact_id", *f.ContactID)
	}
	if f.AccountID != nil {
		addWhere("account_id", *f.AccountID)
	}
	if f.DealID != nil {
		addWhere("deal_id", *f.DealID)
	}
	if f.Q != "" {
		where = append(where, fmt.Sprintf(`(subject ILIKE $%d OR description ILIKE $%d)`, i, i))
		args = append(args, "%"+f.Q+"%")
		i++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM activities WHERE `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sortCol := "created_at"
	allowedSorts := map[string]bool{
		"created_at": true, "updated_at": true,
		"subject": true, "due_date": true, "type": true,
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
			`SELECT %s FROM activities WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
			activityCols, whereClause, sortCol, order, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var activities []*domain.Activity
	for rows.Next() {
		var a domain.Activity
		if err := rows.Scan(
			&a.ID, &a.OrgID, &a.Type, &a.Subject, &a.Description, &a.DueDate, &a.CompletedAt,
			&a.ContactID, &a.AccountID, &a.DealID, &a.OwnerID,
			&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		activities = append(activities, &a)
	}
	return activities, total, rows.Err()
}
