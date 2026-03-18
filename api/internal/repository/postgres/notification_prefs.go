package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// NotificationPrefRepo implements repository.NotificationPrefRepository.
type NotificationPrefRepo struct {
	db *pgxpool.Pool
}

func NewNotificationPrefRepo(db *pgxpool.Pool) *NotificationPrefRepo {
	return &NotificationPrefRepo{db: db}
}

// GetByUser returns the stored pref or a default (all-enabled) struct when no row exists.
func (r *NotificationPrefRepo) GetByUser(ctx context.Context, userID, orgID uuid.UUID) (*domain.UserNotificationPref, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, org_id, email_on_assigned, email_on_resolved, email_on_closed, created_at, updated_at
		FROM user_notification_prefs
		WHERE user_id = $1 AND org_id = $2
	`, userID, orgID)

	p := &domain.UserNotificationPref{}
	err := row.Scan(
		&p.ID, &p.UserID, &p.OrgID,
		&p.EmailOnAssigned, &p.EmailOnResolved, &p.EmailOnClosed,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// Return a default-on pref (not persisted).
		return &domain.UserNotificationPref{
			UserID:          userID,
			OrgID:           orgID,
			EmailOnAssigned: true,
			EmailOnResolved: true,
			EmailOnClosed:   true,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Upsert inserts or updates the notification pref for a user+org.
func (r *NotificationPrefRepo) Upsert(ctx context.Context, pref *domain.UserNotificationPref) (*domain.UserNotificationPref, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO user_notification_prefs (user_id, org_id, email_on_assigned, email_on_resolved, email_on_closed)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, org_id) DO UPDATE SET
			email_on_assigned = EXCLUDED.email_on_assigned,
			email_on_resolved = EXCLUDED.email_on_resolved,
			email_on_closed   = EXCLUDED.email_on_closed,
			updated_at        = NOW()
		RETURNING id, user_id, org_id, email_on_assigned, email_on_resolved, email_on_closed, created_at, updated_at
	`, pref.UserID, pref.OrgID, pref.EmailOnAssigned, pref.EmailOnResolved, pref.EmailOnClosed)

	p := &domain.UserNotificationPref{}
	err := row.Scan(
		&p.ID, &p.UserID, &p.OrgID,
		&p.EmailOnAssigned, &p.EmailOnResolved, &p.EmailOnClosed,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}
