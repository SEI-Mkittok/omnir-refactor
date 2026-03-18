package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type EntityAttachmentRepo struct{ db *pgxpool.Pool }

func NewEntityAttachmentRepo(db *pgxpool.Pool) *EntityAttachmentRepo {
	return &EntityAttachmentRepo{db: db}
}

func (r *EntityAttachmentRepo) Create(ctx context.Context, a *domain.EntityAttachment) (*domain.EntityAttachment, error) {
	return nil, nil
}

func (r *EntityAttachmentRepo) List(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.EntityAttachment, error) {
	return nil, nil
}

func (r *EntityAttachmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EntityAttachment, error) {
	return nil, nil
}

func (r *EntityAttachmentRepo) Delete(ctx context.Context, id uuid.UUID, entityType domain.EntityType, entityID uuid.UUID) error {
	return nil
}
