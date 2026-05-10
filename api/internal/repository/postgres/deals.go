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

type DealRepo struct {
	db *pgxpool.Pool
}

func NewDealRepo(db *pgxpool.Pool) *DealRepo {
	return &DealRepo{db: db}
}

const dealCols = `
	deals.id, deals.org_id, deals.title, deals.value_cents, deals.currency, deals.stage, deals.probability,
	deals.expected_close_date, deals.contact_id, deals.account_id,
	deals.owner_id, deals.pipeline_id, deals.custom_fields,
	deals.created_at, deals.updated_at, deals.deleted_at
`

func scanDeal(row pgx.Row) (*domain.Deal, error) {
	var d domain.Deal
	err := row.Scan(
		&d.ID, &d.OrgID, &d.Title, &d.ValueCents, &d.Currency, &d.Stage, &d.Probability,
		&d.ExpectedCloseDate, &d.ContactID, &d.AccountID,
		&d.OwnerID, &d.PipelineID, &d.CustomFields,
		&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *DealRepo) Create(ctx context.Context, d *domain.Deal) (*domain.Deal, error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		d.OrgID = orgID
	}
	now := time.Now().UTC()
	d.CreatedAt = now
	d.UpdatedAt = now

	if d.Currency == "" {
		cur, err := getOrgDefaultCurrency(ctx, r.db, d.OrgID)
		if err != nil {
			return nil, err
		}
		d.Currency = cur
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO deals
			(id, org_id, title, value_cents, currency, stage, probability,
			 expected_close_date, contact_id, account_id,
			 owner_id, pipeline_id, custom_fields, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, org_id, title, value_cents, currency, stage, probability,
	expected_close_date, contact_id, account_id,
	owner_id, pipeline_id, custom_fields,
	created_at, updated_at, deleted_at`,
		d.ID, d.OrgID, d.Title, d.ValueCents, d.Currency, d.Stage, d.Probability,
		d.ExpectedCloseDate, d.ContactID, d.AccountID,
		d.OwnerID, d.PipelineID, d.CustomFields, d.CreatedAt, d.UpdatedAt,
	)
	return scanDeal(row)
}

func (r *DealRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deal, error) {
	q := `SELECT id, org_id, title, value_cents, currency, stage, probability,
	expected_close_date, contact_id, account_id,
	owner_id, pipeline_id, custom_fields,
	created_at, updated_at, deleted_at FROM deals WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	deal, err := scanDeal(r.db.QueryRow(ctx, q, args...))
	if err != nil {
		return nil, err
	}
	contacts, err := r.ListContacts(ctx, id)
	if err != nil {
		return nil, err
	}
	deal.Contacts = contacts
	return deal, nil
}

// AddContact inserts a row into deal_contacts (upsert on conflict to allow role updates).
func (r *DealRepo) AddContact(ctx context.Context, dealID, contactID uuid.UUID, role string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO deal_contacts (deal_id, contact_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (deal_id, contact_id) DO UPDATE SET role = EXCLUDED.role
	`, dealID, contactID, role)
	return err
}

// ListContacts returns all contacts linked to a deal via deal_contacts.
func (r *DealRepo) ListContacts(ctx context.Context, dealID uuid.UUID) ([]domain.Contact, error) {
	const cols = `
		c.id, c.org_id, c.first_name, c.last_name, c.email, c.phone,
		c.account_id, c.owner_id, c.lead_source, c.stage, c.tags,
		c.custom_fields, c.created_at, c.updated_at, c.deleted_at
	`
	rows, err := r.db.Query(ctx, `
		SELECT `+cols+`
		FROM contacts c
		JOIN deal_contacts dc ON dc.contact_id = c.id
		WHERE dc.deal_id = $1 AND c.deleted_at IS NULL
		ORDER BY dc.created_at ASC
	`, dealID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []domain.Contact
	for rows.Next() {
		var c domain.Contact
		if err := rows.Scan(
			&c.ID, &c.OrgID, &c.FirstName, &c.LastName, &c.Email, &c.Phone,
			&c.AccountID, &c.OwnerID, &c.LeadSource, &c.Stage, &c.Tags,
			&c.CustomFields, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, rows.Err()
}

func (r *DealRepo) Update(ctx context.Context, id uuid.UUID, patch domain.DealPatch) (*domain.Deal, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.Title != nil {
		addArg("title", *patch.Title)
	}
	if patch.ValueCents != nil {
		addArg("value_cents", *patch.ValueCents)
	}
	if patch.Currency != nil {
		addArg("currency", *patch.Currency)
	}
	if patch.Stage != nil {
		addArg("stage", *patch.Stage)
	}
	if patch.Probability != nil {
		addArg("probability", *patch.Probability)
	}
	if patch.ExpectedCloseDate != nil {
		addArg("expected_close_date", *patch.ExpectedCloseDate)
	}
	if patch.ContactID != nil {
		addArg("contact_id", *patch.ContactID)
	} else if patch.ClearContactID {
		addArg("contact_id", nil)
	}
	if patch.AccountID != nil {
		addArg("account_id", *patch.AccountID)
	} else if patch.ClearAccountID {
		addArg("account_id", nil)
	}
	if patch.OwnerID != nil {
		addArg("owner_id", *patch.OwnerID)
	}
	if patch.PipelineID != nil {
		addArg("pipeline_id", *patch.PipelineID)
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
	}

	query := fmt.Sprintf(
		`UPDATE deals SET %s WHERE %s RETURNING id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, custom_fields, created_at, updated_at, deleted_at`,
		strings.Join(sets, ", "), whereClause,
	)
	row := r.db.QueryRow(ctx, query, args...)
	return scanDeal(row)
}

func (r *DealRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE deals SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
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

func (r *DealRepo) List(ctx context.Context, f domain.DealFilter) ([]*domain.Deal, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"deals.deleted_at IS NULL"}
	args := []any{}
	i := 1

	addWhere := func(expr string, val any) {
		where = append(where, fmt.Sprintf("deals.%s = $%d", expr, i))
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

	if f.OwnerID != nil {
		addWhere("owner_id", *f.OwnerID)
	}
	if f.Stage != nil {
		addWhere("stage", *f.Stage)
	}
	if f.AccountID != nil {
		addWhere("account_id", *f.AccountID)
	}
	if f.ContactID != nil {
		addWhere("contact_id", *f.ContactID)
	}
	if f.PipelineID != nil {
		addWhere("pipeline_id", *f.PipelineID)
	}
	if f.Q != "" {
		where = append(where, fmt.Sprintf(
			`to_tsvector('english', deals.title) @@ plainto_tsquery('english', $%d)`, i,
		))
		args = append(args, f.Q)
		i++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM deals WHERE `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sortCol := "created_at"
	allowedSorts := map[string]bool{
		"created_at": true, "updated_at": true,
		"title": true, "value_cents": true, "expected_close_date": true,
	}
	if allowedSorts[f.Sort] {
		sortCol = f.Sort
	}
	order := "DESC"
	if strings.ToUpper(f.Order) == "ASC" {
		order = "ASC"
	}

	sortCol = "deals." + sortCol

	rows, err := r.db.Query(ctx,
		fmt.Sprintf(
			`SELECT %s, a.id, a.name, c.id, c.first_name, c.last_name
			 FROM deals 
			 LEFT JOIN accounts a ON deals.account_id = a.id
			 LEFT JOIN contacts c ON deals.contact_id = c.id
			 WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
			dealCols, whereClause, sortCol, order, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deals []*domain.Deal
	for rows.Next() {
		var d domain.Deal
		var aID *uuid.UUID
		var aName *string
		var cID *uuid.UUID
		var cFirst, cLast *string

		if err := rows.Scan(
			&d.ID, &d.OrgID, &d.Title, &d.ValueCents, &d.Currency, &d.Stage, &d.Probability,
			&d.ExpectedCloseDate, &d.ContactID, &d.AccountID,
			&d.OwnerID, &d.PipelineID, &d.CustomFields,
			&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
			&aID, &aName, &cID, &cFirst, &cLast,
		); err != nil {
			return nil, 0, err
		}

		if aID != nil && aName != nil {
			d.Account = &domain.Account{ID: *aID, Name: *aName}
		}
		if cID != nil && cFirst != nil && cLast != nil {
			d.Contact = &domain.Contact{ID: *cID, FirstName: *cFirst, LastName: *cLast}
		}
		deals = append(deals, &d)
	}
	return deals, total, rows.Err()
}
