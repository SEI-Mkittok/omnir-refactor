package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
)

// MockTicketRepository is a testify mock implementing repository.TicketRepository.
type MockTicketRepository struct {
	mock.Mock
}

func (m *MockTicketRepository) Create(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Ticket), args.Error(1)
}

func (m *MockTicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Ticket), args.Error(1)
}

func (m *MockTicketRepository) GetDetailByID(ctx context.Context, id uuid.UUID) (*domain.TicketDetail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TicketDetail), args.Error(1)
}

func (m *MockTicketRepository) GetByEmailMessageID(ctx context.Context, messageID string) (*domain.Ticket, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Ticket), args.Error(1)
}

func (m *MockTicketRepository) Update(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Ticket), args.Error(1)
}

func (m *MockTicketRepository) UpdateContact(ctx context.Context, id uuid.UUID, contactID *uuid.UUID) (*domain.Ticket, error) {
	args := m.Called(ctx, id, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Ticket), args.Error(1)
}

func (m *MockTicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockTicketRepository) List(ctx context.Context, filter domain.TicketFilter) ([]*domain.Ticket, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Ticket), args.Int(1), args.Error(2)
}

// MockTicketCommentRepository is a testify mock implementing repository.TicketCommentRepository.
type MockTicketCommentRepository struct {
	mock.Mock
}

func (m *MockTicketCommentRepository) Create(ctx context.Context, c *domain.TicketComment) (*domain.TicketComment, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TicketComment), args.Error(1)
}

func (m *MockTicketCommentRepository) List(ctx context.Context, filter domain.TicketCommentFilter) ([]*domain.TicketComment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.TicketComment), args.Error(1)
}

func (m *MockTicketCommentRepository) Delete(ctx context.Context, id, ticketID uuid.UUID) error {
	return m.Called(ctx, id, ticketID).Error(0)
}

// MockTicketAttachmentRepository is a testify mock implementing repository.TicketAttachmentRepository.
type MockTicketAttachmentRepository struct {
	mock.Mock
}

func (m *MockTicketAttachmentRepository) Create(ctx context.Context, a *domain.TicketAttachment) (*domain.TicketAttachment, error) {
	args := m.Called(ctx, a)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TicketAttachment), args.Error(1)
}

func (m *MockTicketAttachmentRepository) GetByID(ctx context.Context, id, ticketID uuid.UUID) (*domain.TicketAttachment, error) {
	args := m.Called(ctx, id, ticketID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TicketAttachment), args.Error(1)
}

func (m *MockTicketAttachmentRepository) List(ctx context.Context, ticketID uuid.UUID) ([]*domain.TicketAttachment, error) {
	args := m.Called(ctx, ticketID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.TicketAttachment), args.Error(1)
}

func (m *MockTicketAttachmentRepository) Delete(ctx context.Context, id, ticketID uuid.UUID) error {
	return m.Called(ctx, id, ticketID).Error(0)
}

// MockNotificationPrefRepository is a testify mock implementing repository.NotificationPrefRepository.
type MockNotificationPrefRepository struct {
	mock.Mock
}

func (m *MockNotificationPrefRepository) GetByUser(ctx context.Context, userID, orgID uuid.UUID) (*domain.UserNotificationPref, error) {
	args := m.Called(ctx, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserNotificationPref), args.Error(1)
}

func (m *MockNotificationPrefRepository) Upsert(ctx context.Context, pref *domain.UserNotificationPref) (*domain.UserNotificationPref, error) {
	args := m.Called(ctx, pref)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserNotificationPref), args.Error(1)
}
