package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockAuditLogRepository is a testify mock implementing repository.AuditLogRepository.
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Append(ctx context.Context, entry domain.AuditEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockAuditLogRepository) List(ctx context.Context, filter domain.AuditLogFilter) ([]*domain.AuditLog, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.AuditLog), args.Int(1), args.Error(2)
}
