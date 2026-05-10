package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type CurrencyRepo struct {
	db *pgxpool.Pool
}

func NewCurrencyRepo(db *pgxpool.Pool) *CurrencyRepo {
	return &CurrencyRepo{db: db}
}

const currencyCols = `id, org_id, code, display_name, symbol, decimal_places, is_active, is_default, created_at, updated_at`

func scanCurrency(row pgx.Row) (*domain.OrgCurrency, error) {
	var c domain.OrgCurrency
	err := row.Scan(
		&c.ID, &c.OrgID, &c.Code, &c.DisplayName, &c.Symbol, &c.DecimalPlaces,
		&c.IsActive, &c.IsDefault, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CurrencyRepo) ensureDefault(ctx context.Context, orgID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO org_currencies (org_id, code, display_name, symbol, decimal_places, is_active, is_default)
		VALUES ($1, 'USD', 'US Dollar', '$', 2, TRUE, TRUE)
		ON CONFLICT (org_id, code) DO UPDATE SET
			is_active = TRUE,
			display_name = EXCLUDED.display_name,
			symbol = EXCLUDED.symbol,
			decimal_places = EXCLUDED.decimal_places
	`, orgID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		UPDATE org_currencies
		SET is_default = (code = 'USD'), updated_at = NOW()
		WHERE org_id = $1 AND is_default = TRUE AND code <> 'USD'
	`, orgID)
	return err
}

func (r *CurrencyRepo) List(ctx context.Context, orgID uuid.UUID) ([]*domain.OrgCurrency, error) {
	if err := r.ensureDefault(ctx, orgID); err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT `+currencyCols+`
		FROM org_currencies
		WHERE org_id = $1
		ORDER BY is_default DESC, code ASC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*domain.OrgCurrency{}
	for rows.Next() {
		c, err := scanCurrency(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CurrencyRepo) Replace(ctx context.Context, orgID uuid.UUID, req domain.OrgCurrencyUpdateRequest) ([]*domain.OrgCurrency, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if len(req.Currencies) == 0 {
		req.Currencies = []domain.OrgCurrencyInput{
			{Code: "USD", DisplayName: "US Dollar", Symbol: "$", DecimalPlaces: 2, IsActive: true},
		}
	}
	defaultCode := strings.ToUpper(strings.TrimSpace(req.DefaultCode))
	if defaultCode == "" {
		defaultCode = strings.ToUpper(strings.TrimSpace(req.Currencies[0].Code))
	}

	seen := map[string]struct{}{}
	for _, c := range req.Currencies {
		code := strings.ToUpper(strings.TrimSpace(c.Code))
		if code == "" {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}

		displayName := strings.TrimSpace(c.DisplayName)
		if displayName == "" {
			displayName = code
		}
		symbol := strings.TrimSpace(c.Symbol)
		if symbol == "" {
			symbol = code
		}
		decimalPlaces := c.DecimalPlaces
		if decimalPlaces < 0 || decimalPlaces > 6 {
			decimalPlaces = 2
		}
		isDefault := code == defaultCode
		isActive := c.IsActive || isDefault

		_, err = tx.Exec(ctx, `
			INSERT INTO org_currencies (org_id, code, display_name, symbol, decimal_places, is_active, is_default, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
			ON CONFLICT (org_id, code) DO UPDATE SET
				display_name = EXCLUDED.display_name,
				symbol = EXCLUDED.symbol,
				decimal_places = EXCLUDED.decimal_places,
				is_active = EXCLUDED.is_active,
				is_default = EXCLUDED.is_default,
				updated_at = NOW()
		`, orgID, code, displayName, symbol, decimalPlaces, isActive, isDefault)
		if err != nil {
			return nil, err
		}
	}

	// Keep rows that are omitted as inactive, except the selected default.
	_, err = tx.Exec(ctx, `
		UPDATE org_currencies
		SET is_active = CASE WHEN code = $2 THEN TRUE ELSE FALSE END,
		    is_default = CASE WHEN code = $2 THEN TRUE ELSE FALSE END,
		    updated_at = NOW()
		WHERE org_id = $1 AND code <> ALL($3)
	`, orgID, defaultCode, keysToSlice(seen))
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE org_currencies
		SET is_default = (code = $2), updated_at = NOW()
		WHERE org_id = $1
	`, orgID, defaultCode)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.List(ctx, orgID)
}

func keysToSlice(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	return out
}

func (r *CurrencyRepo) GetDefaultCode(ctx context.Context, orgID uuid.UUID) (string, error) {
	if err := r.ensureDefault(ctx, orgID); err != nil {
		return "", err
	}
	var code string
	err := r.db.QueryRow(ctx, `
		SELECT code
		FROM org_currencies
		WHERE org_id = $1 AND is_default = TRUE
		LIMIT 1
	`, orgID).Scan(&code)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(strings.TrimSpace(code)), nil
}

func (r *CurrencyRepo) touchAll(ctx context.Context, orgID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE org_currencies SET updated_at = $2 WHERE org_id = $1`, orgID, time.Now().UTC())
	return err
}

