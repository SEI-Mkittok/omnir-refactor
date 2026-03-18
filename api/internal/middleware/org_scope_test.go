package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
)

// orgFromContext returns the org_id that OrgScope stored in the request context.
func orgFromContext(t *testing.T, mode config.OrgMode, orgInToken *uuid.UUID) (uuid.UUID, int) {
	t.Helper()

	var captured uuid.UUID

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := domain.OrgIDFromContext(r.Context())
		captured = id
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.OrgScope(mode)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if orgInToken != nil {
		// Simulate what Authenticate would have put in context.
		ctx := domain.WithOrgID(req.Context(), *orgInToken)
		req = req.WithContext(ctx)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return captured, rec.Code
}

func TestOrgScope_SingleMode_NoOrgInToken(t *testing.T) {
	got, code := orgFromContext(t, config.OrgModeSingle, nil)
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, domain.DefaultOrgID, got, "should fall back to DefaultOrgID")
}

func TestOrgScope_SingleMode_OrgInToken(t *testing.T) {
	orgID := uuid.New()
	got, code := orgFromContext(t, config.OrgModeSingle, &orgID)
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, orgID, got, "should use token org_id when provided")
}

func TestOrgScope_Multitenant_OrgInToken(t *testing.T) {
	orgID := uuid.New()
	got, code := orgFromContext(t, config.OrgModeMultitenant, &orgID)
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, orgID, got)
}

func TestOrgScope_Multitenant_NoOrgInToken(t *testing.T) {
	_, code := orgFromContext(t, config.OrgModeMultitenant, nil)
	assert.Equal(t, http.StatusUnauthorized, code, "multitenant mode requires org_id in JWT")
}

func TestOrgScope_Enterprise_NoOrgInToken(t *testing.T) {
	_, code := orgFromContext(t, config.OrgModeEnterprise, nil)
	assert.Equal(t, http.StatusUnauthorized, code, "enterprise mode requires org_id in JWT")
}

func TestOrgScope_Enterprise_OrgInToken(t *testing.T) {
	orgID := uuid.New()
	got, code := orgFromContext(t, config.OrgModeEnterprise, &orgID)
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, orgID, got)
}

// TestOrgScope_AfterAuthenticate verifies the realistic middleware chain:
// Authenticate → OrgScope → handler.
func TestOrgScope_AfterAuthenticate(t *testing.T) {
	orgID := uuid.New()
	jwtSvc := auth.NewJWTService("test-secret")

	token, err := jwtSvc.Issue(auth.Claims{
		UserID: uuid.New(),
		OrgID:  orgID,
		Role:   "admin",
	}, 10*60*1e9)
	require.NoError(t, err)

	var capturedOrg uuid.UUID
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := domain.OrgIDFromContext(r.Context())
		capturedOrg = id
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(jwtSvc)(
		middleware.OrgScope(config.OrgModeMultitenant)(inner),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, orgID, capturedOrg)
}
