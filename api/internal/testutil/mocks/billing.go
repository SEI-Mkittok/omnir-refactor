package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockBillingRepository is a testify mock implementing repository.BillingRepository.
type MockBillingRepository struct {
	mock.Mock
}

func (m *MockBillingRepository) GetOrCreatePlan(ctx context.Context, orgID uuid.UUID) (*domain.OrgPlanRecord, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrgPlanRecord), args.Error(1)
}

func (m *MockBillingRepository) UpsertPlan(ctx context.Context, orgID uuid.UUID, patch domain.OrgPlanPatch) (*domain.OrgPlanRecord, error) {
	args := m.Called(ctx, orgID, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrgPlanRecord), args.Error(1)
}

func (m *MockBillingRepository) GetPlanByStripeSubscriptionID(ctx context.Context, subID string) (*domain.OrgPlanRecord, error) {
	args := m.Called(ctx, subID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrgPlanRecord), args.Error(1)
}

func (m *MockBillingRepository) GetPlanByStripeCustomerID(ctx context.Context, customerID string) (*domain.OrgPlanRecord, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrgPlanRecord), args.Error(1)
}

func (m *MockBillingRepository) UpsertInvoice(ctx context.Context, inv *domain.Invoice) (*domain.Invoice, error) {
	args := m.Called(ctx, inv)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *MockBillingRepository) ListInvoices(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.Invoice, int, error) {
	args := m.Called(ctx, orgID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Invoice), args.Int(1), args.Error(2)
}

func (m *MockBillingRepository) GetUsageStats(ctx context.Context, orgID uuid.UUID) (int, int, error) {
	args := m.Called(ctx, orgID)
	return args.Int(0), args.Int(1), args.Error(2)
}
