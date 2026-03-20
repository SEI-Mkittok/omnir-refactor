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
			name: "sets cookies on valid credentials",
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
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for missing password",
			body:       map[string]any{"email": "user@example.com"},
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
				// Tokens must be in httpOnly cookies, not the response body
				cookies := rr.Result().Cookies()
				cookieMap := make(map[string]*http.Cookie)
				for _, c := range cookies {
					cookieMap[c.Name] = c
				}
				require.Contains(t, cookieMap, "access_token", "access_token cookie must be set")
				assert.True(t, cookieMap["access_token"].HttpOnly, "access_token must be httpOnly")
				assert.Equal(t, http.SameSiteStrictMode, cookieMap["access_token"].SameSite)
				require.Contains(t, cookieMap, "refresh_token", "refresh_token cookie must be set")
				assert.True(t, cookieMap["refresh_token"].HttpOnly, "refresh_token must be httpOnly")

				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				assert.NotNil(t, resp["user"])
				assert.Nil(t, resp["access_token"], "tokens must not be in response body")
				assert.Nil(t, resp["refresh_token"], "tokens must not be in response body")
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
		name        string
		cookieValue string
		setCookie   bool
		wantStatus  int
	}{
		{
			name:        "rotates cookies for valid refresh token",
			cookieValue: validToken,
			setCookie:   true,
			wantStatus:  http.StatusNoContent,
		},
		{
			name:        "returns 401 for expired token",
			cookieValue: expiredToken,
			setCookie:   true,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "returns 401 for invalid token",
			cookieValue: "not.a.token",
			setCookie:   true,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:       "returns 401 for missing cookie",
			setCookie:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			h := handler.NewAuthHandler(mockRepo, jwtSvc)

			req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
			if tt.setCookie {
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tt.cookieValue})
			}
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantStatus == http.StatusNoContent {
				cookies := rr.Result().Cookies()
				cookieMap := make(map[string]*http.Cookie)
				for _, c := range cookies {
					cookieMap[c.Name] = c
				}
				require.Contains(t, cookieMap, "access_token", "new access_token cookie must be set")
				assert.True(t, cookieMap["access_token"].HttpOnly)
				require.Contains(t, cookieMap, "refresh_token", "new refresh_token cookie must be set")
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	jwtSvc := auth.NewJWTService("test-secret")
	h := handler.NewAuthHandler(mockRepo, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rr := httptest.NewRecorder()

	h.Router().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	cookies := rr.Result().Cookies()
	cookieMap := make(map[string]*http.Cookie)
	for _, c := range cookies {
		cookieMap[c.Name] = c
	}
	require.Contains(t, cookieMap, "access_token")
	assert.Equal(t, -1, cookieMap["access_token"].MaxAge)
	require.Contains(t, cookieMap, "refresh_token")
	assert.Equal(t, -1, cookieMap["refresh_token"].MaxAge)
}
