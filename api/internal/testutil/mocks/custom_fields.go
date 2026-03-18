package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockCustomFieldDefinitionRepository is a testify mock for CustomFieldDefinitionRepository.
type MockCustomFieldDefinitionRepository struct {
	mock.Mock
}

func (m *MockCustomFieldDefinitionRepository) Create(ctx context.Context, def *domain.CustomFieldDefinition) (*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, def)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomFieldDefinition), args.Error(1)
}

func (m *MockCustomFieldDefinitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomFieldDefinition), args.Error(1)
}

func (m *MockCustomFieldDefinitionRepository) Update(ctx context.Context, id uuid.UUID, patch domain.CustomFieldDefinitionPatch) (*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomFieldDefinition), args.Error(1)
}

func (m *MockCustomFieldDefinitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockCustomFieldDefinitionRepository) List(ctx context.Context, filter domain.CustomFieldDefinitionFilter) ([]*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CustomFieldDefinition), args.Error(1)
}
