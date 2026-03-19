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

func TestTicketHandler_List(t *testing.T) {
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
				assert.Contains(t, resp, "total")
				assert.Equal(t, float64(tt.wantTotal), resp["total"])
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
			setupMock:  func(m *mocks.MockTicketRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       nil,
			setupMock:  func(m *mocks.MockTicketRepository) {},
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
				m.On("GetByID", mock.Anything, ticketID).
					Return(&domain.Ticket{ID: ticketID, Subject: "Test ticket"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "returns 404 for unknown id",
			ticketID: uuid.New().String(),
			setupMock: func(m *mocks.MockTicketRepository) {
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid uuid",
			ticketID:   "not-a-uuid",
			setupMock:  func(m *mocks.MockTicketRepository) {},
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
			name:       "returns 400 for invalid uuid",
			ticketID:   "not-a-uuid",
			body:       map[string]any{"status": "closed"},
			setupMock:  func(m *mocks.MockTicketRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			ticketID:   ticketID.String(),
			body:       nil,
			setupMock:  func(m *mocks.MockTicketRepository) {},
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
			setupMock:  func(m *mocks.MockTicketRepository) {},
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
			setupMock:      func(m *mocks.MockTicketCommentRepository) {},
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
			setupMock:  func(m *mocks.MockTicketCommentRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			ticketID:   ticketID.String(),
			claims:     &auth.Claims{UserID: userID, OrgID: orgID, Role: string(domain.UserRoleAdmin)},
			body:       nil,
			setupMock:  func(m *mocks.MockTicketCommentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid ticket id",
			ticketID:   "not-a-uuid",
			claims:     &auth.Claims{UserID: userID, OrgID: orgID, Role: string(domain.UserRoleAdmin)},
			body:       map[string]any{"body": "Comment"},
			setupMock:  func(m *mocks.MockTicketCommentRepository) {},
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
			setupMock:  func(m *mocks.MockTicketCommentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid comment id",
			ticketID:   ticketID.String(),
			commentID:  "not-a-uuid",
			setupMock:  func(m *mocks.MockTicketCommentRepository) {},
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
