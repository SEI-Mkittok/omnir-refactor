package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
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

func TestTicketHandler_List(t *testing.T) {
	accountID := uuid.New()
	ticket1 := &domain.Ticket{
		ID:      uuid.New(),
		Subject: "Test ticket 1",
		Status:  domain.TicketStatusOpen,
	}
	ticket2 := &domain.Ticket{
		ID:      uuid.New(),
		Subject: "Test ticket 2",
		Status:  domain.TicketStatusPending,
	}

	tests := []struct {
		name       string
		query      string
		setupMock  func(*mocks.MockTicketRepository)
		wantStatus int
		wantTotal  int
	}{
		{
			name:  "returns paginated tickets",
			query: "?page=1&per_page=10",
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketFilter) bool {
					return f.Page == 1 && f.Limit == 10
				})).Return([]*domain.Ticket{ticket1, ticket2}, 2, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  2,
		},
		{
			name:  "respects status filter",
			query: "?status=open",
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketFilter) bool {
					return f.Status != nil && *f.Status == domain.TicketStatusOpen
				})).Return([]*domain.Ticket{ticket1}, 1, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name:  "respects priority filter",
			query: "?priority=high",
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketFilter) bool {
					return f.Priority != nil && *f.Priority == domain.TicketPriorityHigh
				})).Return([]*domain.Ticket{ticket1}, 1, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name:  "respects assignee filter",
			query: "?assignee_id=" + ticket1.ID.String(),
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketFilter) bool {
					return f.AssigneeID != nil
				})).Return([]*domain.Ticket{ticket1}, 1, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name:  "respects account filter",
			query: "?account_id=" + accountID.String(),
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketFilter) bool {
					return f.AccountID != nil && *f.AccountID == accountID
				})).Return([]*domain.Ticket{ticket1}, 1, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name:  "uses default pagination when not specified",
			query: "",
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketFilter) bool {
					return f.Page == 1 && f.Limit == 50
				})).Return([]*domain.Ticket{ticket1, ticket2}, 2, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			tt.setupMock(mockTickets)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets"+tt.query, nil)
			w := httptest.NewRecorder()

			h.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				assert.Contains(t, resp, "data")
				assert.Contains(t, resp, "meta")
				meta := resp["meta"].(map[string]any)
				assert.Equal(t, float64(tt.wantTotal), meta["total"])
			}

			mockTickets.AssertExpectations(t)
		})
	}
}

func TestTicketHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockTicketRepository)
		wantStatus int
	}{
		{
			name: "creates ticket successfully",
			body: map[string]any{
				"subject":     "New ticket",
				"description": "Ticket description",
				"priority":    "high",
				"status":      "open",
			},
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Ticket")).
					Return(&domain.Ticket{
						ID:      uuid.New(),
						Subject: "New ticket",
						Status:  domain.TicketStatusOpen,
					}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "returns 422 for missing subject",
			body: map[string]any{
				"description": "No subject",
			},
			setupMock:  func(_ *mocks.MockTicketRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       nil,
			setupMock:  func(_ *mocks.MockTicketRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			tt.setupMock(mockTickets)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte("not-json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockTickets.AssertExpectations(t)
		})
	}
}

func TestTicketHandler_GetByID(t *testing.T) {
	ticketID := uuid.New()

	tests := []struct {
		name       string
		ticketID   string
		setupMock  func(*mocks.MockTicketRepository)
		wantStatus int
	}{
		{
			name:     "returns ticket by id",
			ticketID: ticketID.String(),
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("GetDetailByID", mock.Anything, ticketID).
					Return(&domain.TicketDetail{Ticket: domain.Ticket{ID: ticketID, Subject: "Test ticket"}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "returns 404 for unknown id",
			ticketID: uuid.New().String(),
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("GetDetailByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid uuid",
			ticketID:   "not-a-uuid",
			setupMock:  func(_ *mocks.MockTicketRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			tt.setupMock(mockTickets)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/"+tt.ticketID, nil)
			req = withURLParam(req, "id", tt.ticketID)
			w := httptest.NewRecorder()

			h.GetByID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockTickets.AssertExpectations(t)
		})
	}
}

func TestTicketHandler_Update(t *testing.T) {
	ticketID := uuid.New()

	tests := []struct {
		name       string
		ticketID   string
		body       map[string]any
		setupMock  func(*mocks.MockTicketRepository)
		wantStatus int
	}{
		{
			name:     "updates ticket successfully",
			ticketID: ticketID.String(),
			body: map[string]any{
				"status": "closed",
			},
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("Update", mock.Anything, ticketID, mock.AnythingOfType("domain.TicketPatch")).
					Return(&domain.Ticket{ID: ticketID, Status: domain.TicketStatusClosed}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "updates ticket tags",
			ticketID: ticketID.String(),
			body: map[string]any{
				"tags": []string{"vip", "renewal"},
			},
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("Update", mock.Anything, ticketID, mock.MatchedBy(func(p domain.TicketPatch) bool {
					return len(p.Tags) == 2 && p.Tags[0] == "vip" && p.Tags[1] == "renewal"
				})).Return(&domain.Ticket{ID: ticketID, Tags: []string{"vip", "renewal"}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid uuid",
			ticketID:   "not-a-uuid",
			body:       map[string]any{"status": "closed"},
			setupMock:  func(_ *mocks.MockTicketRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			ticketID:   ticketID.String(),
			body:       nil,
			setupMock:  func(_ *mocks.MockTicketRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			tt.setupMock(mockTickets)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte("not-json")
			}

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/"+tt.ticketID, bytes.NewReader(body))
			req = withURLParam(req, "id", tt.ticketID)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Update(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockTickets.AssertExpectations(t)
		})
	}
}

func TestTicketHandler_Delete(t *testing.T) {
	ticketID := uuid.New()

	tests := []struct {
		name       string
		ticketID   string
		setupMock  func(*mocks.MockTicketRepository)
		wantStatus int
	}{
		{
			name:     "deletes ticket successfully",
			ticketID: ticketID.String(),
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("Delete", mock.Anything, ticketID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:     "returns 404 for unknown ticket",
			ticketID: uuid.New().String(),
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid uuid",
			ticketID:   "not-a-uuid",
			setupMock:  func(_ *mocks.MockTicketRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			tt.setupMock(mockTickets)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/tickets/"+tt.ticketID, nil)
			req = withURLParam(req, "id", tt.ticketID)
			w := httptest.NewRecorder()

			h.Delete(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockTickets.AssertExpectations(t)
		})
	}
}

func TestTicketHandler_UpdateContact(t *testing.T) {
	ticketID := uuid.New()
	contactID := uuid.New()

	tests := []struct {
		name       string
		ticketID   string
		body       map[string]any
		setupMock  func(*mocks.MockTicketRepository)
		wantStatus int
	}{
		{
			name:     "sets contact successfully",
			ticketID: ticketID.String(),
			body:     map[string]any{"contact_id": contactID.String()},
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("UpdateContact", mock.Anything, ticketID, &contactID).
					Return(&domain.Ticket{ID: ticketID, ContactID: &contactID}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "clears contact with null",
			ticketID: ticketID.String(),
			body:     map[string]any{"contact_id": nil},
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("UpdateContact", mock.Anything, ticketID, (*uuid.UUID)(nil)).
					Return(&domain.Ticket{ID: ticketID}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid uuid",
			ticketID:   "not-a-uuid",
			body:       map[string]any{"contact_id": contactID.String()},
			setupMock:  func(_ *mocks.MockTicketRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			ticketID:   ticketID.String(),
			body:       nil,
			setupMock:  func(_ *mocks.MockTicketRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when ticket not found",
			ticketID: ticketID.String(),
			body:     map[string]any{"contact_id": contactID.String()},
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("UpdateContact", mock.Anything, ticketID, mock.Anything).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			tt.setupMock(mockTickets)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte("not-json")
			}

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/"+tt.ticketID+"/contact", bytes.NewReader(body))
			req = withURLParam(req, "id", tt.ticketID)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.UpdateContact(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockTickets.AssertExpectations(t)
		})
	}
}

func TestTicketHandler_UpdateContact_RealignsTicketAccountToReplacementContact(t *testing.T) {
	ticketID := uuid.New()
	currentAccountID := uuid.New()
	replacementAccountID := uuid.New()
	contactID := uuid.New()

	mockTickets := new(mocks.MockTicketRepository)
	mockComments := new(mocks.MockTicketCommentRepository)
	mockAttachments := new(mocks.MockTicketAttachmentRepository)
	mockContacts := new(mocks.MockContactRepository)

	mockTickets.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, AccountID: &currentAccountID}, nil)
	mockContacts.On("IsRelatedToAccount", mock.Anything, contactID, currentAccountID).
		Return(false, nil)
	mockContacts.On("GetByID", mock.Anything, contactID).
		Return(&domain.Contact{ID: contactID, AccountID: &replacementAccountID}, nil)
	mockTickets.On("Update", mock.Anything, ticketID, mock.MatchedBy(func(p domain.TicketPatch) bool {
		return p.ContactID != nil &&
			*p.ContactID == contactID &&
			p.AccountID != nil &&
			*p.AccountID == replacementAccountID &&
			!p.ClearAccountID
	})).Return(&domain.Ticket{ID: ticketID, ContactID: &contactID, AccountID: &replacementAccountID}, nil)

	h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{}).
		WithContacts(mockContacts)

	body, err := json.Marshal(map[string]any{"contact_id": contactID.String()})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/"+ticketID.String()+"/contact", bytes.NewReader(body))
	req = withURLParam(req, "id", ticketID.String())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateContact(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockTickets.AssertNotCalled(t, "UpdateContact", mock.Anything, mock.Anything, mock.Anything)
	mockTickets.AssertExpectations(t)
	mockContacts.AssertExpectations(t)
}

func TestTicketHandler_UpdateContact_ClearsTicketAccountWhenReplacementContactHasNoPrimaryAccount(t *testing.T) {
	ticketID := uuid.New()
	currentAccountID := uuid.New()
	contactID := uuid.New()

	mockTickets := new(mocks.MockTicketRepository)
	mockComments := new(mocks.MockTicketCommentRepository)
	mockAttachments := new(mocks.MockTicketAttachmentRepository)
	mockContacts := new(mocks.MockContactRepository)

	mockTickets.On("GetByID", mock.Anything, ticketID).
		Return(&domain.Ticket{ID: ticketID, AccountID: &currentAccountID}, nil)
	mockContacts.On("IsRelatedToAccount", mock.Anything, contactID, currentAccountID).
		Return(false, nil)
	mockContacts.On("GetByID", mock.Anything, contactID).
		Return(&domain.Contact{ID: contactID}, nil)
	mockTickets.On("Update", mock.Anything, ticketID, mock.MatchedBy(func(p domain.TicketPatch) bool {
		return p.ContactID != nil &&
			*p.ContactID == contactID &&
			p.AccountID == nil &&
			p.ClearAccountID
	})).Return(&domain.Ticket{ID: ticketID, ContactID: &contactID, AccountID: nil}, nil)

	h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{}).
		WithContacts(mockContacts)

	body, err := json.Marshal(map[string]any{"contact_id": contactID.String()})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/"+ticketID.String()+"/contact", bytes.NewReader(body))
	req = withURLParam(req, "id", ticketID.String())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateContact(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockTickets.AssertNotCalled(t, "UpdateContact", mock.Anything, mock.Anything, mock.Anything)
	mockTickets.AssertExpectations(t)
	mockContacts.AssertExpectations(t)
}

func TestTicketHandler_ListComments(t *testing.T) {
	ticketID := uuid.New()
	userID := uuid.New()
	orgID := domain.DefaultOrgID

	publicComment := &domain.TicketComment{
		ID:         uuid.New(),
		TicketID:   ticketID,
		Body:       "Public comment",
		IsInternal: false,
	}
	internalComment := &domain.TicketComment{
		ID:         uuid.New(),
		TicketID:   ticketID,
		Body:       "Internal note",
		IsInternal: true,
	}

	tests := []struct {
		name           string
		ticketID       string
		claims         *auth.Claims
		setupMock      func(*mocks.MockTicketCommentRepository)
		wantStatus     int
		expectInternal bool
	}{
		{
			name:     "admin sees all comments including internal",
			ticketID: ticketID.String(),
			claims: &auth.Claims{
				UserID: userID,
				OrgID:  orgID,
				Role:   string(domain.UserRoleAdmin),
			},
			setupMock: func(m *mocks.MockTicketCommentRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketCommentFilter) bool {
					return f.TicketID == ticketID && f.IsInternal == nil
				})).Return([]*domain.TicketComment{publicComment, internalComment}, nil)
			},
			wantStatus:     http.StatusOK,
			expectInternal: true,
		},
		{
			name:     "client sees only public comments",
			ticketID: ticketID.String(),
			claims: &auth.Claims{
				UserID: userID,
				OrgID:  orgID,
				Role:   string(domain.UserRoleClient),
			},
			setupMock: func(m *mocks.MockTicketCommentRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketCommentFilter) bool {
					return f.TicketID == ticketID && f.IsInternal != nil && *f.IsInternal == false
				})).Return([]*domain.TicketComment{publicComment}, nil)
			},
			wantStatus:     http.StatusOK,
			expectInternal: false,
		},
		{
			name:           "returns 400 for invalid ticket id",
			ticketID:       "not-a-uuid",
			claims:         &auth.Claims{UserID: userID, OrgID: orgID, Role: string(domain.UserRoleAdmin)},
			setupMock:      func(_ *mocks.MockTicketCommentRepository) {},
			wantStatus:     http.StatusBadRequest,
			expectInternal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			tt.setupMock(mockComments)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/"+tt.ticketID+"/comments", nil)
			req = withURLParam(req, "id", tt.ticketID)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			w := httptest.NewRecorder()

			h.ListComments(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockComments.AssertExpectations(t)
		})
	}
}

func TestTicketHandler_ListComments_NormalizesNilSlice(t *testing.T) {
	ticketID := uuid.New()

	mockTickets := new(mocks.MockTicketRepository)
	mockComments := new(mocks.MockTicketCommentRepository)
	mockAttachments := new(mocks.MockTicketAttachmentRepository)

	mockComments.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketCommentFilter) bool {
		return f.TicketID == ticketID
	})).Return(nil, nil)

	h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/"+ticketID.String()+"/comments", nil)
	req = withURLParam(req, "id", ticketID.String())
	w := httptest.NewRecorder()

	h.ListComments(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, "[]", w.Body.String())
	mockComments.AssertExpectations(t)
}

func TestTicketHandler_CreateComment(t *testing.T) {
	ticketID := uuid.New()
	userID := uuid.New()
	orgID := domain.DefaultOrgID

	tests := []struct {
		name       string
		ticketID   string
		claims     *auth.Claims
		body       map[string]any
		setupMock  func(*mocks.MockTicketCommentRepository)
		wantStatus int
	}{
		{
			name:     "admin can create internal comment",
			ticketID: ticketID.String(),
			claims: &auth.Claims{
				UserID: userID,
				OrgID:  orgID,
				Role:   string(domain.UserRoleAdmin),
			},
			body: map[string]any{
				"body":        "Internal note",
				"is_internal": true,
			},
			setupMock: func(m *mocks.MockTicketCommentRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.TicketComment) bool {
					return c.TicketID == ticketID && c.IsInternal == true
				})).Return(&domain.TicketComment{
					ID:         uuid.New(),
					TicketID:   ticketID,
					Body:       "Internal note",
					IsInternal: true,
				}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:     "client cannot create internal comment",
			ticketID: ticketID.String(),
			claims: &auth.Claims{
				UserID: userID,
				OrgID:  orgID,
				Role:   string(domain.UserRoleClient),
			},
			body: map[string]any{
				"body":        "Attempted internal note",
				"is_internal": true,
			},
			setupMock: func(m *mocks.MockTicketCommentRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.TicketComment) bool {
					return c.TicketID == ticketID && c.IsInternal == false
				})).Return(&domain.TicketComment{
					ID:         uuid.New(),
					TicketID:   ticketID,
					Body:       "Attempted internal note",
					IsInternal: false,
				}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:     "returns 422 for missing body",
			ticketID: ticketID.String(),
			claims: &auth.Claims{
				UserID: userID,
				OrgID:  orgID,
				Role:   string(domain.UserRoleAdmin),
			},
			body:       map[string]any{"is_internal": false},
			setupMock:  func(_ *mocks.MockTicketCommentRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			ticketID:   ticketID.String(),
			claims:     &auth.Claims{UserID: userID, OrgID: orgID, Role: string(domain.UserRoleAdmin)},
			body:       nil,
			setupMock:  func(_ *mocks.MockTicketCommentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid ticket id",
			ticketID:   "not-a-uuid",
			claims:     &auth.Claims{UserID: userID, OrgID: orgID, Role: string(domain.UserRoleAdmin)},
			body:       map[string]any{"body": "Comment"},
			setupMock:  func(_ *mocks.MockTicketCommentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			tt.setupMock(mockComments)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte("not-json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/"+tt.ticketID+"/comments", bytes.NewReader(body))
			req = withURLParam(req, "id", tt.ticketID)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.CreateComment(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockComments.AssertExpectations(t)
		})
	}
}

func TestTicketHandler_PublicReadAllowsCommentReadButBlocksNonOwnerCommentCreate(t *testing.T) {
	ticketID := uuid.New()
	userID := uuid.New()
	orgID := domain.DefaultOrgID
	access := &domain.AccessContext{
		UserID: userID,
		OrgID:  orgID,
		Sharing: map[domain.ACLModule]domain.ACLSharingAccess{
			domain.ACLModuleTickets: {Mode: domain.SharingDefaultPublicRO},
		},
	}

	mockTickets := new(mocks.MockTicketRepository)
	mockComments := new(mocks.MockTicketCommentRepository)
	mockAttachments := new(mocks.MockTicketAttachmentRepository)
	mockTickets.On("CanAccess", mock.Anything, ticketID, domain.SharingAccessRead).
		Return(true, nil).
		Once()
	mockTickets.On("CanAccess", mock.Anything, ticketID, domain.SharingAccessWrite).
		Return(false, nil).
		Once()
	mockComments.On("List", mock.Anything, mock.MatchedBy(func(f domain.TicketCommentFilter) bool {
		return f.TicketID == ticketID && f.IsInternal == nil
	})).Return([]*domain.TicketComment{}, nil).Once()

	h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/"+ticketID.String()+"/comments", nil)
	listReq = withURLParam(listReq, "id", ticketID.String())
	listReq = withClaims(listReq, &auth.Claims{UserID: userID, OrgID: orgID, Role: string(domain.UserRoleAgent)})
	listReq = listReq.WithContext(domain.WithAccessContext(listReq.Context(), access))
	listW := httptest.NewRecorder()

	h.ListComments(listW, listReq)

	require.Equal(t, http.StatusOK, listW.Code)

	body, err := json.Marshal(map[string]any{"body": "read-only users should not mutate"})
	require.NoError(t, err)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/"+ticketID.String()+"/comments", bytes.NewReader(body))
	createReq = withURLParam(createReq, "id", ticketID.String())
	createReq = withClaims(createReq, &auth.Claims{UserID: userID, OrgID: orgID, Role: string(domain.UserRoleAgent)})
	createReq = createReq.WithContext(domain.WithAccessContext(createReq.Context(), access))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()

	h.CreateComment(createW, createReq)

	require.Equal(t, http.StatusNotFound, createW.Code)
	mockComments.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockTickets.AssertExpectations(t)
	mockComments.AssertExpectations(t)
}

func TestTicketHandler_CreateCommentAllowsRepositoryWriteVisibility(t *testing.T) {
	ticketID := uuid.New()
	userID := uuid.New()
	orgID := domain.DefaultOrgID
	access := &domain.AccessContext{
		UserID: userID,
		OrgID:  orgID,
		Sharing: map[domain.ACLModule]domain.ACLSharingAccess{
			domain.ACLModuleTickets: {Mode: domain.SharingDefaultPrivate},
		},
	}

	mockTickets := new(mocks.MockTicketRepository)
	mockComments := new(mocks.MockTicketCommentRepository)
	mockAttachments := new(mocks.MockTicketAttachmentRepository)
	mockTickets.On("CanAccess", mock.Anything, ticketID, domain.SharingAccessWrite).
		Return(true, nil).
		Once()
	mockComments.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.TicketComment) bool {
		return c.TicketID == ticketID && c.Body == "manager update"
	})).Return(&domain.TicketComment{ID: uuid.New(), TicketID: ticketID, Body: "manager update"}, nil).Once()

	h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

	body, err := json.Marshal(map[string]any{"body": "manager update"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/"+ticketID.String()+"/comments", bytes.NewReader(body))
	req = withURLParam(req, "id", ticketID.String())
	req = withClaims(req, &auth.Claims{UserID: userID, OrgID: orgID, Role: string(domain.UserRoleAgent)})
	req = req.WithContext(domain.WithAccessContext(req.Context(), access))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateComment(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	mockTickets.AssertExpectations(t)
	mockComments.AssertExpectations(t)
}

func TestTicketHandler_ListAttachments_NormalizesNilSlice(t *testing.T) {
	ticketID := uuid.New()

	mockTickets := new(mocks.MockTicketRepository)
	mockComments := new(mocks.MockTicketCommentRepository)
	mockAttachments := new(mocks.MockTicketAttachmentRepository)

	mockAttachments.On("List", mock.Anything, ticketID).Return(nil, nil)

	h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/"+ticketID.String()+"/attachments", nil)
	req = withURLParam(req, "id", ticketID.String())
	w := httptest.NewRecorder()

	h.ListAttachments(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, "[]", w.Body.String())
	mockAttachments.AssertExpectations(t)
}

func TestTicketHandler_GetAttachment_InlinePreviewAllowlist(t *testing.T) {
	ticketID := uuid.New()
	attachmentID := uuid.New()
	size := int64(32)

	tests := []struct {
		name        string
		contentType string
		wantPrefix  string
	}{
		{
			name:        "keeps html attachments downloadable",
			contentType: "text/html",
			wantPrefix:  "attachment",
		},
		{
			name:        "allows inert text previews",
			contentType: "text/plain; charset=utf-8",
			wantPrefix:  "inline",
		},
		{
			name:        "allows pdf previews",
			contentType: "application/pdf",
			wantPrefix:  "inline",
		},
		{
			name:        "keeps svg attachments downloadable",
			contentType: "image/svg+xml",
			wantPrefix:  "attachment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			mockStorage := new(mocks.MockStorageBackend)

			attachment := &domain.TicketAttachment{
				ID:          attachmentID,
				TicketID:    ticketID,
				Filename:    "preview-test",
				ContentType: tt.contentType,
				SizeBytes:   &size,
				StorageKey:  "tickets/preview-test",
			}
			mockAttachments.On("GetByID", mock.Anything, attachmentID, ticketID).Return(attachment, nil)
			mockStorage.On("PresignURL", mock.Anything, attachment.StorageKey).Return("", nil)
			mockStorage.On("Open", mock.Anything, attachment.StorageKey).
				Return(io.NopCloser(bytes.NewBufferString("preview body")), nil)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mockStorage)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/"+ticketID.String()+"/attachments/"+attachmentID.String()+"?preview=1", nil)
			req = withURLParam(req, "id", ticketID.String())
			req = withURLParam(req, "attachmentID", attachmentID.String())
			w := httptest.NewRecorder()

			h.GetAttachment(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			assert.Contains(t, w.Header().Get("Content-Disposition"), tt.wantPrefix)
			mockAttachments.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestTicketHandler_DeleteComment(t *testing.T) {
	ticketID := uuid.New()
	commentID := uuid.New()

	tests := []struct {
		name       string
		ticketID   string
		commentID  string
		setupMock  func(*mocks.MockTicketCommentRepository)
		wantStatus int
	}{
		{
			name:      "deletes comment successfully",
			ticketID:  ticketID.String(),
			commentID: commentID.String(),
			setupMock: func(m *mocks.MockTicketCommentRepository) {
				m.On("Delete", mock.Anything, commentID, ticketID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:      "returns 404 for unknown comment",
			ticketID:  ticketID.String(),
			commentID: uuid.New().String(),
			setupMock: func(m *mocks.MockTicketCommentRepository) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID"), ticketID).
					Return(domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid ticket id",
			ticketID:   "not-a-uuid",
			commentID:  commentID.String(),
			setupMock:  func(_ *mocks.MockTicketCommentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid comment id",
			ticketID:   ticketID.String(),
			commentID:  "not-a-uuid",
			setupMock:  func(_ *mocks.MockTicketCommentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTickets := new(mocks.MockTicketRepository)
			mockComments := new(mocks.MockTicketCommentRepository)
			mockAttachments := new(mocks.MockTicketAttachmentRepository)
			tt.setupMock(mockComments)

			h := handler.NewTicketHandler(mockTickets, mockComments, mockAttachments, mocks.NoopStorageBackend{})

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/tickets/"+tt.ticketID+"/comments/"+tt.commentID, nil)
			req = withURLParam(req, "id", tt.ticketID)
			req = withURLParam(req, "commentID", tt.commentID)
			w := httptest.NewRecorder()

			h.DeleteComment(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockComments.AssertExpectations(t)
		})
	}
}
