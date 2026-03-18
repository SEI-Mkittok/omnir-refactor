package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	"github.com/omnir/crm-api/internal/repository"
)

// SequenceWorker processes pending email sequence enrollments on a ticker.
type SequenceWorker struct {
	repo     repository.SequenceRepository
	mailer   *email.Mailer
	interval time.Duration
	log      *slog.Logger
	stop     chan struct{}
}

// NewSequenceWorker creates a SequenceWorker that runs every interval.
func NewSequenceWorker(repo repository.SequenceRepository, mailer *email.Mailer, interval time.Duration, log *slog.Logger) *SequenceWorker {
	return &SequenceWorker{
		repo:     repo,
		mailer:   mailer,
		interval: interval,
		log:      log,
		stop:     make(chan struct{}),
	}
}

// Start runs the worker in a background goroutine until Stop is called or ctx is cancelled.
func (w *SequenceWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		w.log.Info("sequence worker started", "interval", w.interval)
		for {
			select {
			case <-ticker.C:
				w.tick(ctx)
			case <-w.stop:
				w.log.Info("sequence worker stopped")
				return
			case <-ctx.Done():
				w.log.Info("sequence worker context cancelled")
				return
			}
		}
	}()
}

// Stop signals the worker to exit after the current tick finishes.
func (w *SequenceWorker) Stop() {
	close(w.stop)
}

func (w *SequenceWorker) tick(ctx context.Context) {
	enrollments, err := w.repo.PendingEnrollments(ctx)
	if err != nil {
		w.log.Error("sequence worker: failed to fetch pending enrollments", "err", err)
		return
	}

	for _, enrollment := range enrollments {
		if err := w.processEnrollment(ctx, enrollment); err != nil {
			w.log.Error("sequence worker: failed to process enrollment",
				"enrollment_id", enrollment.ID, "err", err)
		}
	}
}

func (w *SequenceWorker) processEnrollment(ctx context.Context, enrollment *domain.SequenceEnrollment) error {
	seq, err := w.repo.GetSequence(ctx, enrollment.SequenceID)
	if err != nil {
		return err
	}

	steps := seq.Steps
	currentIdx := enrollment.CurrentStep

	if currentIdx >= len(steps) {
		// No more steps — complete the enrollment.
		return w.repo.CompleteEnrollment(ctx, enrollment.ID)
	}

	step := steps[currentIdx]

	if step.Kind == domain.StepKindEmail {
		if enrollment.ContactEmail != "" && step.Subject != "" {
			if err := w.mailer.SendDirect(enrollment.ContactEmail, step.Subject, step.Body); err != nil {
				w.log.Warn("sequence worker: failed to send email",
					"enrollment_id", enrollment.ID,
					"contact_email", enrollment.ContactEmail,
					"err", err)
				// Don't abort — still advance so a bad address doesn't block the sequence.
			}
		}

		// Record a "sent" event.
		stepID := step.ID
		evt := &domain.SequenceEvent{
			SequenceID:   enrollment.SequenceID,
			StepID:       &stepID,
			EnrollmentID: enrollment.ID,
			ContactID:    enrollment.ContactID,
			OrgID:        enrollment.OrgID,
			Kind:         domain.SequenceEventSent,
		}
		if err := w.repo.RecordEvent(ctx, evt); err != nil {
			w.log.Warn("sequence worker: failed to record sent event",
				"enrollment_id", enrollment.ID, "err", err)
		}
	}

	// Advance to the next step.
	nextIdx := currentIdx + 1
	var nextStepAt *time.Time

	if nextIdx < len(steps) {
		next := steps[nextIdx]
		waitHours := 0
		if next.WaitDurationHours != nil {
			waitHours = *next.WaitDurationHours
		}
		t := time.Now().UTC().Add(time.Duration(waitHours) * time.Hour)
		nextStepAt = &t
	}

	if err := w.repo.AdvanceEnrollment(ctx, enrollment.ID, nextIdx, nextStepAt); err != nil {
		return err
	}

	// If there are no more steps after advancing, complete immediately.
	if nextIdx >= len(steps) {
		return w.repo.CompleteEnrollment(ctx, enrollment.ID)
	}

	return nil
}
