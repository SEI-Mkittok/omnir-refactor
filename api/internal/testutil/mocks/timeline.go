package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockTimelineRepository is a testify mock implementing repository.TimelineRepository.
type MockTimelineRepository struct {
	mock.Mock
}

func (m *MockTimelineRepository) List(ctx context.Context, filter domain.TimelineFilter) ([]*domain.TimelineEvent, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.TimelineEvent), args.Int(1), args.Error(2)
}
