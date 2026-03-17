package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/omnir/crm-api/internal/domain"
)

// MockDealRepository is a testify mock implementing repository.DealRepository.
type MockDealRepository struct {
	mock.Mock
}

func (m *MockDealRepository) Create(ctx context.Context, d *domain.Deal) (*domain.Deal, error) {
	args := m.Called(ctx, d)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Deal), args.Error(1)
}

func (m *MockDealRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Deal), args.Error(1)
}

func (m *MockDealRepository) Update(ctx context.Context, id uuid.UUID, patch domain.DealPatch) (*domain.Deal, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Deal), args.Error(1)
}

func (m *MockDealRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDealRepository) List(ctx context.Context, filter domain.DealFilter) ([]*domain.Deal, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Deal), args.Int(1), args.Error(2)
}

func (m *MockDealRepository) AddContact(ctx context.Context, dealID, contactID uuid.UUID, role string) error {
	args := m.Called(ctx, dealID, contactID, role)
	return args.Error(0)
}

func (m *MockDealRepository) ListContacts(ctx context.Context, dealID uuid.UUID) ([]domain.Contact, error) {
	args := m.Called(ctx, dealID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Contact), args.Error(1)
}
