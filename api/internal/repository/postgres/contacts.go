package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type ContactRepo struct {
	db *pgxpool.Pool
}

func NewContactRepo(db *pgxpool.Pool) *ContactRepo {
	return &ContactRepo{db: db}
}

const contactCols = `
	id, org_id, first_name, last_name, email, phone,
	account_id, owner_id, lead_source, lead_score, stage, tags,
	custom_fields, email_opt_out, bounce_count,
	converted_at, converted_by, converted_deal_id, converted_from_lead_id,
	created_at, updated_at, deleted_at
`

func scanContact(row pgx.Row) (*domain.Contact, error) {
	var c domain.Contact
	err := row.Scan(
		&c.ID, &c.OrgID, &c.FirstName, &c.LastName, &c.Email, &c.Phone,
		&c.AccountID, &c.OwnerID, &c.LeadSource, &c.LeadScore, &c.Stage, &c.Tags,
		&c.CustomFields, &c.EmailOptOut, &c.BounceCount,
		&c.ConvertedAt, &c.ConvertedBy, &c.ConvertedDealID, &c.ConvertedFromLeadID,
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

func (r *ContactRepo) Create(ctx context.Context, c *domain.Contact) (*domain.Contact, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		c.OrgID = orgID
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.Tags == nil {
		c.Tags = []string{}
	}
	if c.Stage == "" {
		c.Stage = domain.ContactStageLead
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO contacts
			(id, org_id, first_name, last_name, email, phone,
			 account_id, owner_id, lead_source, stage, tags,
			 custom_fields, converted_from_lead_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING `+contactCols,
		c.ID, c.OrgID, c.FirstName, c.LastName, c.Email, c.Phone,
		c.AccountID, c.OwnerID, c.LeadSource, c.Stage, c.Tags,
		c.CustomFields, c.ConvertedFromLeadID, c.CreatedAt, c.UpdatedAt,
	)
	return scanContact(row)
}

func (r *ContactRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	q := `SELECT ` + contactCols + ` FROM contacts WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}
	i := 2

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
		i++
	}
	appendAccessVisibilitySQL(ctx, &q, &args, domain.ACLModuleContacts, domain.SharingAccessRead, "owner_id")

	row := r.db.QueryRow(ctx, q, args...)
	return scanContact(row)
}

func (r *ContactRepo) GetByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	q := `SELECT ` + contactCols + ` FROM contacts WHERE email=$1 AND deleted_at IS NULL`
	args := []any{email}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	return scanContact(row)
}

func (r *ContactRepo) Update(ctx context.Context, id uuid.UUID, patch domain.ContactPatch) (*domain.Contact, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.FirstName != nil {
		addArg("first_name", *patch.FirstName)
	}
	if patch.LastName != nil {
		addArg("last_name", *patch.LastName)
	}
	if patch.Email != nil {
		addArg("email", *patch.Email)
	}
	if patch.Phone != nil {
		addArg("phone", *patch.Phone)
	}
	if patch.AccountID != nil {
		addArg("account_id", *patch.AccountID)
	} else if patch.ClearAccountID {
		addArg("account_id", nil)
	}
	if patch.OwnerID != nil {
		addArg("owner_id", *patch.OwnerID)
	}
	if patch.LeadSource != nil {
		addArg("lead_source", *patch.LeadSource)
	}
	if patch.Stage != nil {
		addArg("stage", *patch.Stage)
	}
	if patch.Tags != nil {
		addArg("tags", patch.Tags)
	}
	if patch.CustomFields != nil {
		addArg("custom_fields", patch.CustomFields)
	}

	whereClause := fmt.Sprintf(`id=$%d AND deleted_at IS NULL`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
		i++
	}
	appendAccessVisibilitySQL(ctx, &whereClause, &args, domain.ACLModuleContacts, domain.SharingAccessWrite, "owner_id")

	query := fmt.Sprintf(
		`UPDATE contacts SET %s WHERE %s RETURNING %s`,
		strings.Join(sets, ", "), whereClause, contactCols,
	)
	row := r.db.QueryRow(ctx, query, args...)
	return scanContact(row)
}

func (r *ContactRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE contacts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}
	i := 2

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
		i++
	}
	appendAccessVisibilitySQL(ctx, &q, &args, domain.ACLModuleContacts, domain.SharingAccessWrite, "owner_id")

	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ContactRepo) List(ctx context.Context, f domain.ContactFilter) ([]*domain.Contact, int, error) {
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

	// Always scope by org_id: prefer context, fall back to filter field.
	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if !hasCtxOrg {
		orgID = f.OrgID
	}
	if orgID != uuid.Nil {
		addWhere("org_id", orgID)
	}
	addAccessVisibilityWhere(ctx, &where, &args, &i, domain.ACLModuleContacts, domain.SharingAccessRead, "owner_id")

	if f.OwnerID != nil {
		addWhere("owner_id", *f.OwnerID)
	}
	if f.AccountID != nil {
		where = append(where, fmt.Sprintf(
			`(account_id = $%d OR EXISTS (
				SELECT 1
				FROM account_contacts ac
				WHERE ac.contact_id = contacts.id
				  AND ac.account_id = $%d
			))`, i, i,
		))
		args = append(args, *f.AccountID)
		i++
	}
	if f.Stage != nil {
		addWhere("stage", *f.Stage)
	}
	if f.Source != nil {
		addWhere("lead_source", *f.Source)
	}
	if f.ScoreMin != nil {
		where = append(where, fmt.Sprintf("lead_score >= $%d", i))
		args = append(args, *f.ScoreMin)
		i++
	}
	if f.ScoreMax != nil {
		where = append(where, fmt.Sprintf("lead_score <= $%d", i))
		args = append(args, *f.ScoreMax)
		i++
	}
	if f.Q != "" {
		where = append(where, fmt.Sprintf(
			`to_tsvector('english', first_name || ' ' || last_name || ' ' || coalesce(email, '') || ' ' || coalesce(phone, '')) @@ plainto_tsquery('english', $%d)`, i,
		))
		args = append(args, f.Q)
		i++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM contacts WHERE `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sortCol := "created_at"
	allowedSorts := map[string]bool{
		"created_at": true, "updated_at": true,
		"first_name": true, "last_name": true, "email": true,
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
			`SELECT %s FROM contacts WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
			contactCols, whereClause, sortCol, order, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, 0, err
		}
		contacts = append(contacts, c)
	}
	return contacts, total, rows.Err()
}

