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
	"github.com/omnir/crm-api/internal/repository"
)

// AutomationRepo implements repository.AutomationRepository.
type AutomationRepo struct {
	db *pgxpool.Pool
}

func NewAutomationRepo(db *pgxpool.Pool) *AutomationRepo {
	return &AutomationRepo{db: db}
}

// scanAutomation reads an automation row, unmarshaling the JSONB columns.
func scanAutomation(row pgx.Row) (*domain.Automation, error) {
	var a domain.Automation
	var triggerJSON, conditionsJSON, actionsJSON []byte
	var runCount int

	err := row.Scan(
		&a.ID, &a.OrgID, &a.Name, &a.Description, &a.Status,
		&triggerJSON, &conditionsJSON, &actionsJSON,
		&a.CreatedBy, &runCount,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(triggerJSON, &a.Trigger); err != nil {
		return nil, fmt.Errorf("unmarshal trigger: %w", err)
	}
	if err := json.Unmarshal(conditionsJSON, &a.Conditions); err != nil {
		return nil, fmt.Errorf("unmarshal conditions: %w", err)
	}
	if err := json.Unmarshal(actionsJSON, &a.Actions); err != nil {
		return nil, fmt.Errorf("unmarshal actions: %w", err)
	}
	a.RunCount = runCount
	return &a, nil
}

const automationSelect = `
SELECT a.id, a.org_id, a.name, a.description, a.status,
       a.trigger_config, a.conditions, a.actions,
       a.created_by,
       (SELECT COUNT(*) FROM automation_runs r WHERE r.automation_id = a.id) AS run_count,
       a.created_at, a.updated_at
FROM automations a`

func (r *AutomationRepo) CreateAutomation(ctx context.Context, a *domain.Automation) (*domain.Automation, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		a.OrgID = orgID
	}
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	if a.Status == "" {
		a.Status = domain.AutomationStatusDraft
	}
	if a.Conditions == nil {
		a.Conditions = []domain.AutomationCondition{}
	}

	triggerJSON, err := json.Marshal(a.Trigger)
	if err != nil {
		return nil, err
	}
	condJSON, err := json.Marshal(a.Conditions)
	if err != nil {
		return nil, err
	}
	actJSON, err := json.Marshal(a.Actions)
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO automations
			(id, org_id, name, description, status, trigger_type, trigger_config, conditions, actions, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id`,
		a.ID, a.OrgID, a.Name, a.Description, a.Status,
		a.Trigger.Type, triggerJSON, condJSON, actJSON,
		a.CreatedBy, a.CreatedAt, a.UpdatedAt,
	)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return r.GetAutomation(ctx, id)
}

func (r *AutomationRepo) GetAutomation(ctx context.Context, id uuid.UUID) (*domain.Automation, error) {
	q := automationSelect + ` WHERE a.id = $1`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND a.org_id = $2`
		args = append(args, orgID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	return scanAutomation(row)
}

func (r *AutomationRepo) ListAutomations(ctx context.Context, f domain.AutomationFilter) ([]*domain.Automation, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"1=1"}
	args := []any{}
	i := 1

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		where = append(where, fmt.Sprintf("a.org_id = $%d", i))
		args = append(args, orgID)
		i++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("a.status = $%d", i))
		args = append(args, *f.Status)
		i++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM automations a WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		automationSelect+` WHERE `+whereClause+
			fmt.Sprintf(` ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d`, i, i+1),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*domain.Automation
	for rows.Next() {
		a, err := scanAutomation(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, a)
	}
	return list, total, rows.Err()
}

func (r *AutomationRepo) UpdateAutomation(ctx context.Context, id uuid.UUID, req domain.UpdateAutomationRequest) (*domain.Automation, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if req.Name != nil {
		addArg("name", *req.Name)
	}
	if req.Description != nil {
		addArg("description", *req.Description)
	}
	if req.Status != nil {
		addArg("status", *req.Status)
	}
	if req.Trigger != nil {
		b, err := json.Marshal(req.Trigger)
		if err != nil {
			return nil, err
		}
		addArg("trigger_type", req.Trigger.Type)
		addArg("trigger_config", b)
	}
	if req.Conditions != nil {
		b, err := json.Marshal(req.Conditions)
		if err != nil {
			return nil, err
		}
		addArg("conditions", b)
	}
	if req.Actions != nil {
		b, err := json.Marshal(req.Actions)
		if err != nil {
			return nil, err
		}
		addArg("actions", b)
	}

	whereClause := fmt.Sprintf("id = $%d", i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(" AND org_id = $%d", i)
		args = append(args, orgID)
	}

	_, err := r.db.Exec(ctx,
		fmt.Sprintf(`UPDATE automations SET %s WHERE %s`,
			strings.Join(sets, ", "), whereClause),
		args...,
	)
	if err != nil {
		return nil, err
	}
	return r.GetAutomation(ctx, id)
}

