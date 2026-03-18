package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockNotificationRepository is a testify mock implementing repository.NotificationRepository.
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
	args := m.Called(ctx, n)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAllRead(ctx context.Context, userID, orgID uuid.UUID) error {
	args := m.Called(ctx, userID, orgID)
	return args.Error(0)
}

func (m *MockNotificationRepository) UnreadCount(ctx context.Context, userID, orgID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID, orgID)
	return args.Int(0), args.Error(1)
}

func (m *MockNotificationRepository) ListByUser(ctx context.Context, filter domain.NotificationFilter) ([]*domain.Notification, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GenerateReminders(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
