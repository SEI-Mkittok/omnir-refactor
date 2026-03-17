package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type ReportsRepo struct {
	db *pgxpool.Pool
}

func NewReportsRepo(db *pgxpool.Pool) *ReportsRepo {
	return &ReportsRepo{db: db}
}

// DealsByStage returns deal count and total value_cents grouped by stage, scoped to the org.
func (r *ReportsRepo) DealsByStage(ctx context.Context) ([]domain.DealStageMetric, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}

	rows, err := r.db.Query(ctx, `
		SELECT stage, COUNT(*) AS count, COALESCE(SUM(value_cents), 0) AS total_value_cents
		FROM deals
		WHERE org_id = $1 AND deleted_at IS NULL
		GROUP BY stage
		ORDER BY stage
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []domain.DealStageMetric
	for rows.Next() {
		var m domain.DealStageMetric
		if err := rows.Scan(&m.Stage, &m.Count, &m.TotalValueCents); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	if metrics == nil {
		metrics = []domain.DealStageMetric{}
	}
	return metrics, rows.Err()
}

// ContactsMonthly returns new contact counts per month for the last 12 months.
func (r *ReportsRepo) ContactsMonthly(ctx context.Context) ([]domain.ContactMonthlyMetric, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}

	rows, err := r.db.Query(ctx, `
		SELECT TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*) AS count
		FROM contacts
		WHERE org_id = $1
		  AND deleted_at IS NULL
		  AND created_at >= DATE_TRUNC('month', NOW()) - INTERVAL '11 months'
		GROUP BY month
		ORDER BY month
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []domain.ContactMonthlyMetric
	for rows.Next() {
		var m domain.ContactMonthlyMetric
		if err := rows.Scan(&m.Month, &m.Count); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	if metrics == nil {
		metrics = []domain.ContactMonthlyMetric{}
	}
	return metrics, rows.Err()
}

// ActivitiesByType returns activity counts grouped by type, scoped to the org.
func (r *ReportsRepo) ActivitiesByType(ctx context.Context) ([]domain.ActivityTypeMetric, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}

	rows, err := r.db.Query(ctx, `
		SELECT type, COUNT(*) AS count
		FROM activities
		WHERE org_id = $1 AND deleted_at IS NULL
		GROUP BY type
		ORDER BY type
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []domain.ActivityTypeMetric
	for rows.Next() {
		var m domain.ActivityTypeMetric
		if err := rows.Scan(&m.Type, &m.Count); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	if metrics == nil {
		metrics = []domain.ActivityTypeMetric{}
	}
	return metrics, rows.Err()
}