func (r *AutomationRepo) DeleteAutomation(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM automations WHERE id = $1`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $2`
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

// --- Runs ---

func (r *AutomationRepo) CreateRun(ctx context.Context, run *domain.AutomationRun) (*domain.AutomationRun, error) {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	now := time.Now().UTC()
	run.CreatedAt = now
	run.Status = domain.RunStatusRunning
	run.StartedAt = &now

	row := r.db.QueryRow(ctx, `
		INSERT INTO automation_runs
			(id, automation_id, org_id, status, entity_type, entity_id, started_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, automation_id, org_id, status, entity_type, entity_id,
		          error_message, started_at, finished_at, created_at`,
		run.ID, run.AutomationID, run.OrgID, run.Status,
		run.EntityType, run.EntityID, run.StartedAt, run.CreatedAt,
	)
	return scanRun(row)
}

func (r *AutomationRepo) UpdateRun(ctx context.Context, id uuid.UUID, status domain.AutomationRunStatus, errMsg string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, `
		UPDATE automation_runs
		SET status = $1, error_message = $2, finished_at = $3
		WHERE id = $4`,
		status, errMsg, now, id,
	)
	return err
}

func (r *AutomationRepo) ListRuns(ctx context.Context, f domain.AutomationRunFilter) ([]*domain.AutomationRun, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM automation_runs WHERE automation_id = $1`,
		f.AutomationID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, automation_id, org_id, status, entity_type, entity_id,
		       error_message, started_at, finished_at, created_at
		FROM automation_runs
		WHERE automation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		f.AutomationID, f.Limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*domain.AutomationRun
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, run)
	}
	return list, total, rows.Err()
}

func scanRun(row pgx.Row) (*domain.AutomationRun, error) {
	var run domain.AutomationRun
	var errMsg *string
	err := row.Scan(
		&run.ID, &run.AutomationID, &run.OrgID, &run.Status,
		&run.EntityType, &run.EntityID,
		&errMsg, &run.StartedAt, &run.FinishedAt, &run.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if errMsg != nil {
		run.ErrorMessage = *errMsg
	}
	return &run, nil
}

// --- Worker helpers ---

func (r *AutomationRepo) ListActiveByTrigger(ctx context.Context, orgID uuid.UUID, triggerType domain.TriggerType) ([]*domain.Automation, error) {
	rows, err := r.db.Query(ctx,
		automationSelect+` WHERE a.org_id = $1 AND a.status = 'active' AND a.trigger_type = $2
		ORDER BY a.created_at ASC`,
		orgID, triggerType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Automation
	for rows.Next() {
		a, err := scanAutomation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

// OverdueActivityIDs returns activities overdue before cutoff that have no
// existing activity_overdue automation run in the last 24 hours (dedup window).
func (r *AutomationRepo) OverdueActivityIDs(ctx context.Context, cutoff time.Time, limit int) ([]repository.OverdueActivityRef, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.org_id, a.owner_id
		FROM activities a
		WHERE a.due_date <= $1
		  AND a.completed_at IS NULL
		  AND a.deleted_at IS NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM automation_runs ar
		      WHERE ar.entity_id = a.id
		        AND ar.entity_type = 'activity'
		        AND ar.created_at >= NOW() - INTERVAL '24 hours'
		  )
		ORDER BY a.due_date ASC
		LIMIT $2`,
		cutoff, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []repository.OverdueActivityRef
	for rows.Next() {
		var ref repository.OverdueActivityRef
		if err := rows.Scan(&ref.ActivityID, &ref.OrgID, &ref.OwnerID); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}
