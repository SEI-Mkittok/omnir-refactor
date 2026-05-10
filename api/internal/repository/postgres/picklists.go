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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type PicklistRepo struct {
	db *pgxpool.Pool
}

func NewPicklistRepo(db *pgxpool.Pool) *PicklistRepo {
	return &PicklistRepo{db: db}
}

const picklistCols = `id, org_id, custom_field_id, value, display_label, order_idx, is_active, created_at, updated_at`
const picklistDepCols = `id, org_id, entity_type, source_field_id, target_field_id, mapping, is_active, created_at, updated_at`

func scanPicklistValue(row pgx.Row) (*domain.PicklistValue, error) {
	var p domain.PicklistValue
	err := row.Scan(
		&p.ID, &p.OrgID, &p.CustomFieldID, &p.Value, &p.DisplayLabel, &p.OrderIdx,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func scanPicklistDependency(row pgx.Row) (*domain.PicklistDependency, error) {
	var d domain.PicklistDependency
	var raw []byte
	err := row.Scan(
		&d.ID, &d.OrgID, &d.EntityType, &d.SourceFieldID, &d.TargetFieldID,
		&raw, &d.IsActive, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	d.Mapping = map[string][]string{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &d.Mapping)
	}
	return &d, nil
}

func (r *PicklistRepo) fieldDefinition(ctx context.Context, orgID, customFieldID uuid.UUID) (*domain.CustomFieldDefinition, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, org_id, entity_type, name, label, field_type, options, required, order_idx, created_at, updated_at, deleted_at
		FROM custom_field_definitions
		WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL
	`, customFieldID, orgID)
	var optionsRaw []byte
	def := &domain.CustomFieldDefinition{}
	err := row.Scan(
		&def.ID, &def.OrgID, &def.EntityType, &def.Name, &def.Label, &def.FieldType,
		&optionsRaw, &def.Required, &def.OrderIdx, &def.CreatedAt, &def.UpdatedAt, &def.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(optionsRaw) > 0 {
		_ = json.Unmarshal(optionsRaw, &def.Options)
	}
	return def, nil
}

func (r *PicklistRepo) bootstrapFromDefinition(ctx context.Context, tx pgx.Tx, def *domain.CustomFieldDefinition) error {
	for i, opt := range def.Options {
		value := strings.TrimSpace(opt)
		if value == "" {
			continue
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO picklist_values (org_id, custom_field_id, value, display_label, order_idx, is_active)
			VALUES ($1,$2,$3,$4,$5,TRUE)
			ON CONFLICT (org_id, custom_field_id, value) DO NOTHING
		`, def.OrgID, def.ID, value, value, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PicklistRepo) ListValues(ctx context.Context, orgID, customFieldID uuid.UUID) ([]*domain.PicklistValue, error) {
	def, err := r.fieldDefinition(ctx, orgID, customFieldID)
	if err != nil {
		return nil, err
	}

	if def.FieldType != domain.CustomFieldTypeSelect && def.FieldType != domain.CustomFieldTypeMultiSelect {
		return []*domain.PicklistValue{}, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := r.bootstrapFromDefinition(ctx, tx, def); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT `+picklistCols+`
		FROM picklist_values
		WHERE org_id = $1 AND custom_field_id = $2
		ORDER BY order_idx ASC, created_at ASC
	`, orgID, customFieldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*domain.PicklistValue{}
	for rows.Next() {
		v, err := scanPicklistValue(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func normalizePicklistInputs(values []domain.PicklistValueInput) []domain.PicklistValueInput {
	out := make([]domain.PicklistValueInput, 0, len(values))
	seen := map[string]struct{}{}
	for i, v := range values {
		val := strings.TrimSpace(v.Value)
		if val == "" {
			continue
		}
		if _, exists := seen[val]; exists {
			continue
		}
		seen[val] = struct{}{}
		label := strings.TrimSpace(v.DisplayLabel)
		if label == "" {
			label = val
		}
		out = append(out, domain.PicklistValueInput{
			Value:        val,
			DisplayLabel: label,
			OrderIdx:     i,
			IsActive:     v.IsActive,
		})
	}
	return out
}

func (r *PicklistRepo) UpsertValues(ctx context.Context, orgID, customFieldID uuid.UUID, values []domain.PicklistValueInput) ([]*domain.PicklistValue, error) {
	def, err := r.fieldDefinition(ctx, orgID, customFieldID)
	if err != nil {
		return nil, err
	}
	if def.FieldType != domain.CustomFieldTypeSelect && def.FieldType != domain.CustomFieldTypeMultiSelect {
		return nil, fmt.Errorf("%w: field is not a picklist", domain.ErrValidation)
	}

	values = normalizePicklistInputs(values)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := r.bootstrapFromDefinition(ctx, tx, def); err != nil {
		return nil, err
	}

	for _, v := range values {
		_, err = tx.Exec(ctx, `
			INSERT INTO picklist_values (org_id, custom_field_id, value, display_label, order_idx, is_active, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,NOW())
			ON CONFLICT (org_id, custom_field_id, value) DO UPDATE SET
				display_label = EXCLUDED.display_label,
				order_idx = EXCLUDED.order_idx,
				is_active = EXCLUDED.is_active,
				updated_at = NOW()
		`, orgID, customFieldID, v.Value, v.DisplayLabel, v.OrderIdx, v.IsActive)
		if err != nil {
			return nil, err
		}
	}

	if len(values) > 0 {
		placeholders := make([]string, 0, len(values))
		args := make([]any, 0, len(values)+2)
		args = append(args, orgID, customFieldID)
		for idx, v := range values {
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx+3))
			args = append(args, v.Value)
		}
		_, err = tx.Exec(ctx, `
			UPDATE picklist_values
			SET is_active = FALSE, updated_at = NOW()
			WHERE org_id = $1 AND custom_field_id = $2 AND value NOT IN (`+strings.Join(placeholders, ",")+`)
		`, args...)
		if err != nil {
			return nil, err
		}
	}

	// Keep definition options as the full option set (active+inactive) for backward compatibility validation.
	allRows, err := tx.Query(ctx, `
		SELECT value
		FROM picklist_values
		WHERE org_id = $1 AND custom_field_id = $2
		ORDER BY order_idx ASC, created_at ASC
	`, orgID, customFieldID)
	if err != nil {
		return nil, err
	}
	allValues := []string{}
	for allRows.Next() {
		var value string
		if err := allRows.Scan(&value); err != nil {
			allRows.Close()
			return nil, err
		}
		allValues = append(allValues, value)
	}
	allRows.Close()

	optionsRaw, err := json.Marshal(allValues)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `
		UPDATE custom_field_definitions
		SET options = $3::jsonb, updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, customFieldID, orgID, string(optionsRaw))
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT `+picklistCols+`
		FROM picklist_values
		WHERE org_id = $1 AND custom_field_id = $2
		ORDER BY order_idx ASC, created_at ASC
	`, orgID, customFieldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.PicklistValue{}
	for rows.Next() {
		item, err := scanPicklistValue(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PicklistRepo) usageCount(ctx context.Context, tx pgx.Tx, entityType domain.CustomFieldEntityType, fieldName, value string) (int, error) {
	table := ""
	switch entityType {
	case domain.CustomFieldEntityContact:
		table = "contacts"
	case domain.CustomFieldEntityAccount:
		table = "accounts"
	case domain.CustomFieldEntityLead:
		table = "leads"
	case domain.CustomFieldEntityDeal:
		table = "deals"
	case domain.CustomFieldEntityTicket:
		table = "tickets"
	default:
		return 0, nil
	}

	var count int
	q := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE deleted_at IS NULL
		  AND (
			custom_fields ->> $1 = $2
			OR EXISTS (
				SELECT 1
				FROM jsonb_array_elements_text(COALESCE(custom_fields -> $1, '[]'::jsonb)) AS elem
				WHERE elem = $2
			)
		  )
	`, table)
	if err := tx.QueryRow(ctx, q, fieldName, value).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func replaceValueInJSON(raw json.RawMessage, fieldName, fromValue string, toValue *string) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return raw
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return raw
	}
	current, ok := obj[fieldName]
	if !ok {
		return raw
	}

	switch typed := current.(type) {
	case string:
		if typed == fromValue {
			if toValue == nil {
				delete(obj, fieldName)
			} else {
				obj[fieldName] = *toValue
			}
		}
	case []any:
		next := make([]string, 0, len(typed))
		seen := map[string]struct{}{}
		for _, item := range typed {
			s, ok := item.(string)
			if !ok {
				continue
			}
			if s == fromValue {
				if toValue != nil {
					s = *toValue
				} else {
					continue
				}
			}
			if _, exists := seen[s]; exists {
				continue
			}
			seen[s] = struct{}{}
			next = append(next, s)
		}
		if len(next) == 0 {
			delete(obj, fieldName)
		} else {
			obj[fieldName] = next
		}
	}

	updated, err := json.Marshal(obj)
	if err != nil {
		return raw
	}
	return updated
}

