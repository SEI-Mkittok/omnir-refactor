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

// CRMEntityLinkRepo implements repository.CRMEntityLinkRepository.
type CRMEntityLinkRepo struct {
	db *pgxpool.Pool
}

func NewCRMEntityLinkRepo(db *pgxpool.Pool) *CRMEntityLinkRepo {
	return &CRMEntityLinkRepo{db: db}
}

const crmEntityLinkCols = `id, org_id, from_entity_type, from_entity_id, to_entity_type, to_entity_id, link_type, metadata, created_at, updated_at`

func scanCRMEntityLink(row pgx.Row) (*domain.CRMEntityLink, error) {
	var link domain.CRMEntityLink
	if err := row.Scan(
		&link.ID,
		&link.OrgID,
		&link.FromEntityType,
		&link.FromEntityID,
		&link.ToEntityType,
		&link.ToEntityID,
		&link.LinkType,
		&link.Metadata,
		&link.CreatedAt,
		&link.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &link, nil
}

func (r *CRMEntityLinkRepo) Create(ctx context.Context, link *domain.CRMEntityLink) (*domain.CRMEntityLink, error) {
	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		link.OrgID = orgID
	}
	if err := link.Validate(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	link.CreatedAt = now
	link.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO crm_entity_links
			(id, org_id, from_entity_type, from_entity_id, to_entity_type, to_entity_id, link_type, metadata, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+crmEntityLinkCols,
		link.ID,
		link.OrgID,
		link.FromEntityType,
		link.FromEntityID,
		link.ToEntityType,
		link.ToEntityID,
		link.LinkType,
		link.Metadata,
		link.CreatedAt,
		link.UpdatedAt,
	)
	return scanCRMEntityLink(row)
}

func (r *CRMEntityLinkRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CRMEntityLink, error) {
	q := `SELECT ` + crmEntityLinkCols + ` FROM crm_entity_links WHERE id=$1`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	return scanCRMEntityLink(r.db.QueryRow(ctx, q, args...))
}

func (r *CRMEntityLinkRepo) ListForEntity(ctx context.Context, filter domain.CRMEntityLinkFilter) ([]*domain.CRMEntityLink, error) {
	filter.Normalize()
	args := []any{filter.EntityType, filter.EntityID}
	q := `SELECT ` + crmEntityLinkCols + ` FROM crm_entity_links WHERE (`
	argN := 3
	if filter.IncludeFrom {
		q += `(from_entity_type=$1 AND from_entity_id=$2)`
	}
	if filter.IncludeTo {
		if filter.IncludeFrom {
			q += ` OR `
		}
		q += `(to_entity_type=$1 AND to_entity_id=$2)`
	}
	q += `)`

	if filter.LinkType != "" {
		q += ` AND link_type=$` + itoa(argN)
		args = append(args, filter.LinkType)
		argN++
	}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$` + itoa(argN)
		args = append(args, orgID)
	}

	q += ` ORDER BY created_at ASC, id ASC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]*domain.CRMEntityLink, 0)
	for rows.Next() {
		var link domain.CRMEntityLink
		if err := rows.Scan(
			&link.ID,
			&link.OrgID,
			&link.FromEntityType,
			&link.FromEntityID,
			&link.ToEntityType,
			&link.ToEntityID,
			&link.LinkType,
			&link.Metadata,
			&link.CreatedAt,
			&link.UpdatedAt,
		); err != nil {
			return nil, err
		}
		links = append(links, &link)
	}
	return links, rows.Err()
}

func (r *CRMEntityLinkRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM crm_entity_links WHERE id=$1`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	res, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
