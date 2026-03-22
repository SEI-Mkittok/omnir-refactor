package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// DashboardRepo implements repository.DashboardRepository against PostgreSQL.
type DashboardRepo struct {
	db *pgxpool.Pool
}

func NewDashboardRepo(db *pgxpool.Pool) *DashboardRepo {
	return &DashboardRepo{db: db}
}

// ── custom_dashboards ────────────────────────────────────────────────────────

func (r *DashboardRepo) CreateDashboard(ctx context.Context, d *domain.CustomDashboard) (*domain.CustomDashboard, error) {
	widgets, err := domain.WidgetsToJSON(d.Widgets)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO custom_dashboards (id, org_id, name, widgets, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, org_id, name, widgets, created_by, created_at, updated_at
	`, d.ID, d.OrgID, d.Name, widgets, d.CreatedBy)
	return scanDashboard(row)
}

func (r *DashboardRepo) GetDashboardByID(ctx context.Context, id uuid.UUID) (*domain.CustomDashboard, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	row := r.db.QueryRow(ctx, `
		SELECT id, org_id, name, widgets, created_by, created_at, updated_at
		FROM custom_dashboards
		WHERE id = $1 AND org_id = $2
	`, id, orgID)
	d, err := scanDashboard(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return d, err
}

func (r *DashboardRepo) UpdateDashboard(ctx context.Context, id uuid.UUID, patch domain.CustomDashboardPatch) (*domain.CustomDashboard, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	existing, err := r.GetDashboardByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if patch.Name != nil {
		existing.Name = *patch.Name
	}
	if patch.Widgets != nil {
		existing.Widgets = patch.Widgets
	}
	widgets, err := domain.WidgetsToJSON(existing.Widgets)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		UPDATE custom_dashboards
		SET name = $1, widgets = $2, updated_at = NOW()
		WHERE id = $3 AND org_id = $4
		RETURNING id, org_id, name, widgets, created_by, created_at, updated_at
	`, existing.Name, widgets, id, orgID)
	d, err := scanDashboard(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return d, err
}

func (r *DashboardRepo) DeleteDashboard(ctx context.Context, id uuid.UUID) error {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return domain.ErrNotFound
	}
	tag, err := r.db.Exec(ctx, `
		DELETE FROM custom_dashboards WHERE id = $1 AND org_id = $2
	`, id, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DashboardRepo) ListDashboards(ctx context.Context) ([]*domain.CustomDashboard, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, org_id, name, widgets, created_by, created_at, updated_at
		FROM custom_dashboards
		WHERE org_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.CustomDashboard
	for rows.Next() {
		d, err := scanDashboard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if out == nil {
		out = []*domain.CustomDashboard{}
	}
	return out, rows.Err()
}

func scanDashboard(row pgx.Row) (*domain.CustomDashboard, error) {
	var d domain.CustomDashboard
	var widgetsJSON []byte
	if err := row.Scan(&d.ID, &d.OrgID, &d.Name, &widgetsJSON, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, err
	}
	ws, err := domain.WidgetsFromJSON(widgetsJSON)
	if err != nil {
		return nil, err
	}
	d.Widgets = ws
	return &d, nil
}

// ── scheduled_reports ────────────────────────────────────────────────────────

func (r *DashboardRepo) CreateSchedule(ctx context.Context, s *domain.ScheduledReport) (*domain.ScheduledReport, error) {
	recipients, err := domain.RecipientsToJSON(s.Recipients)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO scheduled_reports (id, org_id, dashboard_id, schedule, recipients, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, org_id, dashboard_id, schedule, recipients, last_sent_at, created_at, updated_at
	`, s.ID, s.OrgID, s.DashboardID, s.Schedule, recipients)
	return scanSchedule(row)
}

func (r *DashboardRepo) GetScheduleByID(ctx context.Context, id uuid.UUID) (*domain.ScheduledReport, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	row := r.db.QueryRow(ctx, `
		SELECT id, org_id, dashboard_id, schedule, recipients, last_sent_at, created_at, updated_at
		FROM scheduled_reports
		WHERE id = $1 AND org_id = $2
	`, id, orgID)
	s, err := scanSchedule(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return s, err
}

func (r *DashboardRepo) UpdateSchedule(ctx context.Context, id uuid.UUID, patch domain.ScheduledReportPatch) (*domain.ScheduledReport, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	existing, err := r.GetScheduleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if patch.DashboardID != nil {
		existing.DashboardID = *patch.DashboardID
	}
	if patch.Schedule != nil {
		existing.Schedule = *patch.Schedule
	}
	if patch.Recipients != nil {
		existing.Recipients = patch.Recipients
	}
	recipients, err := domain.RecipientsToJSON(existing.Recipients)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		UPDATE scheduled_reports
		SET dashboard_id = $1, schedule = $2, recipients = $3, updated_at = NOW()
		WHERE id = $4 AND org_id = $5
		RETURNING id, org_id, dashboard_id, schedule, recipients, last_sent_at, created_at, updated_at
	`, existing.DashboardID, existing.Schedule, recipients, id, orgID)
	s, err := scanSchedule(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return s, err
}

func (r *DashboardRepo) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return domain.ErrNotFound
	}
	tag, err := r.db.Exec(ctx, `
		DELETE FROM scheduled_reports WHERE id = $1 AND org_id = $2
	`, id, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DashboardRepo) ListSchedules(ctx context.Context) ([]*domain.ScheduledReport, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrNotFound
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, org_id, dashboard_id, schedule, recipients, last_sent_at, created_at, updated_at
		FROM scheduled_reports
		WHERE org_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ScheduledReport
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []*domain.ScheduledReport{}
	}
	return out, rows.Err()
}

// ListAllDueSchedules returns all scheduled reports across all orgs.
// The worker applies cron-expression filtering to determine which are due.
func (r *DashboardRepo) ListAllDueSchedules(ctx context.Context, _ time.Time) ([]*domain.ScheduledReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, org_id, dashboard_id, schedule, recipients, last_sent_at, created_at, updated_at
		FROM scheduled_reports
		ORDER BY org_id, dashboard_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ScheduledReport
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *DashboardRepo) MarkScheduleSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE scheduled_reports SET last_sent_at = $1, updated_at = NOW() WHERE id = $2
	`, sentAt, id)
	return err
}

func scanSchedule(row pgx.Row) (*domain.ScheduledReport, error) {
	var s domain.ScheduledReport
	var recipientsJSON []byte
	if err := row.Scan(
		&s.ID, &s.OrgID, &s.DashboardID, &s.Schedule,
		&recipientsJSON, &s.LastSentAt, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, err
	}
	rs, err := domain.RecipientsFromJSON(recipientsJSON)
	if err != nil {
		return nil, err
	}
	s.Recipients = rs
	return &s, nil
}
