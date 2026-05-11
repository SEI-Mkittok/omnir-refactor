package middleware_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
)

type recordAccessCall struct {
	module domain.ACLModule
	id     uuid.UUID
	access domain.SharingAccessLevel
}

type fakeRecordAccessRepo struct {
	allowed bool
	calls   []recordAccessCall
}

func (f *fakeRecordAccessRepo) CanAccessRecord(_ context.Context, module domain.ACLModule, id uuid.UUID, access domain.SharingAccessLevel) (bool, error) {
	f.calls = append(f.calls, recordAccessCall{module: module, id: id, access: access})
	return f.allowed, nil
}

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

func TestRequireNestedParentRecordAccessChecksReadVisibility(t *testing.T) {
	parentID := uuid.New()
	repo := &fakeRecordAccessRepo{allowed: false}
	handler := middleware.RequireNestedParentRecordAccess(repo)(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts/"+parentID.String()+"/notes", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Len(t, repo.calls, 1)
	require.Equal(t, domain.ACLModuleContacts, repo.calls[0].module)
	require.Equal(t, parentID, repo.calls[0].id)
	require.Equal(t, domain.SharingAccessRead, repo.calls[0].access)
}

func TestRequireNestedParentRecordAccessUsesWriteVisibilityForMutations(t *testing.T) {
	parentID := uuid.New()
	repo := &fakeRecordAccessRepo{allowed: true}
	handler := middleware.RequireNestedParentRecordAccess(repo)(okHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/deals/"+parentID.String()+"/quotes", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Len(t, repo.calls, 1)
	require.Equal(t, domain.ACLModuleDeals, repo.calls[0].module)
	require.Equal(t, parentID, repo.calls[0].id)
	require.Equal(t, domain.SharingAccessWrite, repo.calls[0].access)
}

func TestRequireNestedParentRecordAccessSkipsTopLevelRecords(t *testing.T) {
	parentID := uuid.New()
	repo := &fakeRecordAccessRepo{allowed: false}
	handler := middleware.RequireNestedParentRecordAccess(repo)(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts/"+parentID.String(), nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, repo.calls)
}
