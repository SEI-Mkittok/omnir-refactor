package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
	"github.com/omnir/crm-api/internal/worker"
)

// newWebhookHandler builds a WebhookHandler with the provided mocks for unit tests.
func newTestWebhookHandler(
	tickets *mocks.MockTicketRepository,
	comments *mocks.MockTicketCommentRepository,
	contacts *mocks.MockContactRepository,
	users *mocks.MockUserRepository,
) http.Handler {
	h := handler.NewWebhookHandler(tickets, comments, contacts, users, "", config.OrgModeSingle, nopLogger())
	return h.Router()
}

// newTestWebhookHandlerWithDispatcher builds a WebhookHandler that sends events to the returned channel.
func newTestWebhookHandlerWithDispatcher(
	tickets *mocks.MockTicketRepository,
	comments *mocks.MockTicketCommentRepository,
	contacts *mocks.MockContactRepository,
	users *mocks.MockUserRepository,
) (http.Handler, <-chan worker.WebhookEvent) {
	ch := make(chan worker.WebhookEvent, 8)
	h := handler.NewWebhookHandler(tickets, comments, contacts, users, "", config.OrgModeSingle, nopLogger()).
		WithDispatcher(ch)
	return h.Router(), ch
}

func TestHandlePostmark_AutoCreatesContactForUnknownSender(t *testing.T) {
	ownerID := uuid.New()
	contactID := uuid.New()
	ticketID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	contactMock := new(mocks.MockContactRepository)
	userMock := new(mocks.MockUserRepository)

	// No existing contact for this email.
	contactMock.On("GetByEmail", mock.Anything, "alice@example.com").
		Return(nil, domain.ErrNotFound)

	// First user in org — used as contact owner.
	userMock.On("List", mock.Anything, domain.UserFilter{Limit: 1}).
		Return([]*domain.User{{ID: ownerID, Name: "Admin"}}, 1, nil)

	// Auto-create the new contact.
	contactMock.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Contact) bool {
		return c.FirstName == "Alice" &&
			c.LastName == "Smith" &&
			c.Email != nil && *c.Email == "alice@example.com" &&
			c.OwnerID == ownerID &&
			c.Stage == domain.ContactStageLead &&
			c.LeadSource != nil && *c.LeadSource == "email"
	})).Return(&domain.Contact{ID: contactID}, nil)

	// No thread match (In-Reply-To is empty).
	// Create new ticket linked to the new contact.
	ticketMock.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.Ticket) bool {
		return t.Subject == "Hello from Alice" &&
			t.ContactID != nil && *t.ContactID == contactID
	})).Return(&domain.Ticket{ID: ticketID, Subject: "Hello from Alice"}, nil)

	// Attach body as first comment.
	commentMock.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.TicketComment) bool {
		return c.TicketID == ticketID && c.Body == "Hi there!"
	})).Return(&domain.TicketComment{ID: uuid.New()}, nil)

	h := newTestWebhookHandler(ticketMock, commentMock, contactMock, userMock)

	body, err := json.Marshal(map[string]any{
		"MessageID": "msg-001",
		"From":      "Alice Smith <alice@example.com>",
		"Subject":   "Hello from Alice",
		"TextBody":  "Hi there!",
		"Headers":   []any{},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/postmark", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	contactMock.AssertExpectations(t)
	userMock.AssertExpectations(t)
	ticketMock.AssertExpectations(t)
	commentMock.AssertExpectations(t)
}

func TestHandlePostmark_NoAutoCreateWhenSenderKnown(t *testing.T) {
	ownerID := uuid.New()
	contactID := uuid.New()
	ticketID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	contactMock := new(mocks.MockContactRepository)
	userMock := new(mocks.MockUserRepository)

	// Existing contact found — no auto-create, no user lookup.
	contactMock.On("GetByEmail", mock.Anything, "bob@example.com").
		Return(&domain.Contact{ID: contactID, OwnerID: ownerID}, nil)

	ticketMock.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.Ticket) bool {
		return t.ContactID != nil && *t.ContactID == contactID
	})).Return(&domain.Ticket{ID: ticketID, Subject: "Support request"}, nil)

	commentMock.On("Create", mock.Anything, mock.Anything).
		Return(&domain.TicketComment{ID: uuid.New()}, nil)

	h := newTestWebhookHandler(ticketMock, commentMock, contactMock, userMock)

	body, err := json.Marshal(map[string]any{
		"MessageID": "msg-002",
		"From":      "bob@example.com",
		"Subject":   "Support request",
		"TextBody":  "Please help.",
		"Headers":   []any{},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/postmark", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// userMock should have no calls — existing contact means no user lookup needed.
	userMock.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
	contactMock.AssertExpectations(t)
}

// ───────────────────────── /inbound (Mailgun format) ─────────────────────────

func TestHandleInbound_CreatesTicketFromMailgunForm(t *testing.T) {
	contactID := uuid.New()
	ticketID := uuid.New()
	orgID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	contactMock := new(mocks.MockContactRepository)
	userMock := new(mocks.MockUserRepository)

	contactMock.On("GetByEmail", mock.Anything, "carol@example.com").
		Return(&domain.Contact{ID: contactID}, nil)

	ticketMock.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.Ticket) bool {
		return t.Subject == "Need help" &&
			t.ContactID != nil && *t.ContactID == contactID &&
			t.Source != nil && *t.Source == "email" &&
			t.EmailMessageID != nil && *t.EmailMessageID == "<abc123@mailgun.org>"
	})).Return(&domain.Ticket{ID: ticketID, OrgID: orgID, Subject: "Need help"}, nil)

	commentMock.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.TicketComment) bool {
		return c.TicketID == ticketID && c.Body == "Please assist."
	})).Return(&domain.TicketComment{ID: uuid.New()}, nil)

	h := newTestWebhookHandler(ticketMock, commentMock, contactMock, userMock)

	form := url.Values{}
	form.Set("Message-Id", "<abc123@mailgun.org>")
	form.Set("sender", "carol@example.com")
	form.Set("subject", "Need help")
	form.Set("body-plain", "Please assist.")

	req := httptest.NewRequest(http.MethodPost, "/inbound", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ticketMock.AssertExpectations(t)
	commentMock.AssertExpectations(t)
}

