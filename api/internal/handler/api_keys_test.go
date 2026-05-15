package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
)

type fakeAPIKeyHandlerRepo struct {
	createdKey  *domain.APIKey
	revokedID   uuid.UUID
	revokedOrg  uuid.UUID
	createHash  string
	revokeError error
}

func (f *fakeAPIKeyHandlerRepo) Create(_ context.Context, k *domain.APIKey, keyHash string) (*domain.APIKey, error) {
	f.createHash = keyHash
	k.ID = uuid.New()
	f.createdKey = k
	return k, nil
}

func (f *fakeAPIKeyHandlerRepo) GetByHash(context.Context, string) (*domain.APIKey, error) {
	return nil, nil
}

func (f *fakeAPIKeyHandlerRepo) List(context.Context, uuid.UUID) ([]*domain.APIKey, error) {
	return nil, nil
}

func (f *fakeAPIKeyHandlerRepo) Revoke(_ context.Context, id, orgID uuid.UUID) error {
	f.revokedID = id
	f.revokedOrg = orgID
	return f.revokeError
}

func (f *fakeAPIKeyHandlerRepo) UpdateLastUsed(context.Context, uuid.UUID) error {
	return nil
}

type fakeAuditRepo struct {
	entries chan domain.AuditEntry
}

func (f *fakeAuditRepo) Append(_ context.Context, entry domain.AuditEntry) error {
	f.entries <- entry
	return nil
}

func (f *fakeAuditRepo) GetByID(context.Context, uuid.UUID, uuid.UUID) (*domain.AuditLog, error) {
	return nil, nil
}

func (f *fakeAuditRepo) List(context.Context, domain.AuditLogFilter) ([]*domain.AuditLog, int, error) {
	return nil, 0, nil
}

func TestAPIKeyHandlerCreateReturnsPlaintextOnceAndAudits(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	repo := &fakeAPIKeyHandlerRepo{}
	audit := &fakeAuditRepo{entries: make(chan domain.AuditEntry, 1)}
	h := handler.NewAPIKeyHandler(repo).WithAuditLog(audit)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", strings.NewReader(`{"name":"  Partner Sync  ","scopes":["read"]}`))
	req = withClaims(req, &auth.Claims{OrgID: orgID, UserID: userID, Role: string(domain.UserRoleAdmin)})
	w := httptest.NewRecorder()

	h.Create(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotEmpty(t, repo.createHash)
	require.Equal(t, "Partner Sync", repo.createdKey.Name)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.NotEmpty(t, body["key"])
	require.NotContains(t, body, "key_hash")

	select {
	case entry := <-audit.entries:
		require.Equal(t, orgID, entry.OrgID)
		require.Equal(t, userID, *entry.UserID)
		require.Equal(t, domain.AuditActionCreated, entry.Action)
		require.Equal(t, domain.AuditEntityAPIKey, entry.EntityType)
		require.Equal(t, repo.createdKey.ID, *entry.EntityID)
		require.Equal(t, "Partner Sync", *entry.EntityName)
	case <-time.After(time.Second):
		t.Fatal("expected API key creation audit entry")
	}
}

func TestAPIKeyHandlerRevokeAuditsScopedDelete(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	keyID := uuid.New()
	repo := &fakeAPIKeyHandlerRepo{}
	audit := &fakeAuditRepo{entries: make(chan domain.AuditEntry, 1)}
	h := handler.NewAPIKeyHandler(repo).WithAuditLog(audit)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/api-keys/"+keyID.String(), nil)
	req = withClaims(req, &auth.Claims{OrgID: orgID, UserID: userID, Role: string(domain.UserRoleAdmin)})
	req = withURLParam(req, "id", keyID.String())
	w := httptest.NewRecorder()

	h.Revoke(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	require.Equal(t, keyID, repo.revokedID)
	require.Equal(t, orgID, repo.revokedOrg)

	select {
	case entry := <-audit.entries:
		require.Equal(t, orgID, entry.OrgID)
		require.Equal(t, userID, *entry.UserID)
		require.Equal(t, domain.AuditActionDeleted, entry.Action)
		require.Equal(t, domain.AuditEntityAPIKey, entry.EntityType)
		require.Equal(t, keyID, *entry.EntityID)
	case <-time.After(time.Second):
		t.Fatal("expected API key revoke audit entry")
	}
}
