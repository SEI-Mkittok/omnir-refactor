package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

type contextKey string

const claimsKey contextKey = "claims"

// Authenticate validates credentials and injects Claims into the context.
// It accepts credentials in two forms (tried in order):
//  1. access_token httpOnly cookie — standard browser session JWT
//  2. Authorization: Bearer <token> header — either a JWT or an API key (ak_...)
//
// Returns 401 if no valid credential is found.
func Authenticate(jwtSvc *auth.JWTService, apiKeyRepo repository.APIKeyRepository, userRepo repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Try JWT cookie.
			if cookie, err := r.Cookie("access_token"); err == nil {
				if claims, err := jwtSvc.Verify(cookie.Value); err == nil {
					next.ServeHTTP(w, withClaims(r, claims))
					return
				}
			}

			// 2. Try Authorization: Bearer header.
			if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")

				if strings.HasPrefix(token, "ak_") {
					// API key path.
					if apiKeyRepo != nil {
						if ok := authenticateAPIKey(w, r, next, token, apiKeyRepo, userRepo); ok {
							return
						}
					}
				} else {
					// JWT in Authorization header (e.g. test clients, non-browser callers).
					if claims, err := jwtSvc.Verify(token); err == nil {
						next.ServeHTTP(w, withClaims(r, claims))
						return
					}
				}
			}

			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		})
	}
}

// authenticateAPIKey validates an API key token, updates last_used_at, and calls next.
// Returns true if the request was handled (either successfully or with 401).
func authenticateAPIKey(w http.ResponseWriter, r *http.Request, next http.Handler, plaintext string, apiKeyRepo repository.APIKeyRepository, userRepo repository.UserRepository) bool { //nolint:unparam // always true by design — every code path handles the request
	h := sha256.Sum256([]byte(plaintext))
	keyHash := hex.EncodeToString(h[:])

	apiKey, err := apiKeyRepo.GetByHash(r.Context(), keyHash)
	if err != nil || apiKey == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return true
	}

	// Reject expired or revoked keys.
	if apiKey.RevokedAt != nil || (apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt)) {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return true
	}

	// Update last_used_at asynchronously — don't block the request.
	go func() {
		_ = apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)
	}()

	// Resolve the key owner's current role.
	userCtx := domain.WithOrgID(r.Context(), apiKey.OrgID)
	user, err := userRepo.GetByID(userCtx, apiKey.CreatedBy)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return true
	}

	// Enforce scope restrictions before passing the request on.
	// "read" scope (without "write") allows only safe/idempotent methods.
	hasWrite := false
	hasRead := false
	for _, s := range apiKey.Scopes {
		switch s {
		case "write":
			hasWrite = true
		case "read":
			hasRead = true
		}
	}
	if hasRead && !hasWrite {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			// allowed
		default:
			http.Error(w, `{"error":"forbidden","code":"insufficient_scope"}`, http.StatusForbidden)
			return true
		}
	}

	claims := &auth.Claims{
		UserID:       apiKey.CreatedBy,
		OrgID:        apiKey.OrgID,
		Role:         string(user.Role),
		APIKeyScopes: apiKey.Scopes,
	}
	next.ServeHTTP(w, withClaims(r, claims))
	return true
}

// withClaims returns a new request with claims and org_id injected into the context.
func withClaims(r *http.Request, claims *auth.Claims) *http.Request {
	ctx := context.WithValue(r.Context(), claimsKey, claims)
	if claims.OrgID != (uuid.UUID{}) {
		ctx = domain.WithOrgID(ctx, claims.OrgID)
	}
	return r.WithContext(ctx)
}

// ClaimsFromContext retrieves JWT claims stored by the Authenticate middleware.
func ClaimsFromContext(r *http.Request) (*auth.Claims, bool) {
	c, ok := r.Context().Value(claimsKey).(*auth.Claims)
	return c, ok
}

// WithClaims injects claims into a context. Intended for use in tests.
func WithClaims(ctx context.Context, c *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}
