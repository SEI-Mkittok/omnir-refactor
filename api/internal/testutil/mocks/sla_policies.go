package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockSLAPolicyRepository is a testify mock implementing repository.SLAPolicyRepository.
type MockSLAPolicyRepository struct {
	mock.Mock
}

func (m *MockSLAPolicyRepository) Create(ctx context.Context, p *domain.SLAPolicy) (*domain.SLAPolicy, error) {
	args := m.Called(ctx, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SLAPolicy), args.Error(1)
}

func (m *MockSLAPolicyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SLAPolicy, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SLAPolicy), args.Error(1)
}

func (m *MockSLAPolicyRepository) Update(ctx context.Context, id uuid.UUID, patch domain.SLAPolicyPatch) (*domain.SLAPolicy, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SLAPolicy), args.Error(1)
}

func (m *MockSLAPolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSLAPolicyRepository) List(ctx context.Context, orgID uuid.UUID) ([]*domain.SLAPolicy, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SLAPolicy), args.Error(1)
}

func (m *MockSLAPolicyRepository) MatchByPriority(ctx context.Context, priority domain.TicketPriority) (*domain.SLAPolicy, error) {
	args := m.Called(ctx, priority)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SLAPolicy), args.Error(1)
}
