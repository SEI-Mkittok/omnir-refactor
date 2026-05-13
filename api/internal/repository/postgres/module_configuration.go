package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type ModuleLayoutRepo struct {
	db *pgxpool.Pool
}

func NewModuleLayoutRepo(db *pgxpool.Pool) *ModuleLayoutRepo {
	return &ModuleLayoutRepo{db: db}
}

const moduleLayoutCols = `id, org_id, entity_type, blocks, created_at, updated_at`

func scanModuleLayout(row pgx.Row) (*domain.ModuleLayout, error) {
	var layout domain.ModuleLayout
	var blocksRaw []byte
	if err := row.Scan(&layout.ID, &layout.OrgID, &layout.EntityType, &blocksRaw, &layout.CreatedAt, &layout.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(blocksRaw) > 0 {
		if err := json.Unmarshal(blocksRaw, &layout.Blocks); err != nil {
			return nil, err
		}
	}
	return &layout, nil
}

func (r *ModuleLayoutRepo) GetByEntity(ctx context.Context, orgID uuid.UUID, entityType domain.CustomFieldEntityType) (*domain.ModuleLayout, error) {
	if ctxOrgID, ok := domain.OrgIDFromContext(ctx); ok {
		orgID = ctxOrgID
	}
	return scanModuleLayout(r.db.QueryRow(ctx, `
		SELECT `+moduleLayoutCols+`
		FROM module_layouts
		WHERE org_id = $1 AND entity_type = $2
	`, orgID, entityType))
}

func (r *ModuleLayoutRepo) Upsert(ctx context.Context, layout *domain.ModuleLayout) (*domain.ModuleLayout, error) {
	if layout.ID == uuid.Nil {
		layout.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		layout.OrgID = orgID
	}
	if layout.OrgID == uuid.Nil {
		return nil, fmt.Errorf("%w: org_id is required", domain.ErrValidation)
	}
	blocks, err := json.Marshal(layout.Blocks)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	layout.CreatedAt = now
	layout.UpdatedAt = now
	return scanModuleLayout(r.db.QueryRow(ctx, `
		INSERT INTO module_layouts (id, org_id, entity_type, blocks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (org_id, entity_type) DO UPDATE
		SET blocks = EXCLUDED.blocks,
		    updated_at = NOW()
		RETURNING `+moduleLayoutCols,
		layout.ID, layout.OrgID, layout.EntityType, blocks, layout.CreatedAt, layout.UpdatedAt,
	))
}

func (r *ModuleLayoutRepo) Delete(ctx context.Context, orgID uuid.UUID, entityType domain.CustomFieldEntityType) error {
	if ctxOrgID, ok := domain.OrgIDFromContext(ctx); ok {
		orgID = ctxOrgID
	}
	tag, err := r.db.Exec(ctx, `DELETE FROM module_layouts WHERE org_id = $1 AND entity_type = $2`, orgID, entityType)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type ModuleRelationshipDefinitionRepo struct {
	db *pgxpool.Pool
}

func NewModuleRelationshipDefinitionRepo(db *pgxpool.Pool) *ModuleRelationshipDefinitionRepo {
	return &ModuleRelationshipDefinitionRepo{db: db}
}

const moduleRelationshipDefinitionCols = `id, org_id, relationship_key, from_entity_type, to_entity_type, label, cardinality, storage_strategy, is_enabled, system_locked, order_idx, metadata, created_at, updated_at`

func scanModuleRelationshipDefinition(row pgx.Row) (*domain.ModuleRelationshipDefinition, error) {
	var def domain.ModuleRelationshipDefinition
	if err := row.Scan(
		&def.ID,
		&def.OrgID,
		&def.RelationshipKey,
		&def.FromEntityType,
		&def.ToEntityType,
		&def.Label,
		&def.Cardinality,
		&def.StorageStrategy,
		&def.IsEnabled,
		&def.SystemLocked,
		&def.OrderIdx,
		&def.Metadata,
		&def.CreatedAt,
		&def.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(def.Metadata) == 0 {
		def.Metadata = json.RawMessage(`{}`)
	}
	return &def, nil
}

func (r *ModuleRelationshipDefinitionRepo) List(ctx context.Context, filter domain.ModuleRelationshipDefinitionFilter) ([]*domain.ModuleRelationshipDefinition, error) {
	conditions := []string{"1=1"}
	args := []any{}
	i := 1
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		conditions = append(conditions, fmt.Sprintf("org_id = $%d", i))
		args = append(args, orgID)
		i++
	} else if filter.OrgID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("org_id = $%d", i))
		args = append(args, filter.OrgID)
		i++
	}
	if filter.EntityType != nil {
		conditions = append(conditions, fmt.Sprintf("(from_entity_type = $%d OR to_entity_type = $%d)", i, i))
		args = append(args, *filter.EntityType)
		i++
	}
	rows, err := r.db.Query(ctx, `
		SELECT `+moduleRelationshipDefinitionCols+`
		FROM module_relationship_definitions
		WHERE `+strings.Join(conditions, " AND ")+`
		ORDER BY order_idx ASC, label ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var defs []*domain.ModuleRelationshipDefinition
	for rows.Next() {
		def, err := scanModuleRelationshipDefinition(rows)
		if err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}
	return defs, rows.Err()
}

func (r *ModuleRelationshipDefinitionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ModuleRelationshipDefinition, error) {
	q := `SELECT ` + moduleRelationshipDefinitionCols + ` FROM module_relationship_definitions WHERE id = $1`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $2`
		args = append(args, orgID)
	}
	return scanModuleRelationshipDefinition(r.db.QueryRow(ctx, q, args...))
}

