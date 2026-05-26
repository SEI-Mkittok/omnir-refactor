package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

func TestSchedulerRegistryTrackRunWrapsFunction(t *testing.T) {
	registry := NewSchedulerRegistry()
	registry.Register("mail", "Mail sync", time.Minute, true)

	err := registry.TrackRun("mail", time.Minute, func() error { return errors.New("boom") })

	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected wrapped error, got %v", err)
	}
	jobs := registry.Jobs()
	if jobs[0].LastResult == nil || *jobs[0].LastResult != "failed" {
		t.Fatalf("expected failed job result, got %#v", jobs[0].LastResult)
	}
}

func TestSchedulerRegistryTracksRuns(t *testing.T) {
	registry := NewSchedulerRegistry()
	registry.Register("mail", "Mail sync", time.Minute, true)

	runID := registry.Start("mail")
	registry.Finish("mail", runID, nil, time.Minute)

	jobs := registry.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("expected one job, got %d", len(jobs))
	}
	if jobs[0].LastResult == nil || *jobs[0].LastResult != "succeeded" {
		t.Fatalf("expected succeeded result, got %#v", jobs[0].LastResult)
	}
	runs := registry.Runs("mail")
	if len(runs) != 1 || runs[0].Status != "succeeded" {
		t.Fatalf("expected succeeded run, got %#v", runs)
	}
}

func TestSchedulerRegistryTracksErrors(t *testing.T) {
	registry := NewSchedulerRegistry()
	registry.Register("mail", "Mail sync", time.Minute, true)

	runID := registry.Start("mail")
	registry.Finish("mail", runID, errors.New("boom"), time.Minute)

	jobs := registry.Jobs()
	if jobs[0].ErrorText == nil || *jobs[0].ErrorText != "boom" {
		t.Fatalf("expected job error text, got %#v", jobs[0].ErrorText)
	}
}

func TestReminderWorkerRecordsSchedulerSuccess(t *testing.T) {
	registry := NewSchedulerRegistry()
	registry.Register("activity-reminders", "Activity reminders", time.Millisecond, true)
	repo := &fakeReminderNotificationRepo{}
	worker := NewReminderWorker(repo, time.Millisecond, discardLogger()).WithScheduler(registry, "activity-reminders")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)

	waitForSchedulerResult(t, registry, "activity-reminders", "succeeded")
}

func TestReminderWorkerRecordsSchedulerFailure(t *testing.T) {
	registry := NewSchedulerRegistry()
	registry.Register("activity-reminders", "Activity reminders", time.Millisecond, true)
	repo := &fakeReminderNotificationRepo{err: errors.New("reminder failure")}
	worker := NewReminderWorker(repo, time.Millisecond, discardLogger()).WithScheduler(registry, "activity-reminders")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)

	waitForSchedulerResult(t, registry, "activity-reminders", "failed")
	jobs := registry.Jobs()
	if jobs[0].ErrorText == nil || *jobs[0].ErrorText != "reminder failure" {
		t.Fatalf("expected reminder failure error text, got %#v", jobs[0].ErrorText)
	}
}

func waitForSchedulerResult(t *testing.T, registry *SchedulerRegistry, key string, want string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		for _, job := range registry.Jobs() {
			if job.Key == key && job.LastResult != nil && *job.LastResult == want {
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s result %q; jobs=%#v", key, want, registry.Jobs())
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type fakeReminderNotificationRepo struct {
	err error
}

func (r *fakeReminderNotificationRepo) Create(_ context.Context, notification *domain.Notification) (*domain.Notification, error) {
	return notification, nil
}

func (r *fakeReminderNotificationRepo) MarkRead(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (r *fakeReminderNotificationRepo) MarkAllRead(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (r *fakeReminderNotificationRepo) UnreadCount(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	return 0, nil
}

func (r *fakeReminderNotificationRepo) ListByUser(context.Context, domain.NotificationFilter) ([]*domain.Notification, error) {
	return nil, nil
}

func (r *fakeReminderNotificationRepo) GenerateReminders(context.Context) error {
	return r.err
}
