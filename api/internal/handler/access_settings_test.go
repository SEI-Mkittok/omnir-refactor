package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func TestAccessSettingsRoutersRequireAdminForReads(t *testing.T) {
	userID := uuid.New()
	handler := NewAccessSettingsHandler(nil).RolesRouter()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: userID,
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleSettings: {domain.ACLActionRead: true},
		},
	}))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAccessSettingsAdminAllowsSettingsAdminAndPlatformAdmin(t *testing.T) {
	okHandler := requireAccessSettingsAdmin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("settings admin permission", func(t *testing.T) {
		userID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
			UserID: userID,
			Role:   string(domain.UserRoleAgent),
		}))
		req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
			UserID: userID,
			Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
				domain.ACLModuleSettings: {domain.ACLActionAdmin: true},
			},
		}))
		w := httptest.NewRecorder()

		okHandler.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("platform admin role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
			UserID: uuid.New(),
			Role:   string(domain.UserRoleAdmin),
		}))
		w := httptest.NewRecorder()

		okHandler.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAccessSettingsGroupsMemberCandidatesUseSettingsAdmin(t *testing.T) {
	userID := uuid.New()
	memberID := uuid.New()
	users := new(mocks.MockUserRepository)
	users.On("List", mock.Anything, mock.MatchedBy(func(f domain.UserFilter) bool {
		return f.Page == 1 && f.Limit == 500 && f.Sort == "name" && f.Order == "asc"
	})).Return([]*domain.User{
		{ID: memberID, OrgID: domain.DefaultOrgID, Name: "Member One", Email: "member@example.com", Role: domain.UserRoleAgent},
	}, 1, nil)

	handler := NewAccessSettingsHandler(nil, users).GroupsRouter()
	req := httptest.NewRequest(http.MethodGet, "/member-candidates", nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: userID,
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithOrgID(req.Context(), domain.DefaultOrgID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleSettings: {domain.ACLActionAdmin: true},
		},
	}))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data []domain.User `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Data, 1)
	require.Equal(t, memberID, resp.Data[0].ID)
	users.AssertExpectations(t)
}

type nilGroupsAccessRepo struct {
	repository.AccessRepository
}

func (nilGroupsAccessRepo) ListGroups(context.Context, uuid.UUID) ([]*domain.ACLGroup, error) {
	return nil, nil
}

func TestAccessSettingsGroupsListNormalizesNilSlice(t *testing.T) {
	handler := NewAccessSettingsHandler(nilGroupsAccessRepo{}).GroupsRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: uuid.New(),
		OrgID:  domain.DefaultOrgID,
		Role:   string(domain.UserRoleAdmin),
	}))
	req = req.WithContext(domain.WithOrgID(req.Context(), domain.DefaultOrgID))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data []domain.ACLGroup `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotNil(t, resp.Data)
	require.Len(t, resp.Data, 0)
}