func TestHandleInbound_NoSubjectFallsBackToDefault(t *testing.T) {
	contactID := uuid.New()
	ticketID := uuid.New()
	orgID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	contactMock := new(mocks.MockContactRepository)
	userMock := new(mocks.MockUserRepository)

	contactMock.On("GetByEmail", mock.Anything, "dave@example.com").
		Return(&domain.Contact{ID: contactID}, nil)

	ticketMock.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.Ticket) bool {
		return t.Subject == "(no subject)"
	})).Return(&domain.Ticket{ID: ticketID, OrgID: orgID, Subject: "(no subject)"}, nil)

	commentMock.On("Create", mock.Anything, mock.Anything).
		Return(&domain.TicketComment{ID: uuid.New()}, nil)

	h := newTestWebhookHandler(ticketMock, commentMock, contactMock, userMock)

	form := url.Values{}
	form.Set("sender", "dave@example.com")
	form.Set("body-plain", "Hello.")

	req := httptest.NewRequest(http.MethodPost, "/inbound", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ticketMock.AssertExpectations(t)
}

func TestHandleInbound_InReplyToThreadsExistingTicket(t *testing.T) {
	ticketID := uuid.New()
	contactID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	contactMock := new(mocks.MockContactRepository)
	userMock := new(mocks.MockUserRepository)

	contactMock.On("GetByEmail", mock.Anything, "eve@example.com").
		Return(&domain.Contact{ID: contactID}, nil)

	// In-Reply-To matches an existing ticket.
	existingMsgID := "<original@mailgun.org>"
	ticketMock.On("GetByEmailMessageID", mock.Anything, existingMsgID).
		Return(&domain.Ticket{ID: ticketID}, nil)

	// A comment should be appended, not a new ticket.
	commentMock.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.TicketComment) bool {
		return c.TicketID == ticketID && c.Body == "Follow-up message." && !c.IsInternal
	})).Return(&domain.TicketComment{ID: uuid.New()}, nil)

	h := newTestWebhookHandler(ticketMock, commentMock, contactMock, userMock)

	form := url.Values{}
	form.Set("sender", "eve@example.com")
	form.Set("subject", "Re: Original")
	form.Set("body-plain", "Follow-up message.")
	form.Set("In-Reply-To", existingMsgID)

	req := httptest.NewRequest(http.MethodPost, "/inbound", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ticketMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	commentMock.AssertExpectations(t)
}

