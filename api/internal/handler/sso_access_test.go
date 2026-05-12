package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
)

type fakeSSOConfigRepo struct {
	cfg *domain.SSOConfig
}

func (f fakeSSOConfigRepo) Upsert(_ context.Context, cfg *domain.SSOConfig) (*domain.SSOConfig, error) {
	return cfg, nil
}

func (f fakeSSOConfigRepo) GetByOrgID(_ context.Context, _ uuid.UUID) (*domain.SSOConfig, error) {
	return f.cfg, nil
}

func (f fakeSSOConfigRepo) GetByOrgSlug(_ context.Context, _ string) (*domain.SSOConfig, error) {
	return f.cfg, nil
}

func TestSSOHandler_OrgSSORequiresSameOrgAdminAccess(t *testing.T) {
	orgID := uuid.New()
	otherOrgID := uuid.New()
	h := handler.NewSSOHandler(fakeSSOConfigRepo{
		cfg: &domain.SSOConfig{OrgID: orgID, Provider: "google", ClientID: "client", Enabled: true},
	}, nil, nil, nil, "", "")

	t.Run("same org delegated integrations admin can read", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/orgs/"+orgID.String()+"/sso", nil)
		req = withURLParam(req, "orgId", orgID.String())
		req = withClaims(req, &auth.Claims{UserID: uuid.New(), OrgID: orgID, Role: string(domain.UserRoleAgent)})
		req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
		req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
			UserID: uuid.New(),
			OrgID:  orgID,
			Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
				domain.ACLModuleIntegrations: {domain.ACLActionAdmin: true},
			},
		}))
		w := httptest.NewRecorder()

		h.GetOrgSSO(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("same org agent without integrations admin is denied", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/orgs/"+orgID.String()+"/sso", nil)
		req = withURLParam(req, "orgId", orgID.String())
		req = withClaims(req, &auth.Claims{UserID: uuid.New(), OrgID: orgID, Role: string(domain.UserRoleAgent)})
		req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
		w := httptest.NewRecorder()

		h.GetOrgSSO(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("platform admin cannot patch a different org", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/orgs/"+otherOrgID.String()+"/sso", nil)
		req = withURLParam(req, "orgId", otherOrgID.String())
		req = withClaims(req, &auth.Claims{UserID: uuid.New(), OrgID: orgID, Role: string(domain.UserRoleAdmin)})
		req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
		w := httptest.NewRecorder()

		h.UpdateOrgSSO(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
	})
}
