package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type ssoConfigRepo struct{ db *pgxpool.Pool }

func NewSSOConfigRepo(db *pgxpool.Pool) *ssoConfigRepo { //nolint:revive // internal package, unexported type is intentional
	return &ssoConfigRepo{db: db}
}

func (r *ssoConfigRepo) Upsert(ctx context.Context, cfg *domain.SSOConfig) (*domain.SSOConfig, error) {
	mapping, err := json.Marshal(cfg.AttributeMapping)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO sso_configs (org_id, provider, client_id, client_secret, issuer_url, attribute_mapping, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (org_id) DO UPDATE SET
			provider          = EXCLUDED.provider,
			client_id         = EXCLUDED.client_id,
			client_secret     = EXCLUDED.client_secret,
			issuer_url        = EXCLUDED.issuer_url,
			attribute_mapping = EXCLUDED.attribute_mapping,
			enabled           = EXCLUDED.enabled
		RETURNING id, org_id, provider, client_id, client_secret, issuer_url, attribute_mapping, enabled, created_at
	`, cfg.OrgID, cfg.Provider, cfg.ClientID, cfg.ClientSecret, cfg.IssuerURL, mapping, cfg.Enabled)
	return scanSSOConfig(row)
}

func (r *ssoConfigRepo) GetByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.SSOConfig, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, org_id, provider, client_id, client_secret, issuer_url, attribute_mapping, enabled, created_at
		FROM sso_configs WHERE org_id = $1
	`, orgID)
	cfg, err := scanSSOConfig(row)
	if err != nil {
		return nil, nil // not found
	}
	return cfg, nil
}

func (r *ssoConfigRepo) GetByOrgSlug(ctx context.Context, slug string) (*domain.SSOConfig, error) {
	row := r.db.QueryRow(ctx, `
		SELECT s.id, s.org_id, s.provider, s.client_id, s.client_secret, s.issuer_url, s.attribute_mapping, s.enabled, s.created_at
		FROM sso_configs s
		JOIN orgs o ON o.id = s.org_id
		WHERE o.slug = $1 AND s.enabled = true
	`, slug)
	cfg, err := scanSSOConfig(row)
	if err != nil {
		return nil, nil
	}
	return cfg, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanSSOConfig(row scannable) (*domain.SSOConfig, error) {
	var cfg domain.SSOConfig
	var mappingJSON []byte
	if err := row.Scan(&cfg.ID, &cfg.OrgID, &cfg.Provider, &cfg.ClientID, &cfg.ClientSecret,
		&cfg.IssuerURL, &mappingJSON, &cfg.Enabled, &cfg.CreatedAt); err != nil {
		return nil, err
	}
	if len(mappingJSON) > 0 {
		_ = json.Unmarshal(mappingJSON, &cfg.AttributeMapping)
	}
	return &cfg, nil
}
