package postgres

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// ReportsRepo implements repository.ReportsRepository against PostgreSQL.
type ReportsRepo struct {
	db *pgxpool.Pool
}

func NewReportsRepo(db *pgxpool.Pool) *ReportsRepo {
	return &ReportsRepo{db: db}
}

// reportArgs builds the common WHERE clause args for date-ranged report queries.
// It always pins the org_id as $1, then optionally appends args for from/to dates.
// Returns the arg slice and the SQL fragment to append (e.g. " AND t.created_at >= $2").
func reportArgs(orgID interface{}, f domain.ReportFilter, alias string) ([]interface{}, string) {
	args := []interface{}{orgID}
	var sb strings.Builder
	if f.From != nil {
		args = append(args, *f.From)
		sb.WriteString(" AND " + alias + ".created_at >= $" + strconv.Itoa(len(args)))
	}
	if f.To != nil {
		args = append(args, *f.To)
		sb.WriteString(" AND " + alias + ".created_at <= $" + strconv.Itoa(len(args)))
	}
	return args, sb.String()
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

// TicketMetrics returns aggregated ticket stats for the given date/org filter.
func (r *ReportsRepo) TicketMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.TicketReport, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	if filter.OrgID != nil {
		orgID = *filter.OrgID
	}

	args, dateClause := reportArgs(orgID, filter, "t")

	// Per-status counts.
	statusRows, err := r.db.Query(ctx, `
		SELECT status, COUNT(*) AS count
		FROM tickets t
		WHERE t.org_id = $1 AND t.deleted_at IS NULL`+dateClause+`
		GROUP BY status
		ORDER BY status
	`, args...)
	if err != nil {
		return nil, err
	}
	defer statusRows.Close()

	var byStatus []domain.TicketStatusCount
	totalOpen, totalClosed := 0, 0
	for statusRows.Next() {
		var sc domain.TicketStatusCount
		if err := statusRows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, err
		}
		byStatus = append(byStatus, sc)
		switch sc.Status {
		case domain.TicketStatusOpen, domain.TicketStatusInProgress, domain.TicketStatusPending:
			totalOpen += sc.Count
		case domain.TicketStatusResolved, domain.TicketStatusClosed:
			totalClosed += sc.Count
		}
	}
	if err := statusRows.Err(); err != nil {
		return nil, err
	}
	if byStatus == nil {
		byStatus = []domain.TicketStatusCount{}
	}

	// Average resolution time (hours) for resolved/closed tickets.
	var avgHours float64
	if err := r.db.QueryRow(ctx, `
		SELECT COALESCE(
			AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 3600.0),
			0
		)
		FROM tickets t
		WHERE t.org_id = $1
		  AND t.deleted_at IS NULL
		  AND t.status IN ('resolved', 'closed')`+dateClause,
		args...,
	).Scan(&avgHours); err != nil {
		return nil, err
	}

	// Breach rate: % of open tickets older than 48 hours.
	// Build separate args so the breach threshold doesn't conflict with dateClause positions.
	breachArgs, breachDateClause := reportArgs(orgID, filter, "t")
	breachArgs = append(breachArgs, time.Now().Add(-48*time.Hour))
	breachThreshIdx := "$" + strconv.Itoa(len(breachArgs))

	var openTotal, openBreached int
	if err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE t.status IN ('open','in_progress','pending')) AS open_total,
			COUNT(*) FILTER (WHERE t.status IN ('open','in_progress','pending') AND t.created_at <= `+breachThreshIdx+`) AS open_breached
		FROM tickets t
		WHERE t.org_id = $1 AND t.deleted_at IS NULL`+breachDateClause,
		breachArgs...,
	).Scan(&openTotal, &openBreached); err != nil {
		return nil, err
	}

	var breachRate float64
	if openTotal > 0 {
		breachRate = float64(openBreached) / float64(openTotal) * 100
	}

	return &domain.TicketReport{
		TotalOpen:          totalOpen,
		TotalClosed:        totalClosed,
		AvgResolutionHours: avgHours,
		ByStatus:           byStatus,
		BreachRate:         breachRate,
	}, nil
}

// ContactMetrics returns aggregated contact stats for the given date/org filter.
func (r *ReportsRepo) ContactMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.ContactReport, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	if filter.OrgID != nil {
		orgID = *filter.OrgID
	}

	args, dateClause := reportArgs(orgID, filter, "c")

	var total int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM contacts c
		WHERE c.org_id = $1 AND c.deleted_at IS NULL`+dateClause,
		args...,
	).Scan(&total); err != nil {
		return nil, err
	}

	periodRows, err := r.db.Query(ctx, `
		SELECT TO_CHAR(c.created_at, 'YYYY-MM') AS period, COUNT(*) AS count
		FROM contacts c
		WHERE c.org_id = $1 AND c.deleted_at IS NULL`+dateClause+`
		GROUP BY period
		ORDER BY period
	`, args...)
	if err != nil {
		return nil, err
	}
	defer periodRows.Close()

	var byPeriod []domain.ContactPeriodMetric
	for periodRows.Next() {
		var m domain.ContactPeriodMetric
		if err := periodRows.Scan(&m.Period, &m.Count); err != nil {
			return nil, err
		}
		byPeriod = append(byPeriod, m)
	}
	if err := periodRows.Err(); err != nil {
		return nil, err
	}
	if byPeriod == nil {
		byPeriod = []domain.ContactPeriodMetric{}
	}

	return &domain.ContactReport{
		Total:    total,
		ByPeriod: byPeriod,
	}, nil
}