func (r *PicklistRepo) remapTableRows(ctx context.Context, tx pgx.Tx, table, fieldName, fromValue string, toValue *string) error {
	selectSQL := fmt.Sprintf(`
		SELECT id, custom_fields
		FROM %s
		WHERE deleted_at IS NULL
		  AND (
			custom_fields ->> $1 = $2
			OR EXISTS (
				SELECT 1
				FROM jsonb_array_elements_text(COALESCE(custom_fields -> $1, '[]'::jsonb)) AS elem
				WHERE elem = $2
			)
		  )
	`, table)
	rows, err := tx.Query(ctx, selectSQL, fieldName, fromValue)
	if err != nil {
		return err
	}
	defer rows.Close()

	type item struct {
		ID   uuid.UUID
		JSON []byte
	}
	items := []item{}
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.ID, &it.JSON); err != nil {
			return err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	updateSQL := fmt.Sprintf(`UPDATE %s SET custom_fields = $2::jsonb, updated_at = NOW() WHERE id = $1`, table)
	for _, it := range items {
		next := replaceValueInJSON(it.JSON, fieldName, fromValue, toValue)
		_, err := tx.Exec(ctx, updateSQL, it.ID, string(next))
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PicklistRepo) RemapAndDeleteValue(ctx context.Context, orgID, customFieldID uuid.UUID, fromValue string, toValue *string) error {
	def, err := r.fieldDefinition(ctx, orgID, customFieldID)
	if err != nil {
		return err
	}
	fromValue = strings.TrimSpace(fromValue)
	if fromValue == "" {
		return fmt.Errorf("%w: from_value is required", domain.ErrValidation)
	}
	if toValue != nil {
		trimmed := strings.TrimSpace(*toValue)
		toValue = &trimmed
		if trimmed == "" || trimmed == fromValue {
			toValue = nil
		}
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	used, err := r.usageCount(ctx, tx, def.EntityType, def.Name, fromValue)
	if err != nil {
		return err
	}
	if used > 0 && toValue == nil {
		return fmt.Errorf("%w: value is currently in use and must be remapped before delete", domain.ErrValidation)
	}

	table := map[domain.CustomFieldEntityType]string{
		domain.CustomFieldEntityContact: "contacts",
		domain.CustomFieldEntityAccount: "accounts",
		domain.CustomFieldEntityLead:    "leads",
		domain.CustomFieldEntityDeal:    "deals",
		domain.CustomFieldEntityTicket:  "tickets",
	}[def.EntityType]
	if table != "" {
		if err := r.remapTableRows(ctx, tx, table, def.Name, fromValue, toValue); err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM picklist_values
		WHERE org_id = $1 AND custom_field_id = $2 AND value = $3
	`, orgID, customFieldID, fromValue)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `
		SELECT value
		FROM picklist_values
		WHERE org_id = $1 AND custom_field_id = $2
		ORDER BY order_idx ASC, created_at ASC
	`, orgID, customFieldID)
	if err != nil {
		return err
	}
	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			rows.Close()
			return err
		}
		values = append(values, value)
	}
	rows.Close()
	raw, err := json.Marshal(values)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE custom_field_definitions
		SET options = $3::jsonb, updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, customFieldID, orgID, string(raw))
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *PicklistRepo) ListDependencies(ctx context.Context, orgID uuid.UUID, entityType *domain.CustomFieldEntityType) ([]*domain.PicklistDependency, error) {
	q := `
		SELECT ` + picklistDepCols + `
		FROM picklist_dependencies
		WHERE org_id = $1
	`
	args := []any{orgID}
	if entityType != nil {
		q += ` AND entity_type = $2`
		args = append(args, *entityType)
	}
	q += ` ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*domain.PicklistDependency{}
	for rows.Next() {
		d, err := scanPicklistDependency(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *PicklistRepo) UpsertDependency(ctx context.Context, orgID uuid.UUID, input domain.PicklistDependencyInput) (*domain.PicklistDependency, error) {
	if !input.EntityType.IsValid() {
		return nil, fmt.Errorf("%w: invalid entity_type", domain.ErrValidation)
	}
	raw, err := json.Marshal(input.Mapping)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO picklist_dependencies (
			org_id, entity_type, source_field_id, target_field_id, mapping, is_active, updated_at
		)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6,$7)
		ON CONFLICT (org_id, source_field_id, target_field_id) DO UPDATE SET
			entity_type = EXCLUDED.entity_type,
			mapping = EXCLUDED.mapping,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
		RETURNING `+picklistDepCols,
		orgID, input.EntityType, input.SourceFieldID, input.TargetFieldID, string(raw), input.IsActive, time.Now().UTC(),
	)
	return scanPicklistDependency(row)
}

func (r *PicklistRepo) DeleteDependency(ctx context.Context, orgID, dependencyID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM picklist_dependencies WHERE id=$1 AND org_id=$2`, dependencyID, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

