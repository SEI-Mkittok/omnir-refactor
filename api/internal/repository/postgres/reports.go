package postgres

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

func reportDateClause(args []interface{}, f domain.ReportFilter, alias, col string) ([]interface{}, string) {
	var sb strings.Builder
	if f.From != nil {
		args = append(args, *f.From)
		sb.WriteString(" AND " + alias + "." + col + " >= $" + strconv.Itoa(len(args)))
	}
	if f.To != nil {
		args = append(args, *f.To)
		sb.WriteString(" AND " + alias + "." + col + " <= $" + strconv.Itoa(len(args)))
	}
	return args, sb.String()
}

func appendReportReadVisibility(ctx context.Context, q *string, args *[]interface{}, module domain.ACLModule, ownerExprs ...string) {
	appendAccessVisibilitySQL(ctx, q, args, module, domain.SharingAccessRead, ownerExprs...)
}

func reportActivityParentVisibilityClause(ctx context.Context, args *[]interface{}, activityRef string) string {
	where := []string{}
	idx := len(*args) + 1
	addActivityParentVisibilityWhere(ctx, &where, args, &idx, domain.SharingAccessRead, activityRef)
	if len(where) == 0 {
		return ""
	}
	return " AND " + strings.Join(where, " AND ")
}

func cloneReportArgs(args []interface{}) []interface{} {
	cloned := make([]interface{}, len(args))
	copy(cloned, args)
	return cloned
}

func scopedReportOrgID(ctx context.Context, filter domain.ReportFilter) (uuid.UUID, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return uuid.Nil, domain.ErrNotFound
	}
	if filter.OrgID != nil {
		orgID = *filter.OrgID
	}
	return orgID, nil
}

