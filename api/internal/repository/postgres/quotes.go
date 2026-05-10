package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type QuoteRepo struct {
	db *pgxpool.Pool
}

func NewQuoteRepo(db *pgxpool.Pool) *QuoteRepo {
	return &QuoteRepo{db: db}
}

const quoteCols = `
	id, org_id, deal_id, account_id, contact_id, title, status, currency,
	valid_until, notes, sent_at, approved_at, rejected_at,
	total_cents, created_by, number, number_prefix, created_at, updated_at
`

func scanQuote(row pgx.Row) (*domain.Quote, error) {
	var q domain.Quote
	var notes *string
	err := row.Scan(
		&q.ID, &q.OrgID, &q.DealID, &q.AccountID, &q.ContactID,
		&q.Title, &q.Status, &q.Currency,
		&q.ValidUntil, &notes, &q.SentAt, &q.ApprovedAt, &q.RejectedAt,
		&q.TotalCents, &q.CreatedBy, &q.Number, &q.NumberPrefix, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if notes != nil {
		q.Notes = *notes
	}
	return &q, nil
}

func (r *QuoteRepo) Create(ctx context.Context, q *domain.Quote) (*domain.Quote, error) {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	if q.Status == "" {
		q.Status = domain.QuoteStatusDraft
	}
	if q.Currency == "" {
		cur, err := getOrgDefaultCurrency(ctx, r.db, q.OrgID)
		if err != nil {
			return nil, err
		}
		q.Currency = cur
	}
	if err := r.validateDealAccountContext(ctx, q.DealID, q.AccountID, q.OrgID); err != nil {
		return nil, err
	}
	num, err := getNextDocNumber(ctx, r.db, q.OrgID, domain.DocTypeQuote)
	if err != nil {
		return nil, err
	}
	prefix, err := getDocPrefix(ctx, r.db, q.OrgID, domain.DocTypeQuote)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx,
		`INSERT INTO quotes
		 (id, org_id, deal_id, account_id, contact_id, title, status, currency, valid_until, notes, created_by, number, number_prefix)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		 RETURNING `+quoteCols,
		q.ID, q.OrgID, q.DealID, q.AccountID, q.ContactID,
		q.Title, q.Status, q.Currency,
		q.ValidUntil, nilIfEmpty(q.Notes), q.CreatedBy, num, prefix,
	)
	created, err := scanQuote(row)
	if err != nil {
		return nil, err
	}
	if len(q.LineItems) > 0 {
		inputs := make([]domain.QuoteLineItemInput, len(q.LineItems))
		for i, li := range q.LineItems {
			inputs[i] = domain.QuoteLineItemInput{
				ProductID:      li.ProductID,
				ProductName:    li.ProductName,
				Description:    li.Description,
				Quantity:       li.Quantity,
				UnitPriceCents: li.UnitPriceCents,
				DiscountPct:    li.DiscountPct,
				SortOrder:      li.SortOrder,
			}
		}
		items, err := r.ReplaceLineItems(ctx, created.ID, inputs)
		if err != nil {
			return nil, err
		}
		created.LineItems = items
		created.ComputeTotal()
		if err := r.persistTotal(ctx, created.ID, created.TotalCents); err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (r *QuoteRepo) persistTotal(ctx context.Context, id uuid.UUID, total int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE quotes SET total_cents=$1, updated_at=NOW() WHERE id=$2`,
		total, id,
	)
	return err
}

func (r *QuoteRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Quote, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+quoteCols+` FROM quotes WHERE id=$1`, id,
	)
	q, err := scanQuote(row)
	if err != nil {
		return nil, err
	}
	items, err := r.listLineItems(ctx, id)
	if err != nil {
		return nil, err
	}
	q.LineItems = items
	q.ComputeTotal()
	return q, nil
}

func (r *QuoteRepo) Update(ctx context.Context, id uuid.UUID, patch domain.QuotePatch) (*domain.Quote, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	nextDealID := current.DealID
	if patch.DealID != nil {
		nextDealID = patch.DealID
	}
	nextAccountID := current.AccountID
	if patch.AccountID != nil {
		nextAccountID = patch.AccountID
	}
	if err := r.validateDealAccountContext(ctx, nextDealID, nextAccountID, current.OrgID); err != nil {
		return nil, err
	}

	setClauses := []string{}
	args := []any{}
	i := 1

	if patch.Title != nil {
		setClauses = append(setClauses, "title=$"+itoa(i))
		args = append(args, *patch.Title)
		i++
	}
	if patch.Status != nil {
		setClauses = append(setClauses, "status=$"+itoa(i))
		args = append(args, string(*patch.Status))
		i++
	}
	if patch.Currency != nil {
		setClauses = append(setClauses, "currency=$"+itoa(i))
		args = append(args, *patch.Currency)
		i++
	}
	if patch.ValidUntil != nil {
		setClauses = append(setClauses, "valid_until=$"+itoa(i))
		args = append(args, *patch.ValidUntil)
		i++
	}
	if patch.Notes != nil {
		setClauses = append(setClauses, "notes=$"+itoa(i))
		args = append(args, nilIfEmpty(*patch.Notes))
		i++
	}
	if patch.AccountID != nil {
		setClauses = append(setClauses, "account_id=$"+itoa(i))
		args = append(args, *patch.AccountID)
		i++
	}
	if patch.ContactID != nil {
		setClauses = append(setClauses, "contact_id=$"+itoa(i))
		args = append(args, *patch.ContactID)
		i++
	}
	if patch.DealID != nil {
		setClauses = append(setClauses, "deal_id=$"+itoa(i))
		args = append(args, *patch.DealID)
		i++
	}

	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at=NOW()")
		args = append(args, id)
		q := `UPDATE quotes SET ` + strings.Join(setClauses, ",") +
			` WHERE id=$` + itoa(i)
		if _, err := r.db.Exec(ctx, q, args...); err != nil {
			return nil, err
		}
	}

	if patch.LineItems != nil {
		items, err := r.ReplaceLineItems(ctx, id, patch.LineItems)
		if err != nil {
			return nil, err
		}
		var total int64
		for i := range items {
			items[i].ComputeTotal()
			total += items[i].TotalCents
		}
		if err := r.persistTotal(ctx, id, total); err != nil {
			return nil, err
		}
	}

	return r.GetByID(ctx, id)
}

func (r *QuoteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.Exec(ctx, `DELETE FROM quotes WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *QuoteRepo) List(ctx context.Context, filter domain.QuoteFilter) ([]*domain.Quote, int, error) {
	where := []string{"org_id=$1"}
	args := []any{filter.OrgID}
	i := 2

	if filter.DealID != nil {
		where = append(where, "deal_id=$"+itoa(i))
		args = append(args, *filter.DealID)
		i++
	}
	if filter.ContactID != nil {
		where = append(where, "contact_id=$"+itoa(i))
		args = append(args, *filter.ContactID)
		i++
	}
	if filter.AccountID != nil {
		where = append(where, "account_id=$"+itoa(i))
		args = append(args, *filter.AccountID)
		i++
	}
	if filter.Status != nil {
		where = append(where, "status=$"+itoa(i))
		args = append(args, string(*filter.Status))
		i++
	}
	if filter.Q != "" {
		where = append(where, "title ILIKE $"+itoa(i))
		args = append(args, "%"+filter.Q+"%")
		i++
	}

	clause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM quotes `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	sortCol := "created_at"
	allowedSorts := map[string]bool{
		"created_at":  true,
		"updated_at":  true,
		"title":       true,
		"status":      true,
		"total_cents": true,
		"valid_until": true,
		"account_id":  true,
	}
	if allowedSorts[filter.Sort] {
		sortCol = filter.Sort
	}
	order := "DESC"
	if strings.ToUpper(filter.Order) == "ASC" {
		order = "ASC"
	}

	rows, err := r.db.Query(ctx,
		`SELECT `+quoteCols+` FROM quotes `+clause+
			fmt.Sprintf(` ORDER BY %s %s LIMIT $`, sortCol, order)+itoa(i)+` OFFSET $`+itoa(i+1),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var quotes []*domain.Quote
	for rows.Next() {
		q, err := scanQuote(rows)
		if err != nil {
			return nil, 0, err
		}
		quotes = append(quotes, q)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Load line items for each quote.
	for _, q := range quotes {
		items, err := r.listLineItems(ctx, q.ID)
		if err != nil {
			return nil, 0, err
		}
		q.LineItems = items
		q.ComputeTotal()
	}
	return quotes, total, nil
}

func (r *QuoteRepo) validateDealAccountContext(ctx context.Context, dealID, accountID *uuid.UUID, orgID uuid.UUID) error {
	if dealID == nil || accountID == nil {
		return nil
	}

	var dealAccountID *uuid.UUID
	err := r.db.QueryRow(ctx,
		`SELECT account_id FROM deals WHERE id=$1 AND org_id=$2 AND deleted_at IS NULL`,
		*dealID, orgID,
	).Scan(&dealAccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: deal_id not found in org", domain.ErrValidation)
		}
		return err
	}

	if dealAccountID == nil || *dealAccountID != *accountID {
		return fmt.Errorf("%w: deal_id and account_id must reference the same account", domain.ErrValidation)
	}
	return nil
}

func (r *QuoteRepo) ReplaceLineItems(ctx context.Context, quoteID uuid.UUID, inputs []domain.QuoteLineItemInput) ([]domain.QuoteLineItem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM quote_line_items WHERE quote_id=$1`, quoteID); err != nil {
		return nil, err
	}

	items := make([]domain.QuoteLineItem, 0, len(inputs))
	for idx, inp := range inputs {
		li := domain.QuoteLineItem{
			ID:             uuid.New(),
			QuoteID:        quoteID,
			ProductID:      inp.ProductID,
			ProductName:    inp.ProductName,
			Description:    inp.Description,
			Quantity:       inp.Quantity,
			UnitPriceCents: inp.UnitPriceCents,
			DiscountPct:    inp.DiscountPct,
			SortOrder:      idx,
		}
		li.ComputeTotal()

		row := tx.QueryRow(ctx,
			`INSERT INTO quote_line_items
			 (id, quote_id, product_id, product_name, description, quantity, unit_price_cents, discount_pct, total_cents, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			 RETURNING id, quote_id, product_id, product_name, description, quantity, unit_price_cents, discount_pct, total_cents, sort_order, created_at`,
			li.ID, li.QuoteID, li.ProductID, li.ProductName, nilIfEmpty(li.Description),
			li.Quantity, li.UnitPriceCents, li.DiscountPct, li.TotalCents, li.SortOrder,
		)
		var scanned domain.QuoteLineItem
		var desc *string
		if err := row.Scan(
			&scanned.ID, &scanned.QuoteID, &scanned.ProductID, &scanned.ProductName, &desc,
			&scanned.Quantity, &scanned.UnitPriceCents, &scanned.DiscountPct, &scanned.TotalCents,
			&scanned.SortOrder, &scanned.CreatedAt,
		); err != nil {
			return nil, err
		}
		if desc != nil {
			scanned.Description = *desc
		}
		items = append(items, scanned)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *QuoteRepo) MarkSent(ctx context.Context, id uuid.UUID) (*domain.Quote, error) {
	_, err := r.db.Exec(ctx,
		`UPDATE quotes SET status='sent', sent_at=NOW(), updated_at=NOW() WHERE id=$1`, id,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *QuoteRepo) MarkApproved(ctx context.Context, id uuid.UUID) (*domain.Quote, error) {
	res, err := r.db.Exec(ctx,
		`UPDATE quotes SET status='approved', approved_at=NOW(), updated_at=NOW() WHERE id=$1`, id,
	)
	if err != nil {
		return nil, err
	}
	if res.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *QuoteRepo) MarkRejected(ctx context.Context, id uuid.UUID) (*domain.Quote, error) {
	res, err := r.db.Exec(ctx,
		`UPDATE quotes SET status='rejected', rejected_at=NOW(), updated_at=NOW() WHERE id=$1`, id,
	)
	if err != nil {
		return nil, err
	}
	if res.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *QuoteRepo) listLineItems(ctx context.Context, quoteID uuid.UUID) ([]domain.QuoteLineItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, quote_id, product_id, product_name, description, quantity, unit_price_cents, discount_pct, total_cents, sort_order, created_at
		 FROM quote_line_items WHERE quote_id=$1 ORDER BY sort_order ASC`, quoteID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.QuoteLineItem
	for rows.Next() {
		var li domain.QuoteLineItem
		var desc *string
		if err := rows.Scan(
			&li.ID, &li.QuoteID, &li.ProductID, &li.ProductName, &desc,
			&li.Quantity, &li.UnitPriceCents, &li.DiscountPct, &li.TotalCents,
			&li.SortOrder, &li.CreatedAt,
		); err != nil {
			return nil, err
		}
		if desc != nil {
			li.Description = *desc
		}
		items = append(items, li)
	}
	return items, rows.Err()
}
