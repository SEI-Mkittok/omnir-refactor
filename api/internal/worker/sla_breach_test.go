package worker

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
)

type fakeSLAInstanceRepo struct {
	breached   []*domain.SLAInstance
	warnings   []*domain.SLAInstance
	recipients map[uuid.UUID][]uuid.UUID
	warned     []uuid.UUID
}

func (f *fakeSLAInstanceRepo) Create(_ context.Context, inst *domain.SLAInstance) (*domain.SLAInstance, error) {
	return inst, nil
}

func (f *fakeSLAInstanceRepo) GetByID(context.Context, uuid.UUID) (*domain.SLAInstance, error) {
	return nil, domain.ErrNotFound
}

func (f *fakeSLAInstanceRepo) List(context.Context, domain.SLAInstanceFilter) ([]*domain.SLAInstance, error) {
	return nil, nil
}

func (f *fakeSLAInstanceRepo) MarkResponded(context.Context, uuid.UUID, domain.SLAEntityType, time.Time) error {
	return nil
}

func (f *fakeSLAInstanceRepo) MarkResolved(context.Context, uuid.UUID, domain.SLAEntityType, time.Time) error {
	return nil
}

func (f *fakeSLAInstanceRepo) ScanBreaches(context.Context) ([]*domain.SLAInstance, error) {
	return f.breached, nil
}

func (f *fakeSLAInstanceRepo) ScanWarnings(context.Context) ([]*domain.SLAInstance, error) {
	return f.warnings, nil
}

func (f *fakeSLAInstanceRepo) MarkWarned(_ context.Context, id uuid.UUID, _ time.Time) error {
	f.warned = append(f.warned, id)
	return nil
}

func (f *fakeSLAInstanceRepo) Dashboard(context.Context) (*domain.SLADashboard, error) {
	return &domain.SLADashboard{}, nil
}

func (f *fakeSLAInstanceRepo) NotificationRecipients(_ context.Context, inst *domain.SLAInstance) ([]uuid.UUID, error) {
	return f.recipients[inst.ID], nil
}

type fakeNotificationRepo struct {
	created []*domain.Notification
}

func (f *fakeNotificationRepo) Create(_ context.Context, n *domain.Notification) (*domain.Notification, error) {
	f.created = append(f.created, n)
	return n, nil
}

func (f *fakeNotificationRepo) MarkRead(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (f *fakeNotificationRepo) MarkAllRead(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (f *fakeNotificationRepo) UnreadCount(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	return 0, nil
}

func (f *fakeNotificationRepo) ListByUser(context.Context, domain.NotificationFilter) ([]*domain.Notification, error) {
	return nil, nil
}

func (f *fakeNotificationRepo) GenerateReminders(context.Context) error {
	return nil
}

func TestSLABreachWorkerCreatesNotificationsForNewBreaches(t *testing.T) {
	orgID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()
	inst := &domain.SLAInstance{
		ID:         uuid.New(),
		OrgID:      orgID,
		EntityID:   uuid.New(),
		EntityType: domain.SLAEntityTypeTicket,
	}
	instances := &fakeSLAInstanceRepo{
		breached: []*domain.SLAInstance{inst},
		recipients: map[uuid.UUID][]uuid.UUID{
			inst.ID: {userA, userB},
		},
	}
	notifications := &fakeNotificationRepo{}
	worker := NewSLABreachWorker(instances, notifications, time.Minute, slog.New(slog.NewTextHandler(io.Discard, nil)))

	worker.tick(context.Background())

	require.Len(t, notifications.created, 2)
	require.Equal(t, domain.NotificationKindSLABreached, notifications.created[0].Kind)
	require.Equal(t, orgID, notifications.created[0].OrgID)
	require.ElementsMatch(t, []uuid.UUID{userA, userB}, []uuid.UUID{
		notifications.created[0].UserID,
		notifications.created[1].UserID,
	})
	require.Empty(t, instances.warned)
}

func TestSLABreachWorkerMarksWarningOnlyAfterNotificationRecipients(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	inst := &domain.SLAInstance{
		ID:         uuid.New(),
		OrgID:      orgID,
		EntityID:   uuid.New(),
		EntityType: domain.SLAEntityTypeTicket,
	}
	instances := &fakeSLAInstanceRepo{
		warnings: []*domain.SLAInstance{inst},
		recipients: map[uuid.UUID][]uuid.UUID{
			inst.ID: {userID},
		},
	}
	notifications := &fakeNotificationRepo{}
	worker := NewSLABreachWorker(instances, notifications, time.Minute, slog.New(slog.NewTextHandler(io.Discard, nil)))

	worker.tick(context.Background())

	require.Len(t, notifications.created, 1)
	require.Equal(t, domain.NotificationKindSLAWarning, notifications.created[0].Kind)
	require.Equal(t, userID, notifications.created[0].UserID)
	require.Equal(t, []uuid.UUID{inst.ID}, instances.warned)
}