// DealsByStage returns deal count and total value_cents grouped by stage, scoped to the org.
func (r *ReportsRepo) DealsByStage(ctx context.Context) ([]domain.DealStageMetric, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	args := []interface{}{orgID}
	q := `
		SELECT stage, COUNT(*) AS count, COALESCE(SUM(value_cents), 0) AS total_value_cents
		FROM deals
		WHERE org_id = $1 AND deleted_at IS NULL
	`
	appendReportReadVisibility(ctx, &q, &args, domain.ACLModuleDeals, "owner_id")
	q += `
		GROUP BY stage
		ORDER BY stage
	`
	rows, err := r.db.Query(ctx, q, args...)
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
	args := []interface{}{orgID}
	q := `
		SELECT TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*) AS count
		FROM contacts
		WHERE org_id = $1
		  AND deleted_at IS NULL
		  AND created_at >= DATE_TRUNC('month', NOW()) - INTERVAL '11 months'
	`
	appendReportReadVisibility(ctx, &q, &args, domain.ACLModuleContacts, "owner_id")
	q += `
		GROUP BY month
		ORDER BY month
	`
	rows, err := r.db.Query(ctx, q, args...)
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
	args := []interface{}{orgID}
	q := `
		SELECT type, COUNT(*) AS count
		FROM activities
		WHERE activities.org_id = $1 AND activities.deleted_at IS NULL
	`
	q += reportActivityParentVisibilityClause(ctx, &args, "activities")
	q += `
		GROUP BY type
		ORDER BY type
	`
	rows, err := r.db.Query(ctx, q, args...)
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

	baseArgs, dateClause := reportArgs(orgID, filter, "t")

	// Per-status counts.
	statusArgs := cloneReportArgs(baseArgs)
	statusQuery := `
		SELECT status, COUNT(*) AS count
		FROM tickets t
		WHERE t.org_id = $1 AND t.deleted_at IS NULL` + dateClause + `
	`
	appendReportReadVisibility(ctx, &statusQuery, &statusArgs, domain.ACLModuleTickets, "t.assignee_id", "t.submitted_by_user_id")
	statusQuery += `
		GROUP BY status
		ORDER BY status
	`
	statusRows, err := r.db.Query(ctx, statusQuery, statusArgs...)
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
	avgArgs := cloneReportArgs(baseArgs)
	avgQuery := `
		SELECT COALESCE(
			AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 3600.0),
			0
		)
		FROM tickets t
		WHERE t.org_id = $1
		  AND t.deleted_at IS NULL
		  AND t.status IN ('resolved', 'closed')` + dateClause
	appendReportReadVisibility(ctx, &avgQuery, &avgArgs, domain.ACLModuleTickets, "t.assignee_id", "t.submitted_by_user_id")
	if err := r.db.QueryRow(ctx, avgQuery, avgArgs...).Scan(&avgHours); err != nil {
		return nil, err
	}

	// Breach rate: % of open tickets older than 48 hours.
	breachArgs, breachDateClause := reportArgs(orgID, filter, "t")
	breachArgs = append(breachArgs, time.Now().Add(-48*time.Hour))
	breachThreshIdx := "$" + strconv.Itoa(len(breachArgs))

	var openTotal, openBreached int
	breachQuery := `
		SELECT
			COUNT(*) FILTER (WHERE t.status IN ('open','in_progress','pending')) AS open_total,
			COUNT(*) FILTER (WHERE t.status IN ('open','in_progress','pending') AND t.created_at <= ` + breachThreshIdx + `) AS open_breached
		FROM tickets t
		WHERE t.org_id = $1 AND t.deleted_at IS NULL` + breachDateClause
	appendReportReadVisibility(ctx, &breachQuery, &breachArgs, domain.ACLModuleTickets, "t.assignee_id", "t.submitted_by_user_id")
	if err := r.db.QueryRow(ctx, breachQuery, breachArgs...).Scan(&openTotal, &openBreached); err != nil {
		return nil, err
	}

	var breachRate float64
	if openTotal > 0 {
		breachRate = float64(openBreached) / float64(openTotal) * 100
	}

	// Daily ticket count for the Tickets Over Time chart.
	dailyArgs := cloneReportArgs(baseArgs)
	dailyQuery := `
		SELECT TO_CHAR(t.created_at, 'YYYY-MM-DD') AS date, COUNT(*) AS count
		FROM tickets t
		WHERE t.org_id = $1 AND t.deleted_at IS NULL` + dateClause + `
	`
	appendReportReadVisibility(ctx, &dailyQuery, &dailyArgs, domain.ACLModuleTickets, "t.assignee_id", "t.submitted_by_user_id")
	dailyQuery += `
		GROUP BY date
		ORDER BY date
	`
	dailyRows, err := r.db.Query(ctx, dailyQuery, dailyArgs...)
	if err != nil {
		return nil, err
	}
	defer dailyRows.Close()

	var overTime []domain.TicketDailyMetric
	for dailyRows.Next() {
		var m domain.TicketDailyMetric
		if err := dailyRows.Scan(&m.Date, &m.Count); err != nil {
			return nil, err
		}
		overTime = append(overTime, m)
	}
	if err := dailyRows.Err(); err != nil {
		return nil, err
	}
	if overTime == nil {
		overTime = []domain.TicketDailyMetric{}
	}

	return &domain.TicketReport{
		TotalOpen:          totalOpen,
		TotalClosed:        totalClosed,
		AvgResolutionHours: avgHours,
		ByStatus:           byStatus,
		BreachRate:         breachRate,
		OverTime:           overTime,
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

	// All-time total for the org (no date filter).
	var totalCount int
	totalArgs := []interface{}{orgID}
	totalQuery := `
		SELECT COUNT(*)
		FROM contacts c
		WHERE c.org_id = $1 AND c.deleted_at IS NULL`
	appendReportReadVisibility(ctx, &totalQuery, &totalArgs, domain.ACLModuleContacts, "c.owner_id")
	if err := r.db.QueryRow(ctx, totalQuery, totalArgs...).Scan(&totalCount); err != nil {
		return nil, err
	}

	// Contacts created within the date range.
	baseRangeArgs, dateClause := reportArgs(orgID, filter, "c")
	var newCount int
	newCountArgs := cloneReportArgs(baseRangeArgs)
	newCountQuery := `
		SELECT COUNT(*)
		FROM contacts c
		WHERE c.org_id = $1 AND c.deleted_at IS NULL` + dateClause
	appendReportReadVisibility(ctx, &newCountQuery, &newCountArgs, domain.ACLModuleContacts, "c.owner_id")
	if err := r.db.QueryRow(ctx, newCountQuery, newCountArgs...).Scan(&newCount); err != nil {
		return nil, err
	}

	overTimeArgs := cloneReportArgs(baseRangeArgs)
	overTimeQuery := `
		SELECT TO_CHAR(c.created_at, 'YYYY-MM') AS month, COUNT(*) AS count
		FROM contacts c
		WHERE c.org_id = $1 AND c.deleted_at IS NULL` + dateClause + `
	`
	appendReportReadVisibility(ctx, &overTimeQuery, &overTimeArgs, domain.ACLModuleContacts, "c.owner_id")
	overTimeQuery += `
		GROUP BY month
		ORDER BY month
	`
	overTimeRows, err := r.db.Query(ctx, overTimeQuery, overTimeArgs...)
	if err != nil {
		return nil, err
	}
	defer overTimeRows.Close()

	var overTime []domain.ContactOverTimeMetric
	for overTimeRows.Next() {
		var m domain.ContactOverTimeMetric
		if err := overTimeRows.Scan(&m.Month, &m.Count); err != nil {
			return nil, err
		}
		overTime = append(overTime, m)
	}
	if err := overTimeRows.Err(); err != nil {
		return nil, err
	}
	if overTime == nil {
		overTime = []domain.ContactOverTimeMetric{}
	}

	return &domain.ContactReport{
		NewCount:   newCount,
		TotalCount: totalCount,
		OverTime:   overTime,
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

	stageQuery := `
		SELECT stage, COUNT(*) AS count, COALESCE(SUM(value_cents), 0) AS total_value_cents
		FROM deals d
		WHERE d.org_id = $1 AND d.deleted_at IS NULL` + dateClause + `
	`
	appendReportReadVisibility(ctx, &stageQuery, &args, domain.ACLModuleDeals, "d.owner_id")
	stageQuery += `
		GROUP BY stage
		ORDER BY stage
	`
	stageRows, err := r.db.Query(ctx, stageQuery, args...)
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

	baseArgs, dateClause := reportArgs(orgID, filter, "l")

	var newCount, convertedCount int
	countArgs := cloneReportArgs(baseArgs)
	countQuery := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE l.status = 'converted' OR l.converted_contact_id IS NOT NULL) AS converted
		FROM leads l
		WHERE l.org_id = $1 AND l.deleted_at IS NULL` + dateClause
	appendReportReadVisibility(ctx, &countQuery, &countArgs, domain.ACLModuleLeads, "l.owner_id")
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&newCount, &convertedCount); err != nil {
		return nil, err
	}

	var conversionRate float64
	if newCount > 0 {
		conversionRate = float64(convertedCount) / float64(newCount)
	}

	// Funnel: count per lead status for the funnel breakdown chart.
	funnelArgs := cloneReportArgs(baseArgs)
	funnelQuery := `
		SELECT l.status, COUNT(*) AS count
		FROM leads l
		WHERE l.org_id = $1 AND l.deleted_at IS NULL` + dateClause + `
	`
	appendReportReadVisibility(ctx, &funnelQuery, &funnelArgs, domain.ACLModuleLeads, "l.owner_id")
	funnelQuery += `
		GROUP BY l.status
		ORDER BY l.status
	`
	funnelRows, err := r.db.Query(ctx, funnelQuery, funnelArgs...)
	if err != nil {
		return nil, err
	}
	defer funnelRows.Close()

	stageLabels := map[string]string{
		"new":       "New",
		"contacted": "Contacted",
		"qualified": "Qualified",
		"converted": "Converted",
	}

	var funnel []domain.LeadFunnelMetric
	for funnelRows.Next() {
		var stage string
		var count int
		if err := funnelRows.Scan(&stage, &count); err != nil {
			return nil, err
		}
		label, ok := stageLabels[stage]
		if !ok {
			label = stage
		}
		funnel = append(funnel, domain.LeadFunnelMetric{
			Stage: stage,
			Label: label,
			Count: count,
		})
	}
	if err := funnelRows.Err(); err != nil {
		return nil, err
	}
	if funnel == nil {
		funnel = []domain.LeadFunnelMetric{}
	}

	return &domain.LeadReport{
		NewCount:       newCount,
		ConvertedCount: convertedCount,
		ConversionRate: conversionRate,
		Funnel:         funnel,
	}, nil
}

func (r *ReportsRepo) PipelineFunnel(ctx context.Context, pipelineID *uuid.UUID, filter domain.ReportFilter) (*domain.PipelineFunnelReport, error) {
	orgID, err := scopedReportOrgID(ctx, filter)
	if err != nil {
		return nil, err
	}
	stages := []string{"lead", "qualified", "proposal", "negotiation"}

	where := []string{"deleted_at IS NULL", "stage NOT IN ('closed_won','closed_lost')", "org_id = $1"}
	args := []any{orgID}
	i := 2

	if pipelineID != nil {
		where = append(where, fmt.Sprintf("pipeline_id = $%d", i))
		args = append(args, *pipelineID)
		i++
	}
	if filter.From != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", i))
		args = append(args, *filter.From)
		i++
	}
	if filter.To != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", i))
		args = append(args, *filter.To)
		i++
	}
	addAccessVisibilityWhere(ctx, &where, &args, &i, domain.ACLModuleDeals, domain.SharingAccessRead, "owner_id")
	_ = i

	q := `SELECT stage, COUNT(*), COALESCE(SUM(value_cents), 0) FROM deals WHERE ` +
		strings.Join(where, " AND ") + ` GROUP BY stage`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byStage := make(map[string]domain.PipelineFunnelStage)
	for rows.Next() {
		var s domain.PipelineFunnelStage
		if err := rows.Scan(&s.Name, &s.Count, &s.ValueCents); err != nil {
			return nil, err
		}
		byStage[s.Name] = s
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &domain.PipelineFunnelReport{Stages: make([]domain.PipelineFunnelStage, 0, len(stages))}
	for _, name := range stages {
		if s, ok := byStage[name]; ok {
			result.Stages = append(result.Stages, s)
		} else {
			result.Stages = append(result.Stages, domain.PipelineFunnelStage{Name: name})
		}
	}
	return result, nil
}

func (r *ReportsRepo) ConversionRates(ctx context.Context, filter domain.ReportFilter) (*domain.ConversionRatesReport, error) {
	orgID, err := scopedReportOrgID(ctx, filter)
	if err != nil {
		return nil, err
	}
	pairs := []struct{ from, to string }{
		{"lead", "qualified"},
		{"qualified", "proposal"},
		{"proposal", "negotiation"},
	}

	where := []string{"deleted_at IS NULL", "org_id = $1"}
	args := []any{orgID}
	i := 2

	if filter.From != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", i))
		args = append(args, *filter.From)
		i++
	}
	if filter.To != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", i))
		args = append(args, *filter.To)
		i++
	}
	addAccessVisibilityWhere(ctx, &where, &args, &i, domain.ACLModuleDeals, domain.SharingAccessRead, "owner_id")
	_ = i

	q := `SELECT stage, COUNT(*) FROM deals WHERE ` + strings.Join(where, " AND ") + ` GROUP BY stage`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var stage string
		var count int
		if err := rows.Scan(&stage, &count); err != nil {
			return nil, err
		}
		counts[stage] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &domain.ConversionRatesReport{Rates: make([]domain.ConversionRate, 0, len(pairs))}
	for _, p := range pairs {
		fromCount := counts[p.from]
		toCount := counts[p.to]
		var rate float64
		if fromCount > 0 {
			rate = float64(toCount) / float64(fromCount)
		}
		result.Rates = append(result.Rates, domain.ConversionRate{From: p.from, To: p.to, Rate: rate})
	}
	return result, nil
}

func (r *ReportsRepo) RevenueProjection(ctx context.Context, months int) (*domain.RevenueProjectionReport, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	if months <= 0 || months > 12 {
		months = 3
	}

	stageWeights := map[string]float64{
		"lead":        0.10,
		"qualified":   0.25,
		"proposal":    0.50,
		"negotiation": 0.75,
	}

	where := []string{
		"deleted_at IS NULL",
		"org_id = $1",
		"stage NOT IN ('closed_won','closed_lost')",
		"expected_close_date IS NOT NULL",
		fmt.Sprintf("expected_close_date <= NOW() + ('%d months')::interval", months),
	}
	args := []any{orgID}
	i := 2

	addAccessVisibilityWhere(ctx, &where, &args, &i, domain.ACLModuleDeals, domain.SharingAccessRead, "owner_id")
	_ = i

	q := `SELECT stage, expected_close_date, value_cents FROM deals WHERE ` + strings.Join(where, " AND ")
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type bucket struct {
		projected int64
		count     int
	}
	buckets := make(map[string]*bucket)
	for rows.Next() {
		var stage string
		var closeDate time.Time
		var valueCents int64
		if err := rows.Scan(&stage, &closeDate, &valueCents); err != nil {
			return nil, err
		}
		month := closeDate.Format("2006-01")
		w := stageWeights[stage]
		if _, ok := buckets[month]; !ok {
			buckets[month] = &bucket{}
		}
		buckets[month].projected += int64(float64(valueCents) * w)
		buckets[month].count++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build ordered month list
	now := time.Now()
	result := &domain.RevenueProjectionReport{Months: make([]domain.RevenueProjectionMonth, 0, months)}
	for m := 0; m < months; m++ {
		t := now.AddDate(0, m, 0)
		monthStr := t.Format("2006-01")
		b, ok := buckets[monthStr]
		if !ok {
			b = &bucket{}
		}
		result.Months = append(result.Months, domain.RevenueProjectionMonth{
			Month:          monthStr,
			ProjectedCents: b.projected,
			DealCount:      b.count,
		})
	}
	return result, nil
}

func (r *ReportsRepo) ActivitySummary(ctx context.Context, filter domain.ReportFilter) (*domain.ActivitySummaryReport, error) {
	orgID, err := scopedReportOrgID(ctx, filter)
	if err != nil {
		return nil, err
	}
	where := []string{"a.deleted_at IS NULL", "a.org_id = $1"}
	args := []any{orgID}
	i := 2

	if filter.From != nil {
		where = append(where, fmt.Sprintf("a.created_at >= $%d", i))
		args = append(args, *filter.From)
		i++
	}
	if filter.To != nil {
		where = append(where, fmt.Sprintf("a.created_at <= $%d", i))
		args = append(args, *filter.To)
		i++
	}
	addActivityParentVisibilityWhere(ctx, &where, &args, &i, domain.SharingAccessRead, "a")
	_ = i

	whereStr := strings.Join(where, " AND ")

	// By type (called "kind" in the response)
	kindRows, err := r.db.Query(ctx, `SELECT a.type, COUNT(*) FROM activities a WHERE `+whereStr+` GROUP BY a.type ORDER BY a.type`, args...)
	if err != nil {
		return nil, err
	}
	defer kindRows.Close()
	var byKind []domain.ActivityKindCount
	for kindRows.Next() {
		var ak domain.ActivityKindCount
		if err := kindRows.Scan(&ak.Kind, &ak.Count); err != nil {
			return nil, err
		}
		byKind = append(byKind, ak)
	}
	if err := kindRows.Err(); err != nil {
		return nil, err
	}

	// By owner
	ownerRows, err := r.db.Query(ctx, `SELECT a.owner_id, COUNT(*) AS count FROM activities a WHERE `+whereStr+` GROUP BY a.owner_id ORDER BY count DESC LIMIT 20`, args...)
	if err != nil {
		return nil, err
	}
	defer ownerRows.Close()
	var byOwner []domain.ActivityOwnerCount
	for ownerRows.Next() {
		var ao domain.ActivityOwnerCount
		if err := ownerRows.Scan(&ao.OwnerID, &ao.Count); err != nil {
			return nil, err
		}
		byOwner = append(byOwner, ao)
	}
	if err := ownerRows.Err(); err != nil {
		return nil, err
	}

	return &domain.ActivitySummaryReport{
		ByKind:  byKind,
		ByOwner: byOwner,
	}, nil
}

// ManagerDashboard returns a fixed manager dashboard response contract that
// aggregates CRM, help desk, and team activity metrics.
func (r *ReportsRepo) ManagerDashboard(ctx context.Context, filter domain.ReportFilter) (*domain.ManagerDashboardReport, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	if filter.OrgID != nil {
		orgID = *filter.OrgID
	}

	dealReport, err := r.DealMetrics(domain.WithOrgID(ctx, orgID), filter)
	if err != nil {
		return nil, err
	}
	ticketReport, err := r.TicketMetrics(domain.WithOrgID(ctx, orgID), filter)
	if err != nil {
		return nil, err
	}

	// Backlog = open-like tickets older than 48h, respecting the created_at range.
	backlogArgs := []interface{}{orgID}
	backlogArgs, backlogDateClause := reportDateClause(backlogArgs, filter, "t", "created_at")
	backlogArgs = append(backlogArgs, time.Now().Add(-48*time.Hour))
	backlogThresholdIdx := "$" + strconv.Itoa(len(backlogArgs))

	var backlogCount int
	backlogQuery := `
		SELECT COUNT(*)
		FROM tickets t
		WHERE t.org_id = $1
		  AND t.deleted_at IS NULL
		  AND t.status IN ('open','in_progress','pending')
		  AND t.created_at <= ` + backlogThresholdIdx + backlogDateClause
	appendReportReadVisibility(ctx, &backlogQuery, &backlogArgs, domain.ACLModuleTickets, "t.assignee_id", "t.submitted_by_user_id")
	if err := r.db.QueryRow(ctx, backlogQuery, backlogArgs...).Scan(&backlogCount); err != nil {
		return nil, err
	}

	// Resolution trend: resolved/closed ticket counts grouped by updated_at date.
	resolutionArgs := []interface{}{orgID}
	resolutionArgs, resolutionDateClause := reportDateClause(resolutionArgs, filter, "t", "updated_at")
	resolutionQuery := `
		SELECT TO_CHAR(t.updated_at, 'YYYY-MM-DD') AS date, COUNT(*) AS count
		FROM tickets t
		WHERE t.org_id = $1
		  AND t.deleted_at IS NULL
		  AND t.status IN ('resolved','closed')` + resolutionDateClause + `
	`
	appendReportReadVisibility(ctx, &resolutionQuery, &resolutionArgs, domain.ACLModuleTickets, "t.assignee_id", "t.submitted_by_user_id")
	resolutionQuery += `
		GROUP BY date
		ORDER BY date
	`
	resolutionRows, err := r.db.Query(ctx, resolutionQuery, resolutionArgs...)
	if err != nil {
		return nil, err
	}
	defer resolutionRows.Close()

	resolutionTrend := make([]domain.DailyCountMetric, 0)
	for resolutionRows.Next() {
		var point domain.DailyCountMetric
		if err := resolutionRows.Scan(&point.Date, &point.Count); err != nil {
			return nil, err
		}
		resolutionTrend = append(resolutionTrend, point)
	}
	if err := resolutionRows.Err(); err != nil {
		return nil, err
	}

	// Team activity by user: combine created and completed aggregates.
	byOwner := map[uuid.UUID]*domain.TeamActivityByUserMetric{}

	createdByOwnerArgs := []interface{}{orgID}
	createdByOwnerArgs, createdByOwnerClause := reportDateClause(createdByOwnerArgs, filter, "a", "created_at")
	createdByOwnerClause += reportActivityParentVisibilityClause(ctx, &createdByOwnerArgs, "a")
	createdByOwnerRows, err := r.db.Query(ctx, `
		SELECT a.owner_id, COUNT(*) AS created_count
		FROM activities a
		WHERE a.org_id = $1
		  AND a.deleted_at IS NULL`+createdByOwnerClause+`
		GROUP BY a.owner_id
	`, createdByOwnerArgs...)
	if err != nil {
		return nil, err
	}
	for createdByOwnerRows.Next() {
		var ownerID uuid.UUID
		var count int
		if err := createdByOwnerRows.Scan(&ownerID, &count); err != nil {
			createdByOwnerRows.Close()
			return nil, err
		}
		byOwner[ownerID] = &domain.TeamActivityByUserMetric{
			OwnerID:      ownerID,
			CreatedCount: count,
		}
	}
	if err := createdByOwnerRows.Err(); err != nil {
		createdByOwnerRows.Close()
		return nil, err
	}
	createdByOwnerRows.Close()

	completedByOwnerArgs := []interface{}{orgID}
	completedByOwnerArgs, completedByOwnerClause := reportDateClause(completedByOwnerArgs, filter, "a", "completed_at")
	completedByOwnerClause += reportActivityParentVisibilityClause(ctx, &completedByOwnerArgs, "a")
	completedByOwnerRows, err := r.db.Query(ctx, `
		SELECT a.owner_id, COUNT(*) AS completed_count
		FROM activities a
		WHERE a.org_id = $1
		  AND a.deleted_at IS NULL
		  AND a.completed_at IS NOT NULL`+completedByOwnerClause+`
		GROUP BY a.owner_id
	`, completedByOwnerArgs...)
	if err != nil {
		return nil, err
	}
	for completedByOwnerRows.Next() {
		var ownerID uuid.UUID
		var count int
		if err := completedByOwnerRows.Scan(&ownerID, &count); err != nil {
			completedByOwnerRows.Close()
			return nil, err
		}
		if byOwner[ownerID] == nil {
			byOwner[ownerID] = &domain.TeamActivityByUserMetric{OwnerID: ownerID}
		}
		byOwner[ownerID].CompletedCount = count
	}
	if err := completedByOwnerRows.Err(); err != nil {
		completedByOwnerRows.Close()
		return nil, err
	}
	completedByOwnerRows.Close()

	teamByUser := make([]domain.TeamActivityByUserMetric, 0, len(byOwner))
	for _, row := range byOwner {
		teamByUser = append(teamByUser, *row)
	}
	sort.Slice(teamByUser, func(i, j int) bool {
		left := teamByUser[i].CreatedCount + teamByUser[i].CompletedCount
		right := teamByUser[j].CreatedCount + teamByUser[j].CompletedCount
		if left == right {
			return teamByUser[i].OwnerID.String() < teamByUser[j].OwnerID.String()
		}
		return left > right
	})

	createdTrendArgs := []interface{}{orgID}
	createdTrendArgs, createdTrendClause := reportDateClause(createdTrendArgs, filter, "a", "created_at")
	createdTrendClause += reportActivityParentVisibilityClause(ctx, &createdTrendArgs, "a")
	createdTrendRows, err := r.db.Query(ctx, `
		SELECT TO_CHAR(a.created_at, 'YYYY-MM-DD') AS date, COUNT(*) AS count
		FROM activities a
		WHERE a.org_id = $1
		  AND a.deleted_at IS NULL`+createdTrendClause+`
		GROUP BY date
		ORDER BY date
	`, createdTrendArgs...)
	if err != nil {
		return nil, err
	}
	defer createdTrendRows.Close()

	createdTrend := make([]domain.DailyCountMetric, 0)
	for createdTrendRows.Next() {
		var point domain.DailyCountMetric
		if err := createdTrendRows.Scan(&point.Date, &point.Count); err != nil {
			return nil, err
		}
		createdTrend = append(createdTrend, point)
	}
	if err := createdTrendRows.Err(); err != nil {
		return nil, err
	}

	completedTrendArgs := []interface{}{orgID}
	completedTrendArgs, completedTrendClause := reportDateClause(completedTrendArgs, filter, "a", "completed_at")
	completedTrendClause += reportActivityParentVisibilityClause(ctx, &completedTrendArgs, "a")
	completedTrendRows, err := r.db.Query(ctx, `
		SELECT TO_CHAR(a.completed_at, 'YYYY-MM-DD') AS date, COUNT(*) AS count
		FROM activities a
		WHERE a.org_id = $1
		  AND a.deleted_at IS NULL
		  AND a.completed_at IS NOT NULL`+completedTrendClause+`
		GROUP BY date
		ORDER BY date
	`, completedTrendArgs...)
	if err != nil {
		return nil, err
	}
	defer completedTrendRows.Close()

	completedTrend := make([]domain.DailyCountMetric, 0)
	for completedTrendRows.Next() {
		var point domain.DailyCountMetric
		if err := completedTrendRows.Scan(&point.Date, &point.Count); err != nil {
			return nil, err
		}
		completedTrend = append(completedTrend, point)
	}
	if err := completedTrendRows.Err(); err != nil {
		return nil, err
	}

	return &domain.ManagerDashboardReport{
		CRM: domain.DashboardCRMMetrics{
			PipelineValueCents: dealReport.PipelineValueCents,
			WonCount:           dealReport.WonCount,
			LostCount:          dealReport.LostCount,
			StageDistribution:  dealReport.ByStage,
		},
		HelpDesk: domain.DashboardHelpDeskMetrics{
			OpenCount:          ticketReport.TotalOpen,
			BacklogCount:       backlogCount,
			StatusDistribution: ticketReport.ByStatus,
			VolumeTrend:        ticketReport.OverTime,
			ResolutionTrend:    resolutionTrend,
		},
		TeamActivity: domain.DashboardTeamActivityMetrics{
			ByUser:            teamByUser,
			CreatedOverTime:   createdTrend,
			CompletedOverTime: completedTrend,
		},
	}, nil
}
