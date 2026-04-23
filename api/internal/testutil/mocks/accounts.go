package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockAccountRepository is a testify mock implementing repository.AccountRepository.
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(ctx context.Context, a *domain.Account) (*domain.Account, error) {
	args := m.Called(ctx, a)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) Update(ctx context.Context, id uuid.UUID, patch domain.AccountPatch) (*domain.Account, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAccountRepository) GetByName(ctx context.Context, name string) (*domain.Account, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) List(ctx context.Context, filter domain.AccountFilter) ([]*domain.Account, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Account), args.Int(1), args.Error(2)
}

func (m *MockAccountRepository) ListLinkedEntities(ctx context.Context, id uuid.UUID, filter domain.LinkedEntityFilter) ([]domain.LinkedEntity, int, error) {
	args := m.Called(ctx, id, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]domain.LinkedEntity), args.Int(1), args.Error(2)
}

func (m *MockAccountRepository) CreateRelationship(ctx context.Context, rel *domain.AccountRelationship) (*domain.AccountRelationship, error) {
	args := m.Called(ctx, rel)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AccountRelationship), args.Error(1)
}

func (m *MockAccountRepository) DeleteRelationship(ctx context.Context, relationshipID uuid.UUID, deletedBy *uuid.UUID) error {
	args := m.Called(ctx, relationshipID, deletedBy)
	return args.Error(0)
}

func (m *MockAccountRepository) ListDescendants(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockAccountRepository) ListAncestors(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
