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

const apiKeyCols = `id, org_id, created_by, name, key_prefix, scopes, last_used_at, expires_at, revoked_at, created_at`

type APIKeyRepo struct {
	db *pgxpool.Pool
}

func NewAPIKeyRepo(db *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

// Create inserts a new API key. keyHash must be the SHA-256 hex digest of the plaintext key.
func (r *APIKeyRepo) Create(ctx context.Context, k *domain.APIKey, keyHash string) (*domain.APIKey, error) {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	k.CreatedAt = time.Now().UTC()

	row := r.db.QueryRow(ctx, `
		INSERT INTO api_keys (id, org_id, created_by, name, key_hash, key_prefix, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+apiKeyCols,
		k.ID, k.OrgID, k.CreatedBy, k.Name, keyHash, k.KeyPrefix, k.Scopes, k.ExpiresAt, k.CreatedAt,
	)
	return scanAPIKey(row)
}

// GetByHash returns the API key matching the given SHA-256 hex hash, or nil if not found.
func (r *APIKeyRepo) GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+apiKeyCols+` FROM api_keys WHERE key_hash = $1`,
		keyHash,
	)
	k, err := scanAPIKey(row)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, nil
	}
	return k, err
}

// List returns all keys for the given org (including revoked, for audit display).
func (r *APIKeyRepo) List(ctx context.Context, orgID uuid.UUID) ([]*domain.APIKey, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+apiKeyCols+` FROM api_keys WHERE org_id = $1 ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// Revoke sets revoked_at to now for the given key, scoped to orgID.
func (r *APIKeyRepo) Revoke(ctx context.Context, id, orgID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE api_keys SET revoked_at = NOW() WHERE id = $1 AND org_id = $2 AND revoked_at IS NULL`,
		id, orgID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateLastUsed sets last_used_at to now for the given key.
func (r *APIKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

func scanAPIKey(row pgx.Row) (*domain.APIKey, error) {
	var k domain.APIKey
	err := row.Scan(
		&k.ID, &k.OrgID, &k.CreatedBy, &k.Name, &k.KeyPrefix,
		&k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.RevokedAt, &k.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &k, nil
}
