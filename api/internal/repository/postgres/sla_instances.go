package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// SLAInstanceRepo implements repository.SLAInstanceRepository.
type SLAInstanceRepo struct {
	db *pgxpool.Pool
}

func NewSLAInstanceRepo(db *pgxpool.Pool) *SLAInstanceRepo {
	return &SLAInstanceRepo{db: db}
}

const slaInstanceCols = `id, org_id, policy_id, entity_id, entity_type, started_at,
	response_due_at, resolution_due_at, responded_at, resolved_at,
	breached, breach_type, warned_at, created_at`

func scanSLAInstance(row pgx.Row) (*domain.SLAInstance, error) {
	var inst domain.SLAInstance
	err := row.Scan(
		&inst.ID, &inst.OrgID, &inst.PolicyID,
		&inst.EntityID, &inst.EntityType,
		&inst.StartedAt, &inst.ResponseDueAt, &inst.ResolutionDueAt,
		&inst.RespondedAt, &inst.ResolvedAt,
		&inst.Breached, &inst.BreachType, &inst.WarnedAt,
		&inst.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	inst.ComputeStatus()
	return &inst, nil
}

func (r *SLAInstanceRepo) Create(ctx context.Context, inst *domain.SLAInstance) (*domain.SLAInstance, error) {
	if inst.ID == uuid.Nil {
		inst.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		inst.OrgID = orgID
	}
	if inst.BreachType == "" {
		inst.BreachType = domain.SLABreachTypeNone
	}
	now := time.Now().UTC()
	inst.StartedAt = now
	inst.CreatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO sla_instances
			(id, org_id, policy_id, entity_id, entity_type, started_at,
			 response_due_at, resolution_due_at, responded_at, resolved_at,
			 breached, breach_type, warned_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING `+slaInstanceCols,
		inst.ID, inst.OrgID, inst.PolicyID,
		inst.EntityID, string(inst.EntityType),
		inst.StartedAt, inst.ResponseDueAt, inst.ResolutionDueAt,
		inst.RespondedAt, inst.ResolvedAt,
		inst.Breached, string(inst.BreachType), inst.WarnedAt,
		inst.CreatedAt,
	)
	return scanSLAInstance(row)
}

func (r *SLAInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SLAInstance, error) {
	q := `SELECT ` + slaInstanceCols + ` FROM sla_instances WHERE id=$1`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	return scanSLAInstance(r.db.QueryRow(ctx, q, args...))
}

func (r *SLAInstanceRepo) List(ctx context.Context, filter domain.SLAInstanceFilter) ([]*domain.SLAInstance, error) {
	where := []string{"1=1"}
	args := []any{}
	i := 1

	addCond := func(cond string, val any) {
		where = append(where, fmt.Sprintf(cond, i))
		args = append(args, val)
		i++
	}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		filter.OrgID = orgID
	}
	if filter.OrgID != uuid.Nil {
		addCond("org_id=$%d", filter.OrgID)
	}
	if filter.EntityID != nil {
		addCond("entity_id=$%d", *filter.EntityID)
	}
	if filter.EntityType != "" {
		addCond("entity_type=$%d", string(filter.EntityType))
	}
	if filter.Breached != nil {
		addCond("breached=$%d", *filter.Breached)
	}

	q := fmt.Sprintf(`SELECT %s FROM sla_instances WHERE %s ORDER BY created_at DESC`,
		slaInstanceCols, joinWhere(where))

	if filter.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d", i)
		args = append(args, filter.Limit)
		i++
	}
	if filter.Offset > 0 {
		q += fmt.Sprintf(" OFFSET $%d", i)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*domain.SLAInstance
	for rows.Next() {
		var inst domain.SLAInstance
		if err := rows.Scan(
			&inst.ID, &inst.OrgID, &inst.PolicyID,
			&inst.EntityID, &inst.EntityType,
			&inst.StartedAt, &inst.ResponseDueAt, &inst.ResolutionDueAt,
			&inst.RespondedAt, &inst.ResolvedAt,
			&inst.Breached, &inst.BreachType, &inst.WarnedAt,
			&inst.CreatedAt,
		); err != nil {
			return nil, err
		}
		inst.ComputeStatus()
		instances = append(instances, &inst)
	}
	return instances, rows.Err()
}

func (r *SLAInstanceRepo) MarkResponded(ctx context.Context, entityID uuid.UUID, entityType domain.SLAEntityType, t time.Time) error {
	q := `UPDATE sla_instances SET responded_at=$1
		WHERE entity_id=$2 AND entity_type=$3 AND responded_at IS NULL`
	args := []any{t, entityID, string(entityType)}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$4`
		args = append(args, orgID)
	}
	_, err := r.db.Exec(ctx, q, args...)
	return err
}

func (r *SLAInstanceRepo) MarkResolved(ctx context.Context, entityID uuid.UUID, entityType domain.SLAEntityType, t time.Time) error {
	q := `UPDATE sla_instances SET resolved_at=$1
		WHERE entity_id=$2 AND entity_type=$3 AND resolved_at IS NULL`
	args := []any{t, entityID, string(entityType)}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$4`
		args = append(args, orgID)
	}
	_, err := r.db.Exec(ctx, q, args...)
	return err
}

// ScanBreaches marks open instances as breached where response_due_at or resolution_due_at < now.
// Uses a superuser/background context (no RLS org filter), so this must only be called
// from background workers that have admin-level DB access.
func (r *SLAInstanceRepo) ScanBreaches(ctx context.Context) ([]*domain.SLAInstance, error) {
	rows, err := r.db.Query(ctx, `
		UPDATE sla_instances SET
			breached    = TRUE,
			breach_type = CASE
				WHEN responded_at IS NULL AND response_due_at < NOW() THEN 'response'
				ELSE 'resolution'
			END
		WHERE resolved_at IS NULL
		  AND breached = FALSE
		  AND (
			(responded_at IS NULL AND response_due_at < NOW())
			OR resolution_due_at < NOW()
		  )
		RETURNING `+slaInstanceCols)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*domain.SLAInstance
	for rows.Next() {
		inst, err := scanSLAInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// NotificationRecipients returns users who should receive SLA notifications for
// the instance's entity. It intentionally stays org-scoped and deduplicated.
func (r *SLAInstanceRepo) NotificationRecipients(ctx context.Context, inst *domain.SLAInstance) ([]uuid.UUID, error) {
	if inst == nil {
		return nil, nil
	}

	var query string
	switch inst.EntityType {
	case domain.SLAEntityTypeTicket:
		query = `
			SELECT DISTINCT user_id
			FROM (
				SELECT assignee_id AS user_id
				FROM tickets
				WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL AND assignee_id IS NOT NULL
				UNION
				SELECT submitted_by_user_id AS user_id
				FROM tickets
				WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL AND submitted_by_user_id IS NOT NULL
			) recipients`
	case domain.SLAEntityTypeDeal:
		query = `
			SELECT DISTINCT owner_id AS user_id
			FROM deals
			WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL`
	case domain.SLAEntityTypeContact:
		query = `
			SELECT DISTINCT owner_id AS user_id
			FROM contacts
			WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL`
	case domain.SLAEntityTypeActivity:
		query = `
			SELECT DISTINCT owner_id AS user_id
			FROM activities
			WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL`
	default:
		return nil, nil
	}

	rows, err := r.db.Query(ctx, query, inst.EntityID, inst.OrgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipients := []uuid.UUID{}
	seen := map[uuid.UUID]struct{}{}
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		if userID == uuid.Nil {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		recipients = append(recipients, userID)
	}
	return recipients, rows.Err()
}

// ScanWarnings returns open instances where 80% of the resolution window has elapsed and no warning sent yet.
func (r *SLAInstanceRepo) ScanWarnings(ctx context.Context) ([]*domain.SLAInstance, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+slaInstanceCols+`
		FROM sla_instances
		WHERE resolved_at IS NULL
		  AND breached = FALSE
		  AND warned_at IS NULL
		  AND NOW() >= started_at + (resolution_due_at - started_at) * 0.8
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*domain.SLAInstance
	for rows.Next() {
		var inst domain.SLAInstance
		if err := rows.Scan(
			&inst.ID, &inst.OrgID, &inst.PolicyID,
			&inst.EntityID, &inst.EntityType,
			&inst.StartedAt, &inst.ResponseDueAt, &inst.ResolutionDueAt,
			&inst.RespondedAt, &inst.ResolvedAt,
			&inst.Breached, &inst.BreachType, &inst.WarnedAt,
			&inst.CreatedAt,
		); err != nil {
			return nil, err
		}
		inst.ComputeStatus()
		instances = append(instances, &inst)
	}
	return instances, rows.Err()
}

func (r *SLAInstanceRepo) MarkWarned(ctx context.Context, id uuid.UUID, t time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE sla_instances SET warned_at=$1 WHERE id=$2`, t, id)
	return err
}

func (r *SLAInstanceRepo) Dashboard(ctx context.Context) (*domain.SLADashboard, error) {
	q := `
		SELECT
			COUNT(*) FILTER (WHERE NOT breached AND resolved_at IS NULL
				AND NOW() < started_at + (resolution_due_at - started_at) * 0.8) AS on_track,
			COUNT(*) FILTER (WHERE NOT breached AND resolved_at IS NULL
				AND NOW() >= started_at + (resolution_due_at - started_at) * 0.8) AS at_risk,
			COUNT(*) FILTER (WHERE breached = TRUE) AS breached
		FROM sla_instances
	`
	args := []any{}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` WHERE org_id=$1`
		args = append(args, orgID)
	}

	var d domain.SLADashboard
	err := r.db.QueryRow(ctx, q, args...).Scan(&d.OnTrack, &d.AtRisk, &d.Breached)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// joinWhere joins WHERE conditions with AND.
func joinWhere(conds []string) string {
	result := ""
	for i, c := range conds {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return result
}
