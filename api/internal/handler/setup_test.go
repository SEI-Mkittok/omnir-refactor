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
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

var defaultOrg = &domain.Organization{
	ID:   domain.DefaultOrgID,
	Name: "Default",
	Slug: "default",
	Plan: domain.OrgPlanSingle,
}

func TestSetupHandler_Status(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
		wantBody   map[string]bool
	}{
		{
			name: "returns setupRequired true when no admin user exists",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("HasAdminUser", mock.Anything).Return(false, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   map[string]bool{"setupRequired": true},
		},
		{
			name: "returns setupRequired false when admin user exists",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("HasAdminUser", mock.Anything).Return(true, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   map[string]bool{"setupRequired": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsers := new(mocks.MockUserRepository)
			mockOrgs := new(mocks.MockOrgRepository)
			tt.setupMock(mockUsers)

			jwtSvc := auth.NewJWTService("test-secret")
			h := handler.NewSetupHandler(mockUsers, mockOrgs, jwtSvc, config.OrgModeSingle)

			req := httptest.NewRequest(http.MethodGet, "/status", nil)
			rr := httptest.NewRecorder()

			h.Router().ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			var body map[string]bool
			require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
			assert.Equal(t, tt.wantBody, body)
			mockUsers.AssertExpectations(t)
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
		name          string
		body          map[string]any
		setupUserMock func(*mocks.MockUserRepository)
		setupOrgMock  func(*mocks.MockOrgRepository)
		wantStatus    int
	}{
		{
			name: "creates first admin user when org already exists",
			body: map[string]any{
				"adminName": "Admin",
				"email":     "admin@example.com",
				"password":  "securepassword",
			},
			setupUserMock: func(m *mocks.MockUserRepository) {
				m.On("HasAdminUser", mock.Anything).Return(false, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).
					Return(adminUser, nil)
			},
			setupOrgMock: func(m *mocks.MockOrgRepository) {
				m.On("GetByID", mock.Anything, domain.DefaultOrgID).Return(defaultOrg, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "creates org when not found then creates admin user",
			body: map[string]any{
				"adminName": "Admin",
				"email":     "admin@example.com",
				"password":  "securepassword",
			},
			setupUserMock: func(m *mocks.MockUserRepository) {
				m.On("HasAdminUser", mock.Anything).Return(false, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).
					Return(adminUser, nil)
			},
			setupOrgMock: func(m *mocks.MockOrgRepository) {
				m.On("GetByID", mock.Anything, domain.DefaultOrgID).Return(nil, domain.ErrNotFound)
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Organization")).Return(defaultOrg, nil)
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
			setupUserMock: func(m *mocks.MockUserRepository) {
				m.On("HasAdminUser", mock.Anything).Return(true, nil)
			},
			setupOrgMock: func(_ *mocks.MockOrgRepository) {},
			wantStatus:   http.StatusConflict,
		},
		{
			name: "returns 422 for missing name",
			body: map[string]any{
				"email":    "admin@example.com",
				"password": "securepassword",
			},
			setupUserMock: func(_ *mocks.MockUserRepository) {},
			setupOrgMock:  func(_ *mocks.MockOrgRepository) {},
			wantStatus:    http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for missing email",
			body: map[string]any{
				"adminName": "Admin",
				"password":  "securepassword",
			},
			setupUserMock: func(_ *mocks.MockUserRepository) {},
			setupOrgMock:  func(_ *mocks.MockOrgRepository) {},
			wantStatus:    http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for password too short",
			body: map[string]any{
				"adminName": "Admin",
				"email":     "admin@example.com",
				"password":  "short",
			},
			setupUserMock: func(_ *mocks.MockUserRepository) {},
			setupOrgMock:  func(_ *mocks.MockOrgRepository) {},
			wantStatus:    http.StatusUnprocessableEntity,
		},
		{
			name:          "returns 400 for invalid JSON",
			body:          nil,
			setupUserMock: func(_ *mocks.MockUserRepository) {},
			setupOrgMock:  func(_ *mocks.MockOrgRepository) {},
			wantStatus:    http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsers := new(mocks.MockUserRepository)
			mockOrgs := new(mocks.MockOrgRepository)
			tt.setupUserMock(mockUsers)
			tt.setupOrgMock(mockOrgs)

			jwtSvc := auth.NewJWTService("test-secret")
			h := handler.NewSetupHandler(mockUsers, mockOrgs, jwtSvc, config.OrgModeSingle)

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

			mockUsers.AssertExpectations(t)
			mockOrgs.AssertExpectations(t)
		})
	}
}

func TestSetupHandler_Setup_MultiMode(t *testing.T) {
	adminUser := &domain.User{
		ID:    uuid.New(),
		OrgID: domain.DefaultOrgID,
		Email: "admin@example.com",
		Name:  "Admin",
		Role:  domain.UserRoleAdmin,
	}

	t.Run("returns 404 when orgs already exist in multi mode", func(t *testing.T) {
		mockUsers := new(mocks.MockUserRepository)
		mockOrgs := new(mocks.MockOrgRepository)
		mockOrgs.On("HasAny", mock.Anything).Return(true, nil)

		h := handler.NewSetupHandler(mockUsers, mockOrgs, auth.NewJWTService("s"), config.OrgModeSaaS)

		body, _ := json.Marshal(map[string]any{
			"adminName": "Admin", "email": "admin@example.com", "password": "securepassword",
		})
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.Router().ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockOrgs.AssertExpectations(t)
		mockUsers.AssertNotCalled(t, "HasAdminUser", mock.Anything)
	})

	t.Run("allows setup in multi mode when no orgs exist (bootstrap)", func(t *testing.T) {
		mockUsers := new(mocks.MockUserRepository)
		mockOrgs := new(mocks.MockOrgRepository)
		mockOrgs.On("HasAny", mock.Anything).Return(false, nil)
		mockUsers.On("HasAdminUser", mock.Anything).Return(false, nil)
		mockOrgs.On("GetByID", mock.Anything, domain.DefaultOrgID).Return(defaultOrg, nil)
		mockUsers.On("Create", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).
			Return(adminUser, nil)

		h := handler.NewSetupHandler(mockUsers, mockOrgs, auth.NewJWTService("test-secret"), config.OrgModeSaaS)

		body, _ := json.Marshal(map[string]any{
			"adminName": "Admin", "email": "admin@example.com", "password": "securepassword",
		})
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.Router().ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockOrgs.AssertExpectations(t)
	})
}
