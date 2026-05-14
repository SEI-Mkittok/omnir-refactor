package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type AccessSettingsHandler struct {
	repo  repository.AccessRepository
	users repository.UserRepository
	audit Auditor
}

func NewAccessSettingsHandler(repo repository.AccessRepository, users ...repository.UserRepository) *AccessSettingsHandler {
	h := &AccessSettingsHandler{repo: repo}
	if len(users) > 0 {
		h.users = users[0]
	}
	return h
}

func (h *AccessSettingsHandler) WithAuditLog(r repository.AuditLogRepository) *AccessSettingsHandler {
	h.audit = newAuditor(r)
	return h
}

func (h *AccessSettingsHandler) RolesRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(requireAccessSettingsAdmin)
	r.Get("/", h.ListRoles)
	r.Post("/", h.CreateRole)
	r.Patch("/{id}", h.UpdateRole)
	r.Delete("/{id}", h.DeleteRole)
	r.Patch("/{id}/parent", h.MoveRole)
	return r
}

func (h *AccessSettingsHandler) ProfilesRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(requireAccessSettingsAdmin)
	r.Get("/", h.ListProfiles)
	r.Post("/", h.CreateProfile)
	r.Get("/catalog", h.PermissionCatalog)
	r.Patch("/{id}", h.UpdateProfile)
	r.Delete("/{id}", h.DeleteProfile)
	r.Get("/{id}/permissions", h.GetProfilePermissions)
	r.Put("/{id}/permissions", h.ReplaceProfilePermissions)
	return r
}

func (h *AccessSettingsHandler) GroupsRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(requireAccessSettingsAdmin)
	r.Get("/", h.ListGroups)
	r.Get("/member-candidates", h.ListGroupMemberCandidates)
	r.Post("/", h.CreateGroup)
	r.Patch("/{id}", h.UpdateGroup)
	r.Delete("/{id}", h.DeleteGroup)
	r.Put("/{id}/members", h.ReplaceGroupMembers)
	return r
}

func (h *AccessSettingsHandler) SharingRulesRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(requireAccessSettingsAdmin)
	r.Get("/", h.GetSharingRules)
	r.Patch("/", h.ReplaceSharingRules)
	return r
}

func requireAccessSettingsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasAccessSettingsAdmin(r) {
			next.ServeHTTP(w, r)
			return
		}
		writeError(w, http.StatusForbidden, "admin access required")
	})
}

func hasAccessSettingsAdmin(r *http.Request) bool {
	return hasModuleAdminAccess(r, domain.ACLModuleSettings)
}

