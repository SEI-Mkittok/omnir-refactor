package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
)

// injectClaims injects JWT claims directly into the request context,
// simulating what the Authenticate middleware does.
func injectClaims(r *http.Request, role string) *http.Request {
	claims := &auth.Claims{Role: role}
	ctx := context.WithValue(r.Context(), contextKey("claims"), claims)
	return r.WithContext(ctx)
}

// contextKey mirrors the unexported type in middleware — we use the exported
// ClaimsFromContext helper instead of the key directly in tests.
type contextKey string

// okHandler is a simple handler that always returns 200.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestRequireRole(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret")

	// Helper: build a request with a real JWT in an httpOnly cookie.
	makeAuthedRequest := func(role string) *http.Request {
		token, err := jwtSvc.Issue(auth.Claims{Role: role}, 60*1000000000 /* 1 min */)
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		return req
	}

	tests := []struct {
		name         string
		callerRole   string
		allowedRoles []domain.UserRole
		wantStatus   int
	}{
		{
			name:         "admin allowed when admin required",
			callerRole:   "admin",
			allowedRoles: []domain.UserRole{domain.UserRoleAdmin},
			wantStatus:   http.StatusOK,
		},
		{
			name:         "agent allowed when admin or agent required",
			callerRole:   "agent",
			allowedRoles: []domain.UserRole{domain.UserRoleAdmin, domain.UserRoleAgent},
			wantStatus:   http.StatusOK,
		},
		{
			name:         "admin allowed when admin or agent required",
			callerRole:   "admin",
			allowedRoles: []domain.UserRole{domain.UserRoleAdmin, domain.UserRoleAgent},
			wantStatus:   http.StatusOK,
		},
		{
			name:         "client forbidden when admin or agent required",
			callerRole:   "client",
			allowedRoles: []domain.UserRole{domain.UserRoleAdmin, domain.UserRoleAgent},
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "client forbidden when admin required",
			callerRole:   "client",
			allowedRoles: []domain.UserRole{domain.UserRoleAdmin},
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "agent forbidden when admin only required",
			callerRole:   "agent",
			allowedRoles: []domain.UserRole{domain.UserRoleAdmin},
			wantStatus:   http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Chain: Authenticate → RequireRole → okHandler
			chain := middleware.Authenticate(jwtSvc, nil, nil)(
				middleware.RequireRole(tt.allowedRoles...)(okHandler),
			)

			req := makeAuthedRequest(tt.callerRole)
			w := httptest.NewRecorder()
			chain.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestRequireRole_NoClaimsReturns401(t *testing.T) {
	// Call RequireRole directly without Authenticate — no claims in context.
	handler := middleware.RequireRole(domain.UserRoleAdmin)(okHandler)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
