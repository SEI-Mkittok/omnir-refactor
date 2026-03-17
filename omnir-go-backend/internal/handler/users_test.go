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

func adminClaims(userID uuid.UUID) *auth.Claims {
	return &auth.Claims{UserID: userID, Role: string(domain.UserRoleAdmin)}
}

func userClaims(userID uuid.UUID) *auth.Claims {
	return &auth.Claims{UserID: userID, Role: string(domain.UserRoleUser)}
}

func makeUser(id uuid.UUID) *domain.User {
	return &domain.User{ID: id, Email: "user@example.com", Name: "Test User", Role: domain.UserRoleUser}
}

func TestUserHandler_List(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		query      string
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
		wantTotal  int
	}{
		{
			name:  "lists users with defaults",
			query: "",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.UserFilter")).
					Return([]*domain.User{makeUser(userID)}, 1, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name:  "filters by role",
			query: "?role=admin",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.UserFilter) bool {
					return f.Role != nil && *f.Role == domain.UserRoleAdmin
				})).Return([]*domain.User{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
		{
			name:  "search query",
			query: "?q=alice",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.UserFilter) bool {
					return f.Q == "alice"
				})).Return([]*domain.User{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
		{
			name:  "pagination params",
			query: "?page=2&limit=10",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.UserFilter) bool {
					return f.Page == 2 && f.Limit == 10
				})).Return([]*domain.User{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			h := handler.NewUserHandler(mockRepo)
			req := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
			w := httptest.NewRecorder()

			h.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			var resp map[string]any
			require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
			assert.Equal(t, float64(tt.wantTotal), resp["total"])
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUserHandler_Create tests Create via the Router (which applies requireAdmin).
func TestUserHandler_Create(t *testing.T) {
	adminID := uuid.New()

	tests := []struct {
		name       string
		body       any
		claims     *auth.Claims
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name:   "admin creates user successfully",
			claims: adminClaims(adminID),
			body: map[string]any{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "secret123",
				"role":     "user",
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).
					Return(makeUser(uuid.New()), nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "returns 422 for missing name",
			claims: adminClaims(adminID),
			body:   map[string]any{"email": "alice@example.com", "password": "secret"},
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "returns 422 for missing email",
			claims: adminClaims(adminID),
			body:   map[string]any{"name": "Alice", "password": "secret"},
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "returns 422 for missing password",
			claims: adminClaims(adminID),
			body:   map[string]any{"name": "Alice", "email": "alice@example.com"},
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "returns 400 for invalid JSON",
			claims: adminClaims(adminID),
			body:   nil,
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "non-admin gets 403",
			claims: userClaims(uuid.New()),
			body: map[string]any{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "secret",
			},
			setupMock:  func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "no claims gets 403",
			claims: nil,
			body: map[string]any{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "secret",
			},
			setupMock:  func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			h := handler.NewUserHandler(mockRepo)

			var bodyBytes []byte
			if tt.body != nil {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			} else {
				bodyBytes = []byte("invalid-json{")
			}

			// Route through the chi subrouter so requireAdmin wrapping is applied.
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			w := httptest.NewRecorder()

			h.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetMe(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		claims     *auth.Claims
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name:   "returns current user",
			claims: userClaims(userID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).
					Return(makeUser(userID), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "returns 401 without claims",
			claims:    nil,
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "returns 404 if user not found",
			claims: userClaims(userID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			h := handler.NewUserHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			w := httptest.NewRecorder()

			h.GetMe(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetByID(t *testing.T) {
	adminID := uuid.New()
	regularID := uuid.New()
	otherID := uuid.New()

	tests := []struct {
		name       string
		targetID   string
		claims     *auth.Claims
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name:     "admin can fetch any user",
			targetID: otherID.String(),
			claims:   adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, otherID).
					Return(makeUser(otherID), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "user can fetch own record",
			targetID: regularID.String(),
			claims:   userClaims(regularID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, regularID).
					Return(makeUser(regularID), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "user cannot fetch other user",
			targetID:  otherID.String(),
			claims:    userClaims(regularID),
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "returns 403 without claims",
			targetID:  otherID.String(),
			claims:    nil,
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 404 for unknown id",
			targetID: uuid.New().String(),
			claims:   adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "returns 400 for invalid uuid",
			targetID:  "not-a-uuid",
			claims:    adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			h := handler.NewUserHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.targetID, nil)
			req = withURLParam(req, "id", tt.targetID)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			w := httptest.NewRecorder()

			h.GetByID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Update(t *testing.T) {
	adminID := uuid.New()
	regularID := uuid.New()
	otherID := uuid.New()
	adminRole := domain.UserRoleAdmin

	tests := []struct {
		name       string
		targetID   string
		claims     *auth.Claims
		body       map[string]any
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name:     "user can update own name",
			targetID: regularID.String(),
			claims:   userClaims(regularID),
			body:     map[string]any{"name": "New Name"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Update", mock.Anything, regularID, mock.AnythingOfType("domain.UserPatch")).
					Return(makeUser(regularID), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "admin can update any user",
			targetID: otherID.String(),
			claims:   adminClaims(adminID),
			body:     map[string]any{"name": "Updated"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Update", mock.Anything, otherID, mock.AnythingOfType("domain.UserPatch")).
					Return(makeUser(otherID), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "admin can change role",
			targetID: otherID.String(),
			claims:   adminClaims(adminID),
			body:     map[string]any{"role": "admin"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Update", mock.Anything, otherID, mock.MatchedBy(func(p domain.UserPatch) bool {
					return p.Role != nil && *p.Role == adminRole
				})).Return(&domain.User{ID: otherID, Role: domain.UserRoleAdmin}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "non-admin cannot change role",
			targetID:  regularID.String(),
			claims:    userClaims(regularID),
			body:      map[string]any{"role": "admin"},
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "user cannot update other user",
			targetID:  otherID.String(),
			claims:    userClaims(regularID),
			body:      map[string]any{"name": "Updated"},
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "returns 401 without claims",
			targetID:  otherID.String(),
			claims:    nil,
			body:      map[string]any{"name": "Updated"},
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "returns 400 for invalid uuid",
			targetID:  "not-a-uuid",
			claims:    adminClaims(adminID),
			body:      map[string]any{"name": "Updated"},
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 for unknown id",
			targetID: uuid.New().String(),
			claims:   adminClaims(adminID),
			body:     map[string]any{"name": "Updated"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Update", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("domain.UserPatch")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			h := handler.NewUserHandler(mockRepo)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/"+tt.targetID, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = withURLParam(req, "id", tt.targetID)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			w := httptest.NewRecorder()

			h.Update(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUserHandler_Delete tests Delete via the Router (which applies requireAdmin).
func TestUserHandler_Delete(t *testing.T) {
	adminID := uuid.New()
	otherID := uuid.New()

	tests := []struct {
		name       string
		targetID   string
		claims     *auth.Claims
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name:     "admin deletes another user",
			targetID: otherID.String(),
			claims:   adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Delete", mock.Anything, otherID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:      "admin cannot self-delete",
			targetID:  adminID.String(),
			claims:    adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:      "non-admin gets 403",
			targetID:  otherID.String(),
			claims:    userClaims(uuid.New()),
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "no claims gets 403",
			targetID:  otherID.String(),
			claims:    nil,
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 404 for unknown user",
			targetID: uuid.New().String(),
			claims:   adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "returns 400 for invalid uuid",
			targetID:  "not-a-uuid",
			claims:    adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			h := handler.NewUserHandler(mockRepo)

			// Route through the chi subrouter so requireAdmin wrapping is applied.
			// chi parses the /{id} param from the URL path automatically.
			req := httptest.NewRequest(http.MethodDelete, "/"+tt.targetID, nil)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			w := httptest.NewRecorder()

			h.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
