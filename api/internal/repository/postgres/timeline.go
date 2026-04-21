package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// TimelineRepo is the Postgres implementation of repository.TimelineRepository.
type TimelineRepo struct {
	db *pgxpool.Pool
}

func NewTimelineRepo(db *pgxpool.Pool) *TimelineRepo {
	return &TimelineRepo{db: db}
}

func (r *TimelineRepo) List(ctx context.Context, f domain.TimelineFilter) ([]*domain.TimelineEvent, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if !hasCtxOrg {
		orgID = f.OrgID
	}
	if orgID == uuid.Nil {
		return []*domain.TimelineEvent{}, 0, nil
	}

	rows, err := r.db.Query(ctx, timelineListSQL,
		orgID,
		f.AccountID,
		f.ContactID,
		f.OccurredAtGTE,
		f.OccurredAtLTE,
		f.Limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	events := make([]*domain.TimelineEvent, 0, f.Limit)
	total := 0
	for rows.Next() {
		var evt domain.TimelineEvent
		var refsRaw []byte
		if err := rows.Scan(
			&evt.EventType,
			&evt.EventID,
			&evt.OccurredAt,
			&evt.ActorID,
			&refsRaw,
			&evt.Preview,
			&total,
		); err != nil {
			return nil, 0, err
		}
		if len(refsRaw) > 0 {
			if err := json.Unmarshal(refsRaw, &evt.EntityRefs); err != nil {
				return nil, 0, err
			}
		}
		events = append(events, &evt)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

const timelineListSQL = `
WITH p AS (
	SELECT
		$1::uuid AS org_id,
		$2::uuid AS account_id,
		$3::uuid AS contact_id,
		$4::timestamptz AS occurred_from,
		$5::timestamptz AS occurred_to
), timeline_union AS (
	SELECT
		'activity.created'::text AS event_type,
		a.id AS event_id,
		a.created_at AS occurred_at,
		a.owner_id AS actor_id,
		jsonb_strip_nulls(jsonb_build_array(
			CASE WHEN a.account_id IS NOT NULL THEN jsonb_build_object('entity_type', 'account', 'entity_id', a.account_id) END,
			CASE WHEN a.contact_id IS NOT NULL THEN jsonb_build_object('entity_type', 'contact', 'entity_id', a.contact_id) END,
			CASE WHEN a.deal_id IS NOT NULL THEN jsonb_build_object('entity_type', 'deal', 'entity_id', a.deal_id) END,
			jsonb_build_object('entity_type', 'activity', 'entity_id', a.id)
		)) AS entity_refs,
		LEFT(a.subject, 240) AS preview
	FROM activities a, p
	WHERE a.deleted_at IS NULL
		AND a.org_id = p.org_id
		AND (p.account_id IS NULL OR a.account_id = p.account_id)
		AND (p.contact_id IS NULL OR a.contact_id = p.contact_id)
		AND (p.occurred_from IS NULL OR a.created_at >= p.occurred_from)
		AND (p.occurred_to IS NULL OR a.created_at <= p.occurred_to)

	UNION ALL

	SELECT
		'ticket.created'::text,
		t.id,
		t.created_at,
		t.submitted_by_user_id,
		jsonb_strip_nulls(jsonb_build_array(
			CASE WHEN t.account_id IS NOT NULL THEN jsonb_build_object('entity_type', 'account', 'entity_id', t.account_id) END,
			CASE WHEN t.contact_id IS NOT NULL THEN jsonb_build_object('entity_type', 'contact', 'entity_id', t.contact_id) END,
			jsonb_build_object('entity_type', 'ticket', 'entity_id', t.id)
		)),
		LEFT(t.subject, 240)
	FROM tickets t, p
	WHERE t.deleted_at IS NULL
		AND t.org_id = p.org_id
		AND (p.account_id IS NULL OR t.account_id = p.account_id)
		AND (p.contact_id IS NULL OR t.contact_id = p.contact_id)
		AND (p.occurred_from IS NULL OR t.created_at >= p.occurred_from)
		AND (p.occurred_to IS NULL OR t.created_at <= p.occurred_to)

	UNION ALL

	SELECT
		'ticket_comment.created'::text,
		tc.id,
		tc.created_at,
		tc.author_id,
		jsonb_strip_nulls(jsonb_build_array(
			CASE WHEN t.account_id IS NOT NULL THEN jsonb_build_object('entity_type', 'account', 'entity_id', t.account_id) END,
			CASE WHEN t.contact_id IS NOT NULL THEN jsonb_build_object('entity_type', 'contact', 'entity_id', t.contact_id) END,
			jsonb_build_object('entity_type', 'ticket', 'entity_id', t.id),
			jsonb_build_object('entity_type', 'ticket_comment', 'entity_id', tc.id)
		)),
		LEFT(tc.body, 240)
	FROM ticket_comments tc
	JOIN tickets t ON t.id = tc.ticket_id
	JOIN p ON TRUE
	WHERE tc.deleted_at IS NULL
		AND t.deleted_at IS NULL
		AND tc.org_id = p.org_id
		AND (p.account_id IS NULL OR t.account_id = p.account_id)
		AND (p.contact_id IS NULL OR t.contact_id = p.contact_id)
		AND (p.occurred_from IS NULL OR tc.created_at >= p.occurred_from)
		AND (p.occurred_to IS NULL OR tc.created_at <= p.occurred_to)

	UNION ALL

	SELECT
		'note.created'::text,
		n.id,
		n.created_at,
		n.author_id,
		jsonb_build_array(
			jsonb_build_object('entity_type', n.entity_type, 'entity_id', n.entity_id),
			jsonb_build_object('entity_type', 'note', 'entity_id', n.id)
		),
		LEFT(n.content, 240)
	FROM notes n, p
	WHERE n.deleted_at IS NULL
		AND n.org_id = p.org_id
		AND (
			(p.account_id IS NULL AND p.contact_id IS NULL)
			OR (p.account_id IS NOT NULL AND n.entity_type = 'account' AND n.entity_id = p.account_id)
			OR (p.contact_id IS NOT NULL AND n.entity_type = 'contact' AND n.entity_id = p.contact_id)
		)
		AND (p.occurred_from IS NULL OR n.created_at >= p.occurred_from)
		AND (p.occurred_to IS NULL OR n.created_at <= p.occurred_to)

	UNION ALL

	SELECT
		'email.logged'::text,
		ce.id,
		ce.sent_at,
		NULL::uuid,
		jsonb_strip_nulls(jsonb_build_array(
			CASE WHEN ce.contact_id IS NOT NULL THEN jsonb_build_object('entity_type', 'contact', 'entity_id', ce.contact_id) END,
			CASE WHEN ce.deal_id IS NOT NULL THEN jsonb_build_object('entity_type', 'deal', 'entity_id', ce.deal_id) END,
			jsonb_build_object('entity_type', 'email', 'entity_id', ce.id)
		)),
		LEFT(ce.subject, 240)
	FROM contact_emails ce
	LEFT JOIN deals d ON d.id = ce.deal_id
	JOIN p ON TRUE
	WHERE ce.org_id = p.org_id
		AND (p.account_id IS NULL OR d.account_id = p.account_id)
		AND (p.contact_id IS NULL OR ce.contact_id = p.contact_id)
		AND (p.occurred_from IS NULL OR ce.sent_at >= p.occurred_from)
		AND (p.occurred_to IS NULL OR ce.sent_at <= p.occurred_to)

	UNION ALL

	SELECT
		'email.inbox_message'::text,
		eim.id,
		eim.sent_at,
		ec.user_id,
		jsonb_strip_nulls(jsonb_build_array(
			CASE WHEN eim.contact_id IS NOT NULL THEN jsonb_build_object('entity_type', 'contact', 'entity_id', eim.contact_id) END,
			jsonb_build_object('entity_type', 'email_inbox_message', 'entity_id', eim.id)
		)),
		LEFT(eim.subject, 240)
	FROM email_inbox_messages eim
	JOIN email_connections ec ON ec.id = eim.connection_id
	JOIN p ON TRUE
	WHERE eim.org_id = p.org_id
		AND (p.contact_id IS NULL OR eim.contact_id = p.contact_id)
		AND p.account_id IS NULL
		AND (p.occurred_from IS NULL OR eim.sent_at >= p.occurred_from)
		AND (p.occurred_to IS NULL OR eim.sent_at <= p.occurred_to)

	UNION ALL

	SELECT
		'sequence.event'::text,
		se.id,
		se.occurred_at,
		NULL::uuid,
		jsonb_build_array(
			jsonb_build_object('entity_type', 'contact', 'entity_id', se.contact_id),
			jsonb_build_object('entity_type', 'sequence', 'entity_id', se.sequence_id),
			jsonb_build_object('entity_type', 'sequence_event', 'entity_id', se.id)
		),
		LEFT(se.kind::text, 240)
	FROM sequence_events se, p
	WHERE se.org_id = p.org_id
		AND (p.contact_id IS NULL OR se.contact_id = p.contact_id)
		AND p.account_id IS NULL
		AND (p.occurred_from IS NULL OR se.occurred_at >= p.occurred_from)
		AND (p.occurred_to IS NULL OR se.occurred_at <= p.occurred_to)

	UNION ALL

	SELECT
		'quote.created'::text,
		q.id,
		q.created_at,
		q.created_by,
		jsonb_strip_nulls(jsonb_build_array(
			CASE WHEN q.contact_id IS NOT NULL THEN jsonb_build_object('entity_type', 'contact', 'entity_id', q.contact_id) END,
			CASE WHEN q.deal_id IS NOT NULL THEN jsonb_build_object('entity_type', 'deal', 'entity_id', q.deal_id) END,
			CASE WHEN d.account_id IS NOT NULL THEN jsonb_build_object('entity_type', 'account', 'entity_id', d.account_id) END,
			jsonb_build_object('entity_type', 'quote', 'entity_id', q.id)
		)),
		LEFT(q.title, 240)
	FROM quotes q
	LEFT JOIN deals d ON d.id = q.deal_id
	JOIN p ON TRUE
	WHERE q.org_id = p.org_id
		AND (p.account_id IS NULL OR d.account_id = p.account_id)
		AND (p.contact_id IS NULL OR q.contact_id = p.contact_id)
		AND (p.occurred_from IS NULL OR q.created_at >= p.occurred_from)
		AND (p.occurred_to IS NULL OR q.created_at <= p.occurred_to)
)
SELECT
	event_type,
	event_id,
	occurred_at,
	actor_id,
	entity_refs,
	preview,
	COUNT(*) OVER() AS total_count
FROM timeline_union
ORDER BY occurred_at DESC, event_id DESC
LIMIT $6 OFFSET $7;
`
