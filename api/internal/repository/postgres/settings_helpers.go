package postgres

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

const fallbackCurrencyCode = "USD"

func getOrgDefaultCurrency(ctx context.Context, db *pgxpool.Pool, orgID uuid.UUID) (string, error) {
	var code string
	err := db.QueryRow(ctx, `
		SELECT code
		FROM org_currencies
		WHERE org_id = $1 AND is_default = TRUE AND is_active = TRUE
		LIMIT 1
	`, orgID).Scan(&code)
	if err == nil {
		return strings.ToUpper(strings.TrimSpace(code)), nil
	}
	if err != nil && err != pgx.ErrNoRows {
		return "", err
	}

	// Lazy bootstrap on legacy orgs with no row yet.
	_, err = db.Exec(ctx, `
		INSERT INTO org_currencies (org_id, code, display_name, symbol, decimal_places, is_active, is_default)
		VALUES ($1, 'USD', 'US Dollar', '$', 2, TRUE, TRUE)
		ON CONFLICT (org_id, code) DO UPDATE
		SET is_active = TRUE
	`, orgID)
	if err != nil {
		return "", err
	}

	return fallbackCurrencyCode, nil
}

func getDocPrefix(ctx context.Context, db *pgxpool.Pool, orgID uuid.UUID, docType domain.DocType) (string, error) {
	var prefix string
	err := db.QueryRow(ctx, `
		SELECT CASE $2
			WHEN 'quote'      THEN quote_number_prefix
			WHEN 'ticket'     THEN ticket_number_prefix
			WHEN 'kb_article' THEN kb_article_number_prefix
			WHEN 'invoice'    THEN invoice_number_prefix
			ELSE NULL
		END
		FROM org_settings
		WHERE org_id = $1
	`, orgID, string(docType)).Scan(&prefix)
	if err != nil && err != pgx.ErrNoRows {
		return "", err
	}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		switch docType {
		case domain.DocTypeQuote:
			return "QUO", nil
		case domain.DocTypeTicket:
			return "TKT", nil
		case domain.DocTypeKBArticle:
			return "KB", nil
		case domain.DocTypeInvoice:
			return "INV", nil
		default:
			return "DOC", nil
		}
	}
	return strings.ToUpper(prefix), nil
}

