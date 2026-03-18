package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// ProductRepo implements repository.ProductRepository.
type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

const productCols = `id, org_id, name, description, sku, price, currency, is_active, created_at, updated_at`

func scanProduct(row pgx.Row) (*domain.Product, error) {
	var p domain.Product
	err := row.Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Description, &p.SKU,
		&p.Price, &p.Currency, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) CreateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		p.OrgID = orgID
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Currency == "" {
		p.Currency = "USD"
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO products (id, org_id, name, description, sku, price, currency, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+productCols,
		p.ID, p.OrgID, p.Name, p.Description, p.SKU,
		p.Price, p.Currency, p.IsActive, p.CreatedAt, p.UpdatedAt,
	)
	return scanProduct(row)
}

func (r *ProductRepo) GetProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	orgID, _ := domain.OrgIDFromContext(ctx)
	row := r.db.QueryRow(ctx,
		`SELECT `+productCols+` FROM products WHERE id=$1 AND org_id=$2`,
		id, orgID,
	)
	return scanProduct(row)
}

func (r *ProductRepo) ListProducts(ctx context.Context, activeOnly bool) ([]*domain.Product, error) {
	orgID, _ := domain.OrgIDFromContext(ctx)
	q := `SELECT ` + productCols + ` FROM products WHERE org_id=$1`
	args := []interface{}{orgID}
	if activeOnly {
		q += ` AND is_active=TRUE`
	}
	q += ` ORDER BY name`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProductRepo) UpdateProduct(ctx context.Context, id uuid.UUID, patch domain.ProductPatch) (*domain.Product, error) {
	orgID, _ := domain.OrgIDFromContext(ctx)
	row := r.db.QueryRow(ctx, `
		UPDATE products SET
			name        = COALESCE($3, name),
			description = COALESCE($4, description),
			sku         = COALESCE($5, sku),
			price       = COALESCE($6, price),
			currency    = COALESCE($7, currency),
			is_active   = COALESCE($8, is_active),
			updated_at  = NOW()
		WHERE id=$1 AND org_id=$2
		RETURNING `+productCols,
		id, orgID,
		patch.Name, patch.Description, patch.SKU, patch.Price, patch.Currency, patch.IsActive,
	)
	return scanProduct(row)
}

func (r *ProductRepo) DeactivateProduct(ctx context.Context, id uuid.UUID) error {
	orgID, _ := domain.OrgIDFromContext(ctx)
	_, err := r.db.Exec(ctx,
		`UPDATE products SET is_active=FALSE, updated_at=NOW() WHERE id=$1 AND org_id=$2`,
		id, orgID,
	)
	return err
}

const priceBookCols = `id, org_id, name, is_default, created_at, updated_at`

