package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockActivityRepository is a testify mock implementing repository.ActivityRepository.
type MockActivityRepository struct {
	mock.Mock
}

func (m *MockActivityRepository) Create(ctx context.Context, a *domain.Activity) (*domain.Activity, error) {
	args := m.Called(ctx, a)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Activity), args.Error(1)
}

func (m *MockActivityRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Activity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Activity), args.Error(1)
}

func (m *MockActivityRepository) Update(ctx context.Context, id uuid.UUID, patch domain.ActivityPatch) (*domain.Activity, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Activity), args.Error(1)
}

func (m *MockActivityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockActivityRepository) List(ctx context.Context, filter domain.ActivityFilter) ([]*domain.Activity, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Activity), args.Int(1), args.Error(2)
}
