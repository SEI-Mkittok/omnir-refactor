package postgres

import (
	"context"
	"encoding/json"
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
	quote_number_prefix, ticket_number_prefix, kb_article_number_prefix, invoice_number_prefix,
	company_name, company_logo_url, company_website, company_email, company_phone,
	company_address_line1, company_address_line2, company_city, company_state, company_postal_code, company_country,
	portal_enabled, portal_display_name, portal_announcement, portal_default_assignee_id,
	portal_menu, portal_shortcuts, portal_recent_widget_limit,
	smtp_host, smtp_port, smtp_username, smtp_from_email, smtp_from_name, smtp_security, smtp_auth_type, smtp_password_enc,
	config_support_email, config_upload_max_mb, config_default_page_size, config_list_preview_chars,
	menu_config,
	created_at, updated_at
`

func scanOrgSettings(row pgx.Row) (*domain.OrgSettings, error) {
	var s domain.OrgSettings
	var portalMenuRaw []byte
	var portalShortcutsRaw []byte
	var menuConfigRaw []byte
	err := row.Scan(
		&s.ID, &s.OrgID,
		&s.QuoteNumberStart, &s.TicketNumberStart, &s.KBArticleNumberStart, &s.InvoiceNumberStart,
		&s.QuoteNumberPrefix, &s.TicketNumberPrefix, &s.KBArticleNumberPrefix, &s.InvoiceNumberPrefix,
		&s.CompanyName, &s.CompanyLogoURL, &s.CompanyWebsite, &s.CompanyEmail, &s.CompanyPhone,
		&s.CompanyAddressLine1, &s.CompanyAddressLine2, &s.CompanyCity, &s.CompanyState, &s.CompanyPostalCode, &s.CompanyCountry,
		&s.PortalEnabled, &s.PortalDisplayName, &s.PortalAnnouncement, &s.PortalDefaultAssigneeID,
		&portalMenuRaw, &portalShortcutsRaw, &s.PortalRecentWidgetLimit,
		&s.SMTPHost, &s.SMTPPort, &s.SMTPUsername, &s.SMTPFromEmail, &s.SMTPFromName, &s.SMTPSecurity, &s.SMTPAuthType, &s.SMTPPasswordEnc,
		&s.ConfigSupportEmail, &s.ConfigUploadMaxMB, &s.ConfigDefaultPageSize, &s.ConfigListPreviewChars,
		&menuConfigRaw,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if len(portalMenuRaw) > 0 {
		if err := json.Unmarshal(portalMenuRaw, &s.PortalMenu); err != nil {
			return nil, err
		}
	}
	if s.PortalMenu == nil {
		s.PortalMenu = []string{}
	}

	if len(portalShortcutsRaw) > 0 {
		if err := json.Unmarshal(portalShortcutsRaw, &s.PortalShortcuts); err != nil {
			return nil, err
		}
	}
	if s.PortalShortcuts == nil {
		s.PortalShortcuts = []string{}
	}

	if len(menuConfigRaw) > 0 {
		if err := json.Unmarshal(menuConfigRaw, &s.MenuConfig); err != nil {
			return nil, err
		}
	}
	if s.MenuConfig == nil {
		s.MenuConfig = map[string]bool{}
	}

	s.SMTPPasswordSet = s.SMTPPasswordEnc != nil && *s.SMTPPasswordEnc != ""
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
	if patch.QuoteNumberPrefix != nil {
		setClauses = append(setClauses, fmt.Sprintf("quote_number_prefix = $%d", argN))
		args = append(args, *patch.QuoteNumberPrefix)
		argN++
	}
	if patch.TicketNumberPrefix != nil {
		setClauses = append(setClauses, fmt.Sprintf("ticket_number_prefix = $%d", argN))
		args = append(args, *patch.TicketNumberPrefix)
		argN++
	}
	if patch.KBArticleNumberPrefix != nil {
		setClauses = append(setClauses, fmt.Sprintf("kb_article_number_prefix = $%d", argN))
		args = append(args, *patch.KBArticleNumberPrefix)
		argN++
	}
	if patch.InvoiceNumberPrefix != nil {
		setClauses = append(setClauses, fmt.Sprintf("invoice_number_prefix = $%d", argN))
		args = append(args, *patch.InvoiceNumberPrefix)
		argN++
	}
	if patch.CompanyName != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_name = $%d", argN))
		args = append(args, *patch.CompanyName)
		argN++
	}
	if patch.CompanyLogoURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_logo_url = $%d", argN))
		args = append(args, *patch.CompanyLogoURL)
		argN++
	}
	if patch.CompanyWebsite != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_website = $%d", argN))
		args = append(args, *patch.CompanyWebsite)
		argN++
	}
	if patch.CompanyEmail != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_email = $%d", argN))
		args = append(args, *patch.CompanyEmail)
		argN++
	}
	if patch.CompanyPhone != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_phone = $%d", argN))
		args = append(args, *patch.CompanyPhone)
		argN++
	}
	if patch.CompanyAddressLine1 != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_address_line1 = $%d", argN))
		args = append(args, *patch.CompanyAddressLine1)
		argN++
	}
	if patch.CompanyAddressLine2 != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_address_line2 = $%d", argN))
		args = append(args, *patch.CompanyAddressLine2)
		argN++
	}
	if patch.CompanyCity != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_city = $%d", argN))
		args = append(args, *patch.CompanyCity)
		argN++
	}
	if patch.CompanyState != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_state = $%d", argN))
		args = append(args, *patch.CompanyState)
		argN++
	}
	if patch.CompanyPostalCode != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_postal_code = $%d", argN))
		args = append(args, *patch.CompanyPostalCode)
		argN++
	}
	if patch.CompanyCountry != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_country = $%d", argN))
		args = append(args, *patch.CompanyCountry)
		argN++
	}
	if patch.PortalEnabled != nil {
		setClauses = append(setClauses, fmt.Sprintf("portal_enabled = $%d", argN))
		args = append(args, *patch.PortalEnabled)
		argN++
	}
	if patch.PortalDisplayName != nil {
		setClauses = append(setClauses, fmt.Sprintf("portal_display_name = $%d", argN))
		args = append(args, *patch.PortalDisplayName)
		argN++
	}
	if patch.PortalAnnouncement != nil {
		setClauses = append(setClauses, fmt.Sprintf("portal_announcement = $%d", argN))
		args = append(args, *patch.PortalAnnouncement)
		argN++
	}
	if patch.PortalDefaultAssigneeID != nil {
		setClauses = append(setClauses, fmt.Sprintf("portal_default_assignee_id = $%d", argN))
		args = append(args, *patch.PortalDefaultAssigneeID)
		argN++
	}
	if patch.PortalMenu != nil {
		raw, err := json.Marshal(*patch.PortalMenu)
		if err != nil {
			return nil, err
		}
		setClauses = append(setClauses, fmt.Sprintf("portal_menu = $%d::jsonb", argN))
		args = append(args, string(raw))
		argN++
	}
	if patch.PortalShortcuts != nil {
		raw, err := json.Marshal(*patch.PortalShortcuts)
		if err != nil {
			return nil, err
		}
		setClauses = append(setClauses, fmt.Sprintf("portal_shortcuts = $%d::jsonb", argN))
		args = append(args, string(raw))
		argN++
	}
	if patch.PortalRecentWidgetLimit != nil {
		setClauses = append(setClauses, fmt.Sprintf("portal_recent_widget_limit = $%d", argN))
		args = append(args, *patch.PortalRecentWidgetLimit)
		argN++
	}
	if patch.SMTPHost != nil {
		setClauses = append(setClauses, fmt.Sprintf("smtp_host = $%d", argN))
		args = append(args, *patch.SMTPHost)
		argN++
	}
	if patch.SMTPPort != nil {
		setClauses = append(setClauses, fmt.Sprintf("smtp_port = $%d", argN))
		args = append(args, *patch.SMTPPort)
		argN++
	}
	if patch.SMTPUsername != nil {
		setClauses = append(setClauses, fmt.Sprintf("smtp_username = $%d", argN))
		args = append(args, *patch.SMTPUsername)
		argN++
	}
	if patch.SMTPFromEmail != nil {
		setClauses = append(setClauses, fmt.Sprintf("smtp_from_email = $%d", argN))
		args = append(args, *patch.SMTPFromEmail)
		argN++
	}
	if patch.SMTPFromName != nil {
		setClauses = append(setClauses, fmt.Sprintf("smtp_from_name = $%d", argN))
		args = append(args, *patch.SMTPFromName)
		argN++
	}
	if patch.SMTPSecurity != nil {
		setClauses = append(setClauses, fmt.Sprintf("smtp_security = $%d", argN))
		args = append(args, *patch.SMTPSecurity)
		argN++
	}
	if patch.SMTPAuthType != nil {
		setClauses = append(setClauses, fmt.Sprintf("smtp_auth_type = $%d", argN))
		args = append(args, *patch.SMTPAuthType)
		argN++
	}
	if patch.SMTPPasswordEnc != nil {
		setClauses = append(setClauses, fmt.Sprintf("smtp_password_enc = $%d", argN))
		args = append(args, *patch.SMTPPasswordEnc)
		argN++
	}
	if patch.ConfigSupportEmail != nil {
		setClauses = append(setClauses, fmt.Sprintf("config_support_email = $%d", argN))
		args = append(args, *patch.ConfigSupportEmail)
		argN++
	}
	if patch.ConfigUploadMaxMB != nil {
		setClauses = append(setClauses, fmt.Sprintf("config_upload_max_mb = $%d", argN))
		args = append(args, *patch.ConfigUploadMaxMB)
		argN++
	}
	if patch.ConfigDefaultPageSize != nil {
		setClauses = append(setClauses, fmt.Sprintf("config_default_page_size = $%d", argN))
		args = append(args, *patch.ConfigDefaultPageSize)
		argN++
	}
	if patch.ConfigListPreviewChars != nil {
		setClauses = append(setClauses, fmt.Sprintf("config_list_preview_chars = $%d", argN))
		args = append(args, *patch.ConfigListPreviewChars)
		argN++
	}
	if patch.MenuConfig != nil {
		raw, err := json.Marshal(*patch.MenuConfig)
		if err != nil {
			return nil, err
		}
		setClauses = append(setClauses, fmt.Sprintf("menu_config = $%d::jsonb", argN))
		args = append(args, string(raw))
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

// GetCurrentDocNumbers returns current sequence numbers for known document types.
// If a sequence does not exist yet, the current value is start-1.
func (r *OrgSettingsRepo) GetCurrentDocNumbers(ctx context.Context, orgID uuid.UUID) (map[domain.DocType]int64, error) {
	s, err := r.GetOrCreate(ctx, orgID)
	if err != nil {
		return nil, err
	}

	current := map[domain.DocType]int64{
		domain.DocTypeQuote:     s.QuoteNumberStart - 1,
		domain.DocTypeTicket:    s.TicketNumberStart - 1,
		domain.DocTypeKBArticle: s.KBArticleNumberStart - 1,
		domain.DocTypeInvoice:   s.InvoiceNumberStart - 1,
	}

	rows, err := r.db.Query(ctx, `
		SELECT doc_type, next_number - 1
		FROM document_sequences
		WHERE org_id = $1
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var docType string
		var num int64
		if err := rows.Scan(&docType, &num); err != nil {
			return nil, err
		}
		current[domain.DocType(docType)] = num
	}
	return current, rows.Err()
}

func (r *OrgSettingsRepo) SetCurrentDocNumbers(ctx context.Context, orgID uuid.UUID, values map[domain.DocType]int64) error {
	for docType, current := range values {
		if current < 0 {
			current = 0
		}
		_, err := r.db.Exec(ctx, `
			INSERT INTO document_sequences (org_id, doc_type, next_number, updated_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (org_id, doc_type) DO UPDATE SET
				next_number = EXCLUDED.next_number,
				updated_at = NOW()
		`, orgID, string(docType), current+1)
		if err != nil {
			return err
		}
	}
	return nil
}
