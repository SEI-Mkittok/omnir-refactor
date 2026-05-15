package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

// clientClaims returns Claims with role=client for use in portal tests.
func clientClaims(userID uuid.UUID) *auth.Claims {
	return &auth.Claims{UserID: userID, OrgID: uuid.New(), Role: string(domain.UserRoleClient)}
}

// agentClaims returns Claims with role=agent (should be blocked by portal middleware).
func agentClaims(userID uuid.UUID) *auth.Claims {
	return &auth.Claims{UserID: userID, OrgID: uuid.New(), Role: string(domain.UserRoleAgent)}
}

func newTestPortalHandler(
	tickets *mocks.MockTicketRepository,
	comments *mocks.MockTicketCommentRepository,
) http.Handler {
	h := handler.NewPortalHandler(tickets, comments)
	return h.Router()
}

// ───────────────────────── Role enforcement ─────────────────────────

func TestPortal_NonClientRoleIsForbidden(t *testing.T) {
	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	h := newTestPortalHandler(ticketMock, commentMock)

	req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
	req = withClaims(req, agentClaims(uuid.New()))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	ticketMock.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
}

func TestPortal_UnauthenticatedRequestIsUnauthorized(t *testing.T) {
	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	h := newTestPortalHandler(ticketMock, commentMock)

	// No claims injected.
	req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ───────────────────────── POST /portal/tickets ─────────────────────────

func TestPortal_CreateTicket_Success(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	ticketMock.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.Ticket) bool {
		return t.Subject == "My issue" &&
			t.SubmittedByUserID != nil && *t.SubmittedByUserID == userID
	})).Return(&domain.Ticket{ID: ticketID, Subject: "My issue", SubmittedByUserID: &userID}, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	body, err := json.Marshal(map[string]string{"subject": "My issue"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	ticketMock.AssertExpectations(t)
}

func TestPortal_CreateTicket_MissingSubjectReturns422(t *testing.T) {
	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	h := newTestPortalHandler(ticketMock, commentMock)

	body, _ := json.Marshal(map[string]string{"description": "no subject here"})
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, clientClaims(uuid.New()))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	ticketMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestPortal_CreateTicket_InvalidPriorityReturns422(t *testing.T) {
	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	h := newTestPortalHandler(ticketMock, commentMock)

	body, _ := json.Marshal(map[string]string{"subject": "Test", "priority": "urgent"})
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, clientClaims(uuid.New()))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ───────────────────────── GET /portal/tickets ─────────────────────────

func TestPortal_ListTickets_OnlyOwnTickets(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	// Filter must scope to this user's submitted tickets.
	ticketMock.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketFilter) bool {
		return f.SubmittedByUserID != nil && *f.SubmittedByUserID == userID
	})).Return([]*domain.Ticket{{ID: ticketID, Subject: "My ticket", SubmittedByUserID: &userID}}, 1, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	meta := resp["meta"].(map[string]any)
	assert.Equal(t, float64(1), meta["total"])
	ticketMock.AssertExpectations(t)
}

// ───────────────────────── GET /portal/tickets/{id} ─────────────────────────

func TestPortal_GetTicket_ReturnsTicketWithPublicComments(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()
	commentID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	ticketMock.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, Subject: "Help!", SubmittedByUserID: &userID}, nil)

	isInternal := false
	commentMock.On("List", mock.Anything, domain.TicketCommentFilter{
		TicketID:   ticketID,
		IsInternal: &isInternal,
	}).Return([]*domain.TicketComment{
		{ID: commentID, TicketID: ticketID, Body: "We're looking into it.", IsInternal: false},
	}, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	req := httptest.NewRequest(http.MethodGet, "/tickets/"+ticketID.String(), nil)
	req = withURLParam(req, "id", ticketID.String())
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, ticketID.String(), resp["id"])
	assert.Equal(t, "Help!", resp["subject"])
	comments, ok := resp["comments"].([]any)
	require.True(t, ok, "expected comments array in response")
	assert.Len(t, comments, 1)

	ticketMock.AssertExpectations(t)
	commentMock.AssertExpectations(t)
}

