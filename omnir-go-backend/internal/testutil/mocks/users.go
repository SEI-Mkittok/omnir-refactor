package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockUserRepository is a testify mock implementing repository.UserRepository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CountAll(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, u *domain.User, passwordHash string) (*domain.User, error) {
	args := m.Called(ctx, u, passwordHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
