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

type ContactRepo struct {
	db *pgxpool.Pool
}

func NewContactRepo(db *pgxpool.Pool) *ContactRepo {
	return &ContactRepo{db: db}
}

const contactCols = `
	c.id, c.org_id, c.first_name, c.last_name, c.email, c.phone,
	COALESCE(ac.account_id, c.account_id) AS account_id,
	ac.relationship_type, COALESCE(ac.is_primary, false) AS is_primary,
	ac.title_at_account,
	CASE WHEN ac.start_date IS NOT NULL THEN ac.start_date::timestamptz END AS start_date,
	CASE WHEN ac.end_date IS NOT NULL THEN ac.end_date::timestamptz END AS end_date,
	c.owner_id, c.lead_source, c.lead_score, c.stage, c.tags,
	c.custom_fields, c.email_opt_out, c.bounce_count,
	c.converted_at, c.converted_by, c.converted_deal_id, c.converted_from_lead_id,
	c.created_at, c.updated_at, c.deleted_at
`

const contactPrimaryJoin = `
	LEFT JOIN LATERAL (
		SELECT account_id, relationship_type, is_primary, title_at_account, start_date, end_date
		FROM account_contacts ac
		WHERE ac.contact_id = c.id
		  AND ac.org_id = c.org_id
		  AND ac.deleted_at IS NULL
		  AND ac.end_date IS NULL
		ORDER BY ac.is_primary DESC, ac.created_at DESC
		LIMIT 1
	) ac ON TRUE
`

