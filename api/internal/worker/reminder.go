package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/omnir/crm-api/internal/repository"
)

// ReminderWorker checks for upcoming and overdue activities on a ticker and
// generates notification records for activity owners.
type ReminderWorker struct {
	repo     repository.NotificationRepository
	interval time.Duration
	logger   *slog.Logger
	stop     chan struct{}
}

// NewReminderWorker creates a worker that runs every interval.
func NewReminderWorker(repo repository.NotificationRepository, interval time.Duration, logger *slog.Logger) *ReminderWorker {
	return &ReminderWorker{
		repo:     repo,
		interval: interval,
		logger:   logger,
		stop:     make(chan struct{}),
	}
}

// Start runs the worker in a background goroutine until Stop is called or ctx is cancelled.
func (w *ReminderWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		w.logger.Info("reminder worker started", "interval", w.interval)
		for {
			select {
			case <-ticker.C:
				if err := w.repo.GenerateReminders(ctx); err != nil {
					w.logger.Error("reminder generation failed", "err", err)
				}
			case <-w.stop:
				w.logger.Info("reminder worker stopped")
				return
			case <-ctx.Done():
				w.logger.Info("reminder worker context cancelled")
				return
			}
		}
	}()
}

// Stop signals the worker to exit after the current tick finishes.
func (w *ReminderWorker) Stop() {
	close(w.stop)
}