func TestPortal_GetTicket_OtherUserTicketReturns404(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	ticketID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	// Ticket belongs to a different user.
	ticketMock.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, Subject: "Not yours", SubmittedByUserID: &otherUserID}, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	req := httptest.NewRequest(http.MethodGet, "/tickets/"+ticketID.String(), nil)
	req = withURLParam(req, "id", ticketID.String())
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	commentMock.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
}

func TestPortal_GetTicket_InvalidIDReturns400(t *testing.T) {
	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)
	h := newTestPortalHandler(ticketMock, commentMock)

	req := httptest.NewRequest(http.MethodGet, "/tickets/not-a-uuid", nil)
	req = withURLParam(req, "id", "not-a-uuid")
	req = withClaims(req, clientClaims(uuid.New()))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPortal_GetTicket_NoSubmitterReturns404(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	// Ticket has no submitted_by_user_id (e.g. created by an agent).
	ticketMock.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, Subject: "Agent ticket", SubmittedByUserID: nil}, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	req := httptest.NewRequest(http.MethodGet, "/tickets/"+ticketID.String(), nil)
	req = withURLParam(req, "id", ticketID.String())
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPortal_ListComments_ReturnsPublicCommentsForOwnedTicket(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()
	commentID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	ticketMock.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, SubmittedByUserID: &userID}, nil)

	isInternal := false
	commentMock.On("List", mock.Anything, domain.TicketCommentFilter{
		TicketID:   ticketID,
		IsInternal: &isInternal,
	}).Return([]*domain.TicketComment{
		{ID: commentID, TicketID: ticketID, Body: "Public update", IsInternal: false},
	}, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	req := httptest.NewRequest(http.MethodGet, "/tickets/"+ticketID.String()+"/comments", nil)
	req = withURLParam(req, "id", ticketID.String())
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp, 1)
	assert.Equal(t, "Public update", resp[0]["body"])
	ticketMock.AssertExpectations(t)
	commentMock.AssertExpectations(t)
}

func TestPortal_ListComments_OtherUserTicketReturns404(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	ticketID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	ticketMock.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, SubmittedByUserID: &otherUserID}, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	req := httptest.NewRequest(http.MethodGet, "/tickets/"+ticketID.String()+"/comments", nil)
	req = withURLParam(req, "id", ticketID.String())
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	commentMock.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
}

// ───────────────────────── POST /portal/tickets/{id}/comments ─────────────────────────

func TestPortal_CreateComment_Success(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()
	commentID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	ticketMock.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, SubmittedByUserID: &userID}, nil)

	commentMock.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.TicketComment) bool {
		return c.TicketID == ticketID &&
			c.Body == "Thanks for the update!" &&
			!c.IsInternal &&
			c.AuthorID != nil && *c.AuthorID == userID
	})).Return(&domain.TicketComment{ID: commentID, Body: "Thanks for the update!"}, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	body, _ := json.Marshal(map[string]string{"body": "Thanks for the update!"})
	req := httptest.NewRequest(http.MethodPost, "/tickets/"+ticketID.String()+"/comments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", ticketID.String())
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	ticketMock.AssertExpectations(t)
	commentMock.AssertExpectations(t)
}

func TestPortal_CreateComment_OtherUserTicketReturns404(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	ticketID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	ticketMock.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, SubmittedByUserID: &otherUserID}, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	body, _ := json.Marshal(map[string]string{"body": "Trying to post"})
	req := httptest.NewRequest(http.MethodPost, "/tickets/"+ticketID.String()+"/comments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", ticketID.String())
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	commentMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestPortal_CreateComment_EmptyBodyReturns422(t *testing.T) {
	userID := uuid.New()
	ticketID := uuid.New()

	ticketMock := new(mocks.MockTicketRepository)
	commentMock := new(mocks.MockTicketCommentRepository)

	ticketMock.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, SubmittedByUserID: &userID}, nil)

	h := newTestPortalHandler(ticketMock, commentMock)

	body, _ := json.Marshal(map[string]string{"body": ""})
	req := httptest.NewRequest(http.MethodPost, "/tickets/"+ticketID.String()+"/comments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", ticketID.String())
	req = withClaims(req, clientClaims(userID))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	commentMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}
