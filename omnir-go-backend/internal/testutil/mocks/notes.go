package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockNoteRepository is a testify mock implementing repository.NoteRepository.
type MockNoteRepository struct {
	mock.Mock
}

func (m *MockNoteRepository) Create(ctx context.Context, n *domain.Note) (*domain.Note, error) {
	args := m.Called(ctx, n)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNoteRepository) ListByEntity(ctx context.Context, filter domain.NoteFilter) ([]*domain.Note, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Note), args.Int(1), args.Error(2)
}
