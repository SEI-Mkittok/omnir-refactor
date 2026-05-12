package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// PortalLinkRepo is the Postgres implementation of repository.PortalLinkRepository.
type PortalLinkRepo struct {
	db *pgxpool.Pool
}

func NewPortalLinkRepo(db *pgxpool.Pool) *PortalLinkRepo {
	return &PortalLinkRepo{db: db}
}

const portalLinkCols = `id, org_id, deal_id, created_by_user_id, token, label,
    expires_at, revoked_at, view_count, last_viewed_at, created_at`

func scanPortalLink(row pgx.Row) (*domain.PortalLink, error) {
	var l domain.PortalLink
	err := row.Scan(
		&l.ID, &l.OrgID, &l.DealID, &l.CreatedByUserID, &l.Token, &l.Label,
		&l.ExpiresAt, &l.RevokedAt, &l.ViewCount, &l.LastViewedAt, &l.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &l, nil
}

func (r *PortalLinkRepo) Create(ctx context.Context, l *domain.PortalLink) (*domain.PortalLink, error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO portal_links
		    (id, org_id, deal_id, created_by_user_id, token, label, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+portalLinkCols,
		l.ID, l.OrgID, l.DealID, l.CreatedByUserID, l.Token, l.Label, l.ExpiresAt,
	)
	return scanPortalLink(row)
}

func (r *PortalLinkRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PortalLink, error) {
	q := `SELECT ` + portalLinkCols + ` FROM portal_links WHERE id = $1`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $2`
		args = append(args, orgID)
	}
	return scanPortalLink(r.db.QueryRow(ctx, q, args...))
}

// GetByToken looks up a portal link by its unique token. The query uses an
// explicit token = $1 filter, which is independent of the org-scoped session
// variable, so it works correctly for unauthenticated public portal requests.
func (r *PortalLinkRepo) GetByToken(ctx context.Context, token string) (*domain.PortalLink, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+portalLinkCols+` FROM portal_links WHERE token = $1`,
		token,
	)
	return scanPortalLink(row)
}

func (r *PortalLinkRepo) ListByDeal(ctx context.Context, dealID uuid.UUID) ([]*domain.PortalLink, error) {
	q := `SELECT ` + portalLinkCols + `
		 FROM portal_links
		 WHERE deal_id = $1 AND revoked_at IS NULL`
	args := []any{dealID}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $2`
		args = append(args, orgID)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*domain.PortalLink
	for rows.Next() {
		l, err := scanPortalLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (r *PortalLinkRepo) Revoke(ctx context.Context, id, orgID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE portal_links SET revoked_at = NOW() WHERE id = $1 AND org_id = $2 AND revoked_at IS NULL`,
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

func (r *PortalLinkRepo) IncrementView(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE portal_links SET view_count = view_count + 1, last_viewed_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}
