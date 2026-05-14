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

type AuditLogRepo struct {
	db *pgxpool.Pool
}

func NewAuditLogRepo(db *pgxpool.Pool) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Append(ctx context.Context, entry domain.AuditEntry) error {
	var changesJSON []byte
	if len(entry.Changes) > 0 {
		b, err := json.Marshal(entry.Changes)
		if err != nil {
			return err
		}
		changesJSON = b
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO audit_log
			(org_id, user_id, agent_id, action, entity_type, entity_id, entity_name, changes, ip_address, user_agent)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		entry.OrgID, entry.UserID, entry.AgentID,
		entry.Action, entry.EntityType,
		entry.EntityID, entry.EntityName,
		changesJSON, entry.IPAddress, entry.UserAgent,
	)
	return err
}

const auditLogSelect = `
	a.id, a.org_id, a.user_id, a.agent_id,
	CASE
		WHEN a.user_id IS NOT NULL THEN 'user'
		WHEN a.agent_id IS NOT NULL THEN 'agent'
		ELSE 'system'
	END AS actor_type,
	u.name AS actor_name,
	u.email AS actor_email,
	CASE
		WHEN u.id IS NOT NULL THEN COALESCE(NULLIF(u.name, ''), u.email)
		WHEN a.user_id IS NOT NULL THEN 'Deleted user'
		WHEN a.agent_id IS NOT NULL THEN a.agent_id
		ELSE 'System'
	END AS actor_display,
	a.action, a.entity_type, a.entity_id, a.entity_name, a.changes, a.ip_address, a.user_agent, a.created_at
`

func (r *AuditLogRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.AuditLog, error) {
	e, err := scanAuditLog(r.db.QueryRow(ctx, `
		SELECT `+auditLogSelect+`
		FROM audit_log a
		LEFT JOIN users u ON u.id = a.user_id AND u.org_id = a.org_id
		WHERE a.id = $1 AND a.org_id = $2
	`, id, orgID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return e, nil
}

func (r *AuditLogRepo) List(ctx context.Context, filter domain.AuditLogFilter) ([]*domain.AuditLog, int, error) {
	where := []string{"a.org_id = $1"}
	args := []interface{}{filter.OrgID}
	idx := 2

	if filter.EntityType != nil {
		where = append(where, fmt.Sprintf("a.entity_type = $%d", idx))
		args = append(args, *filter.EntityType)
		idx++
	}
	if filter.EntityID != nil {
		where = append(where, fmt.Sprintf("a.entity_id = $%d", idx))
		args = append(args, *filter.EntityID)
		idx++
	}
	if filter.UserID != nil {
		where = append(where, fmt.Sprintf("a.user_id = $%d", idx))
		args = append(args, *filter.UserID)
		idx++
	}
	if filter.Action != nil {
		where = append(where, fmt.Sprintf("a.action = $%d", idx))
		args = append(args, *filter.Action)
		idx++
	}
	if filter.From != nil {
		where = append(where, fmt.Sprintf("a.created_at >= $%d", idx))
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		where = append(where, fmt.Sprintf("a.created_at <= $%d", idx))
		args = append(args, *filter.To)
		idx++
	}
	if strings.TrimSpace(filter.Q) != "" {
		where = append(where, fmt.Sprintf(`(
			a.action ILIKE $%d
			OR a.entity_type ILIKE $%d
			OR a.entity_id::text ILIKE $%d
			OR a.entity_name ILIKE $%d
			OR a.user_id::text ILIKE $%d
			OR a.agent_id ILIKE $%d
			OR a.ip_address ILIKE $%d
			OR u.name ILIKE $%d
			OR u.email ILIKE $%d
		)`, idx, idx, idx, idx, idx, idx, idx, idx, idx))
		args = append(args, "%"+strings.TrimSpace(filter.Q)+"%")
		idx++
	}

	clause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM audit_log a
		LEFT JOIN users u ON u.id = a.user_id AND u.org_id = a.org_id
		WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	query := fmt.Sprintf(`
		SELECT %s
		FROM audit_log a
		LEFT JOIN users u ON u.id = a.user_id AND u.org_id = a.org_id
		WHERE %s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d`, auditLogSelect, clause, idx, idx+1)

	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entries := []*domain.AuditLog{}
	for rows.Next() {
		e, err := scanAuditLog(rows)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, e)
	}
	return entries, total, rows.Err()
}

func scanAuditLog(row pgx.Row) (*domain.AuditLog, error) {
	var e domain.AuditLog
	var changesRaw []byte
	var userID *uuid.UUID
	var entityID *uuid.UUID
	var agentID, actorName, actorEmail, entityName, ipAddr, userAgent *string
	var createdAt time.Time

	err := row.Scan(
		&e.ID, &e.OrgID, &userID, &agentID,
		&e.ActorType, &actorName, &actorEmail, &e.ActorDisplay,
		&e.Action, &e.EntityType,
		&entityID, &entityName,
		&changesRaw, &ipAddr, &userAgent,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	e.UserID = userID
	e.AgentID = agentID
	e.ActorName = actorName
	e.ActorEmail = actorEmail
	e.EntityID = entityID
	e.EntityName = entityName
	e.IPAddress = ipAddr
	e.UserAgent = userAgent
	e.CreatedAt = createdAt

	if len(changesRaw) > 0 {
		if err := json.Unmarshal(changesRaw, &e.Changes); err != nil {
			return nil, err
		}
	}
	return &e, nil
}