// DealMetrics returns aggregated deal stats for the given date/org filter.
func (r *ReportsRepo) DealMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.DealReport, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	if filter.OrgID != nil {
		orgID = *filter.OrgID
	}

	args, dateClause := reportArgs(orgID, filter, "d")

	stageRows, err := r.db.Query(ctx, `
		SELECT stage, COUNT(*) AS count, COALESCE(SUM(value_cents), 0) AS total_value_cents
		FROM deals d
		WHERE d.org_id = $1 AND d.deleted_at IS NULL`+dateClause+`
		GROUP BY stage
		ORDER BY stage
	`, args...)
	if err != nil {
		return nil, err
	}
	defer stageRows.Close()

	var byStage []domain.DealStageCount
	var pipelineValue int64
	var wonCount, lostCount int
	for stageRows.Next() {
		var sc domain.DealStageCount
		if err := stageRows.Scan(&sc.Stage, &sc.Count, &sc.TotalValueCents); err != nil {
			return nil, err
		}
		byStage = append(byStage, sc)
		switch sc.Stage {
		case domain.DealStageClosedWon:
			wonCount = sc.Count
		case domain.DealStageClosedLost:
			lostCount = sc.Count
		default:
			pipelineValue += sc.TotalValueCents
		}
	}
	if err := stageRows.Err(); err != nil {
		return nil, err
	}
	if byStage == nil {
		byStage = []domain.DealStageCount{}
	}

	return &domain.DealReport{
		PipelineValueCents: pipelineValue,
		WonCount:           wonCount,
		LostCount:          lostCount,
		ByStage:            byStage,
	}, nil
}

// LeadMetrics returns aggregated lead stats for the given date/org filter.
func (r *ReportsRepo) LeadMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.LeadReport, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	if filter.OrgID != nil {
		orgID = *filter.OrgID
	}

	args, dateClause := reportArgs(orgID, filter, "l")

	var totalNew, totalConverted int
	if err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE l.status = 'converted' OR l.converted_contact_id IS NOT NULL) AS converted
		FROM leads l
		WHERE l.org_id = $1 AND l.deleted_at IS NULL`+dateClause,
		args...,
	).Scan(&totalNew, &totalConverted); err != nil {
		return nil, err
	}

	var conversionRate float64
	if totalNew > 0 {
		conversionRate = float64(totalConverted) / float64(totalNew) * 100
	}

	return &domain.LeadReport{
		TotalNew:       totalNew,
		TotalConverted: totalConverted,
		ConversionRate: conversionRate,
	}, nil
}
