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

func TestHasModuleAdminAccessAllowsPlatformAdmins(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: uuid.New(),
		Role:   string(domain.UserRoleAdmin),
	}))

	require.True(t, hasModuleAdminAccess(req, domain.ACLModuleSettings))
}

func TestHasModuleAdminAccessAllowsMatchingACLAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: uuid.New(),
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleAPIKeys: {domain.ACLActionAdmin: true},
		},
	}))

	require.True(t, hasModuleAdminAccess(req, domain.ACLModuleAPIKeys))
}

func TestHasModuleAdminAccessRequiresMatchingACLModule(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: uuid.New(),
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleSettings: {domain.ACLActionAdmin: true},
		},
	}))

	require.False(t, hasModuleAdminAccess(req, domain.ACLModuleIntegrations))
}
