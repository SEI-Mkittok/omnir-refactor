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

type ContactRepo struct {
	db *pgxpool.Pool
}

func NewContactRepo(db *pgxpool.Pool) *ContactRepo {
	return &ContactRepo{db: db}
}

const contactCols = `
	id, first_name, last_name, email, phone,
	account_id, owner_id, lead_source, stage, tags,
	custom_fields, created_at, updated_at, deleted_at
`

func scanContact(row pgx.Row) (*domain.Contact, error) {
	var c domain.Contact
	err := row.Scan(
		&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone,
		&c.AccountID, &c.OwnerID, &c.LeadSource, &c.Stage, &c.Tags,
		&c.CustomFields, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *ContactRepo) Create(ctx context.Context, c *domain.Contact) (*domain.Contact, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO contacts
			(id, first_name, last_name, email, phone,
			 account_id, owner_id, lead_source, stage, tags,
			 custom_fields, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING `+contactCols,
		c.ID, c.FirstName, c.LastName, c.Email, c.Phone,
		c.AccountID, c.OwnerID, c.LeadSource, c.Stage, c.Tags,
		c.CustomFields, c.CreatedAt, c.UpdatedAt,
	)
	return scanContact(row)
}

func (r *ContactRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+contactCols+` FROM contacts WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanContact(row)
}

func (r *ContactRepo) Update(ctx context.Context, id uuid.UUID, patch domain.ContactPatch) (*domain.Contact, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.FirstName != nil {
		addArg("first_name", *patch.FirstName)
	}
	if patch.LastName != nil {
		addArg("last_name", *patch.LastName)
	}
	if patch.Email != nil {
		addArg("email", *patch.Email)
	}
	if patch.Phone != nil {
		addArg("phone", *patch.Phone)
	}
	if patch.AccountID != nil {
		addArg("account_id", *patch.AccountID)
	}
	if patch.OwnerID != nil {
		addArg("owner_id", *patch.OwnerID)
	}
	if patch.LeadSource != nil {
		addArg("lead_source", *patch.LeadSource)
	}
	if patch.Stage != nil {
		addArg("stage", *patch.Stage)
	}
	if patch.Tags != nil {
		addArg("tags", patch.Tags)
	}
	if patch.CustomFields != nil {
		addArg("custom_fields", patch.CustomFields)
	}

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE contacts SET %s WHERE id=$%d AND deleted_at IS NULL RETURNING %s`,
		strings.Join(sets, ", "), i, contactCols,
	)
	row := r.db.QueryRow(ctx, query, args...)
	return scanContact(row)
}

func (r *ContactRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE contacts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ContactRepo) List(ctx context.Context, f domain.ContactFilter) ([]*domain.Contact, int, error) {
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
	if f.Stage != nil {
		addWhere("stage", *f.Stage)
	}
	if f.Q != "" {
		where = append(where, fmt.Sprintf(
			`(first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)`, i, i, i,
		))
		args = append(args, "%"+f.Q+"%")
		i++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM contacts WHERE `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sortCol := "created_at"
	allowedSorts := map[string]bool{
		"created_at": true, "updated_at": true,
		"first_name": true, "last_name": true, "email": true,
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
			`SELECT %s FROM contacts WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
			contactCols, whereClause, sortCol, order, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		var c domain.Contact
		if err := rows.Scan(
			&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone,
			&c.AccountID, &c.OwnerID, &c.LeadSource, &c.Stage, &c.Tags,
			&c.CustomFields, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		contacts = append(contacts, &c)
	}
	return contacts, total, rows.Err()
}
