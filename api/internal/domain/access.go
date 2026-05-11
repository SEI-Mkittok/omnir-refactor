package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ACLModule string

const (
	ACLModuleAccounts       ACLModule = "accounts"
	ACLModuleActivities     ACLModule = "activities"
	ACLModuleAPIKeys        ACLModule = "api_keys"
	ACLModuleAuditLog       ACLModule = "audit_log"
	ACLModuleAutomations    ACLModule = "automations"
	ACLModuleBilling        ACLModule = "billing"
	ACLModuleCalendar       ACLModule = "calendar"
	ACLModuleContacts       ACLModule = "contacts"
	ACLModuleCustomFields   ACLModule = "custom_fields"
	ACLModuleDashboards     ACLModule = "dashboards"
	ACLModuleDeals          ACLModule = "deals"
	ACLModuleEmailTemplates ACLModule = "email_templates"
	ACLModuleEmails         ACLModule = "emails"
	ACLModuleExport         ACLModule = "export"
	ACLModuleIntegrations   ACLModule = "integrations"
	ACLModuleKB             ACLModule = "kb"
	ACLModuleLeads          ACLModule = "leads"
	ACLModuleNotifications  ACLModule = "notifications"
	ACLModuleOnboarding     ACLModule = "onboarding"
	ACLModuleOpsFinance     ACLModule = "ops_finance"
	ACLModuleProducts       ACLModule = "products"
	ACLModuleQuotes         ACLModule = "quotes"
	ACLModuleReports        ACLModule = "reports"
	ACLModuleSavedViews     ACLModule = "views"
	ACLModuleSearch         ACLModule = "search"
	ACLModuleSequences      ACLModule = "sequences"
	ACLModuleSettings       ACLModule = "settings"
	ACLModuleSLA            ACLModule = "sla"
	ACLModuleTickets        ACLModule = "tickets"
	ACLModuleTimeline       ACLModule = "timeline"
	ACLModuleUsers          ACLModule = "users"
	ACLModuleWebhooks       ACLModule = "webhooks"
)

type ACLAction string

const (
	ACLActionRead   ACLAction = "read"
	ACLActionCreate ACLAction = "create"
	ACLActionUpdate ACLAction = "update"
	ACLActionDelete ACLAction = "delete"
	ACLActionExport ACLAction = "export"
	ACLActionAdmin  ACLAction = "admin"
)

type SharingDefaultMode string

const (
	SharingDefaultPrivate  SharingDefaultMode = "private"
	SharingDefaultPublicRO SharingDefaultMode = "public_read"
	SharingDefaultPublicRW SharingDefaultMode = "public_rw"
)

type SharingAccessLevel string

const (
	SharingAccessRead  SharingAccessLevel = "read"
	SharingAccessWrite SharingAccessLevel = "write"
)

type SharingGranteeType string

const (
	SharingGranteeRole  SharingGranteeType = "role"
	SharingGranteeGroup SharingGranteeType = "group"
)

var PermissionActions = []ACLAction{
	ACLActionRead,
	ACLActionCreate,
	ACLActionUpdate,
	ACLActionDelete,
	ACLActionExport,
	ACLActionAdmin,
}

var PermissionModules = []ACLModule{
	ACLModuleAccounts,
	ACLModuleActivities,
	ACLModuleAPIKeys,
	ACLModuleAuditLog,
	ACLModuleAutomations,
	ACLModuleBilling,
	ACLModuleCalendar,
	ACLModuleContacts,
	ACLModuleCustomFields,
	ACLModuleDashboards,
	ACLModuleDeals,
	ACLModuleEmailTemplates,
	ACLModuleEmails,
	ACLModuleExport,
	ACLModuleIntegrations,
	ACLModuleKB,
	ACLModuleLeads,
	ACLModuleNotifications,
	ACLModuleOnboarding,
	ACLModuleOpsFinance,
	ACLModuleProducts,
	ACLModuleQuotes,
	ACLModuleReports,
	ACLModuleSavedViews,
	ACLModuleSearch,
	ACLModuleSequences,
	ACLModuleSettings,
	ACLModuleSLA,
	ACLModuleTickets,
	ACLModuleTimeline,
	ACLModuleUsers,
	ACLModuleWebhooks,
}

var SharingModules = []ACLModule{
	ACLModuleAccounts,
	ACLModuleContacts,
	ACLModuleDeals,
	ACLModuleLeads,
	ACLModuleTickets,
}

