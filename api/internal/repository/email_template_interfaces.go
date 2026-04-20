package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

type EmailTemplateRepository interface {
	Create(ctx context.Context, template *domain.EmailTemplate) (*domain.EmailTemplate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailTemplate, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.EmailTemplatePatch) (*domain.EmailTemplate, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.EmailTemplateFilter) ([]*domain.EmailTemplate, int, error)
}
