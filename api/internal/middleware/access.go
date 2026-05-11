package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// ResolveAccess loads the Bundle 4 RBAC/sharing context once per authenticated
// request so handlers and repositories can enforce permissions consistently.
func ResolveAccess(repo repository.AccessRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if repo == nil {
				next.ServeHTTP(w, r)
				return
			}
			claims, ok := ClaimsFromContext(r)
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			orgID, _ := domain.OrgIDFromContext(r.Context())
			access, err := repo.ResolveAccess(r.Context(), claims.UserID, orgID, claims.Role)
			if err != nil {
				http.Error(w, `{"error":"access_resolution_failed"}`, http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r.WithContext(domain.WithAccessContext(r.Context(), access)))
		})
	}
}

// RequireModulePermission enforces profile permissions for authenticated
// /api/v1 modules. Unknown compatibility routes are allowed to keep legacy
// behavior until they are mapped explicitly.
func RequireModulePermission() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			module, action, ok := moduleActionFromRequest(r)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			access, ok := domain.AccessContextFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"forbidden","code":"access_context_required"}`, http.StatusForbidden)
				return
			}
			if !access.HasPermission(module, action) {
				http.Error(w, `{"error":"forbidden","code":"permission_denied"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireFieldWriteAccess rejects JSON mutations that touch fields explicitly
// disabled by the user's profile.
func RequireFieldWriteAccess() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}
			module, _, ok := moduleActionFromRequest(r)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			if len(bytes.TrimSpace(body)) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			var payload map[string]json.RawMessage
			if err := json.Unmarshal(body, &payload); err != nil {
				next.ServeHTTP(w, r)
				return
			}
			access, ok := domain.AccessContextFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"forbidden","code":"access_context_required"}`, http.StatusForbidden)
				return
			}
			for field := range payload {
				if !access.CanWriteField(module, field) {
					http.Error(w, fmt.Sprintf(`{"error":"forbidden","code":"field_write_denied","field":%q}`, field), http.StatusForbidden)
					return
				}
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}

func moduleActionFromRequest(r *http.Request) (domain.ACLModule, domain.ACLAction, bool) {
	module, ok := moduleFromPath(r.URL.Path)
	if !ok {
		return "", "", false
	}
	action := actionFromMethod(r.Method)
	if module == domain.ACLModuleExport {
		action = domain.ACLActionExport
	}
	return module, action, true
}

func actionFromMethod(method string) domain.ACLAction {
	switch method {
	case http.MethodPost:
		return domain.ACLActionCreate
	case http.MethodPut, http.MethodPatch:
		return domain.ACLActionUpdate
	case http.MethodDelete:
		return domain.ACLActionDelete
	default:
		return domain.ACLActionRead
	}
}

func moduleFromPath(path string) (domain.ACLModule, bool) {
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.Trim(path, "/")
	if path == "" {
		return "", false
	}
	parts := strings.Split(path, "/")
	first := parts[0]
	second := ""
	if len(parts) > 1 {
		second = parts[1]
	}

	if first == "users" && (second == "me" || second == "") {
		if second == "me" {
			return "", false
		}
	}
	if first == "auth" || first == "push" || first == "attachments" || first == "portal-links" {
		return "", false
	}
	// Portal routes enforce client role and ownership in the portal handlers.
	if first == "portal" {
		return "", false
	}
	if first == "admin" && second == "audit-log" {
		return domain.ACLModuleAuditLog, true
	}
	if first == "settings" {
		return domain.ACLModuleSettings, true
	}
	if first == "sla-policies" || first == "sla-instances" || first == "sla-dashboard" {
		return domain.ACLModuleSLA, true
	}
	if first == "email-templates" {
		return domain.ACLModuleEmailTemplates, true
	}
	if first == "api-keys" {
		return domain.ACLModuleAPIKeys, true
	}
	if first == "custom-fields" {
		return domain.ACLModuleCustomFields, true
	}
	if first == "ops-finance" {
		return domain.ACLModuleOpsFinance, true
	}
	modules := map[string]domain.ACLModule{
		"accounts":      domain.ACLModuleAccounts,
		"activities":    domain.ACLModuleActivities,
		"automations":   domain.ACLModuleAutomations,
		"billing":       domain.ACLModuleBilling,
		"calendar":      domain.ACLModuleCalendar,
		"contacts":      domain.ACLModuleContacts,
		"dashboards":    domain.ACLModuleDashboards,
		"deals":         domain.ACLModuleDeals,
		"emails":        domain.ACLModuleEmails,
		"export":        domain.ACLModuleExport,
		"import":        domain.ACLModuleContacts,
		"integrations":  domain.ACLModuleIntegrations,
		"kb":            domain.ACLModuleKB,
		"leads":         domain.ACLModuleLeads,
		"notifications": domain.ACLModuleNotifications,
		"onboarding":    domain.ACLModuleOnboarding,
		"products":      domain.ACLModuleProducts,
		"quotes":        domain.ACLModuleQuotes,
		"reports":       domain.ACLModuleReports,
		"search":        domain.ACLModuleSearch,
		"sequences":     domain.ACLModuleSequences,
		"tickets":       domain.ACLModuleTickets,
		"timeline":      domain.ACLModuleTimeline,
		"users":         domain.ACLModuleUsers,
		"views":         domain.ACLModuleSavedViews,
		"webhooks":      domain.ACLModuleWebhooks,
	}
	module, ok := modules[first]
	return module, ok
}
