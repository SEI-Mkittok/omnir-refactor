package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
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
