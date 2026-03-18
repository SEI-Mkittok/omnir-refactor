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

// EntityAttachmentRepo implements repository.EntityAttachmentRepository.
type EntityAttachmentRepo struct {
	db *pgxpool.Pool
}

func NewEntityAttachmentRepo(db *pgxpool.Pool) *EntityAttachmentRepo {
	return &EntityAttachmentRepo{db: db}
}

const entityAttachmentCols = `id, entity_type, entity_id, org_id, uploaded_by, filename, content_type, size_bytes, storage_path, created_at`

func scanEntityAttachment(row pgx.Row) (*domain.EntityAttachment, error) {
	var a domain.EntityAttachment
	err := row.Scan(
		&a.ID, &a.EntityType, &a.EntityID, &a.OrgID, &a.UploadedBy,
		&a.Filename, &a.ContentType, &a.SizeBytes, &a.StoragePath, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *EntityAttachmentRepo) Create(ctx context.Context, a *domain.EntityAttachment) (*domain.EntityAttachment, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		a.OrgID = orgID
	}
	a.CreatedAt = time.Now().UTC()

	row := r.db.QueryRow(ctx, `
		INSERT INTO entity_attachments
			(id, entity_type, entity_id, org_id, uploaded_by, filename, content_type, size_bytes, storage_path, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+entityAttachmentCols,
		a.ID, a.EntityType, a.EntityID, a.OrgID, a.UploadedBy,
		a.Filename, a.ContentType, a.SizeBytes, a.StoragePath, a.CreatedAt,
	)
	return scanEntityAttachment(row)
}

func (r *EntityAttachmentRepo) List(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.EntityAttachment, error) {
	args := []any{entityType, entityID}
	q := `SELECT ` + entityAttachmentCols + ` FROM entity_attachments WHERE entity_type = $1 AND entity_id = $2`

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $3`
		args = append(args, orgID)
	}
	q += ` ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*domain.EntityAttachment
	for rows.Next() {
		var a domain.EntityAttachment
		if err := rows.Scan(
			&a.ID, &a.EntityType, &a.EntityID, &a.OrgID, &a.UploadedBy,
			&a.Filename, &a.ContentType, &a.SizeBytes, &a.StoragePath, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, &a)
	}
	return attachments, rows.Err()
}

func (r *EntityAttachmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EntityAttachment, error) {
	q := `SELECT ` + entityAttachmentCols + ` FROM entity_attachments WHERE id = $1`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $2`
		args = append(args, orgID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	return scanEntityAttachment(row)
}

func (r *EntityAttachmentRepo) Delete(ctx context.Context, id uuid.UUID, entityType domain.EntityType, entityID uuid.UUID) error {
	q := `DELETE FROM entity_attachments WHERE id = $1 AND entity_type = $2 AND entity_id = $3`
	args := []any{id, entityType, entityID}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $4`
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
