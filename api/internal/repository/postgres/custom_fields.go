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

// CustomFieldDefinitionRepo implements repository.CustomFieldDefinitionRepository.
type CustomFieldDefinitionRepo struct {
	db *pgxpool.Pool
}

func NewCustomFieldDefinitionRepo(db *pgxpool.Pool) *CustomFieldDefinitionRepo {
	return &CustomFieldDefinitionRepo{db: db}
}

const cfdCols = `id, org_id, entity_type, name, label, field_type, options, required, order_idx, created_at, updated_at, deleted_at`

func scanCFD(row pgx.Row) (*domain.CustomFieldDefinition, error) {
	var d domain.CustomFieldDefinition
	var optionsRaw []byte
	err := row.Scan(
		&d.ID, &d.OrgID, &d.EntityType, &d.Name, &d.Label, &d.FieldType,
		&optionsRaw, &d.Required, &d.OrderIdx,
		&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(optionsRaw) > 0 {
		_ = json.Unmarshal(optionsRaw, &d.Options)
	}
	return &d, nil
}

func (r *CustomFieldDefinitionRepo) Create(ctx context.Context, def *domain.CustomFieldDefinition) (*domain.CustomFieldDefinition, error) {
	if def.ID == uuid.Nil {
		def.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		def.OrgID = orgID
	}
	now := time.Now().UTC()
	def.CreatedAt = now
	def.UpdatedAt = now

	var optionsJSON any
	if len(def.Options) > 0 {
		b, err := json.Marshal(def.Options)
		if err != nil {
			return nil, fmt.Errorf("marshal options: %w", err)
		}
		optionsJSON = json.RawMessage(b)
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO custom_field_definitions
			(id, org_id, entity_type, name, label, field_type, options, required, order_idx, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+cfdCols,
		def.ID, def.OrgID, def.EntityType, def.Name, def.Label, def.FieldType,
		optionsJSON, def.Required, def.OrderIdx, def.CreatedAt, def.UpdatedAt,
	)
	created, err := scanCFD(row)
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		return nil, fmt.Errorf("%w: custom field name already exists for this entity type", domain.ErrConflict)
	}
	return created, err
}

func (r *CustomFieldDefinitionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomFieldDefinition, error) {
	q := `SELECT ` + cfdCols + ` FROM custom_field_definitions WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	return scanCFD(r.db.QueryRow(ctx, q, args...))
}

func (r *CustomFieldDefinitionRepo) Update(ctx context.Context, id uuid.UUID, patch domain.CustomFieldDefinitionPatch) (*domain.CustomFieldDefinition, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.Label != nil {
		addArg("label", *patch.Label)
	}
	if patch.Required != nil {
		addArg("required", *patch.Required)
	}
	if patch.OrderIdx != nil {
		addArg("order_idx", *patch.OrderIdx)
	}
	if patch.Options != nil {
		b, err := json.Marshal(patch.Options)
		if err != nil {
			return nil, fmt.Errorf("marshal options: %w", err)
		}
		addArg("options", json.RawMessage(b))
	}

	args = append(args, id)
	whereClause := fmt.Sprintf("WHERE id = $%d AND deleted_at IS NULL", i)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(" AND org_id = $%d", i)
		args = append(args, orgID)
	}

	q := fmt.Sprintf(`UPDATE custom_field_definitions SET %s %s RETURNING `+cfdCols,
		strings.Join(sets, ", "), whereClause)
	updated, err := scanCFD(r.db.QueryRow(ctx, q, args...))
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		return nil, fmt.Errorf("%w: custom field name already exists for this entity type", domain.ErrConflict)
	}
	return updated, err
}

func (r *CustomFieldDefinitionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE custom_field_definitions SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $2`
		args = append(args, orgID)
	}
	tag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CustomFieldDefinitionRepo) List(ctx context.Context, filter domain.CustomFieldDefinitionFilter) ([]*domain.CustomFieldDefinition, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []any{}
	i := 1

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		conditions = append(conditions, fmt.Sprintf("org_id = $%d", i))
		args = append(args, orgID)
		i++
	} else if filter.OrgID != (uuid.UUID{}) {
		conditions = append(conditions, fmt.Sprintf("org_id = $%d", i))
		args = append(args, filter.OrgID)
		i++
	}

	if filter.EntityType != nil {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", i))
		args = append(args, *filter.EntityType)
		i++
	}
	_ = i

	q := fmt.Sprintf(`SELECT `+cfdCols+` FROM custom_field_definitions WHERE %s ORDER BY order_idx ASC, created_at ASC`,
		strings.Join(conditions, " AND "))

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defs []*domain.CustomFieldDefinition
	for rows.Next() {
		d, err := scanCFD(rows)
		if err != nil {
			return nil, err
		}
		defs = append(defs, d)
	}
	return defs, rows.Err()
}
