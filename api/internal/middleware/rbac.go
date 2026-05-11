package middleware

import (
	"net/http"

	"github.com/omnir/crm-api/internal/domain"
)

// RequireRole returns a middleware that allows only the specified roles.
// It reads claims injected by Authenticate and returns 403 if the caller's
// role is not in the allowed set. Must be used after Authenticate.
func RequireRole(roles ...domain.UserRole) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[string(r)] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r)
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			_, permitted := allowed[claims.Role]
			if !permitted && claims.Role == string(domain.UserRoleSuperAdmin) {
				_, permitted = allowed[string(domain.UserRoleAdmin)]
			}
			if !permitted {
				if module, action, ok := moduleActionFromRequest(r); ok {
					if access, hasAccess := domain.AccessContextFromContext(r.Context()); hasAccess && access.HasPermission(module, action) {
						permitted = true
					}
				}
			}
			if !permitted {
				http.Error(w, `{"error":"forbidden","code":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}