func (h *AccessSettingsHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	roles, err := h.repo.ListRoles(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if roles == nil {
		roles = []*domain.ACLRole{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": roles})
}

func (h *AccessSettingsHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	var req domain.ACLRole
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	req.OrgID = orgID
	created, err := h.repo.CreateRole(r.Context(), &req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AccessSettingsHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var patch domain.ACLRolePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.UpdateRole(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *AccessSettingsHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	if err := h.repo.DeleteRole(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccessSettingsHandler) MoveRole(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var req struct {
		ParentID *uuid.UUID `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.MoveRole(r.Context(), id, req.ParentID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *AccessSettingsHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	profiles, err := h.repo.ListProfiles(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if profiles == nil {
		profiles = []*domain.ACLProfile{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": profiles})
}

func (h *AccessSettingsHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	var req domain.ACLProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	req.OrgID = orgID
	created, err := h.repo.CreateProfile(r.Context(), &req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AccessSettingsHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var patch domain.ACLProfilePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.UpdateProfile(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *AccessSettingsHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	if err := h.repo.DeleteProfile(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccessSettingsHandler) GetProfilePermissions(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	permissions, fieldPermissions, err := h.repo.ListProfilePermissions(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"permissions":       permissions,
		"field_permissions": fieldPermissions,
	})
}

func (h *AccessSettingsHandler) ReplaceProfilePermissions(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var req struct {
		Permissions      []domain.ACLProfilePermission      `json:"permissions"`
		FieldPermissions []domain.ACLProfileFieldPermission `json:"field_permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.repo.ReplaceProfilePermissions(r.Context(), id, req.Permissions, req.FieldPermissions); err != nil {
		handleDomainErr(w, err)
		return
	}
	h.GetProfilePermissions(w, r)
}

func (h *AccessSettingsHandler) PermissionCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"modules": domain.PermissionModules,
		"actions": domain.PermissionActions,
	})
}

func (h *AccessSettingsHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	groups, err := h.repo.ListGroups(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if groups == nil {
		groups = []*domain.ACLGroup{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": groups})
}

func (h *AccessSettingsHandler) ListGroupMemberCandidates(w http.ResponseWriter, r *http.Request) {
	if h.users == nil {
		writeError(w, http.StatusInternalServerError, "user repository unavailable")
		return
	}
	users, total, err := h.users.List(r.Context(), domain.UserFilter{
		Page:  1,
		Limit: 500,
		Sort:  "name",
		Order: "asc",
	})
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(users, total, 1, 500))
}

func (h *AccessSettingsHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	var req domain.ACLGroup
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	req.OrgID = orgID
	created, err := h.repo.CreateGroup(r.Context(), &req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AccessSettingsHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var patch domain.ACLGroupPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.UpdateGroup(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *AccessSettingsHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	if err := h.repo.DeleteGroup(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccessSettingsHandler) ReplaceGroupMembers(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var req struct {
		UserIDs []uuid.UUID `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.repo.ReplaceGroupMembers(r.Context(), id, req.UserIDs); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccessSettingsHandler) GetSharingRules(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	rules, err := h.repo.GetSharingRules(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (h *AccessSettingsHandler) ReplaceSharingRules(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	var req domain.ACLSharingRules
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	before, err := h.repo.GetSharingRules(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	rules, err := h.repo.ReplaceSharingRules(r.Context(), orgID, &req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	h.logSharingRuleChanges(r, before, rules)
	writeJSON(w, http.StatusOK, rules)
}

func (h *AccessSettingsHandler) logSharingRuleChanges(r *http.Request, before, after *domain.ACLSharingRules) {
	if h.audit.repo == nil {
		return
	}
	entries := sharingRuleAuditEntries(before, after)
	if len(entries) == 0 {
		return
	}
	orgID, _ := domain.OrgIDFromContext(r.Context())
	var userID *uuid.UUID
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		id := claims.UserID
		userID = &id
	}
	ip := realClientIP(r)
	ua := r.UserAgent()
	for _, entry := range entries {
		entry.OrgID = orgID
		entry.UserID = userID
		if ip != "" {
			entry.IPAddress = &ip
		}
		if ua != "" {
			entry.UserAgent = &ua
		}
		go func(entry domain.AuditEntry) {
			if err := h.audit.repo.Append(context.Background(), entry); err != nil {
				slog.Error("audit log write failed", "entity", "sharing_rule", "error", err)
			}
		}(entry)
	}
}

func sharingRuleAuditEntries(before, after *domain.ACLSharingRules) []domain.AuditEntry {
	entries := []domain.AuditEntry{}
	beforeDefaults := map[domain.ACLModule]domain.SharingDefaultMode{}
	afterDefaults := map[domain.ACLModule]domain.SharingDefaultMode{}
	beforeRules := map[string]domain.ACLSharingRule{}
	afterRules := map[string]domain.ACLSharingRule{}

	for _, moduleRule := range safeSharingModuleRules(before) {
		beforeDefaults[moduleRule.Module] = moduleRule.Mode
		for _, rule := range moduleRule.AdvancedRules {
			beforeRules[sharingAuditRuleID(rule)] = rule
		}
	}
	for _, moduleRule := range safeSharingModuleRules(after) {
		afterDefaults[moduleRule.Module] = moduleRule.Mode
		for _, rule := range moduleRule.AdvancedRules {
			afterRules[sharingAuditRuleID(rule)] = rule
		}
	}
	for module, beforeMode := range beforeDefaults {
		if afterMode, ok := afterDefaults[module]; ok && afterMode != beforeMode {
			name := string(module) + " default"
			entries = append(entries, domain.AuditEntry{
				Action:     domain.AuditActionUpdated,
				EntityType: domain.AuditEntitySharingRule,
				EntityName: &name,
				Changes: domain.AuditChanges{
					"mode": {From: beforeMode, To: afterMode},
				},
			})
		}
	}
	for key, afterRule := range afterRules {
		beforeRule, existed := beforeRules[key]
		name := string(afterRule.Module) + " sharing rule"
		entityID := afterRule.ID
		entry := domain.AuditEntry{
			EntityType: domain.AuditEntitySharingRule,
			EntityName: &name,
		}
		if entityID != uuid.Nil {
			entry.EntityID = &entityID
		}
		if !existed {
			entry.Action = domain.AuditActionCreated
			entry.Changes = domain.AuditChanges{"rule": {From: nil, To: sharingAuditRuleValue(afterRule)}}
			entries = append(entries, entry)
			continue
		}
		if sharingAuditRuleID(beforeRule) != sharingAuditRuleID(afterRule) || sharingAuditRuleValue(beforeRule) != sharingAuditRuleValue(afterRule) {
			entry.Action = domain.AuditActionUpdated
			entry.Changes = domain.AuditChanges{"rule": {From: sharingAuditRuleValue(beforeRule), To: sharingAuditRuleValue(afterRule)}}
			entries = append(entries, entry)
		}
	}
	for key, beforeRule := range beforeRules {
		if _, ok := afterRules[key]; ok {
			continue
		}
		name := string(beforeRule.Module) + " sharing rule"
		entityID := beforeRule.ID
		entry := domain.AuditEntry{
			Action:     domain.AuditActionDeleted,
			EntityType: domain.AuditEntitySharingRule,
			EntityName: &name,
			Changes: domain.AuditChanges{
				"rule": {From: sharingAuditRuleValue(beforeRule), To: nil},
			},
		}
		if entityID != uuid.Nil {
			entry.EntityID = &entityID
		}
		entries = append(entries, entry)
	}
	return entries
}

func safeSharingModuleRules(rules *domain.ACLSharingRules) []domain.ACLSharingModuleRule {
	if rules == nil {
		return nil
	}
	return rules.Rules
}

func sharingAuditRuleID(rule domain.ACLSharingRule) string {
	if rule.ID != uuid.Nil {
		return rule.ID.String()
	}
	sourceID := ""
	if rule.SourceID != nil {
		sourceID = rule.SourceID.String()
	}
	return strings.Join([]string{
		string(rule.Module),
		string(rule.SourceType),
		sourceID,
		string(rule.TargetType),
		rule.TargetID.String(),
		string(rule.AccessLevel),
	}, "|")
}

func sharingAuditRuleValue(rule domain.ACLSharingRule) string {
	sourceID := ""
	if rule.SourceID != nil {
		sourceID = rule.SourceID.String()
	}
	return strings.Join([]string{
		string(rule.Module),
		string(rule.SourceType),
		sourceID,
		string(rule.TargetType),
		rule.TargetID.String(),
		string(rule.AccessLevel),
	}, "|")
}

func parseACLID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return uuid.Nil, false
	}
	return id, true
}
