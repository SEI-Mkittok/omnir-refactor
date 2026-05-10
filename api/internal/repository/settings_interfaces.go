package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

type UserPreferenceRepository interface {
	GetByUser(ctx context.Context, userID, orgID uuid.UUID) (*domain.UserPreferences, error)
	Upsert(ctx context.Context, userID, orgID uuid.UUID, patch domain.UserPreferencesPatch) (*domain.UserPreferences, error)
}

type CurrencyRepository interface {
	List(ctx context.Context, orgID uuid.UUID) ([]*domain.OrgCurrency, error)
	Replace(ctx context.Context, orgID uuid.UUID, req domain.OrgCurrencyUpdateRequest) ([]*domain.OrgCurrency, error)
	GetDefaultCode(ctx context.Context, orgID uuid.UUID) (string, error)
}

type PicklistRepository interface {
	ListValues(ctx context.Context, orgID, customFieldID uuid.UUID) ([]*domain.PicklistValue, error)
	UpsertValues(ctx context.Context, orgID, customFieldID uuid.UUID, values []domain.PicklistValueInput) ([]*domain.PicklistValue, error)
	RemapAndDeleteValue(ctx context.Context, orgID, customFieldID uuid.UUID, fromValue string, toValue *string) error
	ListDependencies(ctx context.Context, orgID uuid.UUID, entityType *domain.CustomFieldEntityType) ([]*domain.PicklistDependency, error)
	UpsertDependency(ctx context.Context, orgID uuid.UUID, input domain.PicklistDependencyInput) (*domain.PicklistDependency, error)
	DeleteDependency(ctx context.Context, orgID, dependencyID uuid.UUID) error
}

type LeadConversionMappingRepository interface {
	List(ctx context.Context, orgID uuid.UUID) ([]*domain.LeadConversionMapping, error)
	Replace(ctx context.Context, orgID uuid.UUID, rows []domain.LeadConversionMapping) ([]*domain.LeadConversionMapping, error)
}