func scanContact(row pgx.Row) (*domain.Contact, error) {
	var c domain.Contact
	err := row.Scan(
		&c.ID, &c.OrgID, &c.FirstName, &c.LastName, &c.Email, &c.Phone,
		&c.AccountID, &c.RelationshipType, &c.IsPrimary, &c.TitleAtAccount, &c.StartDate, &c.EndDate,
		&c.OwnerID, &c.LeadSource, &c.LeadScore, &c.Stage, &c.Tags,
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

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO contacts
			(id, org_id, first_name, last_name, email, phone,
			 account_id, owner_id, lead_source, stage, tags,
			 custom_fields, converted_from_lead_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		c.ID, c.OrgID, c.FirstName, c.LastName, c.Email, c.Phone,
		c.AccountID, c.OwnerID, c.LeadSource, c.Stage, c.Tags,
		c.CustomFields, c.ConvertedFromLeadID, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := r.upsertPrimaryAccountContact(ctx, tx, c.OrgID, c.ID, c); err != nil {
		return nil, err
	}

	created, err := scanContact(tx.QueryRow(ctx,
		`SELECT `+contactCols+` FROM contacts c `+contactPrimaryJoin+` WHERE c.id=$1 AND c.deleted_at IS NULL`, c.ID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

func (r *ContactRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	q := `SELECT ` + contactCols + ` FROM contacts c ` + contactPrimaryJoin + ` WHERE c.id=$1 AND c.deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND c.org_id=$2`
		args = append(args, orgID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	return scanContact(row)
}

func (r *ContactRepo) GetByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	q := `SELECT ` + contactCols + ` FROM contacts c ` + contactPrimaryJoin + ` WHERE c.email=$1 AND c.deleted_at IS NULL`
	args := []any{email}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND c.org_id=$2`
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

	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if hasCtxOrg {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := fmt.Sprintf(`UPDATE contacts SET %s WHERE %s`, strings.Join(sets, ", "), whereClause)
	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if result.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	if !hasCtxOrg {
		if err := tx.QueryRow(ctx, `SELECT org_id FROM contacts WHERE id = $1`, id).Scan(&orgID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrNotFound
			}
			return nil, err
		}
	}

	if patch.AccountID != nil || patch.RelationshipType != nil || patch.IsPrimary != nil || patch.TitleAtAccount != nil || patch.StartDate != nil || patch.EndDate != nil {
		if err := r.upsertPrimaryAccountContactPatch(ctx, tx, orgID, id, patch); err != nil {
			return nil, err
		}
	}

	updated, err := scanContact(tx.QueryRow(ctx,
		`SELECT `+contactCols+` FROM contacts c `+contactPrimaryJoin+` WHERE c.id=$1 AND c.deleted_at IS NULL`, id))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *ContactRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE contacts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
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

func (r *ContactRepo) List(ctx context.Context, f domain.ContactFilter) ([]*domain.Contact, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"c.deleted_at IS NULL"}
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
		addWhere("c.org_id", orgID)
	}

	if f.OwnerID != nil {
		addWhere("c.owner_id", *f.OwnerID)
	}
	if f.Stage != nil {
		addWhere("c.stage", *f.Stage)
	}
	if f.Source != nil {
		addWhere("c.lead_source", *f.Source)
	}
	if f.ScoreMin != nil {
		where = append(where, fmt.Sprintf("c.lead_score >= $%d", i))
		args = append(args, *f.ScoreMin)
		i++
	}
	if f.ScoreMax != nil {
		where = append(where, fmt.Sprintf("c.lead_score <= $%d", i))
		args = append(args, *f.ScoreMax)
		i++
	}
	if f.Q != "" {
		where = append(where, fmt.Sprintf(
			`to_tsvector('english', c.first_name || ' ' || c.last_name || ' ' || coalesce(c.email, '') || ' ' || coalesce(c.phone, '')) @@ plainto_tsquery('english', $%d)`, i,
		))
		args = append(args, f.Q)
		i++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM contacts c WHERE `+whereClause, args...).Scan(&total)
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
			`SELECT %s FROM contacts c %s WHERE %s ORDER BY c.%s %s LIMIT $%d OFFSET $%d`,
			contactCols, contactPrimaryJoin, whereClause, sortCol, order, i, i+1,
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

	q := fmt.Sprintf(`UPDATE contacts SET %s, updated_at = NOW() WHERE %s`, scoreExpr, whereClause)
	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	if result.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.GetByID(ctx, id)
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
		WHERE %s`, whereClause)
	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	if result.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.GetByID(ctx, id)
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

func (r *ContactRepo) upsertPrimaryAccountContact(ctx context.Context, tx pgx.Tx, orgID, contactID uuid.UUID, c *domain.Contact) error {
	if c.AccountID == nil {
		return nil
	}

	isPrimary := true
	relationshipType := "champion"
	if c.RelationshipType != nil && *c.RelationshipType != "" {
		relationshipType = *c.RelationshipType
	}

	_, err := tx.Exec(ctx, `
		UPDATE account_contacts
		SET is_primary = false,
		    end_date = COALESCE(end_date, CURRENT_DATE),
		    updated_at = NOW()
		WHERE org_id = $1
		  AND contact_id = $2
		  AND is_primary = true
		  AND deleted_at IS NULL
		  AND end_date IS NULL`, orgID, contactID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO account_contacts
			(id, org_id, account_id, contact_id, relationship_type, is_primary, title_at_account, start_date, end_date, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())`,
		uuid.New(), orgID, *c.AccountID, contactID, relationshipType, isPrimary, c.TitleAtAccount, c.StartDate, c.EndDate)
	return err
}

func (r *ContactRepo) upsertPrimaryAccountContactPatch(ctx context.Context, tx pgx.Tx, orgID, contactID uuid.UUID, patch domain.ContactPatch) error {
	var accountID *uuid.UUID
	if patch.AccountID != nil {
		accountID = patch.AccountID
	} else {
		err := tx.QueryRow(ctx, `
			SELECT COALESCE(ac.account_id, c.account_id)
			FROM contacts c
			LEFT JOIN LATERAL (
				SELECT account_id
				FROM account_contacts ac
				WHERE ac.org_id = c.org_id
				  AND ac.contact_id = c.id
				  AND ac.deleted_at IS NULL
				  AND ac.end_date IS NULL
				ORDER BY ac.is_primary DESC, ac.created_at DESC
				LIMIT 1
			) ac ON TRUE
			WHERE c.id = $1`, contactID).Scan(&accountID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		}
	}
	if accountID == nil {
		return nil
	}

	isPrimary := true
	if patch.IsPrimary != nil {
		isPrimary = *patch.IsPrimary
	}
	relationshipType := "champion"
	if patch.RelationshipType != nil && *patch.RelationshipType != "" {
		relationshipType = *patch.RelationshipType
	}

	_, err := tx.Exec(ctx, `
		UPDATE account_contacts
		SET is_primary = false,
		    end_date = COALESCE(end_date, CURRENT_DATE),
		    updated_at = NOW()
		WHERE org_id = $1
		  AND contact_id = $2
		  AND is_primary = true
		  AND deleted_at IS NULL
		  AND end_date IS NULL`, orgID, contactID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO account_contacts
			(id, org_id, account_id, contact_id, relationship_type, is_primary, title_at_account, start_date, end_date, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())`,
		uuid.New(), orgID, *accountID, contactID, relationshipType, isPrimary, patch.TitleAtAccount, patch.StartDate, patch.EndDate)
	return err
}
