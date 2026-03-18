package domain_test

import (
	"testing"

	"github.com/omnir/crm-api/internal/domain"
)

func TestTicketStatus_IsValid(t *testing.T) {
	valid := []domain.TicketStatus{
		domain.TicketStatusOpen,
		domain.TicketStatusInProgress,
		domain.TicketStatusPending,
		domain.TicketStatusResolved,
		domain.TicketStatusClosed,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected status %q to be valid", s)
		}
	}

	if domain.TicketStatus("bogus").IsValid() {
		t.Error("expected status 'bogus' to be invalid")
	}
}

func TestTicketPriority_IsValid(t *testing.T) {
	valid := []domain.TicketPriority{
		domain.TicketPriorityLow,
		domain.TicketPriorityMedium,
		domain.TicketPriorityHigh,
		domain.TicketPriorityCritical,
	}
	for _, p := range valid {
		if !p.IsValid() {
			t.Errorf("expected priority %q to be valid", p)
		}
	}

	if domain.TicketPriority("bogus").IsValid() {
		t.Error("expected priority 'bogus' to be invalid")
	}
}
