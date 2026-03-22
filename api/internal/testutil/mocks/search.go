package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockSearchRepository is a testify mock implementing repository.SearchRepository.
type MockSearchRepository struct {
	mock.Mock
}

func (m *MockSearchRepository) Search(ctx context.Context, q string, limit int) (*domain.SearchGroupedResult, error) {
	args := m.Called(ctx, q, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SearchGroupedResult), args.Error(1)
}
