package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// KBCategoryRepo is a PostgreSQL-backed KBCategoryRepository.
type KBCategoryRepo struct {
	db *pgxpool.Pool
}

// NewKBCategoryRepo constructs a KBCategoryRepo.
func NewKBCategoryRepo(db *pgxpool.Pool) *KBCategoryRepo {
	return &KBCategoryRepo{db: db}
}

const kbCatCols = `id, org_id, name, slug, sort_order, created_at, updated_at`

func scanKBCategory(row pgx.Row) (*domain.KBCategory, error) {
	var c domain.KBCategory
	err := row.Scan(&c.ID, &c.OrgID, &c.Name, &c.Slug, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// Create inserts a new category.
func (r *KBCategoryRepo) Create(ctx context.Context, c *domain.KBCategory) (*domain.KBCategory, error) {
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
		INSERT INTO article_categories (id, org_id, name, slug, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+kbCatCols,
		c.ID, c.OrgID, c.Name, c.Slug, c.SortOrder, c.CreatedAt, c.UpdatedAt,
	)
	return scanKBCategory(row)
}

// GetByID fetches a category by ID.
func (r *KBCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.KBCategory, error) {
	q := `SELECT ` + kbCatCols + ` FROM article_categories WHERE id=$1`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	return scanKBCategory(r.db.QueryRow(ctx, q, args...))
}

// Update applies a patch to a category.
func (r *KBCategoryRepo) Update(ctx context.Context, id uuid.UUID, patch domain.KBCategoryPatch) (*domain.KBCategory, error) {
	set := []string{"updated_at=now()"}
	args := []any{}
	n := 1

	if patch.Name != nil {
		args = append(args, *patch.Name)
		n++
		set = append(set, "name=$"+itoa(n))
	}
	if patch.Slug != nil {
		args = append(args, *patch.Slug)
		n++
		set = append(set, "slug=$"+itoa(n))
	}
	if patch.SortOrder != nil {
		args = append(args, *patch.SortOrder)
		n++
		set = append(set, "sort_order=$"+itoa(n))
	}

	n++
	args = append(args, id)
	q := `UPDATE article_categories SET ` + strings.Join(set, ",") + ` WHERE id=$` + itoa(n) + ` RETURNING ` + kbCatCols

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		n++
		args = append(args, orgID)
		q = `UPDATE article_categories SET ` + strings.Join(set, ",") + ` WHERE id=$` + itoa(n-1) + ` AND org_id=$` + itoa(n) + ` RETURNING ` + kbCatCols
	}

	return scanKBCategory(r.db.QueryRow(ctx, q, args...))
}

// Delete removes a category.
func (r *KBCategoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM article_categories WHERE id=$1`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	ct, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// List returns paginated categories for the org.
func (r *KBCategoryRepo) List(ctx context.Context, f domain.KBCategoryFilter) ([]*domain.KBCategory, int, error) {
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

	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM article_categories WHERE org_id=$1`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+kbCatCols+` FROM article_categories
		WHERE org_id=$1
		ORDER BY sort_order ASC, name ASC
		LIMIT $2 OFFSET $3`,
		orgID, f.Limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*domain.KBCategory
	for rows.Next() {
		var c domain.KBCategory
		if err := rows.Scan(&c.ID, &c.OrgID, &c.Name, &c.Slug, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, &c)
	}
	return out, total, rows.Err()
}