func (r *ContactRepo) ListLinkedEntities(ctx context.Context, id uuid.UUID, f domain.LinkedEntityFilter) ([]domain.LinkedEntity, int, error) {
	if f.Limit <= 0 {
		f.Limit = 25
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	args := []any{id}
	i := 2
	where := []string{}

	if f.Type != nil && *f.Type != "" {
		where = append(where, fmt.Sprintf("entity_type = $%d", i))
		args = append(args, *f.Type)
		i++
	}
	if f.Role != nil && *f.Role != "" {
		where = append(where, fmt.Sprintf("entity_role = $%d", i))
		args = append(args, *f.Role)
		i++
	}
	if f.Since != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", i))
		args = append(args, *f.Since)
		i++
	}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		where = append(where, fmt.Sprintf("org_id = $%d", i))
		args = append(args, orgID)
		i++
	}

	filterClause := ""
	if len(where) > 0 {
		filterClause = " WHERE " + strings.Join(where, " AND ")
	}

	query := `
		WITH linked AS (
			SELECT c.org_id AS org_id, 'account'::text AS entity_type, c.account_id::text AS entity_id,
				'primary'::text AS entity_role, 'contacts.account_id'::text AS source, c.created_at
			FROM contacts c
			WHERE c.id = $1 AND c.account_id IS NOT NULL AND c.deleted_at IS NULL
			UNION ALL
			SELECT d.org_id AS org_id, 'deal'::text AS entity_type, d.id::text AS entity_id,
				'primary'::text AS entity_role, 'deals.contact_id'::text AS source, d.created_at
			FROM deals d
			WHERE d.contact_id = $1 AND d.deleted_at IS NULL
			UNION ALL
			SELECT d.org_id AS org_id, 'deal'::text AS entity_type, dc.deal_id::text AS entity_id,
				dc.role AS entity_role, 'deal_contacts'::text AS source, dc.created_at
			FROM deal_contacts dc
			JOIN deals d ON d.id = dc.deal_id
			WHERE dc.contact_id = $1 AND d.deleted_at IS NULL
		), filtered AS (
			SELECT * FROM linked` + filterClause + `
		)
		SELECT entity_type, entity_id, entity_role, source, created_at,
			COUNT(*) OVER() AS total
		FROM filtered
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(i) + ` OFFSET $` + strconv.Itoa(i+1)

	args = append(args, f.Limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.LinkedEntity, 0)
	total := 0
	for rows.Next() {
		var e domain.LinkedEntity
		if err := rows.Scan(&e.Type, &e.ID, &e.Role, &e.Source, &e.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *ContactRepo) IsRelatedToAccount(ctx context.Context, contactID, accountID uuid.UUID) (bool, error) {
	var related bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM contacts c
			WHERE c.id = $1
			  AND c.deleted_at IS NULL
			  AND (
				c.account_id = $2
				OR EXISTS (
					SELECT 1
					FROM account_contacts ac
					WHERE ac.contact_id = c.id
					  AND ac.account_id = $2
				)
			  )
		)
	`, contactID, accountID).Scan(&related)
	if err != nil {
		return false, err
	}
	return related, nil
}

