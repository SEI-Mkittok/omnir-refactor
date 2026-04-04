package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockContactRepository is a testify mock implementing domain.ContactRepository.
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) Create(ctx context.Context, c *domain.Contact) (*domain.Contact, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) GetByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) Update(ctx context.Context, id uuid.UUID, patch domain.ContactPatch) (*domain.Contact, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactRepository) List(ctx context.Context, filter domain.ContactFilter) ([]*domain.Contact, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Contact), args.Int(1), args.Error(2)
}

func (m *MockContactRepository) UpdateLeadScore(ctx context.Context, id uuid.UUID, patch domain.LeadScorePatch) (*domain.Contact, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) ConvertLead(ctx context.Context, id, byUserID uuid.UUID, dealID *uuid.UUID) (*domain.Contact, error) {
	args := m.Called(ctx, id, byUserID, dealID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) ListLeadSources(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockContactRepository) SetEmailOptOut(ctx context.Context, contactID, orgID uuid.UUID) error {
	args := m.Called(ctx, contactID, orgID)
	return args.Error(0)
}

func (m *MockContactRepository) IncrementBounceCount(ctx context.Context, contactID, orgID uuid.UUID) error {
	args := m.Called(ctx, contactID, orgID)
	return args.Error(0)
}
