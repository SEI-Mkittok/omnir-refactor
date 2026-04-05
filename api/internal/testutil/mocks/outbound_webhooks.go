package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockOutboundWebhookRepository is a testify mock implementing repository.OutboundWebhookRepository.
type MockOutboundWebhookRepository struct {
	mock.Mock
}

func (m *MockOutboundWebhookRepository) Create(ctx context.Context, w *domain.Webhook) (*domain.Webhook, error) {
	args := m.Called(ctx, w)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Webhook), args.Error(1)
}

func (m *MockOutboundWebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Webhook), args.Error(1)
}

func (m *MockOutboundWebhookRepository) List(ctx context.Context) ([]*domain.Webhook, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Webhook), args.Error(1)
}

func (m *MockOutboundWebhookRepository) ListByEvent(ctx context.Context, orgID uuid.UUID, event domain.WebhookEvent) ([]*domain.Webhook, error) {
	args := m.Called(ctx, orgID, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Webhook), args.Error(1)
}

func (m *MockOutboundWebhookRepository) Update(ctx context.Context, id uuid.UUID, patch domain.WebhookPatch) (*domain.Webhook, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Webhook), args.Error(1)
}

func (m *MockOutboundWebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOutboundWebhookRepository) CreateDelivery(ctx context.Context, d *domain.WebhookDelivery) (*domain.WebhookDelivery, error) {
	args := m.Called(ctx, d)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WebhookDelivery), args.Error(1)
}

func (m *MockOutboundWebhookRepository) UpdateDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	args := m.Called(ctx, d)
	return args.Error(0)
}

func (m *MockOutboundWebhookRepository) PendingDeliveries(ctx context.Context) ([]*domain.WebhookDelivery, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.WebhookDelivery), args.Error(1)
}

func (m *MockOutboundWebhookRepository) ListDeliveries(ctx context.Context, webhookID uuid.UUID, limit int) ([]*domain.WebhookDelivery, error) {
	args := m.Called(ctx, webhookID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.WebhookDelivery), args.Error(1)
}
