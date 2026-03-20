package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// SequenceRepository defines the persistence contract for email sequences.
type SequenceRepository interface {
	// Sequences
	CreateSequence(ctx context.Context, s *domain.EmailSequence, steps []domain.CreateStepRequest) (*domain.EmailSequence, error)
	GetSequence(ctx context.Context, id uuid.UUID) (*domain.EmailSequence, error)
	ListSequences(ctx context.Context, filter domain.SequenceFilter) ([]*domain.EmailSequence, int, error)
	UpdateSequence(ctx context.Context, id uuid.UUID, req domain.UpdateSequenceRequest) (*domain.EmailSequence, error)
	DeleteSequence(ctx context.Context, id uuid.UUID) error

	// Enrollments
	Enroll(ctx context.Context, sequenceID uuid.UUID, req domain.EnrollRequest) (int, error)
	ListEnrollments(ctx context.Context, sequenceID uuid.UUID) ([]*domain.SequenceEnrollment, error)
	UpdateEnrollmentStatus(ctx context.Context, enrollmentID uuid.UUID, status domain.EnrollmentStatus) error

	// Analytics
	GetAnalytics(ctx context.Context, sequenceID uuid.UUID) (*domain.SequenceAnalytics, error)

	// Execution engine — no org scoping (used by background worker)
	PendingEnrollments(ctx context.Context) ([]*domain.SequenceEnrollment, error)
	AdvanceEnrollment(ctx context.Context, id uuid.UUID, nextStep int, nextStepAt *time.Time) error
	CompleteEnrollment(ctx context.Context, id uuid.UUID) error
	RecordEvent(ctx context.Context, e *domain.SequenceEvent) error
}
