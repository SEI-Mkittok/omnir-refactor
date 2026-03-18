package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockReportsRepository is a testify mock implementing repository.ReportsRepository.
type MockReportsRepository struct {
	mock.Mock
}

func (m *MockReportsRepository) DealsByStage(ctx context.Context) ([]domain.DealStageMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DealStageMetric), args.Error(1)
}

func (m *MockReportsRepository) ContactsMonthly(ctx context.Context) ([]domain.ContactMonthlyMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ContactMonthlyMetric), args.Error(1)
}

func (m *MockReportsRepository) ActivitiesByType(ctx context.Context) ([]domain.ActivityTypeMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ActivityTypeMetric), args.Error(1)
}

func (m *MockReportsRepository) TicketMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.TicketReport, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TicketReport), args.Error(1)
}

func (m *MockReportsRepository) ContactMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.ContactReport, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ContactReport), args.Error(1)
}

func (m *MockReportsRepository) DealMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.DealReport, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DealReport), args.Error(1)
}

func (m *MockReportsRepository) LeadMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.LeadReport, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LeadReport), args.Error(1)
}

func (m *MockReportsRepository) PipelineFunnel(ctx context.Context, pipelineID *uuid.UUID, filter domain.ReportFilter) (*domain.PipelineFunnelReport, error) {
	args := m.Called(ctx, pipelineID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PipelineFunnelReport), args.Error(1)
}

func (m *MockReportsRepository) ConversionRates(ctx context.Context, filter domain.ReportFilter) (*domain.ConversionRatesReport, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ConversionRatesReport), args.Error(1)
}

func (m *MockReportsRepository) RevenueProjection(ctx context.Context, months int) (*domain.RevenueProjectionReport, error) {
	args := m.Called(ctx, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RevenueProjectionReport), args.Error(1)
}

func (m *MockReportsRepository) ActivitySummary(ctx context.Context, filter domain.ReportFilter) (*domain.ActivitySummaryReport, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ActivitySummaryReport), args.Error(1)
}
