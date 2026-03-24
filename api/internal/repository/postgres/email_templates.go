package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type EmailTemplateRepo struct {
	db *pgxpool.Pool
}

func NewEmailTemplateRepo(db *pgxpool.Pool) *EmailTemplateRepo {
	return &EmailTemplateRepo{db: db}
}

const emailTemplateCols = `id, org_id, name, subject, body, created_at, updated_at`

func scanEmailTemplate(row pgx.Row) (*domain.EmailTemplate, error) {
	var t domain.EmailTemplate
	err := row.Scan(&t.ID, &t.OrgID, &t.Name, &t.Subject, &t.Body, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *EmailTemplateRepo) Create(ctx context.Context, t *domain.EmailTemplate) (*domain.EmailTemplate, error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	row := r.db.QueryRow(ctx,
		`INSERT INTO email_templates (id, org_id, name, subject, body)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING `+emailTemplateCols,
		t.ID, t.OrgID, t.Name, t.Subject, t.Body,
	)
	return scanEmailTemplate(row)
}

func (r *EmailTemplateRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailTemplate, error) {
	row := r.db.QueryRow(ctx, `SELECT `+emailTemplateCols+` FROM email_templates WHERE id=$1`, id)
	return scanEmailTemplate(row)
}

func (r *EmailTemplateRepo) Update(ctx context.Context, id uuid.UUID, patch domain.EmailTemplatePatch) (*domain.EmailTemplate, error) {
	setClauses := []string{}
	args := []any{}
	i := 1

	if patch.Name != nil {
		setClauses = append(setClauses, "name=$"+itoa(i))
		args = append(args, *patch.Name)
		i++
	}
	if patch.Subject != nil {
		setClauses = append(setClauses, "subject=$"+itoa(i))
		args = append(args, *patch.Subject)
		i++
	}
	if patch.Body != nil {
		setClauses = append(setClauses, "body=$"+itoa(i))
		args = append(args, *patch.Body)
		i++
	}

	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at=NOW()")
		args = append(args, id)
		q := `UPDATE email_templates SET ` + strings.Join(setClauses, ",") + ` WHERE id=$` + itoa(i)
		if _, err := r.db.Exec(ctx, q, args...); err != nil {
			return nil, err
		}
	}
	return r.GetByID(ctx, id)
}

func (r *EmailTemplateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.Exec(ctx, `DELETE FROM email_templates WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *EmailTemplateRepo) List(ctx context.Context, filter domain.EmailTemplateFilter) ([]*domain.EmailTemplate, int, error) {
	where := []string{"org_id=$1"}
	args := []any{filter.OrgID}
	i := 2

	if filter.Q != "" {
		where = append(where, "(name ILIKE $"+itoa(i)+" OR subject ILIKE $"+itoa(i)+")")
		args = append(args, "%"+filter.Q+"%")
		i++
	}

	clause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM email_templates `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx,
		`SELECT `+emailTemplateCols+` FROM email_templates `+clause+
			` ORDER BY created_at DESC LIMIT $`+itoa(i)+` OFFSET $`+itoa(i+1),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var templates []*domain.EmailTemplate
	for rows.Next() {
		t, err := scanEmailTemplate(rows)
		if err != nil {
			return nil, 0, err
		}
		templates = append(templates, t)
	}
	return templates, total, rows.Err()
}
