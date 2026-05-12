package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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

func superAdminClaims(userID uuid.UUID) *auth.Claims {
	return &auth.Claims{UserID: userID, Role: string(domain.UserRoleSuperAdmin)}
}

func userClaims(userID uuid.UUID) *auth.Claims {
	return &auth.Claims{UserID: userID, Role: string(domain.UserRoleAgent)}
}

func makeUser(id uuid.UUID) *domain.User {
	return &domain.User{ID: id, Email: "user@example.com", Name: "Test User", Role: domain.UserRoleAgent}
}

type fakeUserAssignmentRepo struct {
	roles    []*domain.ACLRole
	profiles []*domain.ACLProfile
}

func (f fakeUserAssignmentRepo) ListRoles(_ context.Context, _ uuid.UUID) ([]*domain.ACLRole, error) {
	return f.roles, nil
}

func (f fakeUserAssignmentRepo) ListProfiles(_ context.Context, _ uuid.UUID) ([]*domain.ACLProfile, error) {
	return f.profiles, nil
}

func TestUserHandler_List(t *testing.T) {
	adminID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		query      string
		claims     *auth.Claims
		setupMock  func(*mocks.MockUserRepository)
		wantStatus int
		wantTotal  int
	}{
		{
			name:   "admin lists users with defaults",
			query:  "",
			claims: adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.UserFilter")).
					Return([]*domain.User{makeUser(userID)}, 1, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name:   "super admin lists users with defaults",
			query:  "",
			claims: superAdminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.UserFilter")).
					Return([]*domain.User{makeUser(userID)}, 1, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name:   "admin filters by role",
			query:  "?role=admin",
			claims: adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.UserFilter) bool {
					return f.Role != nil && *f.Role == domain.UserRoleAdmin
				})).Return([]*domain.User{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
		{
			name:   "admin search query",
			query:  "?q=alice",
			claims: adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.UserFilter) bool {
					return f.Q == "alice"
				})).Return([]*domain.User{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
		{
			name:   "admin pagination params",
			query:  "?page=2&limit=10",
			claims: adminClaims(adminID),
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.UserFilter) bool {
					return f.Page == 2 && f.Limit == 10
				})).Return([]*domain.User{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
		{
			name:       "non-admin gets 403",
			query:      "",
			claims:     userClaims(uuid.New()),
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "no claims gets 403",
			query:      "",
			claims:     nil,
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			h := handler.NewUserHandler(mockRepo)
			req := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			w := httptest.NewRecorder()

			h.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				meta := resp["meta"].(map[string]any)
				assert.Equal(t, float64(tt.wantTotal), meta["total"])
			}
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
				"role":     "agent",
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).
					Return(makeUser(uuid.New()), nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "returns 422 for invalid role",
			claims: adminClaims(adminID),
			body: map[string]any{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "secret123",
				"role":     "owner",
			},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "returns 409 for duplicate email from repo",
			claims: adminClaims(adminID),
			body: map[string]any{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "secret123",
				"role":     "agent",
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).
					Return(nil, &pgconn.PgError{Code: "23505"})
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:   "does not hardcode DefaultOrgID — leaves org_id unset for repo to populate from context",
			claims: adminClaims(adminID),
			body: map[string]any{
				"name":     "Bob",
				"email":    "bob@example.com",
				"password": "secret123",
				"role":     "agent",
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.OrgID == uuid.Nil
				}), mock.AnythingOfType("string")).
					Return(makeUser(uuid.New()), nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 422 for missing name",
			claims:     adminClaims(adminID),
			body:       map[string]any{"email": "alice@example.com", "password": "secret"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for missing email",
			claims:     adminClaims(adminID),
			body:       map[string]any{"name": "Alice", "password": "secret"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for missing password",
			claims:     adminClaims(adminID),
			body:       map[string]any{"name": "Alice", "email": "alice@example.com"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			claims:     adminClaims(adminID),
			body:       nil,
			setupMock:  func(_ *mocks.MockUserRepository) {},
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
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "admin cannot assign super admin",
			claims: adminClaims(adminID),
			body: map[string]any{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "secret123",
				"role":     "super_admin",
			},
			setupMock:  func(_ *mocks.MockUserRepository) {},
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
			setupMock:  func(_ *mocks.MockUserRepository) {},
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
			name:       "returns 401 without claims",
			claims:     nil,
			setupMock:  func(_ *mocks.MockUserRepository) {},
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

func TestUserHandler_GetMeIncludesResolvedPermissions(t *testing.T) {
	userID := uuid.New()
	mockRepo := new(mocks.MockUserRepository)
	mockRepo.On("GetByID", mock.Anything, userID).Return(makeUser(userID), nil)
	h := handler.NewUserHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req = withClaims(req, userClaims(userID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleUsers: {
				domain.ACLActionAdmin: true,
			},
			domain.ACLModuleSettings: {
				domain.ACLActionAdmin: true,
			},
		},
	}))
	w := httptest.NewRecorder()

	h.GetMe(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var got domain.User
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	require.True(t, got.Permissions[domain.ACLModuleUsers][domain.ACLActionAdmin])
	require.True(t, got.Permissions[domain.ACLModuleSettings][domain.ACLActionAdmin])
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_AssignmentOptionsAllowsDelegatedUserAdmin(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	profileID := uuid.New()
	mockRepo := new(mocks.MockUserRepository)
	assignments := fakeUserAssignmentRepo{
		roles: []*domain.ACLRole{
			{ID: roleID, OrgID: domain.DefaultOrgID, Name: "Sales Manager"},
		},
		profiles: []*domain.ACLProfile{
			{ID: profileID, OrgID: domain.DefaultOrgID, Name: "Delegated Admin"},
		},
	}
	h := handler.NewUserHandler(mockRepo, assignments)

	req := httptest.NewRequest(http.MethodGet, "/assignment-options", nil)
	req = withClaims(req, userClaims(userID))
	req = req.WithContext(domain.WithOrgID(req.Context(), domain.DefaultOrgID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleUsers: {domain.ACLActionAdmin: true},
		},
	}))
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Roles    []domain.ACLRole    `json:"roles"`
		Profiles []domain.ACLProfile `json:"profiles"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Roles, 1)
	require.Equal(t, roleID, resp.Roles[0].ID)
	require.Len(t, resp.Profiles, 1)
	require.Equal(t, profileID, resp.Profiles[0].ID)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_CreateRequiresSettingsAdminForACLAssignments(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()
	mockRepo := new(mocks.MockUserRepository)
	h := handler.NewUserHandler(mockRepo)

	bodyBytes, err := json.Marshal(map[string]any{
		"name":       "Delegated Target",
		"email":      "target@example.com",
		"password":   "secret123",
		"role":       "agent",
		"profile_id": profileID.String(),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, userClaims(userID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleUsers: {domain.ACLActionAdmin: true},
		},
	}))
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func TestUserHandler_CreateDelegatedUserAdminCannotAssignPlatformAdmin(t *testing.T) {
	userID := uuid.New()
	mockRepo := new(mocks.MockUserRepository)
	h := handler.NewUserHandler(mockRepo)

	bodyBytes, err := json.Marshal(map[string]any{
		"name":     "Promoted User",
		"email":    "promoted@example.com",
		"password": "secret123",
		"role":     "admin",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, userClaims(userID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleUsers: {domain.ACLActionAdmin: true},
		},
	}))
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
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
			name:       "user cannot fetch other user",
			targetID:   otherID.String(),
			claims:     userClaims(regularID),
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "returns 403 without claims",
			targetID:   otherID.String(),
			claims:     nil,
			setupMock:  func(_ *mocks.MockUserRepository) {},
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
			name:       "returns 400 for invalid uuid",
			targetID:   "not-a-uuid",
			claims:     adminClaims(adminID),
			setupMock:  func(_ *mocks.MockUserRepository) {},
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
	superAdminRole := domain.UserRoleSuperAdmin

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
				m.On("GetByID", mock.Anything, otherID).
					Return(&domain.User{ID: otherID, Role: domain.UserRoleAgent}, nil)
				m.On("Update", mock.Anything, otherID, mock.MatchedBy(func(p domain.UserPatch) bool {
					return p.Role != nil && *p.Role == adminRole
				})).Return(&domain.User{ID: otherID, Role: domain.UserRoleAdmin}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "admin can clear bundle4 assignments",
			targetID: otherID.String(),
			claims:   adminClaims(adminID),
			body:     map[string]any{"role_id": nil, "profile_id": nil},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Update", mock.Anything, otherID, mock.MatchedBy(func(p domain.UserPatch) bool {
					return p.RoleID == nil &&
						p.ProfileID == nil &&
						p.ClearRoleID &&
						p.ClearProfileID
				})).Return(makeUser(otherID), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "admin cannot demote super admin",
			targetID: otherID.String(),
			claims:   adminClaims(adminID),
			body:     map[string]any{"role": "admin"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, otherID).
					Return(&domain.User{ID: otherID, Role: domain.UserRoleSuperAdmin}, nil)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "non-admin cannot change role",
			targetID:   regularID.String(),
			claims:     userClaims(regularID),
			body:       map[string]any{"role": "admin"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "non-admin cannot clear bundle4 assignments",
			targetID:   regularID.String(),
			claims:     userClaims(regularID),
			body:       map[string]any{"role_id": nil, "profile_id": nil},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "admin cannot assign super admin",
			targetID:   otherID.String(),
			claims:     adminClaims(adminID),
			body:       map[string]any{"role": "super_admin"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "super admin can assign super admin",
			targetID: otherID.String(),
			claims:   superAdminClaims(adminID),
			body:     map[string]any{"role": "super_admin"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Update", mock.Anything, otherID, mock.MatchedBy(func(p domain.UserPatch) bool {
					return p.Role != nil && *p.Role == superAdminRole
				})).Return(&domain.User{ID: otherID, Role: domain.UserRoleSuperAdmin}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 422 for invalid role",
			targetID:   otherID.String(),
			claims:     adminClaims(adminID),
			body:       map[string]any{"role": "owner"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:     "returns 422 for role constraint violations from repo",
			targetID: otherID.String(),
			claims:   superAdminClaims(adminID),
			body:     map[string]any{"role": "admin"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Update", mock.Anything, otherID, mock.AnythingOfType("domain.UserPatch")).
					Return(nil, &pgconn.PgError{Code: "23514", ConstraintName: "users_role_check"})
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "user cannot update other user",
			targetID:   otherID.String(),
			claims:     userClaims(regularID),
			body:       map[string]any{"name": "Updated"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "returns 401 without claims",
			targetID:   otherID.String(),
			claims:     nil,
			body:       map[string]any{"name": "Updated"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid uuid",
			targetID:   "not-a-uuid",
			claims:     adminClaims(adminID),
			body:       map[string]any{"name": "Updated"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
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

func TestUserHandler_UpdateDelegatedUserAdminCannotManagePlatformAdmins(t *testing.T) {
	userID := uuid.New()
	targetID := uuid.New()

	t.Run("cannot assign admin role", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		h := handler.NewUserHandler(mockRepo)

		bodyBytes, err := json.Marshal(map[string]any{"role": "admin"})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPatch, "/"+targetID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = withURLParam(req, "id", targetID.String())
		req = withClaims(req, userClaims(userID))
		req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
			UserID: userID,
			OrgID:  domain.DefaultOrgID,
			Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
				domain.ACLModuleUsers: {domain.ACLActionAdmin: true},
			},
		}))
		w := httptest.NewRecorder()

		h.Update(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		mockRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
		mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("cannot change existing admin role", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		mockRepo.On("GetByID", mock.Anything, targetID).
			Return(&domain.User{ID: targetID, Role: domain.UserRoleAdmin}, nil)
		h := handler.NewUserHandler(mockRepo)

		bodyBytes, err := json.Marshal(map[string]any{"role": "agent"})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPatch, "/"+targetID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = withURLParam(req, "id", targetID.String())
		req = withClaims(req, userClaims(userID))
		req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
			UserID: userID,
			OrgID:  domain.DefaultOrgID,
			Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
				domain.ACLModuleUsers: {domain.ACLActionAdmin: true},
			},
		}))
		w := httptest.NewRecorder()

		h.Update(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestUserHandler_UpdateSelfIncludesResolvedPermissions(t *testing.T) {
	userID := uuid.New()
	mockRepo := new(mocks.MockUserRepository)
	mockRepo.On("Update", mock.Anything, userID, mock.MatchedBy(func(p domain.UserPatch) bool {
		return p.Name != nil && *p.Name == "Delegated Admin"
	})).Return(&domain.User{ID: userID, Email: "delegate@example.com", Name: "Delegated Admin", Role: domain.UserRoleAgent}, nil)
	h := handler.NewUserHandler(mockRepo)

	bodyBytes, err := json.Marshal(map[string]any{"name": "Delegated Admin"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPatch, "/"+userID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", userID.String())
	req = withClaims(req, userClaims(userID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleUsers: {
				domain.ACLActionAdmin: true,
			},
			domain.ACLModuleSettings: {
				domain.ACLActionAdmin: true,
			},
		},
	}))
	w := httptest.NewRecorder()

	h.Update(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var got domain.User
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	require.True(t, got.Permissions[domain.ACLModuleUsers][domain.ACLActionAdmin])
	require.True(t, got.Permissions[domain.ACLModuleSettings][domain.ACLActionAdmin])
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_UpdateSelfCannotEscalateACLAssignmentWithOnlyUserAdmin(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()
	mockRepo := new(mocks.MockUserRepository)
	h := handler.NewUserHandler(mockRepo)

	bodyBytes, err := json.Marshal(map[string]any{"profile_id": profileID.String()})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPatch, "/"+userID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", userID.String())
	req = withClaims(req, userClaims(userID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleUsers: {domain.ACLActionAdmin: true},
		},
	}))
	w := httptest.NewRecorder()

	h.Update(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestUserHandler_UpdateSelfCannotEscalatePlatformRoleWithUserAdmin(t *testing.T) {
	userID := uuid.New()
	mockRepo := new(mocks.MockUserRepository)
	h := handler.NewUserHandler(mockRepo)

	bodyBytes, err := json.Marshal(map[string]any{"role": "admin"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPatch, "/"+userID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", userID.String())
	req = withClaims(req, userClaims(userID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleUsers: {domain.ACLActionAdmin: true},
		},
	}))
	w := httptest.NewRecorder()

	h.Update(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
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
			name:       "admin cannot self-delete",
			targetID:   adminID.String(),
			claims:     adminClaims(adminID),
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "non-admin gets 403",
			targetID:   otherID.String(),
			claims:     userClaims(uuid.New()),
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "no claims gets 403",
			targetID:   otherID.String(),
			claims:     nil,
			setupMock:  func(_ *mocks.MockUserRepository) {},
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
			name:       "returns 400 for invalid uuid",
			targetID:   "not-a-uuid",
			claims:     adminClaims(adminID),
			setupMock:  func(_ *mocks.MockUserRepository) {},
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