type ACLRole struct {
	ID          uuid.UUID  `json:"id"`
	OrgID       uuid.UUID  `json:"org_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	SystemKey   *string    `json:"system_key,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ACLRolePatch struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

type ACLProfile struct {
	ID          uuid.UUID `json:"id"`
	OrgID       uuid.UUID `json:"org_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	SystemKey   *string   `json:"system_key,omitempty"`
}

type ACLProfilePatch struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ACLProfilePermission struct {
	ProfileID uuid.UUID `json:"profile_id"`
	Module    ACLModule `json:"module"`
	Action    ACLAction `json:"action"`
	Allowed   bool      `json:"allowed"`
}

type ACLProfileFieldPermission struct {
	ProfileID uuid.UUID `json:"profile_id"`
	Module    ACLModule `json:"module"`
	FieldName string    `json:"field_name"`
	CanWrite  bool      `json:"can_write"`
}

type ACLGroup struct {
	ID          uuid.UUID   `json:"id"`
	OrgID       uuid.UUID   `json:"org_id"`
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	UserIDs     []uuid.UUID `json:"user_ids,omitempty"`
}

type ACLGroupPatch struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ACLSharingDefault struct {
	OrgID  uuid.UUID          `json:"org_id"`
	Module ACLModule          `json:"module"`
	Mode   SharingDefaultMode `json:"mode"`
}

type ACLSharingGrant struct {
	ID          uuid.UUID          `json:"id"`
	OrgID       uuid.UUID          `json:"org_id"`
	Module      ACLModule          `json:"module"`
	GranteeType SharingGranteeType `json:"grantee_type"`
	GranteeID   uuid.UUID          `json:"grantee_id"`
	AccessLevel SharingAccessLevel `json:"access_level"`
}

type ACLSharingModuleRule struct {
	Module ACLModule          `json:"module"`
	Mode   SharingDefaultMode `json:"mode"`
	Grants []ACLSharingGrant  `json:"grants"`
}

type ACLSharingRules struct {
	Rules []ACLSharingModuleRule `json:"rules"`
}

type ACLSharingAccess struct {
	Mode     SharingDefaultMode
	ReadAll  bool
	WriteAll bool
}

type AccessContext struct {
	UserID           uuid.UUID
	OrgID            uuid.UUID
	PlatformRole     string
	RoleID           *uuid.UUID
	ProfileID        *uuid.UUID
	GroupIDs         []uuid.UUID
	RoleLineageIDs   []uuid.UUID
	Permissions      map[ACLModule]map[ACLAction]bool
	FieldWrite       map[ACLModule]map[string]bool
	Sharing          map[ACLModule]ACLSharingAccess
	SuperAdminBypass bool
}

func (a *AccessContext) HasPermission(module ACLModule, action ACLAction) bool {
	if a == nil {
		return false
	}
	if a.SuperAdminBypass {
		return true
	}
	if actions, ok := a.Permissions[module]; ok {
		if actions[ACLActionAdmin] || actions[action] {
			return true
		}
	}
	if action == ACLActionExport {
		if actions, ok := a.Permissions[module]; ok && actions[ACLActionRead] {
			return true
		}
	}
	return false
}

func (a *AccessContext) CanWriteField(module ACLModule, field string) bool {
	if a == nil || a.SuperAdminBypass {
		return true
	}
	fields, ok := a.FieldWrite[module]
	if !ok {
		return true
	}
	allowed, configured := fields[field]
	if !configured {
		return true
	}
	return allowed
}

func (a *AccessContext) CanAccessAllRecords(module ACLModule, access SharingAccessLevel) bool {
	if a == nil || a.SuperAdminBypass {
		return true
	}
	rule, ok := a.Sharing[module]
	if !ok {
		return false
	}
	if access == SharingAccessWrite {
		return rule.Mode == SharingDefaultPublicRW || rule.WriteAll
	}
	return rule.Mode == SharingDefaultPublicRO || rule.Mode == SharingDefaultPublicRW || rule.ReadAll || rule.WriteAll
}

type accessContextKey struct{}

func WithAccessContext(ctx context.Context, access *AccessContext) context.Context {
	return context.WithValue(ctx, accessContextKey{}, access)
}

func AccessContextFromContext(ctx context.Context) (*AccessContext, bool) {
	access, ok := ctx.Value(accessContextKey{}).(*AccessContext)
	return access, ok
}
