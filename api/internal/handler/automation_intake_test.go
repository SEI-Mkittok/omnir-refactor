package handler

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

func TestPreviewWebformMapsRequiredFields(t *testing.T) {
	form := &domain.Webform{
		TargetModule: domain.WebformTargetLead,
		Fields: []domain.WebformField{
			{Key: "email", Label: "Email", Required: true, TargetField: "email"},
			{Key: "company", Label: "Company", TargetField: "company"},
		},
	}

	preview := previewWebform(form, map[string]interface{}{"email": "ada@example.com", "company": "Analytical Engines"})

	if len(preview.MissingFields) != 0 {
		t.Fatalf("expected no missing fields, got %v", preview.MissingFields)
	}
	if preview.MappedFields["email"] != "ada@example.com" {
		t.Fatalf("expected mapped email, got %v", preview.MappedFields["email"])
	}
}

func TestMailConditionMatchesSubjectContains(t *testing.T) {
	msg := &domain.EmailInboxMessage{Subject: "Urgent billing issue"}
	condition := domain.MailConverterCondition{Field: "subject", Operator: "contains", Value: "billing"}

	if !mailConditionMatches(condition, msg) {
		t.Fatal("expected subject condition to match")
	}
}

func TestMailConverterCreateTicketOmitsEmptyEmailMessageID(t *testing.T) {
	tickets := &captureMailConverterTicketRepo{}
	h := &MailConverterHandler{tickets: tickets}
	msg := &domain.EmailInboxMessage{Subject: "Missing message id"}

	_, _, err := h.executeAction(context.Background(), &domain.MailConverterRule{}, msg, domain.MailConverterAction{Type: "create_ticket"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tickets.created == nil {
		t.Fatal("expected ticket to be created")
	}
	if tickets.created.EmailMessageID != nil {
		t.Fatalf("expected nil email_message_id, got %q", *tickets.created.EmailMessageID)
	}
}

func TestMailConverterCreateTicketPersistsNonEmptyEmailMessageID(t *testing.T) {
	tickets := &captureMailConverterTicketRepo{}
	h := &MailConverterHandler{tickets: tickets}
	msg := &domain.EmailInboxMessage{Subject: "Has message id", MessageID: "msg-123"}

	_, _, err := h.executeAction(context.Background(), &domain.MailConverterRule{}, msg, domain.MailConverterAction{Type: "create_ticket"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tickets.created == nil || tickets.created.EmailMessageID == nil || *tickets.created.EmailMessageID != "msg-123" {
		t.Fatalf("expected persisted message id, got %#v", tickets.created)
	}
}

type captureMailConverterTicketRepo struct {
	created *domain.Ticket
}

func (r *captureMailConverterTicketRepo) Create(_ context.Context, ticket *domain.Ticket) (*domain.Ticket, error) {
	ticket.ID = uuid.New()
	r.created = ticket
	return ticket, nil
}

func (r *captureMailConverterTicketRepo) GetByID(context.Context, uuid.UUID) (*domain.Ticket, error) {
	return nil, domain.ErrNotFound
}

func (r *captureMailConverterTicketRepo) CanAccess(context.Context, uuid.UUID, domain.SharingAccessLevel) (bool, error) {
	return false, nil
}

func (r *captureMailConverterTicketRepo) GetDetailByID(context.Context, uuid.UUID) (*domain.TicketDetail, error) {
	return nil, domain.ErrNotFound
}

func (r *captureMailConverterTicketRepo) GetByEmailMessageID(context.Context, string) (*domain.Ticket, error) {
	return nil, domain.ErrNotFound
}

func (r *captureMailConverterTicketRepo) Update(context.Context, uuid.UUID, domain.TicketPatch) (*domain.Ticket, error) {
	return nil, domain.ErrNotFound
}

func (r *captureMailConverterTicketRepo) UpdateContact(context.Context, uuid.UUID, *uuid.UUID) (*domain.Ticket, error) {
	return nil, domain.ErrNotFound
}

func (r *captureMailConverterTicketRepo) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r *captureMailConverterTicketRepo) List(context.Context, domain.TicketFilter) ([]*domain.Ticket, int, error) {
	return nil, 0, nil
}
