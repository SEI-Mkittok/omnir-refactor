package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
)

func TestRequireModulePermissionUsesResolvedProfilePermissions(t *testing.T) {
	handler := middleware.RequireModulePermission()(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionRead: true},
		},
	}))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/contacts/contact-1", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionRead: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireFieldWriteAccessRejectsDeniedFieldAndRestoresAllowedBody(t *testing.T) {
	access := &domain.AccessContext{
		FieldWrite: map[domain.ACLModule]map[string]bool{
			domain.ACLModuleContacts: {"email": false},
		},
	}
	handler := middleware.RequireFieldWriteAccess()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"first_name":"Ada"}`, string(body))
		w.WriteHeader(http.StatusCreated)
	}))

	allowedReq := httptest.NewRequest(http.MethodPost, "/api/v1/contacts", strings.NewReader(`{"first_name":"Ada"}`))
	allowedReq.Header.Set("Content-Type", "application/json")
	allowedReq = allowedReq.WithContext(domain.WithAccessContext(allowedReq.Context(), access))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, allowedReq)
	require.Equal(t, http.StatusCreated, w.Code)

	deniedReq := httptest.NewRequest(http.MethodPost, "/api/v1/contacts", strings.NewReader(`{"email":"ada@example.test"}`))
	deniedReq.Header.Set("Content-Type", "application/json")
	deniedReq = deniedReq.WithContext(domain.WithAccessContext(deniedReq.Context(), access))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, deniedReq)
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "field_write_denied")
}

func TestPortalRoutesBypassInternalModuleACL(t *testing.T) {
	moduleHandler := middleware.RequireModulePermission()(okHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/portal/tickets", nil)
	w := httptest.NewRecorder()
	moduleHandler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	fieldHandler := middleware.RequireFieldWriteAccess()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"status":"open"}`, string(body))
		w.WriteHeader(http.StatusCreated)
	}))
	fieldReq := httptest.NewRequest(http.MethodPost, "/api/v1/portal/tickets", strings.NewReader(`{"status":"open"}`))
	fieldReq.Header.Set("Content-Type", "application/json")
	fieldReq = fieldReq.WithContext(domain.WithAccessContext(fieldReq.Context(), &domain.AccessContext{
		FieldWrite: map[domain.ACLModule]map[string]bool{
			domain.ACLModuleTickets: {"status": false},
		},
	}))
	w = httptest.NewRecorder()
	fieldHandler.ServeHTTP(w, fieldReq)
	require.Equal(t, http.StatusCreated, w.Code)
}
