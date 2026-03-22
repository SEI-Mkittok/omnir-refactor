package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockTeamsConnectionRepository is a testify mock implementing repository.TeamsConnectionRepository.
type MockTeamsConnectionRepository struct {
	mock.Mock
}

func (m *MockTeamsConnectionRepository) Upsert(ctx context.Context, c *domain.TeamsConnection) (*domain.TeamsConnection, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TeamsConnection), args.Error(1)
}

func (m *MockTeamsConnectionRepository) GetByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.TeamsConnection, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TeamsConnection), args.Error(1)
}

func (m *MockTeamsConnectionRepository) Delete(ctx context.Context, orgID uuid.UUID) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}
