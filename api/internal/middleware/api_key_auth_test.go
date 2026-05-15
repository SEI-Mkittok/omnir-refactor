package middleware_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
)

type fakeAPIKeyRepo struct {
	keys     map[string]*domain.APIKey
	lastUsed chan uuid.UUID
}

func (f *fakeAPIKeyRepo) Create(context.Context, *domain.APIKey, string) (*domain.APIKey, error) {
	return nil, nil
}

func (f *fakeAPIKeyRepo) GetByHash(_ context.Context, keyHash string) (*domain.APIKey, error) {
	return f.keys[keyHash], nil
}

func (f *fakeAPIKeyRepo) List(context.Context, uuid.UUID) ([]*domain.APIKey, error) {
	return nil, nil
}

func (f *fakeAPIKeyRepo) Revoke(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (f *fakeAPIKeyRepo) UpdateLastUsed(_ context.Context, id uuid.UUID) error {
	if f.lastUsed != nil {
		f.lastUsed <- id
	}
	return nil
}

type fakeAuthUserRepo struct {
	user *domain.User
	err  error
}

func (f *fakeAuthUserRepo) CountAll(context.Context) (int, error) {
	return 0, nil
}

func (f *fakeAuthUserRepo) HasAdminUser(context.Context) (bool, error) {
	return false, nil
}

func (f *fakeAuthUserRepo) Create(context.Context, *domain.User, string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeAuthUserRepo) FindByEmail(context.Context, string) (*domain.User, string, error) {
	return nil, "", nil
}

func (f *fakeAuthUserRepo) GetByID(context.Context, uuid.UUID) (*domain.User, error) {
	return f.user, f.err
}

func (f *fakeAuthUserRepo) Update(context.Context, uuid.UUID, domain.UserPatch) (*domain.User, error) {
	return nil, nil
}

func (f *fakeAuthUserRepo) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (f *fakeAuthUserRepo) List(context.Context, domain.UserFilter) ([]*domain.User, int, error) {
	return nil, 0, nil
}

func apiKeyHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func TestAuthenticateAPIKeyRejectsRevokedAndExpiredKeys(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-time.Minute)
	for _, tt := range []struct {
		name string
		key  *domain.APIKey
	}{
		{
			name: "revoked",
			key: &domain.APIKey{
				ID:        uuid.New(),
				OrgID:     uuid.New(),
				CreatedBy: uuid.New(),
				RevokedAt: &now,
			},
		},
		{
			name: "expired",
			key: &domain.APIKey{
				ID:        uuid.New(),
				OrgID:     uuid.New(),
				CreatedBy: uuid.New(),
				ExpiresAt: &past,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			token := "ak_" + tt.name
			apiKeys := &fakeAPIKeyRepo{keys: map[string]*domain.APIKey{apiKeyHash(token): tt.key}}
			users := &fakeAuthUserRepo{user: &domain.User{ID: tt.key.CreatedBy, OrgID: tt.key.OrgID, Role: domain.UserRoleAdmin}}
			called := false
			chain := middleware.Authenticate(auth.NewJWTService("test"), apiKeys, users)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			chain.ServeHTTP(w, req)

			require.Equal(t, http.StatusUnauthorized, w.Code)
			require.False(t, called)
		})
	}
}

func TestAuthenticateAPIKeyEnforcesReadOnlyScope(t *testing.T) {
	token := "ak_readonly"
	orgID := uuid.New()
	userID := uuid.New()
	apiKeyID := uuid.New()
	apiKeys := &fakeAPIKeyRepo{
		keys: map[string]*domain.APIKey{
			apiKeyHash(token): {
				ID:        apiKeyID,
				OrgID:     orgID,
				CreatedBy: userID,
				Scopes:    []string{"read"},
			},
		},
		lastUsed: make(chan uuid.UUID, 1),
	}
	users := &fakeAuthUserRepo{user: &domain.User{ID: userID, OrgID: orgID, Role: domain.UserRoleAdmin}}
	chain := middleware.Authenticate(auth.NewJWTService("test"), apiKeys, users)(okHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	select {
	case usedID := <-apiKeys.lastUsed:
		require.Equal(t, apiKeyID, usedID)
	case <-time.After(time.Second):
		t.Fatal("expected last_used_at update for authenticated API key attempt")
	}
}

func TestAuthenticateAPIKeyInjectsClaimsForValidKey(t *testing.T) {
	token := "ak_valid"
	orgID := uuid.New()
	userID := uuid.New()
	apiKeys := &fakeAPIKeyRepo{
		keys: map[string]*domain.APIKey{
			apiKeyHash(token): {
				ID:        uuid.New(),
				OrgID:     orgID,
				CreatedBy: userID,
				Scopes:    []string{"read", "write"},
			},
		},
		lastUsed: make(chan uuid.UUID, 1),
	}
	users := &fakeAuthUserRepo{user: &domain.User{ID: userID, OrgID: orgID, Role: domain.UserRoleAgent}}
	chain := middleware.Authenticate(auth.NewJWTService("test"), apiKeys, users)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.ClaimsFromContext(r)
		require.True(t, ok)
		require.Equal(t, userID, claims.UserID)
		require.Equal(t, orgID, claims.OrgID)
		require.Equal(t, string(domain.UserRoleAgent), claims.Role)
		require.Equal(t, []string{"read", "write"}, claims.APIKeyScopes)
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}
