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

// EmailRepo is the Postgres implementation of repository.EmailRepository.
type EmailRepo struct {
	db *pgxpool.Pool
}

func NewEmailRepo(db *pgxpool.Pool) *EmailRepo {
	return &EmailRepo{db: db}
}

const emailCols = `
	id, org_id, contact_id, deal_id, direction,
	from_addr, to_addr, subject, body, thread_id, message_id,
	sent_at, created_at
`

func scanEmail(row pgx.Row) (*domain.ContactEmail, error) {
	var e domain.ContactEmail
	err := row.Scan(
		&e.ID, &e.OrgID, &e.ContactID, &e.DealID, &e.Direction,
		&e.FromAddr, &e.ToAddr, &e.Subject, &e.Body, &e.ThreadID, &e.MessageID,
		&e.SentAt, &e.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *EmailRepo) Create(ctx context.Context, e *domain.ContactEmail) (*domain.ContactEmail, error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		e.OrgID = orgID
	}
	now := time.Now().UTC()
	e.CreatedAt = now
	if e.SentAt.IsZero() {
		e.SentAt = now
	}
	if e.ThreadID == "" {
		e.ThreadID = e.ID.String()
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO contact_emails
			(id, org_id, contact_id, deal_id, direction,
			 from_addr, to_addr, subject, body, thread_id, message_id,
			 sent_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING `+emailCols,
		e.ID, e.OrgID, e.ContactID, e.DealID, e.Direction,
		e.FromAddr, e.ToAddr, e.Subject, e.Body, e.ThreadID, e.MessageID,
		e.SentAt, e.CreatedAt,
	)
	return scanEmail(row)
}

func (r *EmailRepo) List(ctx context.Context, f domain.EmailFilter) ([]*domain.ContactEmail, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"1=1"}
	args := []any{}
	i := 1

	addWhere := func(expr string, val any) {
		where = append(where, fmt.Sprintf("%s = $%d", expr, i))
		args = append(args, val)
		i++
	}

	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if !hasCtxOrg {
		orgID = f.OrgID
	}
	if orgID != uuid.Nil {
		addWhere("org_id", orgID)
	}

	if f.ContactID != nil {
		addWhere("contact_id", *f.ContactID)
	}
	if f.DealID != nil {
		addWhere("deal_id", *f.DealID)
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM contact_emails WHERE `+whereClause, args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		fmt.Sprintf(
			`SELECT %s FROM contact_emails WHERE %s ORDER BY sent_at DESC LIMIT $%d OFFSET $%d`,
			emailCols, whereClause, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var emails []*domain.ContactEmail
	for rows.Next() {
		var e domain.ContactEmail
		if err := rows.Scan(
			&e.ID, &e.OrgID, &e.ContactID, &e.DealID, &e.Direction,
			&e.FromAddr, &e.ToAddr, &e.Subject, &e.Body, &e.ThreadID, &e.MessageID,
			&e.SentAt, &e.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		emails = append(emails, &e)
	}
	return emails, total, rows.Err()
}
