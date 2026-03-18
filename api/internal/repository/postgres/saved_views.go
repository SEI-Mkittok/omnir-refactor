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

type SavedViewRepo struct {
	db *pgxpool.Pool
}

func NewSavedViewRepo(db *pgxpool.Pool) *SavedViewRepo {
	return &SavedViewRepo{db: db}
}

const savedViewCols = `
	id, org_id, created_by, entity_type, name, filters,
	sort_by, sort_dir, is_shared, is_pinned, pinned_order,
	created_at, updated_at, deleted_at
`

func scanSavedView(row pgx.Row) (*domain.SavedView, error) {
	var v domain.SavedView
	err := row.Scan(
		&v.ID, &v.OrgID, &v.CreatedBy, &v.EntityType, &v.Name, &v.Filters,
		&v.SortBy, &v.SortDir, &v.IsShared, &v.IsPinned, &v.PinnedOrder,
		&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *SavedViewRepo) Create(ctx context.Context, v *domain.SavedView) (*domain.SavedView, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		v.OrgID = orgID
	}
	now := time.Now().UTC()
	v.CreatedAt = now
	v.UpdatedAt = now
	if v.Filters == nil {
		v.Filters = []byte("[]")
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO saved_views
			(id, org_id, created_by, entity_type, name, filters,
			 sort_by, sort_dir, is_shared, is_pinned, pinned_order, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING `+savedViewCols,
		v.ID, v.OrgID, v.CreatedBy, v.EntityType, v.Name, v.Filters,
		v.SortBy, v.SortDir, v.IsShared, v.IsPinned, v.PinnedOrder, v.CreatedAt, v.UpdatedAt,
	)
	return scanSavedView(row)
}

func (r *SavedViewRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SavedView, error) {
	q := `SELECT ` + savedViewCols + ` FROM saved_views WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	return scanSavedView(row)
}

func (r *SavedViewRepo) Update(ctx context.Context, id uuid.UUID, patch domain.SavedViewPatch) (*domain.SavedView, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.Name != nil {
		addArg("name", *patch.Name)
	}
	if patch.Filters != nil {
		addArg("filters", patch.Filters)
	}
	if patch.SortBy != nil {
		addArg("sort_by", *patch.SortBy)
	}
	if patch.SortDir != nil {
		addArg("sort_dir", *patch.SortDir)
	}
	if patch.IsShared != nil {
		addArg("is_shared", *patch.IsShared)
	}
	if patch.IsPinned != nil {
		addArg("is_pinned", *patch.IsPinned)
	}
	if patch.PinnedOrder != nil {
		addArg("pinned_order", *patch.PinnedOrder)
	}

	whereClause := fmt.Sprintf(`id=$%d AND deleted_at IS NULL`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
	}

	query := fmt.Sprintf(
		`UPDATE saved_views SET %s WHERE %s RETURNING %s`,
		strings.Join(sets, ", "), whereClause, savedViewCols,
	)
	row := r.db.QueryRow(ctx, query, args...)
	return scanSavedView(row)
}

func (r *SavedViewRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE saved_views SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
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

func (r *SavedViewRepo) List(ctx context.Context, filter domain.SavedViewFilter) ([]*domain.SavedView, error) {
	q := `SELECT ` + savedViewCols + `
		FROM saved_views
		WHERE org_id=$1
		  AND (created_by=$2 OR is_shared=true)
		  AND entity_type=$3
		  AND deleted_at IS NULL
		ORDER BY is_pinned DESC, pinned_order ASC NULLS LAST, name ASC`

	rows, err := r.db.Query(ctx, q, filter.OrgID, filter.UserID, filter.EntityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []*domain.SavedView
	for rows.Next() {
		v, err := scanSavedView(rows)
		if err != nil {
			return nil, err
		}
		views = append(views, v)
	}
	return views, rows.Err()
}

func (r *SavedViewRepo) Pin(ctx context.Context, id uuid.UUID, isPinned bool) (*domain.SavedView, error) {
	if !isPinned {
		// Unpin: clear pinned_order
		q := `UPDATE saved_views SET is_pinned=false, pinned_order=NULL, updated_at=NOW()
			WHERE id=$1 AND deleted_at IS NULL RETURNING ` + savedViewCols
		args := []any{id}
		if orgID, ok := domain.OrgIDFromContext(ctx); ok {
			q = `UPDATE saved_views SET is_pinned=false, pinned_order=NULL, updated_at=NOW()
				WHERE id=$1 AND org_id=$2 AND deleted_at IS NULL RETURNING ` + savedViewCols
			args = append(args, orgID)
		}
		return scanSavedView(r.db.QueryRow(ctx, q, args...))
	}

	// Pin: assign next pinned_order within org+entity_type
	var nextOrder int
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(MAX(pinned_order), 0) + 1
		FROM saved_views
		WHERE org_id=$1 AND entity_type=$2 AND is_pinned=true AND deleted_at IS NULL`,
		current.OrgID, current.EntityType,
	).Scan(&nextOrder)
	if err != nil {
		return nil, err
	}

	q := `UPDATE saved_views SET is_pinned=true, pinned_order=$1, updated_at=NOW()
		WHERE id=$2 AND deleted_at IS NULL RETURNING ` + savedViewCols
	args := []any{nextOrder, id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q = `UPDATE saved_views SET is_pinned=true, pinned_order=$1, updated_at=NOW()
			WHERE id=$2 AND org_id=$3 AND deleted_at IS NULL RETURNING ` + savedViewCols
		args = append(args, orgID)
	}
	return scanSavedView(r.db.QueryRow(ctx, q, args...))
}
