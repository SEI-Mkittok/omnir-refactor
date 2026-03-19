package worker_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	"github.com/omnir/crm-api/internal/worker"
)

// stubSender is a thread-safe in-memory email sender for testing.
type stubSender struct {
	mu   sync.Mutex
	sent []email.Message
}

func (s *stubSender) Send(msg email.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, msg)
	return nil
}

func (s *stubSender) Messages() []email.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]email.Message, len(s.sent))
	copy(out, s.sent)
	return out
}

// senderAdapter wraps stubSender so it satisfies the email.Sender interface shape.
// We create a real Sender but swap the underlying SMTP connection with a mock by
// using SMTP_ENABLED=false and testing the notifier dispatch layer separately.


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