func (r *ModuleRelationshipDefinitionRepo) GetByKey(ctx context.Context, orgID uuid.UUID, key string) (*domain.ModuleRelationshipDefinition, error) {
	if ctxOrgID, ok := domain.OrgIDFromContext(ctx); ok {
		orgID = ctxOrgID
	}
	return scanModuleRelationshipDefinition(r.db.QueryRow(ctx, `
		SELECT `+moduleRelationshipDefinitionCols+`
		FROM module_relationship_definitions
		WHERE org_id = $1 AND relationship_key = $2
	`, orgID, key))
}

func (r *ModuleRelationshipDefinitionRepo) Upsert(ctx context.Context, def *domain.ModuleRelationshipDefinition) (*domain.ModuleRelationshipDefinition, error) {
	if def.ID == uuid.Nil {
		def.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		def.OrgID = orgID
	}
	if def.OrgID == uuid.Nil {
		return nil, fmt.Errorf("%w: org_id is required", domain.ErrValidation)
	}
	if len(def.Metadata) == 0 {
		def.Metadata = json.RawMessage(`{}`)
	}
	now := time.Now().UTC()
	def.CreatedAt = now
	def.UpdatedAt = now
	created, err := scanModuleRelationshipDefinition(r.db.QueryRow(ctx, `
		INSERT INTO module_relationship_definitions
			(id, org_id, relationship_key, from_entity_type, to_entity_type, label, cardinality, storage_strategy, is_enabled, system_locked, order_idx, metadata, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (org_id, relationship_key) DO UPDATE
		SET label = EXCLUDED.label,
		    is_enabled = EXCLUDED.is_enabled,
		    order_idx = EXCLUDED.order_idx,
		    metadata = EXCLUDED.metadata,
		    from_entity_type = EXCLUDED.from_entity_type,
		    to_entity_type = EXCLUDED.to_entity_type,
		    cardinality = EXCLUDED.cardinality,
		    storage_strategy = EXCLUDED.storage_strategy,
		    system_locked = EXCLUDED.system_locked,
		    updated_at = NOW()
		RETURNING `+moduleRelationshipDefinitionCols,
		def.ID,
		def.OrgID,
		def.RelationshipKey,
		def.FromEntityType,
		def.ToEntityType,
		def.Label,
		def.Cardinality,
		def.StorageStrategy,
		def.IsEnabled,
		def.SystemLocked,
		def.OrderIdx,
		def.Metadata,
		def.CreatedAt,
		def.UpdatedAt,
	))
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		return nil, fmt.Errorf("%w: relationship key already exists", domain.ErrConflict)
	}
	return created, err
}

func (r *ModuleRelationshipDefinitionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM module_relationship_definitions WHERE id = $1 AND system_locked = FALSE`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $2`
		args = append(args, orgID)
	}
	tag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			return fmt.Errorf("%w: relationship definition has existing links", domain.ErrConflict)
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
