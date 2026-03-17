package handler_test

import (
	"net/http"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/middleware"
)

// withClaims injects JWT claims into the request context, mimicking the Authenticate middleware.
func withClaims(r *http.Request, claims *auth.Claims) *http.Request {
	return r.WithContext(middleware.WithClaims(r.Context(), claims))
}
