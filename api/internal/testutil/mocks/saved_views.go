package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockSavedViewRepository is a testify mock implementing repository.SavedViewRepository.
type MockSavedViewRepository struct {
	mock.Mock
}

func (m *MockSavedViewRepository) Create(ctx context.Context, v *domain.SavedView) (*domain.SavedView, error) {
	args := m.Called(ctx, v)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavedView), args.Error(1)
}

func (m *MockSavedViewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SavedView, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavedView), args.Error(1)
}

func (m *MockSavedViewRepository) Update(ctx context.Context, id uuid.UUID, patch domain.SavedViewPatch) (*domain.SavedView, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavedView), args.Error(1)
}

func (m *MockSavedViewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSavedViewRepository) List(ctx context.Context, filter domain.SavedViewFilter) ([]*domain.SavedView, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SavedView), args.Error(1)
}

func (m *MockSavedViewRepository) Pin(ctx context.Context, id uuid.UUID, isPinned bool) (*domain.SavedView, error) {
	args := m.Called(ctx, id, isPinned)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavedView), args.Error(1)
}
