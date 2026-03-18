package middleware

import (
	"net/http"

	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
)

// OrgScope validates that a request carries a resolvable org_id and enforces
// deployment-mode rules:
//
//   - single:      if no org_id in context, fall back to domain.DefaultOrgID.
//   - saas:        org_id must be present in the JWT claims; 401 otherwise.
//   - multitenant: alias for saas (legacy mode name).
//   - enterprise:  same as saas.
//
// OrgScope must run after Authenticate so that JWT claims (and hence org_id)
// are already stored in the context.
func OrgScope(mode config.OrgMode) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, hasOrg := domain.OrgIDFromContext(r.Context())

			if !hasOrg {
				if mode == config.OrgModeSingle {
					// Self-hosted single-tenant: pin every request to the default org.
					ctx := domain.WithOrgID(r.Context(), domain.DefaultOrgID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				// SaaS / enterprise modes require the JWT to carry an org_id.
				http.Error(w, `{"error":"org_id required"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
