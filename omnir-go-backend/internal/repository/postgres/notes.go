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

const noteCols = `id, content, entity_type, entity_id, author_id, created_at, updated_at, deleted_at`

func scanNote(row pgx.Row) (*domain.Note, error) {
	var n domain.Note
	err := row.Scan(
		&n.ID, &n.Content, &n.EntityType, &n.EntityID,
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
	now := time.Now().UTC()
	n.CreatedAt = now
	n.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO notes (id, content, entity_type, entity_id, author_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+noteCols,
		n.ID, n.Content, n.EntityType, n.EntityID, n.AuthorID, n.CreatedAt, n.UpdatedAt,
	)
	return scanNote(row)
}

// GetByID fetches a single note by ID.
func (r *NoteRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+noteCols+` FROM notes WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanNote(row)
}

// Delete soft-deletes a note.
func (r *NoteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`UPDATE notes SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListByEntity returns paginated notes for a given entity.
func (r *NoteRepo) ListByEntity(ctx context.Context, f domain.NoteFilter) ([]*domain.Note, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notes WHERE entity_type=$1 AND entity_id=$2 AND deleted_at IS NULL`,
		f.EntityType, f.EntityID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT `+noteCols+` FROM notes
		 WHERE entity_type=$1 AND entity_id=$2 AND deleted_at IS NULL
		 ORDER BY created_at DESC
		 LIMIT $3 OFFSET $4`,
		f.EntityType, f.EntityID, f.Limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notes []*domain.Note
	for rows.Next() {
		var n domain.Note
		if err := rows.Scan(
			&n.ID, &n.Content, &n.EntityType, &n.EntityID,
			&n.AuthorID, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		notes = append(notes, &n)
	}
	return notes, total, rows.Err()
}
