package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type AuditLogRepo struct{ db *pgxpool.Pool }

func NewAuditLogRepo(db *pgxpool.Pool) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Create(ctx context.Context, entry *domain.AuditLogEntry) error {
	return nil
}

func (r *AuditLogRepo) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.AuditLogEntry, int, error) {
	return nil, 0, nil
}
