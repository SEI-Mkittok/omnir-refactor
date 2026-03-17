package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// NotificationRepo is the Postgres implementation of repository.NotificationRepository.
type NotificationRepo struct {
	db *pgxpool.Pool
}

func NewNotificationRepo(db *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{db: db}
}

const notificationCols = `id, org_id, user_id, activity_id, type, read_at, created_at`

func scanNotification(row pgx.Row) (*domain.Notification, error) {
	var n domain.Notification
	err := row.Scan(&n.ID, &n.OrgID, &n.UserID, &n.ActivityID, &n.Type, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &n, nil
}

func (r *NotificationRepo) Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO notifications (id, org_id, user_id, activity_id, type)
		VALUES ($1, $2, $3, $4, $5::notification_type)
		ON CONFLICT (activity_id, type) DO NOTHING
		RETURNING `+notificationCols,
		n.ID, n.OrgID, n.UserID, n.ActivityID, n.Type,
	)
	result, err := scanNotification(row)
	if err != nil {
		// ON CONFLICT DO NOTHING returns no rows — treat as a no-op success.
		if errors.Is(err, domain.ErrNotFound) {
			return n, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *NotificationRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	q := `UPDATE notifications SET read_at = NOW() WHERE id = $1 AND user_id = $2 AND read_at IS NULL`
	result, err := r.db.Exec(ctx, q, id, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NotificationRepo) ListByUser(ctx context.Context, f domain.NotificationFilter) ([]*domain.Notification, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{}
	args := []any{}
	i := 1

	addWhere := func(expr string, val any) {
		where = append(where, fmt.Sprintf("%s = $%d", expr, i))
		args = append(args, val)
		i++
	}

	addWhere("user_id", f.UserID)
	if f.OrgID != uuid.Nil {
		addWhere("org_id", f.OrgID)
	}
	if f.Unread {
		where = append(where, "read_at IS NULL")
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		fmt.Sprintf(
			`SELECT %s FROM notifications WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			notificationCols, whereClause, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.OrgID, &n.UserID, &n.ActivityID, &n.Type, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, &n)
	}
	return notifications, total, rows.Err()
}

// GenerateReminders scans activities with upcoming or overdue due dates and inserts
// missing notifications. The unique index on (activity_id, type) prevents duplicates.
func (r *NotificationRepo) GenerateReminders(ctx context.Context) error {
	// Each entry: (notification_type, interval expression for upcoming window).
	upcomingTypes := []struct {
		nType    string
		interval string
	}{
		{"upcoming_15m", "15 minutes"},
		{"upcoming_1h", "1 hour"},
		{"upcoming_1d", "1 day"},
	}

	for _, ut := range upcomingTypes {
		_, err := r.db.Exec(ctx, `
			INSERT INTO notifications (org_id, user_id, activity_id, type)
			SELECT a.org_id, a.owner_id, a.id, $1::notification_type
			FROM activities a
			WHERE a.due_date IS NOT NULL
			  AND a.completed_at IS NULL
			  AND a.deleted_at IS NULL
			  AND a.due_date BETWEEN NOW() AND NOW() + $2::interval
			ON CONFLICT (activity_id, type) DO NOTHING
		`, ut.nType, ut.interval)
		if err != nil {
			return fmt.Errorf("generate %s reminders: %w", ut.nType, err)
		}
	}

	// Overdue: due_date is in the past and activity is not completed.
	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications (org_id, user_id, activity_id, type)
		SELECT a.org_id, a.owner_id, a.id, 'overdue'::notification_type
		FROM activities a
		WHERE a.due_date IS NOT NULL
		  AND a.completed_at IS NULL
		  AND a.deleted_at IS NULL
		  AND a.due_date < NOW()
		ON CONFLICT (activity_id, type) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("generate overdue reminders: %w", err)
	}

	return nil
}
