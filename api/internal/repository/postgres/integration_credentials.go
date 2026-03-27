package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// IntegrationCredentialRepo is the Postgres implementation of repository.IntegrationCredentialRepository.
type IntegrationCredentialRepo struct {
	db *pgxpool.Pool
}

func NewIntegrationCredentialRepo(db *pgxpool.Pool) *IntegrationCredentialRepo {
	return &IntegrationCredentialRepo{db: db}
}

const integrationCredCols = `
	id, org_id, provider, client_id, client_secret_enc, api_key_enc, webhook_secret,
	created_at, updated_at
`

func scanIntegrationCredential(row pgx.Row) (*domain.IntegrationCredential, error) {
	var c domain.IntegrationCredential
	err := row.Scan(
		&c.ID, &c.OrgID, &c.Provider,
		&c.ClientID, &c.ClientSecretEnc, &c.APIKeyEnc, &c.WebhookSecret,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *IntegrationCredentialRepo) Upsert(ctx context.Context, cred *domain.IntegrationCredential) (*domain.IntegrationCredential, error) {
	if cred.ID == uuid.Nil {
		cred.ID = uuid.New()
	}
	now := time.Now().UTC()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO integration_credentials
			(id, org_id, provider, client_id, client_secret_enc, api_key_enc, webhook_secret, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (org_id, provider) DO UPDATE SET
			client_id         = EXCLUDED.client_id,
			client_secret_enc = EXCLUDED.client_secret_enc,
			api_key_enc       = EXCLUDED.api_key_enc,
			webhook_secret    = EXCLUDED.webhook_secret,
			updated_at        = EXCLUDED.updated_at
		RETURNING `+integrationCredCols,
		cred.ID, cred.OrgID, cred.Provider,
		cred.ClientID, cred.ClientSecretEnc, cred.APIKeyEnc, cred.WebhookSecret,
		cred.CreatedAt, cred.UpdatedAt,
	)
	return scanIntegrationCredential(row)
}

func (r *IntegrationCredentialRepo) GetByProvider(ctx context.Context, orgID uuid.UUID, provider domain.IntegrationProvider) (*domain.IntegrationCredential, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+integrationCredCols+`
		FROM integration_credentials
		WHERE org_id = $1 AND provider = $2
	`, orgID, provider)
	return scanIntegrationCredential(row)
}

func (r *IntegrationCredentialRepo) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*domain.IntegrationCredential, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+integrationCredCols+`
		FROM integration_credentials
		WHERE org_id = $1
		ORDER BY provider
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []*domain.IntegrationCredential
	for rows.Next() {
		c, err := scanIntegrationCredential(rows)
		if err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}
	return creds, rows.Err()
}

func (r *IntegrationCredentialRepo) Delete(ctx context.Context, orgID uuid.UUID, provider domain.IntegrationProvider) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM integration_credentials WHERE org_id = $1 AND provider = $2
	`, orgID, provider)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
