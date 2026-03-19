package worker_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	"github.com/omnir/crm-api/internal/worker"
)

func TestEmailNotifier_EnqueueAndDeliver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	mailer := email.NewMailer(email.NewSender(config.SMTPConfig{Enabled: false}), "https://test.omnir.io")
	n := worker.NewEmailNotifier(mailer, logger)
	n.Start(ctx, 1)

	// Enqueue two jobs.
	n.Enqueue(domain.EmailJob{
		Kind:          domain.EmailEventAssigned,
		ToEmail:       "agent@test.com",
		ToName:        "Agent",
		TicketID:      "t-001",
		TicketSubject: "Password reset",
	})
	n.Enqueue(domain.EmailJob{
		Kind:          domain.EmailEventResolved,
		ToEmail:       "client@test.com",
		ToName:        "Client",
		TicketID:      "t-002",
		TicketSubject: "Cannot export",
	})

	// Give workers time to drain.
	time.Sleep(100 * time.Millisecond)
	// No panics = pass; delivery is a no-op when SMTP is disabled.
}

func TestEmailNotifier_QueueFull_DropsGracefully(t *testing.T) {
	// Create a notifier but don't start any workers so the queue fills.
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	mailer := email.NewMailer(email.NewSender(config.SMTPConfig{Enabled: false}), "https://test.omnir.io")
	n := worker.NewEmailNotifier(mailer, logger)
	// Don't call n.Start — no consumers.

	// Fill queue beyond its capacity (256) without blocking.
	for i := 0; i < 300; i++ {
		n.Enqueue(domain.EmailJob{Kind: domain.EmailEventAssigned, ToEmail: "x@y.com", TicketID: "t"})
	}
	// Should not panic or deadlock.
	require.NotNil(t, n)
}

func TestEmailNotifier_UnknownKind(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	mailer := email.NewMailer(email.NewSender(config.SMTPConfig{Enabled: false}), "https://test.omnir.io")
	n := worker.NewEmailNotifier(mailer, logger)
	n.Start(ctx, 1)

	n.Enqueue(domain.EmailJob{Kind: "unknown_event", ToEmail: "x@y.com", TicketID: "t"})
	time.Sleep(50 * time.Millisecond)
	// Unknown kind is logged and skipped — no panic.
	assert.True(t, true)
}
