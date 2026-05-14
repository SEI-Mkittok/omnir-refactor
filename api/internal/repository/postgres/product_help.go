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

// ProductHelpRepo is a PostgreSQL-backed ProductHelpRepository.
type ProductHelpRepo struct {
	db *pgxpool.Pool
}

// NewProductHelpRepo constructs a ProductHelpRepo.
func NewProductHelpRepo(db *pgxpool.Pool) *ProductHelpRepo {
	return &ProductHelpRepo{db: db}
}

const productHelpArticleCols = `a.id, a.category_id, coalesce(c.slug, ''), coalesce(c.name, ''), a.title, a.slug, a.body, a.excerpt, a.tags, a.status, a.source_path, a.wiki_url, a.edit_url, a.sort_order, a.view_count, a.created_at, a.updated_at, a.deleted_at`

func scanProductHelpArticle(row pgx.Row) (*domain.ProductHelpArticle, error) {
	var a domain.ProductHelpArticle
	err := row.Scan(
		&a.ID, &a.CategoryID, &a.CategorySlug, &a.CategoryName, &a.Title, &a.Slug,
		&a.Body, &a.Excerpt, &a.Tags, &a.Status, &a.SourcePath, &a.WikiURL, &a.EditURL,
		&a.SortOrder, &a.ViewCount, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if a.Tags == nil {
		a.Tags = []string{}
	}
	return &a, nil
}

func scanProductHelpSyncRun(row pgx.Row) (*domain.ProductHelpSyncRun, error) {
	var run domain.ProductHelpSyncRun
	err := row.Scan(
		&run.ID, &run.Status, &run.Message, &run.CategoriesCount,
		&run.ArticlesCount, &run.StartedAt, &run.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

// ListCategories returns active product-help categories with published article counts.
func (r *ProductHelpRepo) ListCategories(ctx context.Context) ([]*domain.ProductHelpCategory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.name, c.slug, c.sort_order, c.created_at, c.updated_at, c.deleted_at,
		       COUNT(a.id) FILTER (WHERE a.status='published' AND a.deleted_at IS NULL) AS article_count
		FROM product_help_categories c
		LEFT JOIN product_help_articles a ON a.category_id = c.id
		WHERE c.deleted_at IS NULL
		GROUP BY c.id
		ORDER BY c.sort_order ASC, c.name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.ProductHelpCategory
	for rows.Next() {
		var c domain.ProductHelpCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.ArticleCount); err != nil {
			return nil, err
		}
		out = append(out, &c)
	}
	return out, rows.Err()
}

// ListArticles returns active product-help articles, optionally filtered by query/category/status.
func (r *ProductHelpRepo) ListArticles(ctx context.Context, f domain.ProductHelpArticleFilter) ([]*domain.ProductHelpArticle, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"a.deleted_at IS NULL"}
	args := []any{}
	n := 0

	status := domain.ProductHelpArticleStatusPublished
	if f.Status != nil {
		status = *f.Status
	}
	n++
	args = append(args, string(status))
	where = append(where, "a.status=$"+itoa(n))

	if f.CategorySlug != "" {
		n++
		args = append(args, f.CategorySlug)
		where = append(where, "c.slug=$"+itoa(n))
	}

	var rankSel, rankJoin, orderBy string
	orderBy = "a.sort_order ASC, a.title ASC"
	if f.Query != "" {
		n++
		args = append(args, f.Query)
		rankSel = `, ts_rank(a.search_vec, query) AS rank`
		rankJoin = `, plainto_tsquery('english', $` + itoa(n) + `) query`
		where = append(where, "a.search_vec @@ query")
		orderBy = "rank DESC, a.sort_order ASC, a.title ASC"
	}

	whereStr := strings.Join(where, " AND ")
	from := ` FROM product_help_articles a LEFT JOIN product_help_categories c ON c.id = a.category_id` + rankJoin

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*)`+from+` WHERE `+whereStr, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	n++
	args = append(args, f.Limit)
	limPos := itoa(n)
	n++
	args = append(args, offset)
	offPos := itoa(n)

	rows, err := r.db.Query(ctx, `SELECT `+productHelpArticleCols+rankSel+from+` WHERE `+whereStr+
		` ORDER BY `+orderBy+` LIMIT $`+limPos+` OFFSET $`+offPos, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*domain.ProductHelpArticle
	for rows.Next() {
		var a domain.ProductHelpArticle
		dest := []any{
			&a.ID, &a.CategoryID, &a.CategorySlug, &a.CategoryName, &a.Title, &a.Slug,
			&a.Body, &a.Excerpt, &a.Tags, &a.Status, &a.SourcePath, &a.WikiURL, &a.EditURL,
			&a.SortOrder, &a.ViewCount, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
		}
		if f.Query != "" {
			var rank float64
			dest = append(dest, &rank)
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, 0, err
		}
		if a.Tags == nil {
			a.Tags = []string{}
		}
		if !f.IncludeBody {
			a.Body = ""
		}
		out = append(out, &a)
	}
	return out, total, rows.Err()
}

// GetArticleBySlug fetches a published product-help article by slug.
func (r *ProductHelpRepo) GetArticleBySlug(ctx context.Context, slug string) (*domain.ProductHelpArticle, error) {
	return scanProductHelpArticle(r.db.QueryRow(ctx, `SELECT `+productHelpArticleCols+`
		FROM product_help_articles a
		LEFT JOIN product_help_categories c ON c.id = a.category_id
		WHERE a.slug=$1 AND a.status='published' AND a.deleted_at IS NULL`, slug))
}

// IncrementViewCount increments a product-help article's view count.
func (r *ProductHelpRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE product_help_articles SET view_count=view_count+1 WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

// LatestSyncRun returns the most recent product-help sync run.
func (r *ProductHelpRepo) LatestSyncRun(ctx context.Context) (*domain.ProductHelpSyncRun, error) {
	return scanProductHelpSyncRun(r.db.QueryRow(ctx, `
		SELECT id, status, message, categories_count, articles_count, started_at, finished_at
		FROM product_help_sync_runs
		ORDER BY started_at DESC
		LIMIT 1`))
}

// RecordSyncRun inserts a product-help sync run without content changes.
func (r *ProductHelpRepo) RecordSyncRun(ctx context.Context, run *domain.ProductHelpSyncRun) (*domain.ProductHelpSyncRun, error) {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.FinishedAt == nil {
		t := time.Now().UTC()
		run.FinishedAt = &t
	}
	return scanProductHelpSyncRun(r.db.QueryRow(ctx, `
		INSERT INTO product_help_sync_runs (id, status, message, categories_count, articles_count, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, status, message, categories_count, articles_count, started_at, finished_at`,
		run.ID, string(run.Status), run.Message, run.CategoriesCount, run.ArticlesCount, run.StartedAt, run.FinishedAt,
	))
}

// ReplaceContent upserts product-help content and soft-removes entries missing from the latest wiki manifest.
func (r *ProductHelpRepo) ReplaceContent(ctx context.Context, categories []*domain.ProductHelpCategory, articles []*domain.ProductHelpArticle, run *domain.ProductHelpSyncRun) (*domain.ProductHelpSyncRun, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	categoryIDs := map[string]uuid.UUID{}
	categorySlugs := make([]string, 0, len(categories))
	for _, c := range categories {
		if c.ID == uuid.Nil {
			c.ID = uuid.New()
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO product_help_categories (id, name, slug, sort_order, deleted_at, updated_at)
			VALUES ($1, $2, $3, $4, NULL, now())
			ON CONFLICT (slug) DO UPDATE SET
				name=excluded.name,
				sort_order=excluded.sort_order,
				deleted_at=NULL,
				updated_at=now()
			RETURNING id`,
			c.ID, c.Name, c.Slug, c.SortOrder,
		)
		if err := row.Scan(&c.ID); err != nil {
			return nil, err
		}
		categoryIDs[c.Slug] = c.ID
		categorySlugs = append(categorySlugs, c.Slug)
	}
	if len(categorySlugs) == 0 {
		if _, err := tx.Exec(ctx, `UPDATE product_help_categories SET deleted_at=now(), updated_at=now() WHERE deleted_at IS NULL`); err != nil {
			return nil, err
		}
	} else if _, err := tx.Exec(ctx, `UPDATE product_help_categories SET deleted_at=now(), updated_at=now() WHERE deleted_at IS NULL AND NOT (slug = ANY($1))`, categorySlugs); err != nil {
		return nil, err
	}

	articleSlugs := make([]string, 0, len(articles))
	for _, a := range articles {
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		if a.CategorySlug != "" {
			if id, ok := categoryIDs[a.CategorySlug]; ok {
				a.CategoryID = &id
			}
		}
		var categoryID any
		if a.CategoryID != nil {
			categoryID = *a.CategoryID
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO product_help_articles (id, category_id, title, slug, body, excerpt, tags, status, source_path, wiki_url, edit_url, sort_order, deleted_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NULL, now())
			ON CONFLICT (slug) DO UPDATE SET
				category_id=excluded.category_id,
				title=excluded.title,
				body=excluded.body,
				excerpt=excluded.excerpt,
				tags=excluded.tags,
				status=excluded.status,
				source_path=excluded.source_path,
				wiki_url=excluded.wiki_url,
				edit_url=excluded.edit_url,
				sort_order=excluded.sort_order,
				deleted_at=NULL,
				updated_at=now()
			RETURNING id`,
			a.ID, categoryID, a.Title, a.Slug, a.Body, a.Excerpt, a.Tags, string(a.Status), a.SourcePath, a.WikiURL, a.EditURL, a.SortOrder,
		)
		if err := row.Scan(&a.ID); err != nil {
			return nil, err
		}
		articleSlugs = append(articleSlugs, a.Slug)
	}
	if len(articleSlugs) == 0 {
		if _, err := tx.Exec(ctx, `UPDATE product_help_articles SET deleted_at=now(), updated_at=now() WHERE deleted_at IS NULL`); err != nil {
			return nil, err
		}
	} else if _, err := tx.Exec(ctx, `UPDATE product_help_articles SET deleted_at=now(), updated_at=now() WHERE deleted_at IS NULL AND NOT (slug = ANY($1))`, articleSlugs); err != nil {
		return nil, err
	}

	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.FinishedAt == nil {
		t := time.Now().UTC()
		run.FinishedAt = &t
	}
	synced, err := scanProductHelpSyncRun(tx.QueryRow(ctx, `
		INSERT INTO product_help_sync_runs (id, status, message, categories_count, articles_count, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, status, message, categories_count, articles_count, started_at, finished_at`,
		run.ID, string(run.Status), run.Message, run.CategoriesCount, run.ArticlesCount, run.StartedAt, run.FinishedAt,
	))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return synced, nil
}
