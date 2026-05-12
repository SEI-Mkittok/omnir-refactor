package handler

import (
	"net/http"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
)

func hasModuleAdminAccess(r *http.Request, module domain.ACLModule) bool {
	if claims, ok := middleware.ClaimsFromContext(r); ok && domain.IsAdminRole(claims.Role) {
		return true
	}
	if access, ok := domain.AccessContextFromContext(r.Context()); ok {
		return access.HasPermission(module, domain.ACLActionAdmin)
	}
	return false
}
