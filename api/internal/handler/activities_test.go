package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func TestActivityHandler_Create(t *testing.T) {
	ownerID := uuid.New()

	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockActivityRepository)
		claims     *auth.Claims
		wantStatus int
	}{
		{
			name: "creates activity successfully",
			body: map[string]any{
				"type":     "call",
				"subject":  "Discovery call",
				"owner_id": ownerID.String(),
			},
			setupMock: func(m *mocks.MockActivityRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Activity")).
					Return(&domain.Activity{
						ID:      uuid.New(),
						Type:    domain.ActivityTypeCall,
						Subject: "Discovery call",
						OwnerID: ownerID,
					}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "defaults owner_id from auth claims when omitted",
			body: map[string]any{
				"type":    "meeting",
				"subject": "Demo prep",
			},
			setupMock: func(m *mocks.MockActivityRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(a *domain.Activity) bool {
					return a.Type == domain.ActivityTypeMeeting && a.OwnerID != uuid.Nil
				})).Return(&domain.Activity{
					ID:      uuid.New(),
					Type:    domain.ActivityTypeMeeting,
					Subject: "Demo prep",
					OwnerID: ownerID,
				}, nil)
			},
			claims: &auth.Claims{
				UserID: ownerID,
				Role:   string(domain.UserRoleAgent),
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "creates task type successfully",
			body: map[string]any{
				"type":     "task",
				"subject":  "Follow up with client",
				"owner_id": ownerID.String(),
			},
			setupMock: func(m *mocks.MockActivityRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(a *domain.Activity) bool {
					return a.Type == domain.ActivityTypeTask && a.Subject == "Follow up with client"
				})).Return(&domain.Activity{
					ID:      uuid.New(),
					Type:    domain.ActivityTypeTask,
					Subject: "Follow up with client",
					OwnerID: ownerID,
				}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "accepts date-only due_date payload",
			body: map[string]any{
				"type":     "task",
				"subject":  "Date only",
				"owner_id": ownerID.String(),
				"due_date": "2026-05-10",
			},
			setupMock: func(m *mocks.MockActivityRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(a *domain.Activity) bool {
					return a.DueDate != nil && a.DueDate.UTC().Format("2006-01-02") == "2026-05-10"
				})).Return(&domain.Activity{
					ID:      uuid.New(),
					Type:    domain.ActivityTypeTask,
					Subject: "Date only",
					OwnerID: ownerID,
				}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "returns 422 when end_at is before start_at",
			body: map[string]any{
				"type":     "meeting",
				"subject":  "Broken schedule",
				"owner_id": ownerID.String(),
				"start_at": "2026-05-10T12:00:00Z",
				"end_at":   "2026-05-10T11:30:00Z",
			},
			setupMock:  func(_ *mocks.MockActivityRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for missing subject",
			body: map[string]any{
				"type":     "call",
				"owner_id": ownerID.String(),
			},
			setupMock:  func(_ *mocks.MockActivityRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for invalid type",
			body: map[string]any{
				"type":     "webinar",
				"subject":  "Webinar",
				"owner_id": ownerID.String(),
			},
			setupMock:  func(_ *mocks.MockActivityRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for missing owner_id",
			body: map[string]any{
				"type":    "task",
				"subject": "Follow up",
			},
			setupMock:  func(_ *mocks.MockActivityRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockActivityRepository)
			tt.setupMock(mockRepo)

			h := handler.NewActivityHandler(mockRepo)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/activities", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.claims != nil {
				req = req.WithContext(middleware.WithClaims(req.Context(), tt.claims))
			}
			w := httptest.NewRecorder()

			h.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestActivityHandler_GetByID(t *testing.T) {
	activityID := uuid.New()

	tests := []struct {
		name       string
		activityID string
		setupMock  func(*mocks.MockActivityRepository)
		wantStatus int
	}{
		{
			name:       "returns activity by id",
			activityID: activityID.String(),
			setupMock: func(m *mocks.MockActivityRepository) {
				m.On("GetByID", mock.Anything, activityID).
					Return(&domain.Activity{
						ID:      activityID,
						Type:    domain.ActivityTypeTask,
						Subject: "Send invoice",
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 404 for unknown id",
			activityID: uuid.New().String(),
			setupMock: func(m *mocks.MockActivityRepository) {
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid uuid",
			activityID: "not-a-uuid",
			setupMock:  func(_ *mocks.MockActivityRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockActivityRepository)
			tt.setupMock(mockRepo)

			h := handler.NewActivityHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/activities/"+tt.activityID, nil)
			req = withURLParam(req, "id", tt.activityID)
			w := httptest.NewRecorder()

			h.GetByID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestActivityHandler_Delete(t *testing.T) {
	activityID := uuid.New()

	tests := []struct {
		name       string
		activityID string
		setupMock  func(*mocks.MockActivityRepository)
		wantStatus int
	}{
		{
			name:       "soft-deletes successfully",
			activityID: activityID.String(),
			setupMock: func(m *mocks.MockActivityRepository) {
				m.On("Delete", mock.Anything, activityID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 404 when not found",
			activityID: uuid.New().String(),
			setupMock: func(m *mocks.MockActivityRepository) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockActivityRepository)
			tt.setupMock(mockRepo)

			h := handler.NewActivityHandler(mockRepo)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/activities/"+tt.activityID, nil)
			req = withURLParam(req, "id", tt.activityID)
			w := httptest.NewRecorder()

			h.Delete(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestActivityHandler_Update_CompletedFalseClearsCompletedAt(t *testing.T) {
	activityID := uuid.New()
	mockRepo := new(mocks.MockActivityRepository)
	mockRepo.On("Update", mock.Anything, activityID, mock.MatchedBy(func(p domain.ActivityPatch) bool {
		return p.CompletedAt != nil && p.CompletedAt.IsZero()
	})).Return(&domain.Activity{
		ID:      activityID,
		Type:    domain.ActivityTypeTask,
		Subject: "Follow up",
	}, nil)

	h := handler.NewActivityHandler(mockRepo)

	body, err := json.Marshal(map[string]any{
		"completed": false,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/activities/"+activityID.String(), bytes.NewReader(body))
	req = withURLParam(req, "id", activityID.String())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestActivityHandler_Update_RejectsEndBeforeStart(t *testing.T) {
	activityID := uuid.New()
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	mockRepo := new(mocks.MockActivityRepository)
	mockRepo.On("GetByID", mock.Anything, activityID).Return(&domain.Activity{
		ID:      activityID,
		Type:    domain.ActivityTypeMeeting,
		Subject: "Demo",
		DueDate: &now,
		OwnerID: uuid.New(),
	}, nil)

	h := handler.NewActivityHandler(mockRepo)

	body, err := json.Marshal(map[string]any{
		"start_at": "2026-05-10T12:00:00Z",
		"end_at":   "2026-05-10T11:00:00Z",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/activities/"+activityID.String(), bytes.NewReader(body))
	req = withURLParam(req, "id", activityID.String())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Update(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	mockRepo.AssertExpectations(t)
}

func TestActivityHandler_Create_RejectsUnrelatedContactAccountPair(t *testing.T) {
	ownerID := uuid.New()
	contactID := uuid.New()
	accountID := uuid.New()

	mockRepo := new(mocks.MockActivityRepository)
	mockContacts := new(mocks.MockContactRepository)
	mockContacts.On("IsRelatedToAccount", mock.Anything, contactID, accountID).Return(false, nil)

	h := handler.NewActivityHandler(mockRepo).WithContacts(mockContacts)

	body, err := json.Marshal(map[string]any{
		"type":       "call",
		"subject":    "Discovery call",
		"owner_id":   ownerID.String(),
		"contact_id": contactID.String(),
		"account_id": accountID.String(),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/activities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Create(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockContacts.AssertExpectations(t)
}
