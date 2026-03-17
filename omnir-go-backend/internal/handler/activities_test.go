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
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func TestActivityHandler_Create(t *testing.T) {
	ownerID := uuid.New()

	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockActivityRepository)
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
			name: "returns 422 for missing subject",
			body: map[string]any{
				"type":     "call",
				"owner_id": ownerID.String(),
			},
			setupMock:  func(m *mocks.MockActivityRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for invalid type",
			body: map[string]any{
				"type":     "webinar",
				"subject":  "Webinar",
				"owner_id": ownerID.String(),
			},
			setupMock:  func(m *mocks.MockActivityRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for missing owner_id",
			body: map[string]any{
				"type":    "task",
				"subject": "Follow up",
			},
			setupMock:  func(m *mocks.MockActivityRepository) {},
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
			setupMock:  func(m *mocks.MockActivityRepository) {},
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
