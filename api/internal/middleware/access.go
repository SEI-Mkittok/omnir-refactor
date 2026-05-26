package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

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
			if isSelfServiceUserUpdate(r) {
				next.ServeHTTP(w, r)
				return
			}
			if isSelfServicePreferenceRequest(r) {
				next.ServeHTTP(w, r)
				return
			}
			if claims, ok := ClaimsFromContext(r); ok && domain.IsAdminRole(claims.Role) {
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
			if childModule, childAction, hasChildModule := childModuleActionFromRequest(r); hasChildModule {
				if !access.HasPermission(childModule, childAction) {
					http.Error(w, `{"error":"forbidden","code":"permission_denied"}`, http.StatusForbidden)
					return
				}
			}
			targetModules := conversionCreateModulesFromRequest(r)
			for _, targetModule := range targetModules {
				if !access.HasPermission(targetModule, domain.ACLActionCreate) {
					http.Error(w, `{"error":"forbidden","code":"permission_denied"}`, http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isSelfServiceUserUpdate(r *http.Request) bool {
	if r.Method != http.MethodPatch {
		return false
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] != "users" {
		return false
	}
	claims, ok := ClaimsFromContext(r)
	if !ok {
		return false
	}
	userID, err := uuid.Parse(parts[1])
	return err == nil && userID == claims.UserID
}

func isSelfServicePreferenceRequest(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodPatch {
		return false
	}
	if _, ok := ClaimsFromContext(r); !ok {
		return false
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] != "settings" {
		return false
	}
	return parts[1] == "preferences" || parts[1] == "calendar-preferences"
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
			if childModule, _, hasChildModule := childModuleActionFromRequest(r); hasChildModule {
				module = childModule
			}
			if isSelfServicePreferenceRequest(r) {
				next.ServeHTTP(w, r)
				return
			}
			if claims, ok := ClaimsFromContext(r); ok && domain.IsAdminRole(claims.Role) {
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
			if deniedField := deniedWriteField(access, module, payload); deniedField != "" {
				http.Error(w, fmt.Sprintf(`{"error":"forbidden","code":"field_write_denied","field":%q}`, deniedField), http.StatusForbidden)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}

func deniedWriteField(access *domain.AccessContext, module domain.ACLModule, payload map[string]json.RawMessage) string {
	for field, raw := range payload {
		if !access.CanWriteField(module, field) {
			return field
		}
		if field != "custom_fields" {
			continue
		}
		var customFields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &customFields); err != nil {
			continue
		}
		for customField := range customFields {
			qualified := "custom_fields." + customField
			if !access.CanWriteField(module, qualified) || !access.CanWriteField(module, customField) {
				return qualified
			}
		}
	}
	return ""
}

// RequireNestedParentRecordAccess ensures child resources inherit the sharing
// visibility of their parent CRM record.
func RequireNestedParentRecordAccess(repo repository.RecordAccessRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			relationshipID, relationshipAccessLevel, isRelationshipRoute, err := accountRelationshipFromRequest(r)
			if err != nil {
				http.Error(w, `{"error":"bad_request","code":"invalid_relationship_id"}`, http.StatusBadRequest)
				return
			}
			if isRelationshipRoute {
				if repo == nil {
					next.ServeHTTP(w, r)
					return
				}
				canAccess, err := repo.CanAccessAccountRelationship(r.Context(), relationshipID, relationshipAccessLevel)
				if err != nil {
					http.Error(w, `{"error":"access_check_failed"}`, http.StatusInternalServerError)
					return
				}
				if !canAccess {
					http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			module, id, accessLevel, ok := nestedParentRecordFromRequest(r)
			if !ok || repo == nil {
				next.ServeHTTP(w, r)
				return
			}
			canAccess, err := repo.CanAccessRecord(r.Context(), module, id, accessLevel)
			if err != nil {
				http.Error(w, `{"error":"access_check_failed"}`, http.StatusInternalServerError)
				return
			}
			if !canAccess {
				http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func accountRelationshipFromRequest(r *http.Request) (uuid.UUID, domain.SharingAccessLevel, bool, error) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	path = strings.Trim(path, "/")
	if path == "" {
		return uuid.Nil, "", false, nil
	}
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[0] != "accounts" || parts[1] != "relationships" {
		return uuid.Nil, "", false, nil
	}
	id, err := uuid.Parse(parts[2])
	if err != nil {
		return uuid.Nil, "", true, err
	}
	return id, sharingAccessFromMethod(r.Method), true, nil
}

func moduleActionFromRequest(r *http.Request) (domain.ACLModule, domain.ACLAction, bool) {
	module, ok := moduleFromPath(r.URL.Path)
	if !ok {
		return "", "", false
	}
	action := actionFromMethod(r.Method)
	if commandAction, ok := commandActionFromPath(r.Method, r.URL.Path); ok {
		action = commandAction
	}
	if isNestedRecordSubroute(r.URL.Path) && action != domain.ACLActionRead {
		action = domain.ACLActionUpdate
	}
	if module == domain.ACLModuleExport {
		action = domain.ACLActionExport
	}
	if isOrgSSOPath(r.URL.Path) {
		action = domain.ACLActionAdmin
	}
	return module, action, true
}

func commandActionFromPath(method, path string) (domain.ACLAction, bool) {
	if method != http.MethodPost {
		return "", false
	}
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		return "", false
	}
	if _, err := uuid.Parse(parts[1]); err != nil {
		return "", false
	}

	updateCommands := map[string]map[string]bool{
		"quotes": {
			"approve": true,
			"reject":  true,
			"send":    true,
		},
		"views": {
			"pin": true,
		},
	}
	if commands, ok := updateCommands[parts[0]]; ok && commands[parts[2]] {
		return domain.ACLActionUpdate, true
	}
	return "", false
}

func childModuleActionFromRequest(r *http.Request) (domain.ACLModule, domain.ACLAction, bool) {
	module, ok := childModuleFromPath(r.URL.Path)
	if !ok {
		return "", "", false
	}
	return module, actionFromMethod(r.Method), true
}

func conversionCreateModulesFromRequest(r *http.Request) []domain.ACLModule {
	if r.Method != http.MethodPost {
		return nil
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[2] != "convert" {
		return nil
	}
	if _, err := uuid.Parse(parts[1]); err != nil {
		return nil
	}

	switch parts[0] {
	case "leads":
		return []domain.ACLModule{
			domain.ACLModuleContacts,
			domain.ACLModuleAccounts,
			domain.ACLModuleDeals,
		}
	case "contacts":
		if !contactConversionCreatesDeal(r) {
			return nil
		}
		return []domain.ACLModule{domain.ACLModuleDeals}
	default:
		return nil
	}
}

func contactConversionCreatesDeal(r *http.Request) bool {
	if r.Body == nil {
		return false
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(nil))
		return false
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	if len(bytes.TrimSpace(body)) == 0 {
		return false
	}
	var payload struct {
		CreateDeal bool `json:"create_deal"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	return payload.CreateDeal
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

func isNestedRecordSubroute(path string) bool {
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.Trim(path, "/")
	if path == "" {
		return false
	}
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return false
	}
	if _, ok := parentRecordModule(parts[0]); !ok {
		return false
	}
	_, err := uuid.Parse(parts[1])
	return err == nil
}

func nestedParentRecordFromRequest(r *http.Request) (domain.ACLModule, uuid.UUID, domain.SharingAccessLevel, bool) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	path = strings.Trim(path, "/")
	if path == "" {
		return "", uuid.Nil, "", false
	}
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return "", uuid.Nil, "", false
	}

	module, ok := parentRecordModule(parts[0])
	if !ok {
		return "", uuid.Nil, "", false
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		// Some routers expose collection subroutes under a record module prefix,
		// e.g. /accounts/relationships/{relationshipID}; those are not parent
		// record child routes and should be left to their handlers.
		return "", uuid.Nil, "", false
	}
	return module, id, sharingAccessFromMethod(r.Method), true
}

func parentRecordModule(segment string) (domain.ACLModule, bool) {
	switch segment {
	case "accounts":
		return domain.ACLModuleAccounts, true
	case "contacts":
		return domain.ACLModuleContacts, true
	case "deals":
		return domain.ACLModuleDeals, true
	case "leads":
		return domain.ACLModuleLeads, true
	case "tickets":
		return domain.ACLModuleTickets, true
	default:
		return "", false
	}
}

func sharingAccessFromMethod(method string) domain.SharingAccessLevel {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return domain.SharingAccessRead
	default:
		return domain.SharingAccessWrite
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
	if first == "enrich" {
		return domain.ACLModuleContacts, true
	}
	if first == "orgs" && len(parts) > 2 && parts[2] == "sso" {
		return domain.ACLModuleIntegrations, true
	}
	if first == "ops-finance" {
		return domain.ACLModuleOpsFinance, true
	}
	if first == "import" {
		return importModuleFromPath(second)
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

func childModuleFromPath(path string) (domain.ACLModule, bool) {
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.Trim(path, "/")
	if path == "" {
		return "", false
	}
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return "", false
	}
	if _, err := uuid.Parse(parts[1]); err != nil {
		return "", false
	}

	switch parts[0] {
	case "contacts":
		if parts[2] == "emails" {
			return domain.ACLModuleEmails, true
		}
	case "deals":
		if parts[2] == "quotes" {
			return domain.ACLModuleQuotes, true
		}
	}
	return "", false
}

func isOrgSSOPath(path string) bool {
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[0] != "orgs" || parts[2] != "sso" {
		return false
	}
	_, err := uuid.Parse(parts[1])
	return err == nil
}

func importModuleFromPath(segment string) (domain.ACLModule, bool) {
	switch segment {
	case "contacts":
		return domain.ACLModuleContacts, true
	case "accounts":
		return domain.ACLModuleAccounts, true
	case "leads":
		return domain.ACLModuleLeads, true
	default:
		return "", false
	}
}
