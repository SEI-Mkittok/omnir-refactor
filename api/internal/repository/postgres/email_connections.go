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

// EmailConnectionRepo is the Postgres implementation of repository.EmailConnectionRepository.
type EmailConnectionRepo struct {
	db *pgxpool.Pool
}

func NewEmailConnectionRepo(db *pgxpool.Pool) *EmailConnectionRepo {
	return &EmailConnectionRepo{db: db}
}

const emailConnCols = `
	id, org_id, user_id, provider, email_address,
	access_token, refresh_token, token_expiry,
	sync_cursor, last_synced_at,
	created_at, updated_at
`

func scanEmailConnection(row pgx.Row) (*domain.EmailConnection, error) {
	var c domain.EmailConnection
	err := row.Scan(
		&c.ID, &c.OrgID, &c.UserID, &c.Provider, &c.EmailAddress,
		&c.AccessToken, &c.RefreshToken, &c.TokenExpiry,
		&c.SyncCursor, &c.LastSyncedAt,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *EmailConnectionRepo) Upsert(ctx context.Context, c *domain.EmailConnection) (*domain.EmailConnection, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO email_connections
			(id, org_id, user_id, provider, email_address,
			 access_token, refresh_token, token_expiry,
			 sync_cursor, last_synced_at,
			 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (org_id, user_id, provider) DO UPDATE SET
			email_address = EXCLUDED.email_address,
			access_token  = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_expiry  = EXCLUDED.token_expiry,
			updated_at    = NOW()
		RETURNING `+emailConnCols,
		c.ID, c.OrgID, c.UserID, c.Provider, c.EmailAddress,
		c.AccessToken, c.RefreshToken, c.TokenExpiry,
		c.SyncCursor, c.LastSyncedAt,
		c.CreatedAt, c.UpdatedAt,
	)
	return scanEmailConnection(row)
}

func (r *EmailConnectionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailConnection, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+emailConnCols+` FROM email_connections WHERE id=$1`, id)
	return scanEmailConnection(row)
}

func (r *EmailConnectionRepo) GetByUserAndProvider(
	ctx context.Context,
	orgID, userID uuid.UUID,
	provider domain.EmailProvider,
) (*domain.EmailConnection, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+emailConnCols+`
		 FROM email_connections
		 WHERE org_id=$1 AND user_id=$2 AND provider=$3`,
		orgID, userID, provider)
	return scanEmailConnection(row)
}

func (r *EmailConnectionRepo) List(ctx context.Context, filter domain.EmailConnectionFilter) ([]*domain.EmailConnection, error) {
	q := `SELECT ` + emailConnCols + ` FROM email_connections WHERE org_id=$1`
	args := []any{filter.OrgID}

	if filter.UserID != nil {
		args = append(args, *filter.UserID)
		q += ` AND user_id=$` + itoa(len(args))
	}
	if filter.Provider != nil {
		args = append(args, *filter.Provider)
		q += ` AND provider=$` + itoa(len(args))
	}
	q += ` ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.EmailConnection
	for rows.Next() {
		c, err := scanEmailConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *EmailConnectionRepo) Update(ctx context.Context, id uuid.UUID, patch domain.EmailConnectionPatch) (*domain.EmailConnection, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE email_connections SET
			access_token  = COALESCE($2, access_token),
			refresh_token = COALESCE($3, refresh_token),
			token_expiry  = COALESCE($4, token_expiry),
			sync_cursor   = COALESCE($5, sync_cursor),
			last_synced_at = COALESCE($6, last_synced_at),
			email_address = COALESCE($7, email_address),
			updated_at    = NOW()
		WHERE id=$1
		RETURNING `+emailConnCols,
		id,
		patch.AccessToken, patch.RefreshToken, patch.TokenExpiry,
		patch.SyncCursor, patch.LastSyncedAt, patch.EmailAddress,
	)
	return scanEmailConnection(row)
}

func (r *EmailConnectionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM email_connections WHERE id=$1`, id)
	return err
}

func (r *EmailConnectionRepo) ListAllActive(ctx context.Context) ([]*domain.EmailConnection, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+emailConnCols+` FROM email_connections ORDER BY last_synced_at ASC NULLS FIRST`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.EmailConnection
	for rows.Next() {
		c, err := scanEmailConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ─── EmailInboxRepo ───────────────────────────────────────────────────────────

// EmailInboxRepo is the Postgres implementation of repository.EmailInboxRepository.
type EmailInboxRepo struct {
	db *pgxpool.Pool
}

func NewEmailInboxRepo(db *pgxpool.Pool) *EmailInboxRepo {
	return &EmailInboxRepo{db: db}
}

func scanInboxMessage(row pgx.Row) (*domain.EmailInboxMessage, error) {
	var m domain.EmailInboxMessage
	var toAddrsJSON []byte
	err := row.Scan(
		&m.ID, &m.OrgID, &m.ConnectionID, &m.MessageID, &m.ThreadID,
		&m.FromAddr, &toAddrsJSON, &m.Subject,
		&m.BodyText, &m.BodyHTML, &m.ContactID,
		&m.Direction, &m.SentAt, &m.ReadAt, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(toAddrsJSON) > 0 {
		_ = json.Unmarshal(toAddrsJSON, &m.ToAddrs)
	}
	if m.ToAddrs == nil {
		m.ToAddrs = []string{}
	}
	return &m, nil
}

const inboxMsgCols = `
	id, org_id, connection_id, message_id, thread_id,
	from_addr, to_addrs, subject,
	body_text, body_html, contact_id,
	direction, sent_at, read_at, created_at
`

func (r *EmailInboxRepo) Upsert(ctx context.Context, msg *domain.EmailInboxMessage) (*domain.EmailInboxMessage, error) {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	msg.CreatedAt = time.Now().UTC()

	toAddrsJSON, err := json.Marshal(msg.ToAddrs)
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO email_inbox_messages
			(id, org_id, connection_id, message_id, thread_id,
			 from_addr, to_addrs, subject,
			 body_text, body_html, contact_id,
			 direction, sent_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (connection_id, message_id) DO NOTHING
		RETURNING `+inboxMsgCols,
		msg.ID, msg.OrgID, msg.ConnectionID, msg.MessageID, msg.ThreadID,
		msg.FromAddr, toAddrsJSON, msg.Subject,
		msg.BodyText, msg.BodyHTML, msg.ContactID,
		msg.Direction, msg.SentAt, msg.CreatedAt,
	)
	result, err := scanInboxMessage(row)
	if err != nil {
		// ON CONFLICT DO NOTHING returns no rows — treat as already-exists, not an error.
		if errors.Is(err, domain.ErrNotFound) {
			return msg, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *EmailInboxRepo) List(ctx context.Context, filter domain.EmailInboxFilter) ([]*domain.EmailInboxMessage, int, error) {
	q := `SELECT ` + inboxMsgCols + ` FROM email_inbox_messages WHERE org_id=$1`
	args := []any{filter.OrgID}

	if filter.ConnectionID != nil {
		args = append(args, *filter.ConnectionID)
		q += ` AND connection_id=$` + itoa(len(args))
	}
	if filter.ContactID != nil {
		args = append(args, *filter.ContactID)
		q += ` AND contact_id=$` + itoa(len(args))
	}
	if filter.ThreadID != nil {
		args = append(args, *filter.ThreadID)
		q += ` AND thread_id=$` + itoa(len(args))
	}

	// Count query
	countQ := `SELECT COUNT(*) FROM email_inbox_messages WHERE org_id=$1`
	countArgs := args[:1]
	if filter.ConnectionID != nil {
		countArgs = append(countArgs, *filter.ConnectionID)
		countQ += ` AND connection_id=$` + itoa(len(countArgs))
	}
	if filter.ContactID != nil {
		countArgs = append(countArgs, *filter.ContactID)
		countQ += ` AND contact_id=$` + itoa(len(countArgs))
	}
	if filter.ThreadID != nil {
		countArgs = append(countArgs, *filter.ThreadID)
		countQ += ` AND thread_id=$` + itoa(len(countArgs))
	}

	var total int
	if err := r.db.QueryRow(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*domain.EmailInboxMessage{}, 0, nil
	}

	q += ` ORDER BY sent_at DESC`
	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		q += ` LIMIT $` + itoa(len(args))
	}
	if filter.Page > 0 && filter.Limit > 0 {
		args = append(args, (filter.Page-1)*filter.Limit)
		q += ` OFFSET $` + itoa(len(args))
	}

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*domain.EmailInboxMessage
	for rows.Next() {
		m, err := scanInboxMessage(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func (r *EmailInboxRepo) ListThreads(ctx context.Context, filter domain.EmailInboxFilter) ([]*domain.EmailInboxThreadSummary, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	where := []string{"org_id = $1"}
	args := []any{filter.OrgID}
	i := 2

	if filter.ConnectionID != nil {
		where = append(where, fmt.Sprintf("connection_id = $%d", i))
		args = append(args, *filter.ConnectionID)
		i++
	}
	if filter.ContactID != nil {
		where = append(where, fmt.Sprintf("contact_id = $%d", i))
		args = append(args, *filter.ContactID)
		i++
	}
	if filter.UnreadOnly {
		where = append(where, "direction = 'inbound' AND read_at IS NULL")
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := `
		WITH filtered AS (
			SELECT thread_id
			FROM email_inbox_messages
			WHERE ` + whereClause + `
		)
		SELECT COUNT(DISTINCT thread_id) FROM filtered`
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*domain.EmailInboxThreadSummary{}, 0, nil
	}

	offset := (filter.Page - 1) * filter.Limit
	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, filter.Limit, offset)
	limitIdx := len(queryArgs) - 1
	offsetIdx := len(queryArgs)

	query := `
		WITH filtered AS (
			SELECT *
			FROM email_inbox_messages
			WHERE ` + whereClause + `
		),
		latest AS (
			SELECT DISTINCT ON (thread_id)
				thread_id,
				org_id,
				connection_id,
				subject,
				COALESCE(NULLIF(body_text, ''), '') AS snippet,
				sent_at,
				contact_id
			FROM filtered
			ORDER BY thread_id, sent_at DESC, created_at DESC
		),
		aggregated AS (
			SELECT
				thread_id,
				org_id,
				connection_id,
				COUNT(*) AS message_count,
				BOOL_OR(direction = 'inbound' AND read_at IS NULL) AS unread,
				MAX(sent_at) AS last_message_at,
				(array_agg(contact_id) FILTER (WHERE contact_id IS NOT NULL))[1] AS contact_id
			FROM filtered
			GROUP BY thread_id, org_id, connection_id
		),
		participants AS (
			SELECT
				thread_id,
				array_agg(DISTINCT participant ORDER BY participant) AS participants
			FROM (
				SELECT thread_id, from_addr AS participant FROM filtered
				UNION ALL
				SELECT f.thread_id, addr.value AS participant
				FROM filtered f
				CROSS JOIN LATERAL jsonb_array_elements_text(f.to_addrs) AS addr(value)
			) p
			GROUP BY thread_id
		)
		SELECT
			a.thread_id,
			a.org_id,
			a.connection_id,
			l.subject,
			COALESCE(p.participants, ARRAY[]::text[]),
			l.snippet,
			a.unread,
			a.message_count,
			a.last_message_at,
			COALESCE(a.contact_id, l.contact_id) AS contact_id
		FROM aggregated a
		JOIN latest l ON l.thread_id = a.thread_id
		LEFT JOIN participants p ON p.thread_id = a.thread_id
		ORDER BY a.last_message_at DESC
		LIMIT $` + itoa(limitIdx) + ` OFFSET $` + itoa(offsetIdx)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*domain.EmailInboxThreadSummary
	for rows.Next() {
		var summary domain.EmailInboxThreadSummary
		if err := rows.Scan(
			&summary.ThreadID,
			&summary.OrgID,
			&summary.ConnectionID,
			&summary.Subject,
			&summary.Participants,
			&summary.Snippet,
			&summary.Unread,
			&summary.MessageCount,
			&summary.LastMessageAt,
			&summary.ContactID,
		); err != nil {
			return nil, 0, err
		}
		if summary.Participants == nil {
			summary.Participants = []string{}
		}
		out = append(out, &summary)
	}
	return out, total, rows.Err()
}

func (r *EmailInboxRepo) GetThread(ctx context.Context, orgID uuid.UUID, threadID string) ([]*domain.EmailInboxMessage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+inboxMsgCols+`
		FROM email_inbox_messages
		WHERE org_id=$1 AND thread_id=$2
		ORDER BY sent_at ASC`,
		orgID, threadID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.EmailInboxMessage
	for rows.Next() {
		m, err := scanInboxMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *EmailInboxRepo) LinkContact(ctx context.Context, orgID uuid.UUID, addr string, contactID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE email_inbox_messages
		SET contact_id=$3
		WHERE org_id=$1 AND from_addr=$2 AND contact_id IS NULL`,
		orgID, addr, contactID,
	)
	return err
}

// MarkThreadRead sets read_at = NOW() on all unread messages in the thread.
func (r *EmailInboxRepo) MarkThreadRead(ctx context.Context, orgID uuid.UUID, threadID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE email_inbox_messages
		SET read_at = NOW()
		WHERE org_id = $1 AND thread_id = $2 AND read_at IS NULL`,
		orgID, threadID,
	)
	return err
}
