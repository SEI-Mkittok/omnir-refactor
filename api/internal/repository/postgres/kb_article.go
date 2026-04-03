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

// KBArticleRepo is a PostgreSQL-backed KBArticleRepository.
type KBArticleRepo struct {
	db *pgxpool.Pool
}

// NewKBArticleRepo constructs a KBArticleRepo.
func NewKBArticleRepo(db *pgxpool.Pool) *KBArticleRepo {
	return &KBArticleRepo{db: db}
}

const kbArtCols = `id, org_id, title, body, category_id, tags, status, author_id, view_count, number, number_prefix, created_at, updated_at, deleted_at`

func scanKBArticle(row pgx.Row) (*domain.KBArticle, error) {
	var a domain.KBArticle
	err := row.Scan(
		&a.ID, &a.OrgID, &a.Title, &a.Body, &a.CategoryID,
		&a.Tags, &a.Status, &a.AuthorID, &a.ViewCount,
		&a.Number, &a.NumberPrefix, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
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

// Create inserts a new article.
func (r *KBArticleRepo) Create(ctx context.Context, a *domain.KBArticle) (*domain.KBArticle, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		a.OrgID = orgID
	}
	if a.Status == "" {
		a.Status = domain.KBArticleStatusDraft
	}
	if a.Tags == nil {
		a.Tags = []string{}
	}
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now

	num, err := getNextDocNumber(ctx, r.db, a.OrgID, domain.DocTypeKBArticle)
	if err != nil {
		return nil, err
	}
	const prefix = "KB"

	row := r.db.QueryRow(ctx, `
		INSERT INTO articles (id, org_id, title, body, category_id, tags, status, author_id, number, number_prefix, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING `+kbArtCols,
		a.ID, a.OrgID, a.Title, a.Body, a.CategoryID, a.Tags, string(a.Status), a.AuthorID, num, prefix, a.CreatedAt, a.UpdatedAt,
	)
	return scanKBArticle(row)
}

// GetByID fetches an article by ID.
func (r *KBArticleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.KBArticle, error) {
	q := `SELECT ` + kbArtCols + ` FROM articles WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}
	return scanKBArticle(r.db.QueryRow(ctx, q, args...))
}

// Update applies a patch to an article.
func (r *KBArticleRepo) Update(ctx context.Context, id uuid.UUID, patch domain.KBArticlePatch) (*domain.KBArticle, error) {
	set := []string{"updated_at=now()"}
	args := []any{}
	n := 0

	addArg := func(col string, val any) {
		n++
		args = append(args, val)
		set = append(set, col+"=$"+itoa(n))
	}

	if patch.Title != nil {
		addArg("title", *patch.Title)
	}
	if patch.Body != nil {
		addArg("body", *patch.Body)
	}
	if patch.CategoryID != nil {
		addArg("category_id", *patch.CategoryID)
	}
	if patch.Tags != nil {
		addArg("tags", patch.Tags)
	}
	if patch.Status != nil {
		addArg("status", string(*patch.Status))
	}

	n++
	args = append(args, id)
	idPos := itoa(n)

	orgClause := ""
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		n++
		args = append(args, orgID)
		orgClause = ` AND org_id=$` + itoa(n)
	}

	q := `UPDATE articles SET ` + strings.Join(set, ",") +
		` WHERE id=$` + idPos + ` AND deleted_at IS NULL` + orgClause +
		` RETURNING ` + kbArtCols

	return scanKBArticle(r.db.QueryRow(ctx, q, args...))
}

// Delete soft-deletes an article.
func (r *KBArticleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE articles SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL`
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

// List returns paginated articles, optionally filtered/searched.
func (r *KBArticleRepo) List(ctx context.Context, f domain.KBArticleFilter) ([]*domain.KBArticle, int, error) {
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

	where := []string{"a.org_id=$1", "a.deleted_at IS NULL"}
	args := []any{orgID}
	n := 1

	if f.Status != nil {
		n++
		args = append(args, string(*f.Status))
		where = append(where, "a.status=$"+itoa(n))
	}
	if f.CategoryID != nil {
		n++
		args = append(args, *f.CategoryID)
		where = append(where, "a.category_id=$"+itoa(n))
	}

	whereStr := strings.Join(where, " AND ")

	orderBy := "a.created_at DESC"
	var rankSel, rankJoin string
	if f.Query != "" {
		n++
		args = append(args, f.Query)
		rankSel = `, ts_rank(a.search_vec, query) AS rank`
		rankJoin = `, plainto_tsquery('english', $` + itoa(n) + `) query`
		whereStr += ` AND a.search_vec @@ query`
		orderBy = "rank DESC, a.created_at DESC"
	}

	var total int
	countQ := `SELECT COUNT(*) FROM articles a` + rankJoin + ` WHERE ` + whereStr
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	n++
	args = append(args, f.Limit)
	limPos := itoa(n)
	n++
	args = append(args, offset)
	offPos := itoa(n)

	listQ := `SELECT ` + kbArtCols + rankSel + ` FROM articles a` + rankJoin +
		` WHERE ` + whereStr +
		` ORDER BY ` + orderBy +
		` LIMIT $` + limPos + ` OFFSET $` + offPos

	rows, err := r.db.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*domain.KBArticle
	for rows.Next() {
		var a domain.KBArticle
		dest := []any{
			&a.ID, &a.OrgID, &a.Title, &a.Body, &a.CategoryID,
			&a.Tags, &a.Status, &a.AuthorID, &a.ViewCount,
			&a.Number, &a.NumberPrefix, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
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
		out = append(out, &a)
	}
	return out, total, rows.Err()
}

// IncrementViewCount increments the view_count for an article.
func (r *KBArticleRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE articles SET view_count=view_count+1 WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

// Suggest returns top KB articles matching the subject string via FTS.
func (r *KBArticleRepo) Suggest(ctx context.Context, orgID uuid.UUID, subject string, limit int) ([]*domain.KBSuggestResult, error) {
	if limit <= 0 {
		limit = 5
	}
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.title
		FROM articles a, plainto_tsquery('english', $1) query
		WHERE a.org_id=$2
		  AND a.status='published'
		  AND a.deleted_at IS NULL
		  AND a.search_vec @@ query
		ORDER BY ts_rank(a.search_vec, query) DESC
		LIMIT $3`,
		subject, orgID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.KBSuggestResult
	for rows.Next() {
		var s domain.KBSuggestResult
		if err := rows.Scan(&s.ID, &s.Title); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}
