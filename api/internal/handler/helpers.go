package handler

import (
	"net/http"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/middleware"
)

// mustClaims extracts JWT claims from the request context (set by middleware.Authenticate).
func mustClaims(r *http.Request) *auth.Claims {
	claims, _ := middleware.ClaimsFromContext(r)
	return claims
}