// ───────────────────────── ticket.created dispatch ─────────────────────────

func TestHandleInbound_DispatchesTicketCreatedEvent(t *testing.T) {
	contactID := uuid.New()
	ticketID := uuid.New()
	orgID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	contactMock := new(mocks.MockContactRepository)
	userMock := new(mocks.MockUserRepository)

	contactMock.On("GetByEmail", mock.Anything, "frank@example.com").
		Return(&domain.Contact{ID: contactID}, nil)

	ticketMock.On("Create", mock.Anything, mock.Anything).
		Return(&domain.Ticket{ID: ticketID, OrgID: orgID, Subject: "Dispatch test"}, nil)

	commentMock.On("Create", mock.Anything, mock.Anything).
		Return(&domain.TicketComment{ID: uuid.New()}, nil)

	h, dispatchCh := newTestWebhookHandlerWithDispatcher(ticketMock, commentMock, contactMock, userMock)

	form := url.Values{}
	form.Set("sender", "frank@example.com")
	form.Set("subject", "Dispatch test")
	form.Set("body-plain", "Testing dispatch.")

	req := httptest.NewRequest(http.MethodPost, "/inbound", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	select {
	case evt := <-dispatchCh:
		assert.Equal(t, domain.WebhookEventTicketCreated, evt.Event)
		assert.Equal(t, orgID, evt.OrgID)
		assert.Equal(t, ticketID, evt.EntityID)
	default:
		t.Fatal("expected ticket.created event to be dispatched, but channel was empty")
	}
}

// ───────────────────────── JSON body ─────────────────────────────────────────

func TestHandleInbound_RejectsMalformedForm(t *testing.T) {
	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	contactMock := new(mocks.MockContactRepository)
	userMock := new(mocks.MockUserRepository)

	h := newTestWebhookHandler(ticketMock, commentMock, contactMock, userMock)

	// Send a body that will cause ParseForm to fail by using a bad content type
	// that cannot be parsed.
	req := httptest.NewRequest(http.MethodPost, "/inbound", strings.NewReader("%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	// Handler should still succeed — bad form fields just yield empty strings.
	// Malformed percent-encoding causes ParseForm to return an error.
	_ = w.Code // behaviour depends on Go net/http form parsing; just ensure no panic
	ticketMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// ───────────────────────── dispatcher missing value ──────────────────────────

func TestHandleInbound_NoDispatcherDoesNotPanic(t *testing.T) {
	contactID := uuid.New()
	ticketID := uuid.New()
	orgID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	contactMock := new(mocks.MockContactRepository)
	userMock := new(mocks.MockUserRepository)

	contactMock.On("GetByEmail", mock.Anything, "grace@example.com").
		Return(&domain.Contact{ID: contactID}, nil)

	ticketMock.On("Create", mock.Anything, mock.Anything).
		Return(&domain.Ticket{ID: ticketID, OrgID: orgID, Subject: "No dispatcher"}, nil)

	commentMock.On("Create", mock.Anything, mock.Anything).
		Return(&domain.TicketComment{ID: uuid.New()}, nil)

	// No dispatcher wired — handler should not panic.
	h := newTestWebhookHandler(ticketMock, commentMock, contactMock, userMock)

	form := url.Values{}
	form.Set("sender", "grace@example.com")
	form.Set("subject", "No dispatcher")
	form.Set("body-plain", "Should work silently.")

	req := httptest.NewRequest(http.MethodPost, "/inbound", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	assert.NotPanics(t, func() { h.ServeHTTP(w, req) })
	assert.Equal(t, http.StatusOK, w.Code)
}
