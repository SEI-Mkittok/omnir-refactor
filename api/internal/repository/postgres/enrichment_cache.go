package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// EnrichmentCacheRepo implements repository.EnrichmentCacheRepository using pgx.
type EnrichmentCacheRepo struct {
	db *pgxpool.Pool
}

func NewEnrichmentCacheRepo(db *pgxpool.Pool) *EnrichmentCacheRepo {
	return &EnrichmentCacheRepo{db: db}
}

func (r *EnrichmentCacheRepo) GetByDomain(ctx context.Context, d string) (*domain.EnrichmentCache, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, org_id, domain, data, fetched_at, created_at
		FROM enrichment_cache
		WHERE domain = $1
		LIMIT 1
	`, d)

	return scanEnrichmentCache(row)
}

func (r *EnrichmentCacheRepo) Upsert(ctx context.Context, entry *domain.EnrichmentCache) (*domain.EnrichmentCache, error) {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	raw, err := json.Marshal(entry.Data)
	if err != nil {
		return nil, err
	}

	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		orgID = domain.DefaultOrgID
	}
	entry.OrgID = orgID

	row := r.db.QueryRow(ctx, `
		INSERT INTO enrichment_cache (id, org_id, domain, data, fetched_at, created_at)
		VALUES ($1, $2, $3, $4, now(), now())
		ON CONFLICT (org_id, domain) DO UPDATE
		SET data = EXCLUDED.data, fetched_at = now()
		RETURNING id, org_id, domain, data, fetched_at, created_at
	`, entry.ID, orgID, entry.Domain, raw)

	return scanEnrichmentCache(row)
}

func scanEnrichmentCache(row pgx.Row) (*domain.EnrichmentCache, error) {
	var e domain.EnrichmentCache
	var raw json.RawMessage
	err := row.Scan(&e.ID, &e.OrgID, &e.Domain, &raw, &e.FetchedAt, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(raw, &e.Data); err != nil {
		return nil, err
	}
	return &e, nil
}
