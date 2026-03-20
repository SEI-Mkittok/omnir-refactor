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

func TestSetupHandler_Status(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
		wantBody   map[string]bool
	}{
		{
			name: "returns setupRequired true when no users exist",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("CountAll", mock.Anything).Return(0, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   map[string]bool{"setupRequired": true},
		},
		{
			name: "returns setupRequired false when users exist",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("CountAll", mock.Anything).Return(1, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   map[string]bool{"setupRequired": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			jwtSvc := auth.NewJWTService("test-secret")
			h := handler.NewSetupHandler(mockRepo, jwtSvc)

			req := httptest.NewRequest(http.MethodGet, "/status", nil)
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			var body map[string]bool
			require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
			assert.Equal(t, tt.wantBody, body)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSetupHandler_Setup(t *testing.T) {
	adminUser := &domain.User{
		ID:    uuid.New(),
		OrgID: domain.DefaultOrgID,
		Email: "admin@example.com",
		Name:  "Admin",
		Role:  domain.UserRoleAdmin,
	}

	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name: "creates first admin user and sets auth cookies",
			body: map[string]any{
				"adminName": "Admin",
				"email":     "admin@example.com",
				"password":  "securepassword",
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("CountAll", mock.Anything).Return(0, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).
					Return(adminUser, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "returns 409 when setup already completed",
			body: map[string]any{
				"adminName": "Admin",
				"email":     "admin@example.com",
				"password":  "securepassword",
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("CountAll", mock.Anything).Return(1, nil)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "returns 422 for missing name",
			body: map[string]any{
				"email":    "admin@example.com",
				"password": "securepassword",
			},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for missing email",
			body: map[string]any{
				"adminName": "Admin",
				"password":  "securepassword",
			},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for password too short",
			body: map[string]any{
				"adminName": "Admin",
				"email":     "admin@example.com",
				"password":  "short",
			},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       nil,
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			jwtSvc := auth.NewJWTService("test-secret")
			h := handler.NewSetupHandler(mockRepo, jwtSvc)

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte("not-json")
			}

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantStatus == http.StatusCreated {
				cookies := rr.Result().Cookies()
				cookieMap := make(map[string]*http.Cookie)
				for _, c := range cookies {
					cookieMap[c.Name] = c
				}
				require.Contains(t, cookieMap, "access_token")
				assert.True(t, cookieMap["access_token"].HttpOnly)
				require.Contains(t, cookieMap, "refresh_token")

				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				assert.NotNil(t, resp["user"])
				assert.Nil(t, resp["access_token"])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
