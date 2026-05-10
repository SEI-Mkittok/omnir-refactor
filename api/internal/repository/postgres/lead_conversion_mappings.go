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

type LeadConversionMappingRepo struct {
	db *pgxpool.Pool
}

func NewLeadConversionMappingRepo(db *pgxpool.Pool) *LeadConversionMappingRepo {
	return &LeadConversionMappingRepo{db: db}
}

const leadConversionCols = `id, org_id, lead_field, target_entity, target_field, is_active, created_at, updated_at`

func scanLeadConversionMapping(row pgx.Row) (*domain.LeadConversionMapping, error) {
	var m domain.LeadConversionMapping
	err := row.Scan(
		&m.ID, &m.OrgID, &m.LeadField, &m.TargetEntity, &m.TargetField, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *LeadConversionMappingRepo) List(ctx context.Context, orgID uuid.UUID) ([]*domain.LeadConversionMapping, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+leadConversionCols+`
		FROM lead_conversion_mappings
		WHERE org_id = $1
		ORDER BY created_at ASC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*domain.LeadConversionMapping{}
	for rows.Next() {
		item, err := scanLeadConversionMapping(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *LeadConversionMappingRepo) Replace(ctx context.Context, orgID uuid.UUID, rows []domain.LeadConversionMapping) ([]*domain.LeadConversionMapping, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `DELETE FROM lead_conversion_mappings WHERE org_id = $1`, orgID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	for _, row := range rows {
		leadField := strings.TrimSpace(row.LeadField)
		targetField := strings.TrimSpace(row.TargetField)
		if leadField == "" || targetField == "" || !row.TargetEntity.IsValid() {
			continue
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO lead_conversion_mappings (
				org_id, lead_field, target_entity, target_field, is_active, created_at, updated_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, orgID, leadField, row.TargetEntity, targetField, row.IsActive, now, now)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.List(ctx, orgID)
}

