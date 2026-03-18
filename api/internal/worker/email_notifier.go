package worker

import (
	"context"
	"log/slog"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
)

const emailQueueSize = 256

// EmailNotifier receives EmailJob items from a channel and delivers them via Mailer.
// It runs a fixed pool of worker goroutines so delivery never blocks the hot path.
type EmailNotifier struct {
	queue  chan domain.EmailJob
	mailer *email.Mailer
	logger *slog.Logger
}

// NewEmailNotifier creates the notifier. Call Start to begin processing.
func NewEmailNotifier(mailer *email.Mailer, logger *slog.Logger) *EmailNotifier {
	return &EmailNotifier{
		queue:  make(chan domain.EmailJob, emailQueueSize),
		mailer: mailer,
		logger: logger,
	}
}

// Enqueue adds a job to the delivery queue without blocking.
// If the queue is full the job is dropped and a warning is logged.
func (n *EmailNotifier) Enqueue(job domain.EmailJob) {
	select {
	case n.queue <- job:
	default:
		n.logger.Warn("email queue full, dropping notification",
			"kind", job.Kind,
			"to", job.ToEmail,
			"ticket", job.TicketID,
		)
	}
}

// Start launches workerCount goroutines that drain the queue until ctx is cancelled.
func (n *EmailNotifier) Start(ctx context.Context, workerCount int) {
	if workerCount <= 0 {
		workerCount = 3
	}
	for i := 0; i < workerCount; i++ {
		go n.run(ctx)
	}
}

func (n *EmailNotifier) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-n.queue:
			n.deliver(job)
		}
	}
}

func (n *EmailNotifier) deliver(job domain.EmailJob) {
	var err error
	switch job.Kind {
	case domain.EmailEventAssigned:
		err = n.mailer.SendAssigned(job.ToEmail, job.ToName, job.TicketID, job.TicketSubject)
	case domain.EmailEventResolved:
		err = n.mailer.SendResolved(job.ToEmail, job.ToName, job.TicketID, job.TicketSubject, "resolved")
	case domain.EmailEventClosed:
		err = n.mailer.SendResolved(job.ToEmail, job.ToName, job.TicketID, job.TicketSubject, "closed")
	default:
		n.logger.Warn("email_notifier: unknown event kind", "kind", job.Kind)
		return
	}
	if err != nil {
		n.logger.Error("email_notifier: delivery failed",
			"kind", job.Kind,
			"to", job.ToEmail,
			"ticket", job.TicketID,
			"err", err,
		)
	} else {
		n.logger.Debug("email_notifier: sent",
			"kind", job.Kind,
			"to", job.ToEmail,
			"ticket", job.TicketID,
		)
	}
}
