package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// OverdueActivityRef carries the minimal fields the automation worker needs
// to fire activity_overdue triggers.
type OverdueActivityRef struct {
	ActivityID uuid.UUID
	OrgID      uuid.UUID
	OwnerID    uuid.UUID
}

// AutomationRepository defines the persistence contract for workflow automations.
type AutomationRepository interface {
	// Automations
	CreateAutomation(ctx context.Context, a *domain.Automation) (*domain.Automation, error)
	GetAutomation(ctx context.Context, id uuid.UUID) (*domain.Automation, error)
	ListAutomations(ctx context.Context, filter domain.AutomationFilter) ([]*domain.Automation, int, error)
	UpdateAutomation(ctx context.Context, id uuid.UUID, req domain.UpdateAutomationRequest) (*domain.Automation, error)
	DeleteAutomation(ctx context.Context, id uuid.UUID) error

	// Runs
	CreateRun(ctx context.Context, run *domain.AutomationRun) (*domain.AutomationRun, error)
	UpdateRun(ctx context.Context, id uuid.UUID, status domain.AutomationRunStatus, errMsg string) error
	ListRuns(ctx context.Context, filter domain.AutomationRunFilter) ([]*domain.AutomationRun, int, error)

	// Worker helpers

	// ListActiveByTrigger returns all active automations for an org and trigger type.
	ListActiveByTrigger(ctx context.Context, orgID uuid.UUID, triggerType domain.TriggerType) ([]*domain.Automation, error)

	// OverdueActivityIDs returns activities with due_date <= cutoff and
	// completed_at IS NULL that have no existing automation_run recorded.
	OverdueActivityIDs(ctx context.Context, cutoff time.Time, limit int) ([]OverdueActivityRef, error)
}
