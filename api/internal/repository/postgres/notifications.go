package postgres

import (
	"context"
	"errors"
	"strconv"

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

const notifCols = `id, org_id, user_id, actor_id, kind, entity_type, entity_id, title, body, read_at, emailed_at, created_at`

func scanNotification(row pgx.Row) (*domain.Notification, error) {
	var n domain.Notification
	err := row.Scan(
		&n.ID, &n.OrgID, &n.UserID, &n.ActorID, &n.Kind,
		&n.EntityType, &n.EntityID, &n.Title, &n.Body,
		&n.ReadAt, &n.EmailedAt, &n.CreatedAt,
	)
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
		INSERT INTO notifications (id, org_id, user_id, actor_id, kind, entity_type, entity_id, title, body)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+notifCols,
		n.ID, n.OrgID, n.UserID, n.ActorID, n.Kind,
		n.EntityType, n.EntityID, n.Title, n.Body,
	)
	return scanNotification(row)
}

func (r *NotificationRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE notifications SET read_at = NOW() WHERE id = $1 AND user_id = $2 AND read_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID, orgID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE notifications SET read_at = NOW() WHERE user_id = $1 AND org_id = $2 AND read_at IS NULL`,
		userID, orgID,
	)
	return err
}

func (r *NotificationRepo) UnreadCount(ctx context.Context, userID, orgID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND org_id = $2 AND read_at IS NULL`,
		userID, orgID,
	).Scan(&count)
	return count, err
}

func (r *NotificationRepo) ListByUser(ctx context.Context, f domain.NotificationFilter) ([]*domain.Notification, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}

	args := []any{f.UserID, f.OrgID}
	q := `SELECT ` + notifCols + ` FROM notifications WHERE user_id = $1 AND org_id = $2`

	if f.UnreadOnly {
		q += ` AND read_at IS NULL`
	}
	if f.Before != nil {
		q += ` AND created_at < $3`
		args = append(args, *f.Before)
	}
	q += ` ORDER BY created_at DESC`
	args = append(args, f.Limit+1)
	q += ` LIMIT $` + itoa(len(args))

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

// GenerateReminders scans activities with upcoming or overdue due dates and creates
// activity_reminder notifications.
func (r *NotificationRepo) GenerateReminders(ctx context.Context) error {
	upcomingTypes := []struct {
		label    string
		interval string
	}{
		{"15 minutes", "15 minutes"},
		{"1 hour", "1 hour"},
		{"1 day", "1 day"},
	}

	for _, ut := range upcomingTypes {
		_, err := r.db.Exec(ctx, `
			INSERT INTO notifications (org_id, user_id, kind, entity_type, entity_id, title)
			SELECT a.org_id, a.owner_id, 'activity_reminder', 'activity', a.id,
			       'Upcoming activity in `+ut.label+`'
			FROM activities a
			WHERE a.due_date IS NOT NULL
			  AND a.completed_at IS NULL
			  AND a.deleted_at IS NULL
			  AND a.due_date BETWEEN NOW() AND NOW() + $1::interval
			  AND NOT EXISTS (
			    SELECT 1 FROM notifications n2
			    WHERE n2.entity_id = a.id
			      AND n2.kind = 'activity_reminder'
			      AND n2.title = 'Upcoming activity in `+ut.label+`'
			      AND n2.created_at > NOW() - $1::interval
			  )
		`, ut.interval)
		if err != nil {
			return err
		}
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications (org_id, user_id, kind, entity_type, entity_id, title)
		SELECT a.org_id, a.owner_id, 'activity_reminder', 'activity', a.id, 'Overdue activity'
		FROM activities a
		WHERE a.due_date IS NOT NULL
		  AND a.completed_at IS NULL
		  AND a.deleted_at IS NULL
		  AND a.due_date < NOW()
		  AND NOT EXISTS (
		    SELECT 1 FROM notifications n2
		    WHERE n2.entity_id = a.id
		      AND n2.kind = 'activity_reminder'
		      AND n2.title = 'Overdue activity'
		      AND n2.created_at > NOW() - interval '1 hour'
		  )
	`)
	return err
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
