package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// PushSubscriptionRepo implements repository.PushSubscriptionRepository.
type PushSubscriptionRepo struct {
	db *pgxpool.Pool
}

func NewPushSubscriptionRepo(db *pgxpool.Pool) *PushSubscriptionRepo {
	return &PushSubscriptionRepo{db: db}
}

func (r *PushSubscriptionRepo) Upsert(ctx context.Context, s *domain.PushSubscription) (*domain.PushSubscription, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO push_subscriptions (user_id, org_id, endpoint, p256dh, auth)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, endpoint) DO UPDATE SET
			p256dh = EXCLUDED.p256dh,
			auth   = EXCLUDED.auth
		RETURNING id, user_id, org_id, endpoint, p256dh, auth, created_at
	`, s.UserID, s.OrgID, s.Endpoint, s.P256dh, s.Auth)

	out := &domain.PushSubscription{}
	err := row.Scan(&out.ID, &out.UserID, &out.OrgID, &out.Endpoint, &out.P256dh, &out.Auth, &out.CreatedAt)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PushSubscriptionRepo) DeleteByEndpoint(ctx context.Context, userID uuid.UUID, endpoint string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2
	`, userID, endpoint)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PushSubscriptionRepo) ListByUser(ctx context.Context, userID, orgID uuid.UUID) ([]*domain.PushSubscription, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, org_id, endpoint, p256dh, auth, created_at
		FROM push_subscriptions
		WHERE user_id = $1 AND org_id = $2
		ORDER BY created_at ASC
	`, userID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*domain.PushSubscription
	for rows.Next() {
		s := &domain.PushSubscription{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.OrgID, &s.Endpoint, &s.P256dh, &s.Auth, &s.CreatedAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				break
			}
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}