func scanPriceBook(row pgx.Row) (*domain.PriceBook, error) {
	var pb domain.PriceBook
	err := row.Scan(&pb.ID, &pb.OrgID, &pb.Name, &pb.IsDefault, &pb.CreatedAt, &pb.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &pb, nil
}

func (r *ProductRepo) CreatePriceBook(ctx context.Context, pb *domain.PriceBook) (*domain.PriceBook, error) {
	if pb.ID == uuid.Nil {
		pb.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		pb.OrgID = orgID
	}
	now := time.Now().UTC()
	pb.CreatedAt = now
	pb.UpdatedAt = now
	row := r.db.QueryRow(ctx, `
		INSERT INTO price_books (id, org_id, name, is_default, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING `+priceBookCols,
		pb.ID, pb.OrgID, pb.Name, pb.IsDefault, pb.CreatedAt, pb.UpdatedAt,
	)
	return scanPriceBook(row)
}

func (r *ProductRepo) GetPriceBook(ctx context.Context, id uuid.UUID) (*domain.PriceBook, error) {
	orgID, _ := domain.OrgIDFromContext(ctx)
	row := r.db.QueryRow(ctx,
		`SELECT `+priceBookCols+` FROM price_books WHERE id=$1 AND org_id=$2`,
		id, orgID,
	)
	pb, err := scanPriceBook(row)
	if err != nil {
		return nil, err
	}
	entries, err := r.listEntries(ctx, pb.ID)
	if err != nil {
		return nil, err
	}
	pb.Entries = entries
	return pb, nil
}

func (r *ProductRepo) ListPriceBooks(ctx context.Context) ([]*domain.PriceBook, error) {
	orgID, _ := domain.OrgIDFromContext(ctx)
	rows, err := r.db.Query(ctx,
		`SELECT `+priceBookCols+` FROM price_books WHERE org_id=$1 ORDER BY is_default DESC, name`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.PriceBook
	for rows.Next() {
		pb, err := scanPriceBook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, pb)
	}
	return out, rows.Err()
}

func (r *ProductRepo) listEntries(ctx context.Context, priceBookID uuid.UUID) ([]*domain.PriceBookEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			e.id, e.price_book_id, e.product_id, e.price_override, e.created_at,
			`+productCols+`
		FROM price_book_entries e
		JOIN products p ON p.id = e.product_id
		WHERE e.price_book_id = $1
		ORDER BY p.name
	`, priceBookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.PriceBookEntry
	for rows.Next() {
		var e domain.PriceBookEntry
		var prod domain.Product
		if err := rows.Scan(
			&e.ID, &e.PriceBookID, &e.ProductID, &e.PriceOverride, &e.CreatedAt,
			&prod.ID, &prod.OrgID, &prod.Name, &prod.Description, &prod.SKU,
			&prod.Price, &prod.Currency, &prod.IsActive, &prod.CreatedAt, &prod.UpdatedAt,
		); err != nil {
			return nil, err
		}
		e.Product = &prod
		out = append(out, &e)
	}
	return out, rows.Err()
}

func (r *ProductRepo) UpsertPriceBookEntry(ctx context.Context, e *domain.PriceBookEntry) (*domain.PriceBookEntry, error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	e.CreatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx, `
		INSERT INTO price_book_entries (id, price_book_id, product_id, price_override, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (price_book_id, product_id) DO UPDATE
			SET price_override = EXCLUDED.price_override
	`, e.ID, e.PriceBookID, e.ProductID, e.PriceOverride, e.CreatedAt)
	return e, err
}

func (r *ProductRepo) DeletePriceBookEntry(ctx context.Context, priceBookID, productID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM price_book_entries WHERE price_book_id=$1 AND product_id=$2`,
		priceBookID, productID,
	)
	return err
}

// DealLineItemRepo implements repository.DealLineItemRepository.
type DealLineItemRepo struct {
	db *pgxpool.Pool
}

func NewDealLineItemRepo(db *pgxpool.Pool) *DealLineItemRepo {
	return &DealLineItemRepo{db: db}
}

const lineItemCols = `id, deal_id, product_id, name, quantity, unit_price, discount_pct, subtotal, position, created_at, updated_at`

func scanLineItem(row pgx.Row) (*domain.DealLineItem, error) {
	var li domain.DealLineItem
	err := row.Scan(
		&li.ID, &li.DealID, &li.ProductID, &li.Name,
		&li.Quantity, &li.UnitPrice, &li.DiscountPct, &li.Subtotal,
		&li.Position, &li.CreatedAt, &li.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &li, nil
}

func computeSubtotal(li *domain.DealLineItem) float64 {
	return li.Quantity * li.UnitPrice * (1 - li.DiscountPct/100)
}

func (r *DealLineItemRepo) List(ctx context.Context, dealID uuid.UUID) ([]*domain.DealLineItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+lineItemCols+` FROM deal_line_items WHERE deal_id=$1 ORDER BY position, created_at`,
		dealID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.DealLineItem
	for rows.Next() {
		li, err := scanLineItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, li)
	}
	return out, rows.Err()
}

func (r *DealLineItemRepo) Upsert(ctx context.Context, item *domain.DealLineItem) (*domain.DealLineItem, error) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	item.Subtotal = computeSubtotal(item)
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	row := r.db.QueryRow(ctx, `
		INSERT INTO deal_line_items
			(id, deal_id, product_id, name, quantity, unit_price, discount_pct, subtotal, position, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			name         = EXCLUDED.name,
			product_id   = EXCLUDED.product_id,
			quantity     = EXCLUDED.quantity,
			unit_price   = EXCLUDED.unit_price,
			discount_pct = EXCLUDED.discount_pct,
			subtotal     = EXCLUDED.subtotal,
			position     = EXCLUDED.position,
			updated_at   = NOW()
		RETURNING `+lineItemCols,
		item.ID, item.DealID, item.ProductID, item.Name,
		item.Quantity, item.UnitPrice, item.DiscountPct, item.Subtotal,
		item.Position, item.CreatedAt, item.UpdatedAt,
	)
	return scanLineItem(row)
}

func (r *DealLineItemRepo) Delete(ctx context.Context, id, dealID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM deal_line_items WHERE id=$1 AND deal_id=$2`,
		id, dealID,
	)
	return err
}

func (r *DealLineItemRepo) Reorder(ctx context.Context, dealID uuid.UUID, ids []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	for i, id := range ids {
		if _, err := tx.Exec(ctx,
			`UPDATE deal_line_items SET position=$1, updated_at=NOW() WHERE id=$2 AND deal_id=$3`,
			i, id, dealID,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
