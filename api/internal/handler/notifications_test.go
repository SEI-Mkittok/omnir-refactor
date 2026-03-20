package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func TestNotificationHandler_List(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	claims := &auth.Claims{UserID: userID, OrgID: orgID}

	tests := []struct {
		name       string
		setupMock  func(*mocks.MockNotificationRepository)
		query      string
		wantStatus int
	}{
		{
			name: "returns notifications for user",
			setupMock: func(m *mocks.MockNotificationRepository) {
				m.On("ListByUser", mock.Anything, mock.MatchedBy(func(f domain.NotificationFilter) bool {
					return f.UserID == userID && !f.UnreadOnly
				})).Return([]*domain.Notification{
					{ID: uuid.New(), UserID: userID, Kind: domain.NotificationKindActivityReminder},
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "filters unread notifications",
			setupMock: func(m *mocks.MockNotificationRepository) {
				m.On("ListByUser", mock.Anything, mock.MatchedBy(func(f domain.NotificationFilter) bool {
					return f.UserID == userID && f.UnreadOnly
				})).Return([]*domain.Notification{}, nil)
			},
			query:      "?unread_only=true",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockNotificationRepository)
			tt.setupMock(mockRepo)

			h := handler.NewNotificationHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/notifications"+tt.query, nil)
			req = withClaims(req, claims)
			w := httptest.NewRecorder()

			h.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationHandler_UnreadCount(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	claims := &auth.Claims{UserID: userID, OrgID: orgID}

	t.Run("returns unread count", func(t *testing.T) {
		mockRepo := new(mocks.MockNotificationRepository)
		mockRepo.On("UnreadCount", mock.Anything, userID, orgID).Return(5, nil)

		h := handler.NewNotificationHandler(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
		req = withClaims(req, claims)
		w := httptest.NewRecorder()

		h.UnreadCount(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestNotificationHandler_MarkRead(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	notifID := uuid.New()

	claims := &auth.Claims{UserID: userID, OrgID: orgID}

	tests := []struct {
		name       string
		notifID    string
		setupMock  func(*mocks.MockNotificationRepository)
		wantStatus int
	}{
		{
			name:    "marks notification as read",
			notifID: notifID.String(),
			setupMock: func(m *mocks.MockNotificationRepository) {
				m.On("MarkRead", mock.Anything, notifID, userID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "returns 404 when notification not found",
			notifID: uuid.New().String(),
			setupMock: func(m *mocks.MockNotificationRepository) {
				m.On("MarkRead", mock.Anything, mock.AnythingOfType("uuid.UUID"), userID).
					Return(domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid uuid",
			notifID:    "not-a-uuid",
			setupMock:  func(_ *mocks.MockNotificationRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockNotificationRepository)
			tt.setupMock(mockRepo)

			h := handler.NewNotificationHandler(mockRepo)

			req := httptest.NewRequest(http.MethodPost, "/api/notifications/"+tt.notifID+"/read", nil)
			req = withClaims(req, claims)
			req = withURLParam(req, "id", tt.notifID)
			w := httptest.NewRecorder()

			h.MarkRead(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationHandler_MarkAllRead(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	claims := &auth.Claims{UserID: userID, OrgID: orgID}

	t.Run("marks all notifications as read", func(t *testing.T) {
		mockRepo := new(mocks.MockNotificationRepository)
		mockRepo.On("MarkAllRead", mock.Anything, userID, orgID).Return(nil)

		h := handler.NewNotificationHandler(mockRepo)

		req := httptest.NewRequest(http.MethodPost, "/api/notifications/read-all", nil)
		req = withClaims(req, claims)
		w := httptest.NewRecorder()

		h.MarkAllRead(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
