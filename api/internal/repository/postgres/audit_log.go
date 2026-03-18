package postgres

import (
	"context"
	"encoding/json"
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

func (r *AuditLogRepo) List(ctx context.Context, filter domain.AuditLogFilter) ([]*domain.AuditLog, int, error) {
	where := []string{"org_id = $1"}
	args := []interface{}{filter.OrgID}
	idx := 2

	if filter.EntityType != nil {
		where = append(where, fmt.Sprintf("entity_type = $%d", idx))
		args = append(args, *filter.EntityType)
		idx++
	}
	if filter.EntityID != nil {
		where = append(where, fmt.Sprintf("entity_id = $%d", idx))
		args = append(args, *filter.EntityID)
		idx++
	}
	if filter.UserID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", idx))
		args = append(args, *filter.UserID)
		idx++
	}
	if filter.Action != nil {
		where = append(where, fmt.Sprintf("action = $%d", idx))
		args = append(args, *filter.Action)
		idx++
	}
	if filter.From != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", idx))
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", idx))
		args = append(args, *filter.To)
		idx++
	}

	clause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM audit_log WHERE "+clause, args...).Scan(&total); err != nil {
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
		SELECT id, org_id, user_id, agent_id, action, entity_type, entity_id, entity_name, changes, ip_address, user_agent, created_at
		FROM audit_log
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, clause, idx, idx+1)

	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []*domain.AuditLog
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
	var agentID, entityName, ipAddr, userAgent *string
	var createdAt time.Time

	err := row.Scan(
		&e.ID, &e.OrgID, &userID, &agentID,
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
