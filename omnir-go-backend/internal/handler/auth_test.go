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
	"golang.org/x/crypto/bcrypt"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func TestAuthHandler_Login(t *testing.T) {
	password := "securepassword"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	user := &domain.User{
		ID:    uuid.New(),
		OrgID: domain.DefaultOrgID,
		Email: "user@example.com",
		Name:  "Test User",
		Role:  domain.UserRoleAdmin,
	}

	tests := []struct {
		name       string
		body       any
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name: "returns tokens on valid credentials",
			body: map[string]any{"email": "user@example.com", "password": password},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "user@example.com").Return(user, string(hash), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "returns 401 for wrong password",
			body: map[string]any{"email": "user@example.com", "password": "wrongpassword"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "user@example.com").Return(user, string(hash), nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns 401 for unknown email",
			body: map[string]any{"email": "nobody@example.com", "password": password},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("FindByEmail", mock.Anything, "nobody@example.com").Return(nil, "", nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 422 for missing email",
			body:       map[string]any{"password": password},
			setupMock:  func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for missing password",
			body:       map[string]any{"email": "user@example.com"},
			setupMock:  func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       nil,
			setupMock:  func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			jwtSvc := auth.NewJWTService("test-secret")
			h := handler.NewAuthHandler(mockRepo, jwtSvc)

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte("not-json")
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantStatus == http.StatusOK {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				assert.NotEmpty(t, resp["access_token"])
				assert.NotEmpty(t, resp["refresh_token"])
				assert.NotNil(t, resp["user"])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret")

	validClaims := auth.Claims{
		UserID: uuid.New(),
		OrgID:  domain.DefaultOrgID,
		Role:   string(domain.UserRoleAdmin),
	}
	validToken, _ := jwtSvc.Issue(validClaims, 7*24*time.Hour)

	expiredJwtSvc := auth.NewJWTService("test-secret")
	expiredToken, _ := expiredJwtSvc.Issue(validClaims, -1*time.Second)

	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{
			name:       "returns new access token for valid refresh token",
			body:       map[string]any{"refresh_token": validToken},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 for expired token",
			body:       map[string]any{"refresh_token": expiredToken},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 for invalid token",
			body:       map[string]any{"refresh_token": "not.a.token"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 422 for missing refresh_token",
			body:       map[string]any{},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			h := handler.NewAuthHandler(mockRepo, jwtSvc)

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte("not-json")
			}

			req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantStatus == http.StatusOK {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				assert.NotEmpty(t, resp["access_token"])
				assert.Nil(t, resp["refresh_token"])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
