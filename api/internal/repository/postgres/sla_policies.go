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

// SLAPolicyRepo implements repository.SLAPolicyRepository.
type SLAPolicyRepo struct {
	db *pgxpool.Pool
}

func NewSLAPolicyRepo(db *pgxpool.Pool) *SLAPolicyRepo {
	return &SLAPolicyRepo{db: db}
}

const slaPolicyCols = `id, org_id, name, response_time_hours, resolution_time_hours, priority_filter, created_at, updated_at`

func scanSLAPolicy(row pgx.Row) (*domain.SLAPolicy, error) {
	var p domain.SLAPolicy
	var priorityJSON []byte
	err := row.Scan(
		&p.ID, &p.OrgID, &p.Name,
		&p.ResponseTimeHours, &p.ResolutionTimeHours,
		&priorityJSON, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(priorityJSON, &p.PriorityFilter); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *SLAPolicyRepo) Create(ctx context.Context, p *domain.SLAPolicy) (*domain.SLAPolicy, error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		p.OrgID = orgID
	}
	if p.PriorityFilter == nil {
		p.PriorityFilter = []domain.TicketPriority{}
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	priorityJSON, err := json.Marshal(p.PriorityFilter)
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO sla_policies
			(id, org_id, name, response_time_hours, resolution_time_hours, priority_filter, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING `+slaPolicyCols,
		p.ID, p.OrgID, p.Name,
		p.ResponseTimeHours, p.ResolutionTimeHours,
		priorityJSON, p.CreatedAt, p.UpdatedAt,
	)
	return scanSLAPolicy(row)
}

func (r *SLAPolicyRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SLAPolicy, error) {
	q := `SELECT ` + slaPolicyCols + ` FROM sla_policies WHERE id=$1`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	return scanSLAPolicy(r.db.QueryRow(ctx, q, args...))
}

func (r *SLAPolicyRepo) Update(ctx context.Context, id uuid.UUID, patch domain.SLAPolicyPatch) (*domain.SLAPolicy, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.Name != nil {
		addArg("name", *patch.Name)
	}
	if patch.ResponseTimeHours != nil {
		addArg("response_time_hours", *patch.ResponseTimeHours)
	}
	if patch.ResolutionTimeHours != nil {
		addArg("resolution_time_hours", *patch.ResolutionTimeHours)
	}
	if patch.PriorityFilter != nil {
		priorityJSON, err := json.Marshal(patch.PriorityFilter)
		if err != nil {
			return nil, err
		}
		addArg("priority_filter", priorityJSON)
	}

	whereClause := fmt.Sprintf(`id=$%d`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
	}

	query := fmt.Sprintf(
		`UPDATE sla_policies SET %s WHERE %s RETURNING %s`,
		strings.Join(sets, ", "), whereClause, slaPolicyCols,
	)
	return scanSLAPolicy(r.db.QueryRow(ctx, query, args...))
}

func (r *SLAPolicyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM sla_policies WHERE id=$1`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SLAPolicyRepo) List(ctx context.Context, orgID uuid.UUID) ([]*domain.SLAPolicy, error) {
	ctxOrg, hasCtx := domain.OrgIDFromContext(ctx)
	if hasCtx {
		orgID = ctxOrg
	}

	q := `SELECT ` + slaPolicyCols + ` FROM sla_policies`
	args := []any{}
	if orgID != uuid.Nil {
		q += ` WHERE org_id=$1`
		args = append(args, orgID)
	}
	q += ` ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*domain.SLAPolicy
	for rows.Next() {
		var p domain.SLAPolicy
		var priorityJSON []byte
		if err := rows.Scan(
			&p.ID, &p.OrgID, &p.Name,
			&p.ResponseTimeHours, &p.ResolutionTimeHours,
			&priorityJSON, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(priorityJSON, &p.PriorityFilter); err != nil {
			return nil, err
		}
		policies = append(policies, &p)
	}
	return policies, rows.Err()
}

func (r *SLAPolicyRepo) MatchByPriority(ctx context.Context, priority domain.TicketPriority) (*domain.SLAPolicy, error) {
	q := `SELECT ` + slaPolicyCols + `
		FROM sla_policies
		WHERE priority_filter @> jsonb_build_array($1::text)`
	args := []any{string(priority)}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	q += ` ORDER BY created_at LIMIT 1`

	p, err := scanSLAPolicy(r.db.QueryRow(ctx, q, args...))
	if errors.Is(err, domain.ErrNotFound) {
		return nil, nil
	}
	return p, err
}
