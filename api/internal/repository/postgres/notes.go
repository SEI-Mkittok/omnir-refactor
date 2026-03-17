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

// NoteRepo is a PostgreSQL-backed NoteRepository.
type NoteRepo struct {
	db *pgxpool.Pool
}

// NewNoteRepo constructs a NoteRepo.
func NewNoteRepo(db *pgxpool.Pool) *NoteRepo {
	return &NoteRepo{db: db}
}

const noteCols = `id, org_id, content, entity_type, entity_id, author_id, created_at, updated_at, deleted_at`

func scanNote(row pgx.Row) (*domain.Note, error) {
	var n domain.Note
	err := row.Scan(
		&n.ID, &n.OrgID, &n.Content, &n.EntityType, &n.EntityID,
		&n.AuthorID, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &n, nil
}

// Create inserts a new note and returns the persisted record.
func (r *NoteRepo) Create(ctx context.Context, n *domain.Note) (*domain.Note, error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		n.OrgID = orgID
	}
	now := time.Now().UTC()
	n.CreatedAt = now
	n.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO notes (id, org_id, content, entity_type, entity_id, author_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING `+noteCols,
		n.ID, n.OrgID, n.Content, n.EntityType, n.EntityID, n.AuthorID, n.CreatedAt, n.UpdatedAt,
	)
	return scanNote(row)
}

// GetByID fetches a single note by ID.
func (r *NoteRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	q := `SELECT ` + noteCols + ` FROM notes WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	return scanNote(row)
}

// Delete soft-deletes a note.
func (r *NoteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE notes SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
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

// ListByEntity returns paginated notes for a given entity, scoped to org.
func (r *NoteRepo) ListByEntity(ctx context.Context, f domain.NoteFilter) ([]*domain.Note, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	// Build org condition.
	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if !hasCtxOrg {
		orgID = f.OrgID
	}

	var countQ, listQ string
	var countArgs, listArgs []any

	if orgID != uuid.Nil {
		countQ = `SELECT COUNT(*) FROM notes WHERE entity_type=$1 AND entity_id=$2 AND deleted_at IS NULL AND org_id=$3`
		countArgs = []any{f.EntityType, f.EntityID, orgID}
		listQ = `SELECT ` + noteCols + ` FROM notes
			WHERE entity_type=$1 AND entity_id=$2 AND deleted_at IS NULL AND org_id=$3
			ORDER BY created_at DESC LIMIT $4 OFFSET $5`
		listArgs = []any{f.EntityType, f.EntityID, orgID, f.Limit, offset}
	} else {
		countQ = `SELECT COUNT(*) FROM notes WHERE entity_type=$1 AND entity_id=$2 AND deleted_at IS NULL`
		countArgs = []any{f.EntityType, f.EntityID}
		listQ = `SELECT ` + noteCols + ` FROM notes
			WHERE entity_type=$1 AND entity_id=$2 AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		listArgs = []any{f.EntityType, f.EntityID, f.Limit, offset}
	}

	var total int
	if err := r.db.QueryRow(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notes []*domain.Note
	for rows.Next() {
		var n domain.Note
		if err := rows.Scan(
			&n.ID, &n.OrgID, &n.Content, &n.EntityType, &n.EntityID,
			&n.AuthorID, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		notes = append(notes, &n)
	}
	return notes, total, rows.Err()
}