// UpdateLeadScore adjusts or sets the lead_score on a contact.
// If patch.Score is set, it is applied as an absolute value.
// If patch.Delta is set (and Score is nil), it is added to the current score (clamped to >= 0).
func (r *ContactRepo) UpdateLeadScore(ctx context.Context, id uuid.UUID, patch domain.LeadScorePatch) (*domain.Contact, error) {
	var scoreExpr string
	var args []any
	i := 1

	if patch.Score != nil {
		scoreExpr = fmt.Sprintf("lead_score = $%d", i)
		args = append(args, *patch.Score)
		i++
	} else if patch.Delta != nil {
		scoreExpr = fmt.Sprintf("lead_score = GREATEST(0, lead_score + $%d)", i)
		args = append(args, *patch.Delta)
		i++
	} else {
		return nil, fmt.Errorf("lead score patch must provide score or delta")
	}

	whereClause := fmt.Sprintf(`id = $%d AND deleted_at IS NULL`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id = $%d`, i)
		args = append(args, orgID)
	}

	q := fmt.Sprintf(`UPDATE contacts SET %s, updated_at = NOW() WHERE %s RETURNING %s`,
		scoreExpr, whereClause, contactCols)
	return scanContact(r.db.QueryRow(ctx, q, args...))
}

// ConvertLead transitions a contact from stage='lead' to stage='prospect',
// recording who converted it and optionally which deal was created.
// The operation is idempotent: if already converted, it returns the contact unchanged.
func (r *ContactRepo) ConvertLead(ctx context.Context, id, byUserID uuid.UUID, dealID *uuid.UUID) (*domain.Contact, error) {
	args := []any{byUserID, dealID, id}
	i := 4

	whereClause := `id = $3 AND deleted_at IS NULL`
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id = $%d`, i)
		args = append(args, orgID)
	}

	q := fmt.Sprintf(`
		UPDATE contacts
		SET stage = 'prospect',
		    converted_at = COALESCE(converted_at, NOW()),
		    converted_by = COALESCE(converted_by, $1),
		    converted_deal_id = COALESCE(converted_deal_id, $2),
		    updated_at = NOW()
		WHERE %s
		RETURNING %s`, whereClause, contactCols)
	return scanContact(r.db.QueryRow(ctx, q, args...))
}

// SetEmailOptOut sets email_opt_out=true for a contact, scoped to orgID.
func (r *ContactRepo) SetEmailOptOut(ctx context.Context, contactID, orgID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE contacts SET email_opt_out = TRUE, updated_at = NOW() WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL`,
		contactID, orgID)
	return err
}

// IncrementBounceCount atomically increments bounce_count for a contact, scoped to orgID.
func (r *ContactRepo) IncrementBounceCount(ctx context.Context, contactID, orgID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE contacts SET bounce_count = bounce_count + 1, updated_at = NOW() WHERE id = $1 AND org_id = $2 AND deleted_at IS NULL`,
		contactID, orgID)
	return err
}

// ListLeadSources returns distinct non-null lead_source values for contacts
// with stage='lead', scoped to the org in context.
func (r *ContactRepo) ListLeadSources(ctx context.Context) ([]string, error) {
	q := `SELECT DISTINCT lead_source FROM contacts WHERE stage = 'lead' AND lead_source IS NOT NULL AND deleted_at IS NULL`
	args := []any{}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $1`
		args = append(args, orgID)
	}
	q += ` ORDER BY lead_source`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	if sources == nil {
		sources = []string{}
	}
	return sources, rows.Err()
}
