package postgres

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// OutboundWebhookRepo implements repository.OutboundWebhookRepository.
type OutboundWebhookRepo struct {
	db *pgxpool.Pool
}

func NewOutboundWebhookRepo(db *pgxpool.Pool) *OutboundWebhookRepo {
	return &OutboundWebhookRepo{db: db}
}

func scanWebhook(row pgx.Row) (*domain.Webhook, error) {
	var w domain.Webhook
	var events []string
	err := row.Scan(
		&w.ID, &w.OrgID, &w.URL, &events, &w.Secret, &w.Active, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	for _, e := range events {
		w.Events = append(w.Events, domain.WebhookEvent(e))
	}
	return &w, nil
}

func (r *OutboundWebhookRepo) Create(ctx context.Context, w *domain.Webhook) (*domain.Webhook, error) {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		w.OrgID = orgID
	}
	now := time.Now().UTC()
	w.CreatedAt = now
	w.UpdatedAt = now
	w.Active = true

	events := make([]string, len(w.Events))
	for i, e := range w.Events {
		events[i] = string(e)
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO outbound_webhooks (id, org_id, url, events, secret, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, org_id, url, events, secret, active, created_at, updated_at`,
		w.ID, w.OrgID, w.URL, events, w.Secret, w.Active, w.CreatedAt, w.UpdatedAt,
	)
	return scanWebhook(row)
}

func (r *OutboundWebhookRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	q := `SELECT id, org_id, url, events, secret, active, created_at, updated_at
		FROM outbound_webhooks WHERE id=$1`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	row := r.db.QueryRow(ctx, q, args...)
	return scanWebhook(row)
}

func (r *OutboundWebhookRepo) List(ctx context.Context) ([]*domain.Webhook, error) {
	q := `SELECT id, org_id, url, events, secret, active, created_at, updated_at
		FROM outbound_webhooks`
	args := []any{}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` WHERE org_id=$1`
		args = append(args, orgID)
	}
	q += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Webhook
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *OutboundWebhookRepo) ListByEvent(ctx context.Context, orgID uuid.UUID, event domain.WebhookEvent) ([]*domain.Webhook, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, org_id, url, events, secret, active, created_at, updated_at
		FROM outbound_webhooks
		WHERE org_id=$1 AND active=TRUE AND $2=ANY(events)
		ORDER BY created_at`,
		orgID, string(event),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Webhook
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *OutboundWebhookRepo) Update(ctx context.Context, id uuid.UUID, patch domain.WebhookPatch) (*domain.Webhook, error) {
	sets := []string{"updated_at=NOW()"}
	args := []any{}
	n := 1

	if patch.URL != nil {
		sets = append(sets, "url="+phN(n))
		args = append(args, *patch.URL)
		n++
	}
	if len(patch.Events) > 0 {
		events := make([]string, len(patch.Events))
		for i, e := range patch.Events {
			events[i] = string(e)
		}
		sets = append(sets, "events="+phN(n))
		args = append(args, events)
		n++
	}
	if patch.Active != nil {
		sets = append(sets, "active="+phN(n))
		args = append(args, *patch.Active)
		n++
	}

	args = append(args, id)
	q := `UPDATE outbound_webhooks SET ` + strings.Join(sets, ", ") +
		` WHERE id=` + phN(n) +
		` RETURNING id, org_id, url, events, secret, active, created_at, updated_at`
	row := r.db.QueryRow(ctx, q, args...)
	return scanWebhook(row)
}

func (r *OutboundWebhookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM outbound_webhooks WHERE id=$1`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	_, err := r.db.Exec(ctx, q, args...)
	return err
}

func (r *OutboundWebhookRepo) CreateDelivery(ctx context.Context, d *domain.WebhookDelivery) (*domain.WebhookDelivery, error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	d.CreatedAt = time.Now().UTC()
	d.Status = domain.DeliveryStatusPending

	row := r.db.QueryRow(ctx, `
		INSERT INTO webhook_deliveries (id, webhook_id, event, payload, status, attempts, next_retry_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, webhook_id, event, payload, status, attempts, next_retry_at, delivered_at, last_error, created_at`,
		d.ID, d.WebhookID, string(d.Event), d.Payload, string(d.Status), d.Attempts, d.NextRetryAt, d.CreatedAt,
	)
	return scanDelivery(row)
}

func (r *OutboundWebhookRepo) UpdateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	_, err := r.db.Exec(ctx, `
		UPDATE webhook_deliveries
		SET status=$2, attempts=$3, next_retry_at=$4, delivered_at=$5, last_error=$6
		WHERE id=$1`,
		d.ID, string(d.Status), d.Attempts, d.NextRetryAt, d.DeliveredAt, d.LastError,
	)
	return err
}

func (r *OutboundWebhookRepo) PendingDeliveries(ctx context.Context) ([]*domain.WebhookDelivery, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, webhook_id, event, payload, status, attempts, next_retry_at, delivered_at, last_error, created_at
		FROM webhook_deliveries
		WHERE status='pending' AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at
		LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.WebhookDelivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func scanDelivery(row pgx.Row) (*domain.WebhookDelivery, error) {
	var d domain.WebhookDelivery
	var event string
	var status string
	err := row.Scan(
		&d.ID, &d.WebhookID, &event, &d.Payload, &status,
		&d.Attempts, &d.NextRetryAt, &d.DeliveredAt, &d.LastError, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	d.Event = domain.WebhookEvent(event)
	d.Status = domain.DeliveryStatus(status)
	return &d, nil
}

// ListDeliveries returns recent delivery attempts for a webhook.
func (r *OutboundWebhookRepo) ListDeliveries(ctx context.Context, webhookID uuid.UUID, limit int) ([]*domain.WebhookDelivery, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, webhook_id, event, payload, status, attempts, next_retry_at, delivered_at, last_error, created_at
		FROM webhook_deliveries
		WHERE webhook_id=$1
		ORDER BY created_at DESC
		LIMIT $2`,
		webhookID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.WebhookDelivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// phN returns "$N" for SQL parameter placeholders.
func phN(n int) string {
	return "$" + strconv.Itoa(n)
}
