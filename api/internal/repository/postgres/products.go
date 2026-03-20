package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

const productCols = `id, org_id, name, sku, description, unit_price_cents, currency, is_active, created_at, updated_at`

func scanProduct(row pgx.Row) (*domain.Product, error) {
	var p domain.Product
	var sku, description *string
	err := row.Scan(
		&p.ID, &p.OrgID, &p.Name, &sku, &description,
		&p.UnitPriceCents, &p.Currency, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if sku != nil {
		p.SKU = *sku
	}
	if description != nil {
		p.Description = *description
	}
	return &p, nil
}

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Currency == "" {
		p.Currency = "USD"
	}
	p.IsActive = true
	row := r.db.QueryRow(ctx,
		`INSERT INTO products (id, org_id, name, sku, description, unit_price_cents, currency, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING `+productCols,
		p.ID, p.OrgID, p.Name, nilIfEmpty(p.SKU), nilIfEmpty(p.Description),
		p.UnitPriceCents, p.Currency, p.IsActive,
	)
	return scanProduct(row)
}

func (r *ProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+productCols+` FROM products WHERE id=$1`, id,
	)
	return scanProduct(row)
}

func (r *ProductRepo) Update(ctx context.Context, id uuid.UUID, patch domain.ProductPatch) (*domain.Product, error) {
	setClauses := []string{}
	args := []any{}
	i := 1

	if patch.Name != nil {
		setClauses = append(setClauses, "name=$"+itoa(i))
		args = append(args, *patch.Name)
		i++
	}
	if patch.SKU != nil {
		setClauses = append(setClauses, "sku=$"+itoa(i))
		args = append(args, nilIfEmpty(*patch.SKU))
		i++
	}
	if patch.Description != nil {
		setClauses = append(setClauses, "description=$"+itoa(i))
		args = append(args, nilIfEmpty(*patch.Description))
		i++
	}
	if patch.UnitPriceCents != nil {
		setClauses = append(setClauses, "unit_price_cents=$"+itoa(i))
		args = append(args, *patch.UnitPriceCents)
		i++
	}
	if patch.Currency != nil {
		setClauses = append(setClauses, "currency=$"+itoa(i))
		args = append(args, *patch.Currency)
		i++
	}
	if patch.IsActive != nil {
		setClauses = append(setClauses, "is_active=$"+itoa(i))
		args = append(args, *patch.IsActive)
		i++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at=NOW()")
	args = append(args, id)
	q := `UPDATE products SET ` + strings.Join(setClauses, ",") +
		` WHERE id=$` + itoa(i) + ` RETURNING ` + productCols
	row := r.db.QueryRow(ctx, q, args...)
	return scanProduct(row)
}

func (r *ProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductRepo) List(ctx context.Context, filter domain.ProductFilter) ([]*domain.Product, int, error) {
	where := []string{"org_id=$1"}
	args := []any{filter.OrgID}
	i := 2

	if filter.Q != "" {
		where = append(where, "(name ILIKE $"+itoa(i)+" OR sku ILIKE $"+itoa(i)+")")
		args = append(args, "%"+filter.Q+"%")
		i++
	}
	if filter.IsActive != nil {
		where = append(where, "is_active=$"+itoa(i))
		args = append(args, *filter.IsActive)
		i++
	}

	clause := "WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM products `+clause, args...).Scan(&total); err != nil {
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
	rows, err := r.db.Query(ctx,
		`SELECT `+productCols+` FROM products `+clause+
			` ORDER BY name ASC LIMIT $`+itoa(i)+` OFFSET $`+itoa(i+1),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, rows.Err()
}

// nilIfEmpty converts an empty string to nil for optional TEXT columns.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
