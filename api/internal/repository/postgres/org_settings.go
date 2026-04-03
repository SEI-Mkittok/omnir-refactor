package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// OrgSettingsRepo is the Postgres implementation of repository.OrgSettingsRepository.
type OrgSettingsRepo struct {
	db *pgxpool.Pool
}

func NewOrgSettingsRepo(db *pgxpool.Pool) *OrgSettingsRepo {
	return &OrgSettingsRepo{db: db}
}

const orgSettingsCols = `
	id, org_id, quote_number_start, ticket_number_start, kb_article_number_start, invoice_number_start,
	created_at, updated_at
`

func scanOrgSettings(row pgx.Row) (*domain.OrgSettings, error) {
	var s domain.OrgSettings
	err := row.Scan(
		&s.ID, &s.OrgID,
		&s.QuoteNumberStart, &s.TicketNumberStart, &s.KBArticleNumberStart, &s.InvoiceNumberStart,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *OrgSettingsRepo) GetOrCreate(ctx context.Context, orgID uuid.UUID) (*domain.OrgSettings, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO org_settings (org_id)
		VALUES ($1)
		ON CONFLICT (org_id) DO UPDATE SET updated_at = org_settings.updated_at
		RETURNING `+orgSettingsCols,
		orgID,
	)
	return scanOrgSettings(row)
}

func (r *OrgSettingsRepo) Update(ctx context.Context, orgID uuid.UUID, patch domain.OrgSettingsPatch) (*domain.OrgSettings, error) {
	setClauses := []string{}
	args := []any{orgID}
	argN := 2

	if patch.QuoteNumberStart != nil {
		setClauses = append(setClauses, fmt.Sprintf("quote_number_start = $%d", argN))
		args = append(args, *patch.QuoteNumberStart)
		argN++
	}
	if patch.TicketNumberStart != nil {
		setClauses = append(setClauses, fmt.Sprintf("ticket_number_start = $%d", argN))
		args = append(args, *patch.TicketNumberStart)
		argN++
	}
	if patch.KBArticleNumberStart != nil {
		setClauses = append(setClauses, fmt.Sprintf("kb_article_number_start = $%d", argN))
		args = append(args, *patch.KBArticleNumberStart)
		argN++
	}
	if patch.InvoiceNumberStart != nil {
		setClauses = append(setClauses, fmt.Sprintf("invoice_number_start = $%d", argN))
		args = append(args, *patch.InvoiceNumberStart)
		argN++
	}
	if len(setClauses) == 0 {
		return r.GetOrCreate(ctx, orgID)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argN))
	args = append(args, time.Now().UTC())

	q := "UPDATE org_settings SET "
	for i, c := range setClauses {
		if i > 0 {
			q += ", "
		}
		q += c
	}
	q += " WHERE org_id = $1 RETURNING " + orgSettingsCols

	row := r.db.QueryRow(ctx, q, args...)
	result, err := scanOrgSettings(row)
	if errors.Is(err, domain.ErrNotFound) {
		// Row didn't exist yet — create it then apply the patch.
		if _, err2 := r.GetOrCreate(ctx, orgID); err2 != nil {
			return nil, err2
		}
		row = r.db.QueryRow(ctx, q, args...)
		return scanOrgSettings(row)
	}
	return result, err
}

// getNextDocNumber atomically allocates the next number for (orgID, docType).
// On first call it lazily reads the org's configured start value from org_settings.
// The function is safe for concurrent callers.
func getNextDocNumber(ctx context.Context, db *pgxpool.Pool, orgID uuid.UUID, docType domain.DocType) (int64, error) {
	var num int64
	err := db.QueryRow(ctx, `
		WITH start AS (
			SELECT CASE $2
				WHEN 'quote'      THEN COALESCE((SELECT quote_number_start      FROM org_settings WHERE org_id = $1), 1)
				WHEN 'ticket'     THEN COALESCE((SELECT ticket_number_start     FROM org_settings WHERE org_id = $1), 1)
				WHEN 'kb_article' THEN COALESCE((SELECT kb_article_number_start FROM org_settings WHERE org_id = $1), 1)
				WHEN 'invoice'    THEN COALESCE((SELECT invoice_number_start    FROM org_settings WHERE org_id = $1), 1)
				ELSE 1
			END AS val
		)
		INSERT INTO document_sequences (org_id, doc_type, next_number)
		SELECT $1, $2, (SELECT val FROM start) + 1
		ON CONFLICT (org_id, doc_type) DO UPDATE
			SET next_number = document_sequences.next_number + 1,
			    updated_at  = NOW()
		RETURNING next_number - 1
	`, orgID, string(docType)).Scan(&num)
	return num, err
}
