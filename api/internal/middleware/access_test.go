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

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
)

type recordAccessCall struct {
	module domain.ACLModule
	id     uuid.UUID
	access domain.SharingAccessLevel
}

type relationshipAccessCall struct {
	id     uuid.UUID
	access domain.SharingAccessLevel
}

type fakeRecordAccessRepo struct {
	allowed             bool
	relationshipAllowed bool
	calls               []recordAccessCall
	relationshipCalls   []relationshipAccessCall
}

func (f *fakeRecordAccessRepo) CanAccessRecord(_ context.Context, module domain.ACLModule, id uuid.UUID, access domain.SharingAccessLevel) (bool, error) {
	f.calls = append(f.calls, recordAccessCall{module: module, id: id, access: access})
	return f.allowed, nil
}

func (f *fakeRecordAccessRepo) CanAccessAccountRelationship(_ context.Context, relationshipID uuid.UUID, access domain.SharingAccessLevel) (bool, error) {
	f.relationshipCalls = append(f.relationshipCalls, relationshipAccessCall{id: relationshipID, access: access})
	return f.relationshipAllowed, nil
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

func TestRequireModulePermissionAllowsSelfServiceUserUpdate(t *testing.T) {
	userID := uuid.New()
	handler := middleware.RequireModulePermission()(okHandler)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+userID.String(), nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: userID,
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleUsers: {domain.ACLActionRead: true},
		},
	}))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	otherID := uuid.New()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+otherID.String(), nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: userID,
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleUsers: {domain.ACLActionRead: true},
		},
	}))
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireModulePermissionAllowsSelfServicePreferenceRequests(t *testing.T) {
	userID := uuid.New()
	handler := middleware.RequireModulePermission()(okHandler)
	noSettingsAccess := &domain.AccessContext{
		UserID:      userID,
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{},
	}

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/settings/preferences"},
		{http.MethodPatch, "/api/v1/settings/preferences"},
		{http.MethodGet, "/api/v1/settings/calendar-preferences"},
		{http.MethodPatch, "/api/v1/settings/calendar-preferences"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
			UserID: userID,
			Role:   string(domain.UserRoleAgent),
		}))
		req = req.WithContext(domain.WithAccessContext(req.Context(), noSettingsAccess))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, tc.method+" "+tc.path)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings/company", nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: userID,
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithAccessContext(req.Context(), noSettingsAccess))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/settings/company", nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: userID,
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithAccessContext(req.Context(), noSettingsAccess))
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireModulePermissionChecksImportTargetModule(t *testing.T) {
	handler := middleware.RequireModulePermission()(okHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/accounts", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionCreate: true},
		},
	}))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/import/accounts", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleAccounts: {domain.ACLActionCreate: true},
		},
	}))
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/import/leads", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleLeads: {domain.ACLActionCreate: true},
		},
	}))
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequireModulePermissionChecksNestedChildModules(t *testing.T) {
	handler := middleware.RequireModulePermission()(okHandler)
	dealID := uuid.New()
	contactID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/deals/"+dealID.String()+"/quotes", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleDeals: {domain.ACLActionRead: true},
		},
	}))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/deals/"+dealID.String()+"/quotes", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleDeals:  {domain.ACLActionRead: true},
			domain.ACLModuleQuotes: {domain.ACLActionRead: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/contacts/"+contactID.String()+"/emails", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionRead: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/contacts/"+contactID.String()+"/emails", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionRead: true},
			domain.ACLModuleEmails:   {domain.ACLActionRead: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/deals/"+dealID.String()+"/quotes", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleQuotes: {domain.ACLActionRead: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireModulePermissionTreatsNestedChildMutationsAsParentUpdate(t *testing.T) {
	handler := middleware.RequireModulePermission()(okHandler)
	contactID := uuid.New()
	noteID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts/"+contactID.String()+"/notes", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionUpdate: true},
		},
	}))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/contacts/"+contactID.String()+"/notes", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionCreate: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/contacts/"+contactID.String()+"/notes/"+noteID.String(), nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionUpdate: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/contacts", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionUpdate: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireModulePermissionChecksConversionCreateTargets(t *testing.T) {
	handler := middleware.RequireModulePermission()(okHandler)
	leadID := uuid.New()
	contactID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/leads/"+leadID.String()+"/convert", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleLeads:    {domain.ACLActionUpdate: true},
			domain.ACLModuleContacts: {domain.ACLActionCreate: true},
			domain.ACLModuleAccounts: {domain.ACLActionCreate: true},
		},
	}))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/leads/"+leadID.String()+"/convert", nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleLeads:    {domain.ACLActionUpdate: true},
			domain.ACLModuleContacts: {domain.ACLActionCreate: true},
			domain.ACLModuleAccounts: {domain.ACLActionCreate: true},
			domain.ACLModuleDeals:    {domain.ACLActionCreate: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/contacts/"+contactID.String()+"/convert", strings.NewReader(`{"create_deal":true}`))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionUpdate: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/contacts/"+contactID.String()+"/convert", strings.NewReader(`{"create_deal":false}`))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionUpdate: true},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
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

func TestRequireFieldWriteAccessChecksNestedChildModuleFields(t *testing.T) {
	dealID := uuid.New()
	handler := middleware.RequireFieldWriteAccess()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"amount":100}`, string(body))
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/deals/"+dealID.String()+"/quotes", strings.NewReader(`{"amount":100}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		FieldWrite: map[domain.ACLModule]map[string]bool{
			domain.ACLModuleQuotes: {"amount": false},
		},
	}))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "field_write_denied")

	req = httptest.NewRequest(http.MethodPost, "/api/v1/deals/"+dealID.String()+"/quotes", strings.NewReader(`{"amount":100}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		FieldWrite: map[domain.ACLModule]map[string]bool{
			domain.ACLModuleDeals: {"amount": false},
		},
	}))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestRequireFieldWriteAccessAllowsSelfServicePreferenceUpdates(t *testing.T) {
	userID := uuid.New()
	handler := middleware.RequireFieldWriteAccess()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"calendar_default_view":"week"}`, string(body))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings/preferences", strings.NewReader(`{"calendar_default_view":"week"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: userID,
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		FieldWrite: map[domain.ACLModule]map[string]bool{
			domain.ACLModuleSettings: {"calendar_default_view": false},
		},
	}))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodPatch, "/api/v1/settings/company", strings.NewReader(`{"calendar_default_view":"week"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: userID,
		Role:   string(domain.UserRoleAgent),
	}))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: userID,
		FieldWrite: map[domain.ACLModule]map[string]bool{
			domain.ACLModuleSettings: {"calendar_default_view": false},
		},
	}))
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
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

func TestRequireNestedParentRecordAccessChecksAccountRelationshipRoutes(t *testing.T) {
	relationshipID := uuid.New()
	repo := &fakeRecordAccessRepo{relationshipAllowed: false}
	handler := middleware.RequireNestedParentRecordAccess(repo)(okHandler)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/relationships/"+relationshipID.String(), nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Empty(t, repo.calls)
	require.Len(t, repo.relationshipCalls, 1)
	require.Equal(t, relationshipID, repo.relationshipCalls[0].id)
	require.Equal(t, domain.SharingAccessWrite, repo.relationshipCalls[0].access)

	repo = &fakeRecordAccessRepo{relationshipAllowed: true}
	handler = middleware.RequireNestedParentRecordAccess(repo)(okHandler)
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/relationships/"+relationshipID.String(), nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, repo.calls)
	require.Len(t, repo.relationshipCalls, 1)
}
