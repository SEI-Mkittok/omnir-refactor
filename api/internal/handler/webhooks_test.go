package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
