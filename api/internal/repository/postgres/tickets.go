package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// TicketRepo implements repository.TicketRepository.
type TicketRepo struct {
	db *pgxpool.Pool
}

func NewTicketRepo(db *pgxpool.Pool) *TicketRepo {
	return &TicketRepo{db: db}
}

const ticketCols = `
	id, org_id, subject, description, status, priority,
	assignee_id, contact_id, account_id, source, email_message_id, tags,
	custom_fields, sla_policy_id, first_responded_at, submitted_by_user_id, created_at, updated_at, deleted_at
`

func scanTicket(row pgx.Row) (*domain.Ticket, error) {
	var t domain.Ticket
	err := row.Scan(
		&t.ID, &t.OrgID, &t.Subject, &t.Description, &t.Status, &t.Priority,
		&t.AssigneeID, &t.ContactID, &t.AccountID, &t.Source, &t.EmailMessageID, &t.Tags,
		&t.CustomFields, &t.SLAPolicyID, &t.FirstRespondedAt, &t.SubmittedByUserID, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepo) Create(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		t.OrgID = orgID
	}
	if t.Status == "" {
		t.Status = domain.TicketStatusOpen
	}
	if t.Priority == "" {
		t.Priority = domain.TicketPriorityMedium
	}
	if t.Tags == nil {
		t.Tags = []string{}
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now

	// Auto-assign a matching SLA policy based on ticket priority.
	var slaID *uuid.UUID
	{
		var policyID uuid.UUID
		err := r.db.QueryRow(ctx, `
			SELECT id FROM sla_policies
			WHERE org_id = $1
			  AND priority_filter @> jsonb_build_array($2::text)
			ORDER BY created_at
			LIMIT 1`,
			t.OrgID, string(t.Priority),
		).Scan(&policyID)
		if err == nil {
			slaID = &policyID
		}
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO tickets
			(id, org_id, subject, description, status, priority,
			 assignee_id, contact_id, account_id, source, email_message_id, tags,
			 custom_fields, sla_policy_id, submitted_by_user_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING `+ticketCols,
		t.ID, t.OrgID, t.Subject, t.Description, t.Status, t.Priority,
		t.AssigneeID, t.ContactID, t.AccountID, t.Source, t.EmailMessageID, t.Tags,
		t.CustomFields, slaID, t.SubmittedByUserID, t.CreatedAt, t.UpdatedAt,
	)
	return scanTicket(row)
}

func (r *TicketRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	q := `SELECT ` + ticketCols + ` FROM tickets WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	return scanTicket(r.db.QueryRow(ctx, q, args...))
}

func (r *TicketRepo) GetByEmailMessageID(ctx context.Context, messageID string) (*domain.Ticket, error) {
	q := `SELECT ` + ticketCols + ` FROM tickets WHERE email_message_id=$1 AND deleted_at IS NULL`
	args := []any{messageID}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	return scanTicket(r.db.QueryRow(ctx, q, args...))
}

func (r *TicketRepo) Update(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.Subject != nil {
		addArg("subject", *patch.Subject)
	}
	if patch.Description != nil {
		addArg("description", *patch.Description)
	}
	if patch.Status != nil {
		addArg("status", *patch.Status)
	}
	if patch.Priority != nil {
		addArg("priority", *patch.Priority)
	}
	if patch.AssigneeID != nil {
		addArg("assignee_id", *patch.AssigneeID)
	}
	if patch.ContactID != nil {
		addArg("contact_id", *patch.ContactID)
	}
	if patch.AccountID != nil {
		addArg("account_id", *patch.AccountID)
	}
	if patch.Source != nil {
		addArg("source", *patch.Source)
	}
	if len(patch.CustomFields) > 0 {
		addArg("custom_fields", []byte(patch.CustomFields))
	}

	whereClause := fmt.Sprintf(`id=$%d AND deleted_at IS NULL`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
	}

	query := fmt.Sprintf(
		`UPDATE tickets SET %s WHERE %s RETURNING %s`,
		strings.Join(sets, ", "), whereClause, ticketCols,
	)
	return scanTicket(r.db.QueryRow(ctx, query, args...))
}

func (r *TicketRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE tickets SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TicketRepo) List(ctx context.Context, f domain.TicketFilter) ([]*domain.Ticket, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"deleted_at IS NULL"}
	args := []any{}
	i := 1

	addWhere := func(expr string, val any) {
		where = append(where, fmt.Sprintf("%s = $%d", expr, i))
		args = append(args, val)
		i++
	}

	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if !hasCtxOrg {
		orgID = f.OrgID
	}
	if orgID != uuid.Nil {
		addWhere("org_id", orgID)
	}
	if f.Status != nil {
		addWhere("status", *f.Status)
	}
	if f.Priority != nil {
		addWhere("priority", *f.Priority)
	}
	if f.AssigneeID != nil {
		addWhere("assignee_id", *f.AssigneeID)
	}
	if f.ContactID != nil {
		addWhere("contact_id", *f.ContactID)
	}
	if f.SubmittedByUserID != nil {
		addWhere("submitted_by_user_id", *f.SubmittedByUserID)
	}
	if f.Q != "" {
		where = append(where, fmt.Sprintf(
			`to_tsvector('english', subject || ' ' || coalesce(description, '')) @@ plainto_tsquery('english', $%d)`, i,
		))
		args = append(args, f.Q)
		i++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM tickets WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortCol := "created_at"
	allowedSorts := map[string]bool{
		"created_at": true, "updated_at": true,
		"subject": true, "status": true, "priority": true,
	}
	if allowedSorts[f.Sort] {
		sortCol = f.Sort
	}
	order := "DESC"
	if strings.ToUpper(f.Order) == "ASC" {
		order = "ASC"
	}

	rows, err := r.db.Query(ctx,
		fmt.Sprintf(
			`SELECT %s FROM tickets WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
			ticketCols, whereClause, sortCol, order, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(
			&t.ID, &t.OrgID, &t.Subject, &t.Description, &t.Status, &t.Priority,
			&t.AssigneeID, &t.ContactID, &t.AccountID, &t.Source, &t.EmailMessageID, &t.Tags,
			&t.CustomFields, &t.SLAPolicyID, &t.FirstRespondedAt, &t.SubmittedByUserID, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, total, rows.Err()
}

// TicketCommentRepo implements repository.TicketCommentRepository.
type TicketCommentRepo struct {
	db *pgxpool.Pool
}

func NewTicketCommentRepo(db *pgxpool.Pool) *TicketCommentRepo {
	return &TicketCommentRepo{db: db}
}

const commentCols = `id, ticket_id, org_id, author_id, body, is_internal, created_at, updated_at, deleted_at`

func scanComment(row pgx.Row) (*domain.TicketComment, error) {
	var c domain.TicketComment
	err := row.Scan(
		&c.ID, &c.TicketID, &c.OrgID, &c.AuthorID, &c.Body, &c.IsInternal,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *TicketCommentRepo) Create(ctx context.Context, c *domain.TicketComment) (*domain.TicketComment, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		c.OrgID = orgID
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO ticket_comments
			(id, ticket_id, org_id, author_id, body, is_internal, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING `+commentCols,
		c.ID, c.TicketID, c.OrgID, c.AuthorID, c.Body, c.IsInternal, c.CreatedAt, c.UpdatedAt,
	)
	comment, err := scanComment(row)
	if err != nil {
		return nil, err
	}

	// Set first_responded_at on the parent ticket when a public (non-internal)
	// comment is posted and the ticket has not yet been responded to.
	if !c.IsInternal {
		_, _ = r.db.Exec(ctx, `
			UPDATE tickets
			SET first_responded_at = $1
			WHERE id = $2 AND first_responded_at IS NULL`,
			now, c.TicketID,
		)
	}

	return comment, nil
}

func (r *TicketCommentRepo) List(ctx context.Context, f domain.TicketCommentFilter) ([]*domain.TicketComment, error) {
	where := []string{"ticket_id = $1", "deleted_at IS NULL"}
	args := []any{f.TicketID}
	i := 2

	if orgID, hasCtx := domain.OrgIDFromContext(ctx); hasCtx {
		where = append(where, fmt.Sprintf("org_id = $%d", i))
		args = append(args, orgID)
		i++
	} else if f.OrgID != uuid.Nil {
		where = append(where, fmt.Sprintf("org_id = $%d", i))
		args = append(args, f.OrgID)
		i++
	}

	if f.IsInternal != nil {
		where = append(where, fmt.Sprintf("is_internal = $%d", i))
		args = append(args, *f.IsInternal)
	}

	rows, err := r.db.Query(ctx,
		`SELECT `+commentCols+` FROM ticket_comments WHERE `+
			strings.Join(where, " AND ")+
			` ORDER BY created_at ASC`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.TicketComment
	for rows.Next() {
		var c domain.TicketComment
		if err := rows.Scan(
			&c.ID, &c.TicketID, &c.OrgID, &c.AuthorID, &c.Body, &c.IsInternal,
			&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	return comments, rows.Err()
}

func (r *TicketCommentRepo) Delete(ctx context.Context, id, ticketID uuid.UUID) error {
	q := `UPDATE ticket_comments SET deleted_at=NOW() WHERE id=$1 AND ticket_id=$2 AND deleted_at IS NULL`
	args := []any{id, ticketID}

	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// TicketAttachmentRepo implements repository.TicketAttachmentRepository.
type TicketAttachmentRepo struct {
	db *pgxpool.Pool
}

func NewTicketAttachmentRepo(db *pgxpool.Pool) *TicketAttachmentRepo {
	return &TicketAttachmentRepo{db: db}
}

const attachmentCols = `id, ticket_id, org_id, uploaded_by, filename, content_type, size_bytes, storage_url, storage_key, storage_backend, created_at`

func scanAttachment(row pgx.Row) (*domain.TicketAttachment, error) {
	var a domain.TicketAttachment
	var storageKey, storageBackend *string
	err := row.Scan(
		&a.ID, &a.TicketID, &a.OrgID, &a.UploadedBy, &a.Filename,
		&a.ContentType, &a.SizeBytes, &a.StorageURL, &storageKey, &storageBackend, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if storageKey != nil {
		a.StorageKey = *storageKey
	}
	if storageBackend != nil {
		a.StorageBackend = *storageBackend
	}
	return &a, nil
}

func (r *TicketAttachmentRepo) Create(ctx context.Context, a *domain.TicketAttachment) (*domain.TicketAttachment, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		a.OrgID = orgID
	}
	a.CreatedAt = time.Now().UTC()

	var storageKey, storageBackend *string
	if a.StorageKey != "" {
		storageKey = &a.StorageKey
	}
	if a.StorageBackend != "" {
		storageBackend = &a.StorageBackend
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO ticket_attachments
			(id, ticket_id, org_id, uploaded_by, filename, content_type, size_bytes, storage_url, storage_key, storage_backend, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+attachmentCols,
		a.ID, a.TicketID, a.OrgID, a.UploadedBy, a.Filename,
		a.ContentType, a.SizeBytes, a.StorageURL, storageKey, storageBackend, a.CreatedAt,
	)
	return scanAttachment(row)
}

func (r *TicketAttachmentRepo) GetByID(ctx context.Context, id, ticketID uuid.UUID) (*domain.TicketAttachment, error) {
	args := []any{id, ticketID}
	q := `SELECT ` + attachmentCols + ` FROM ticket_attachments WHERE id = $1 AND ticket_id = $2`

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $3`
		args = append(args, orgID)
	}
	return scanAttachment(r.db.QueryRow(ctx, q, args...))
}

func (r *TicketAttachmentRepo) List(ctx context.Context, ticketID uuid.UUID) ([]*domain.TicketAttachment, error) {
	args := []any{ticketID}
	q := `SELECT ` + attachmentCols + ` FROM ticket_attachments WHERE ticket_id = $1`

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $2`
		args = append(args, orgID)
	}
	q += ` ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*domain.TicketAttachment
	for rows.Next() {
		var a domain.TicketAttachment
		var storageKey, storageBackend *string
		if err := rows.Scan(
			&a.ID, &a.TicketID, &a.OrgID, &a.UploadedBy, &a.Filename,
			&a.ContentType, &a.SizeBytes, &a.StorageURL, &storageKey, &storageBackend, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		if storageKey != nil {
			a.StorageKey = *storageKey
		}
		if storageBackend != nil {
			a.StorageBackend = *storageBackend
		}
		attachments = append(attachments, &a)
	}
	return attachments, rows.Err()
}

func (r *TicketAttachmentRepo) Delete(ctx context.Context, id, ticketID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM ticket_attachments WHERE id=$1 AND ticket_id=$2`,
		id, ticketID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
