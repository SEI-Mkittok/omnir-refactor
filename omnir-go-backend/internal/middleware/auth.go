package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
)

type contextKey string

const claimsKey contextKey = "claims"

// Authenticate validates the Bearer token and injects JWT claims + org_id into the context.
// Returns 401 if the token is missing or invalid.
func Authenticate(svc *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := svc.Verify(token)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			if claims.OrgID != (uuid.UUID{}) {
				ctx = domain.WithOrgID(ctx, claims.OrgID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
