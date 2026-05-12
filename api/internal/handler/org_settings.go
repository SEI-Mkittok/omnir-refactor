package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// OrgSettingsHandler handles org-level settings such as document numbering configuration.
type OrgSettingsHandler struct {
	settings repository.OrgSettingsRepository
	encKey   string
	auditor  Auditor
}

func NewOrgSettingsHandler(settings repository.OrgSettingsRepository, encKey string) *OrgSettingsHandler {
	return &OrgSettingsHandler{settings: settings, encKey: encKey}
}

func (h *OrgSettingsHandler) WithAuditLog(r repository.AuditLogRepository) *OrgSettingsHandler {
	h.auditor = newAuditor(r)
	return h
}

func (h *OrgSettingsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAdmin)
	r.Get("/", h.Get)
	r.Patch("/", h.Update)
	return r
}

func (h *OrgSettingsHandler) CompanyRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAdmin)
	r.Get("/", h.GetCompany)
	r.Patch("/", h.UpdateCompany)
	return r
}

func (h *OrgSettingsHandler) PortalRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAdmin)
	r.Get("/", h.GetPortal)
	r.Patch("/", h.UpdatePortal)
	return r
}

func (h *OrgSettingsHandler) OutgoingServerRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAdmin)
	r.Get("/", h.GetOutgoingServer)
	r.Patch("/", h.UpdateOutgoingServer)
	return r
}

func (h *OrgSettingsHandler) ConfigEditorRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAdmin)
	r.Get("/", h.GetConfigEditor)
	r.Patch("/", h.UpdateConfigEditor)
	return r
}

func (h *OrgSettingsHandler) MenuConfigRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetMenuConfig)
	r.With(h.requireAdmin).Patch("/", h.UpdateMenuConfig)
	return r
}

func (h *OrgSettingsHandler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type numberingSequenceAccessor interface {
	GetCurrentDocNumbers(ctx context.Context, orgID uuid.UUID) (map[domain.DocType]int64, error)
	SetCurrentDocNumbers(ctx context.Context, orgID uuid.UUID, values map[domain.DocType]int64) error
}

type numberingSettingsPayload struct {
	QuoteNumberStart       *int64  `json:"quote_number_start,omitempty"`
	TicketNumberStart      *int64  `json:"ticket_number_start,omitempty"`
	KBArticleNumberStart   *int64  `json:"kb_article_number_start,omitempty"`
	InvoiceNumberStart     *int64  `json:"invoice_number_start,omitempty"`
	QuoteNumberPrefix      *string `json:"quote_number_prefix,omitempty"`
	TicketNumberPrefix     *string `json:"ticket_number_prefix,omitempty"`
	KBArticleNumberPrefix  *string `json:"kb_article_number_prefix,omitempty"`
	InvoiceNumberPrefix    *string `json:"invoice_number_prefix,omitempty"`
	QuoteNumberCurrent     *int64  `json:"quote_number_current,omitempty"`
	TicketNumberCurrent    *int64  `json:"ticket_number_current,omitempty"`
	KBArticleNumberCurrent *int64  `json:"kb_article_number_current,omitempty"`
	InvoiceNumberCurrent   *int64  `json:"invoice_number_current,omitempty"`
}

type numberingSettingsResponse struct {
	OrgID                  uuid.UUID `json:"org_id"`
	QuoteNumberStart       int64     `json:"quote_number_start"`
	TicketNumberStart      int64     `json:"ticket_number_start"`
	KBArticleNumberStart   int64     `json:"kb_article_number_start"`
	InvoiceNumberStart     int64     `json:"invoice_number_start"`
	QuoteNumberPrefix      string    `json:"quote_number_prefix"`
	TicketNumberPrefix     string    `json:"ticket_number_prefix"`
	KBArticleNumberPrefix  string    `json:"kb_article_number_prefix"`
	InvoiceNumberPrefix    string    `json:"invoice_number_prefix"`
	QuoteNumberCurrent     int64     `json:"quote_number_current"`
	TicketNumberCurrent    int64     `json:"ticket_number_current"`
	KBArticleNumberCurrent int64     `json:"kb_article_number_current"`
	InvoiceNumberCurrent   int64     `json:"invoice_number_current"`
}

// Get returns current org settings (including numbering start values).
// GET /api/v1/settings/numbering
func (h *OrgSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, s, ok := h.loadSettings(w, r)
	if !ok {
		return
	}

	currents := map[domain.DocType]int64{
		domain.DocTypeQuote:     s.QuoteNumberStart - 1,
		domain.DocTypeTicket:    s.TicketNumberStart - 1,
		domain.DocTypeKBArticle: s.KBArticleNumberStart - 1,
		domain.DocTypeInvoice:   s.InvoiceNumberStart - 1,
	}
	if seqRepo, ok := h.settings.(numberingSequenceAccessor); ok {
		loaded, err := seqRepo.GetCurrentDocNumbers(r.Context(), claims.OrgID)
		if err == nil {
			for k, v := range loaded {
				currents[k] = v
			}
		}
	}
	writeJSON(w, http.StatusOK, numberingSettingsResponse{
		OrgID:                  claims.OrgID,
		QuoteNumberStart:       s.QuoteNumberStart,
		TicketNumberStart:      s.TicketNumberStart,
		KBArticleNumberStart:   s.KBArticleNumberStart,
		InvoiceNumberStart:     s.InvoiceNumberStart,
		QuoteNumberPrefix:      s.QuoteNumberPrefix,
		TicketNumberPrefix:     s.TicketNumberPrefix,
		KBArticleNumberPrefix:  s.KBArticleNumberPrefix,
		InvoiceNumberPrefix:    s.InvoiceNumberPrefix,
		QuoteNumberCurrent:     currents[domain.DocTypeQuote],
		TicketNumberCurrent:    currents[domain.DocTypeTicket],
		KBArticleNumberCurrent: currents[domain.DocTypeKBArticle],
		InvoiceNumberCurrent:   currents[domain.DocTypeInvoice],
	})
}

// Update patches org settings numbering fields.
// PATCH /api/v1/settings/numbering
func (h *OrgSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req numberingSettingsPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	for _, v := range []*int64{
		req.QuoteNumberStart, req.TicketNumberStart,
		req.KBArticleNumberStart, req.InvoiceNumberStart,
	} {
		if v != nil && *v < 1 {
			writeError(w, http.StatusUnprocessableEntity, "starting numbers must be >= 1")
			return
		}
	}
	for _, v := range []*int64{
		req.QuoteNumberCurrent, req.TicketNumberCurrent, req.KBArticleNumberCurrent, req.InvoiceNumberCurrent,
	} {
		if v != nil && *v < 0 {
			writeError(w, http.StatusUnprocessableEntity, "current numbers must be >= 0")
			return
		}
	}

	patch := domain.OrgSettingsPatch{
		QuoteNumberStart:      req.QuoteNumberStart,
		TicketNumberStart:     req.TicketNumberStart,
		KBArticleNumberStart:  req.KBArticleNumberStart,
		InvoiceNumberStart:    req.InvoiceNumberStart,
		QuoteNumberPrefix:     req.QuoteNumberPrefix,
		TicketNumberPrefix:    req.TicketNumberPrefix,
		KBArticleNumberPrefix: req.KBArticleNumberPrefix,
		InvoiceNumberPrefix:   req.InvoiceNumberPrefix,
	}

	claims, before, updated, ok := h.applyPatch(w, r, patch)
	if !ok {
		return
	}

	if seqRepo, ok := h.settings.(numberingSequenceAccessor); ok {
		next := map[domain.DocType]int64{}
		if req.QuoteNumberCurrent != nil {
			next[domain.DocTypeQuote] = *req.QuoteNumberCurrent
		}
		if req.TicketNumberCurrent != nil {
			next[domain.DocTypeTicket] = *req.TicketNumberCurrent
		}
		if req.KBArticleNumberCurrent != nil {
			next[domain.DocTypeKBArticle] = *req.KBArticleNumberCurrent
		}
		if req.InvoiceNumberCurrent != nil {
			next[domain.DocTypeInvoice] = *req.InvoiceNumberCurrent
		}
		if len(next) > 0 {
			if err := seqRepo.SetCurrentDocNumbers(r.Context(), claims.OrgID, next); err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
		}
	}

	h.logSettingsMutation(r, claims, "settings.numbering", auditChanges(
		auditField("quote_number_start", before.QuoteNumberStart, updated.QuoteNumberStart),
		auditField("ticket_number_start", before.TicketNumberStart, updated.TicketNumberStart),
		auditField("kb_article_number_start", before.KBArticleNumberStart, updated.KBArticleNumberStart),
		auditField("invoice_number_start", before.InvoiceNumberStart, updated.InvoiceNumberStart),
		auditField("quote_number_prefix", before.QuoteNumberPrefix, updated.QuoteNumberPrefix),
		auditField("ticket_number_prefix", before.TicketNumberPrefix, updated.TicketNumberPrefix),
		auditField("kb_article_number_prefix", before.KBArticleNumberPrefix, updated.KBArticleNumberPrefix),
		auditField("invoice_number_prefix", before.InvoiceNumberPrefix, updated.InvoiceNumberPrefix),
	))

	h.Get(w, r)
}

type companySettingsPayload struct {
	CompanyName         *string `json:"company_name"`
	CompanyLogoURL      *string `json:"company_logo_url"`
	CompanyWebsite      *string `json:"company_website"`
	CompanyEmail        *string `json:"company_email"`
	CompanyPhone        *string `json:"company_phone"`
	CompanyAddressLine1 *string `json:"company_address_line1"`
	CompanyAddressLine2 *string `json:"company_address_line2"`
	CompanyCity         *string `json:"company_city"`
	CompanyState        *string `json:"company_state"`
	CompanyPostalCode   *string `json:"company_postal_code"`
	CompanyCountry      *string `json:"company_country"`
}

func (h *OrgSettingsHandler) GetCompany(w http.ResponseWriter, r *http.Request) {
	_, s, ok := h.loadSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, companySettingsPayload{
		CompanyName:         s.CompanyName,
		CompanyLogoURL:      s.CompanyLogoURL,
		CompanyWebsite:      s.CompanyWebsite,
		CompanyEmail:        s.CompanyEmail,
		CompanyPhone:        s.CompanyPhone,
		CompanyAddressLine1: s.CompanyAddressLine1,
		CompanyAddressLine2: s.CompanyAddressLine2,
		CompanyCity:         s.CompanyCity,
		CompanyState:        s.CompanyState,
		CompanyPostalCode:   s.CompanyPostalCode,
		CompanyCountry:      s.CompanyCountry,
	})
}

func (h *OrgSettingsHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	var req companySettingsPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	patch := domain.OrgSettingsPatch{
		CompanyName:         req.CompanyName,
		CompanyLogoURL:      req.CompanyLogoURL,
		CompanyWebsite:      req.CompanyWebsite,
		CompanyEmail:        req.CompanyEmail,
		CompanyPhone:        req.CompanyPhone,
		CompanyAddressLine1: req.CompanyAddressLine1,
		CompanyAddressLine2: req.CompanyAddressLine2,
		CompanyCity:         req.CompanyCity,
		CompanyState:        req.CompanyState,
		CompanyPostalCode:   req.CompanyPostalCode,
		CompanyCountry:      req.CompanyCountry,
	}

	claims, before, updated, ok := h.applyPatch(w, r, patch)
	if !ok {
		return
	}

	h.logSettingsMutation(r, claims, "settings.company", auditChanges(
		auditField("company_name", valueStr(before.CompanyName), valueStr(updated.CompanyName)),
		auditField("company_logo_url", valueStr(before.CompanyLogoURL), valueStr(updated.CompanyLogoURL)),
		auditField("company_website", valueStr(before.CompanyWebsite), valueStr(updated.CompanyWebsite)),
		auditField("company_email", valueStr(before.CompanyEmail), valueStr(updated.CompanyEmail)),
		auditField("company_phone", valueStr(before.CompanyPhone), valueStr(updated.CompanyPhone)),
		auditField("company_address_line1", valueStr(before.CompanyAddressLine1), valueStr(updated.CompanyAddressLine1)),
		auditField("company_address_line2", valueStr(before.CompanyAddressLine2), valueStr(updated.CompanyAddressLine2)),
		auditField("company_city", valueStr(before.CompanyCity), valueStr(updated.CompanyCity)),
		auditField("company_state", valueStr(before.CompanyState), valueStr(updated.CompanyState)),
		auditField("company_postal_code", valueStr(before.CompanyPostalCode), valueStr(updated.CompanyPostalCode)),
		auditField("company_country", valueStr(before.CompanyCountry), valueStr(updated.CompanyCountry)),
	))

	h.GetCompany(w, r)
}

type portalSettingsPayload struct {
	PortalEnabled           *bool      `json:"portal_enabled"`
	PortalDisplayName       *string    `json:"portal_display_name"`
	PortalAnnouncement      *string    `json:"portal_announcement"`
	PortalDefaultAssigneeID *uuid.UUID `json:"portal_default_assignee_id"`
	PortalMenu              *[]string  `json:"portal_menu"`
	PortalShortcuts         *[]string  `json:"portal_shortcuts"`
	PortalRecentWidgetLimit *int       `json:"portal_recent_widget_limit"`
}

type portalSettingsResponse struct {
	PortalEnabled           bool       `json:"portal_enabled"`
	PortalDisplayName       *string    `json:"portal_display_name,omitempty"`
	PortalAnnouncement      *string    `json:"portal_announcement,omitempty"`
	PortalDefaultAssigneeID *uuid.UUID `json:"portal_default_assignee_id,omitempty"`
	PortalMenu              []string   `json:"portal_menu"`
	PortalShortcuts         []string   `json:"portal_shortcuts"`
	PortalRecentWidgetLimit int        `json:"portal_recent_widget_limit"`
}

func (h *OrgSettingsHandler) GetPortal(w http.ResponseWriter, r *http.Request) {
	_, s, ok := h.loadSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, portalSettingsResponse{
		PortalEnabled:           s.PortalEnabled,
		PortalDisplayName:       s.PortalDisplayName,
		PortalAnnouncement:      s.PortalAnnouncement,
		PortalDefaultAssigneeID: s.PortalDefaultAssigneeID,
		PortalMenu:              s.PortalMenu,
		PortalShortcuts:         s.PortalShortcuts,
		PortalRecentWidgetLimit: s.PortalRecentWidgetLimit,
	})
}

func (h *OrgSettingsHandler) UpdatePortal(w http.ResponseWriter, r *http.Request) {
	var req portalSettingsPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PortalRecentWidgetLimit != nil && (*req.PortalRecentWidgetLimit < 0 || *req.PortalRecentWidgetLimit > 50) {
		writeError(w, http.StatusUnprocessableEntity, "portal_recent_widget_limit must be between 0 and 50")
		return
	}

	patch := domain.OrgSettingsPatch{
		PortalEnabled:           req.PortalEnabled,
		PortalDisplayName:       req.PortalDisplayName,
		PortalAnnouncement:      req.PortalAnnouncement,
		PortalDefaultAssigneeID: req.PortalDefaultAssigneeID,
		PortalMenu:              req.PortalMenu,
		PortalShortcuts:         req.PortalShortcuts,
		PortalRecentWidgetLimit: req.PortalRecentWidgetLimit,
	}

	claims, before, updated, ok := h.applyPatch(w, r, patch)
	if !ok {
		return
	}

	h.logSettingsMutation(r, claims, "settings.portal", auditChanges(
		auditField("portal_enabled", before.PortalEnabled, updated.PortalEnabled),
		auditField("portal_display_name", valueStr(before.PortalDisplayName), valueStr(updated.PortalDisplayName)),
		auditField("portal_announcement", valueStr(before.PortalAnnouncement), valueStr(updated.PortalAnnouncement)),
		auditField("portal_default_assignee_id", valueUUID(before.PortalDefaultAssigneeID), valueUUID(updated.PortalDefaultAssigneeID)),
		auditField("portal_menu", before.PortalMenu, updated.PortalMenu),
		auditField("portal_shortcuts", before.PortalShortcuts, updated.PortalShortcuts),
		auditField("portal_recent_widget_limit", before.PortalRecentWidgetLimit, updated.PortalRecentWidgetLimit),
	))

	h.GetPortal(w, r)
}

type outgoingServerSettingsPayload struct {
	SMTPHost      *string `json:"smtp_host"`
	SMTPPort      *int    `json:"smtp_port"`
	SMTPUsername  *string `json:"smtp_username"`
	SMTPFromEmail *string `json:"smtp_from_email"`
	SMTPFromName  *string `json:"smtp_from_name"`
	SMTPSecurity  *string `json:"smtp_security"`
	SMTPAuthType  *string `json:"smtp_auth_type"`
	SMTPPassword  *string `json:"smtp_password"`
	ClearPassword bool    `json:"clear_password"`
}

type outgoingServerSettingsResponse struct {
	SMTPHost        *string `json:"smtp_host,omitempty"`
	SMTPPort        *int    `json:"smtp_port,omitempty"`
	SMTPUsername    *string `json:"smtp_username,omitempty"`
	SMTPFromEmail   *string `json:"smtp_from_email,omitempty"`
	SMTPFromName    *string `json:"smtp_from_name,omitempty"`
	SMTPSecurity    *string `json:"smtp_security,omitempty"`
	SMTPAuthType    *string `json:"smtp_auth_type,omitempty"`
	SMTPPasswordSet bool    `json:"smtp_password_set"`
}

func (h *OrgSettingsHandler) GetOutgoingServer(w http.ResponseWriter, r *http.Request) {
	_, s, ok := h.loadSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, outgoingServerSettingsResponse{
		SMTPHost:        s.SMTPHost,
		SMTPPort:        s.SMTPPort,
		SMTPUsername:    s.SMTPUsername,
		SMTPFromEmail:   s.SMTPFromEmail,
		SMTPFromName:    s.SMTPFromName,
		SMTPSecurity:    s.SMTPSecurity,
		SMTPAuthType:    s.SMTPAuthType,
		SMTPPasswordSet: s.SMTPPasswordSet,
	})
}

func (h *OrgSettingsHandler) UpdateOutgoingServer(w http.ResponseWriter, r *http.Request) {
	var req outgoingServerSettingsPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SMTPPort != nil && (*req.SMTPPort < 1 || *req.SMTPPort > 65535) {
		writeError(w, http.StatusUnprocessableEntity, "smtp_port must be between 1 and 65535")
		return
	}
	if req.SMTPSecurity != nil {
		v := strings.ToLower(strings.TrimSpace(*req.SMTPSecurity))
		if v != "" && v != "none" && v != "starttls" && v != "tls" {
			writeError(w, http.StatusUnprocessableEntity, "smtp_security must be one of: none, starttls, tls")
			return
		}
		*req.SMTPSecurity = v
	}
	if req.SMTPAuthType != nil {
		v := strings.ToLower(strings.TrimSpace(*req.SMTPAuthType))
		if v != "" && v != "password" && v != "oauth2" {
			writeError(w, http.StatusUnprocessableEntity, "smtp_auth_type must be one of: password, oauth2")
			return
		}
		*req.SMTPAuthType = v
	}

	patch := domain.OrgSettingsPatch{
		SMTPHost:      req.SMTPHost,
		SMTPPort:      req.SMTPPort,
		SMTPUsername:  req.SMTPUsername,
		SMTPFromEmail: req.SMTPFromEmail,
		SMTPFromName:  req.SMTPFromName,
		SMTPSecurity:  req.SMTPSecurity,
		SMTPAuthType:  req.SMTPAuthType,
	}

	if req.SMTPPassword != nil && strings.TrimSpace(*req.SMTPPassword) != "" {
		enc, err := auth.Encrypt(h.encKey, *req.SMTPPassword)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		patch.SMTPPasswordEnc = &enc
	} else if req.ClearPassword {
		empty := ""
		patch.SMTPPasswordEnc = &empty
	}

	claims, before, updated, ok := h.applyPatch(w, r, patch)
	if !ok {
		return
	}

	h.logSettingsMutation(r, claims, "settings.outgoing_server", auditChanges(
		auditField("smtp_host", valueStr(before.SMTPHost), valueStr(updated.SMTPHost)),
		auditField("smtp_port", valueInt(before.SMTPPort), valueInt(updated.SMTPPort)),
		auditField("smtp_username", valueStr(before.SMTPUsername), valueStr(updated.SMTPUsername)),
		auditField("smtp_from_email", valueStr(before.SMTPFromEmail), valueStr(updated.SMTPFromEmail)),
		auditField("smtp_from_name", valueStr(before.SMTPFromName), valueStr(updated.SMTPFromName)),
		auditField("smtp_security", valueStr(before.SMTPSecurity), valueStr(updated.SMTPSecurity)),
		auditField("smtp_auth_type", valueStr(before.SMTPAuthType), valueStr(updated.SMTPAuthType)),
		auditField("smtp_password_set", before.SMTPPasswordSet, updated.SMTPPasswordSet),
	))

	h.GetOutgoingServer(w, r)
}

type configEditorPayload struct {
	ConfigSupportEmail     *string `json:"config_support_email"`
	ConfigUploadMaxMB      *int    `json:"config_upload_max_mb"`
	ConfigDefaultPageSize  *int    `json:"config_default_page_size"`
	ConfigListPreviewChars *int    `json:"config_list_preview_chars"`
}

func (h *OrgSettingsHandler) GetConfigEditor(w http.ResponseWriter, r *http.Request) {
	_, s, ok := h.loadSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, configEditorPayload{
		ConfigSupportEmail:     s.ConfigSupportEmail,
		ConfigUploadMaxMB:      &s.ConfigUploadMaxMB,
		ConfigDefaultPageSize:  &s.ConfigDefaultPageSize,
		ConfigListPreviewChars: &s.ConfigListPreviewChars,
	})
}

func (h *OrgSettingsHandler) UpdateConfigEditor(w http.ResponseWriter, r *http.Request) {
	var req configEditorPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ConfigUploadMaxMB != nil && (*req.ConfigUploadMaxMB < 1 || *req.ConfigUploadMaxMB > 1024) {
		writeError(w, http.StatusUnprocessableEntity, "config_upload_max_mb must be between 1 and 1024")
		return
	}
	if req.ConfigDefaultPageSize != nil && (*req.ConfigDefaultPageSize < 1 || *req.ConfigDefaultPageSize > 500) {
		writeError(w, http.StatusUnprocessableEntity, "config_default_page_size must be between 1 and 500")
		return
	}
	if req.ConfigListPreviewChars != nil && (*req.ConfigListPreviewChars < 20 || *req.ConfigListPreviewChars > 2000) {
		writeError(w, http.StatusUnprocessableEntity, "config_list_preview_chars must be between 20 and 2000")
		return
	}

	patch := domain.OrgSettingsPatch{
		ConfigSupportEmail:     req.ConfigSupportEmail,
		ConfigUploadMaxMB:      req.ConfigUploadMaxMB,
		ConfigDefaultPageSize:  req.ConfigDefaultPageSize,
		ConfigListPreviewChars: req.ConfigListPreviewChars,
	}

	claims, before, updated, ok := h.applyPatch(w, r, patch)
	if !ok {
		return
	}

	h.logSettingsMutation(r, claims, "settings.config_editor", auditChanges(
		auditField("config_support_email", valueStr(before.ConfigSupportEmail), valueStr(updated.ConfigSupportEmail)),
		auditField("config_upload_max_mb", before.ConfigUploadMaxMB, updated.ConfigUploadMaxMB),
		auditField("config_default_page_size", before.ConfigDefaultPageSize, updated.ConfigDefaultPageSize),
		auditField("config_list_preview_chars", before.ConfigListPreviewChars, updated.ConfigListPreviewChars),
	))

	h.GetConfigEditor(w, r)
}

type menuConfigPayload struct {
	MenuConfig *map[string]bool `json:"menu_config"`
}

type menuConfigResponse struct {
	MenuConfig map[string]bool `json:"menu_config"`
}

func (h *OrgSettingsHandler) GetMenuConfig(w http.ResponseWriter, r *http.Request) {
	_, s, ok := h.loadSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, menuConfigResponse{MenuConfig: s.MenuConfig})
}

func (h *OrgSettingsHandler) UpdateMenuConfig(w http.ResponseWriter, r *http.Request) {
	var req menuConfigPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	patch := domain.OrgSettingsPatch{MenuConfig: req.MenuConfig}
	claims, before, updated, ok := h.applyPatch(w, r, patch)
	if !ok {
		return
	}

	h.logSettingsMutation(r, claims, "settings.menu", auditChanges(
		auditField("menu_config", before.MenuConfig, updated.MenuConfig),
	))

	h.GetMenuConfig(w, r)
}

func (h *OrgSettingsHandler) loadSettings(w http.ResponseWriter, r *http.Request) (*auth.Claims, *domain.OrgSettings, bool) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return nil, nil, false
	}

	s, err := h.settings.GetOrCreate(r.Context(), claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return nil, nil, false
	}
	return claims, s, true
}

func (h *OrgSettingsHandler) applyPatch(w http.ResponseWriter, r *http.Request, patch domain.OrgSettingsPatch) (*auth.Claims, *domain.OrgSettings, *domain.OrgSettings, bool) {
	claims, before, ok := h.loadSettings(w, r)
	if !ok {
		return nil, nil, nil, false
	}

	updated, err := h.settings.Update(r.Context(), claims.OrgID, patch)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return nil, nil, nil, false
	}
	return claims, before, updated, true
}

func (h *OrgSettingsHandler) logSettingsMutation(r *http.Request, claims *auth.Claims, entityName string, changes domain.AuditChanges) {
	if h.auditor.repo == nil || len(changes) == 0 {
		return
	}
	uid := claims.UserID
	name := entityName
	entry := domain.AuditEntry{
		OrgID:      claims.OrgID,
		UserID:     &uid,
		Action:     domain.AuditActionUpdated,
		EntityType: domain.AuditEntityUser,
		EntityName: &name,
		Changes:    changes,
	}
	if ip := realClientIP(r); ip != "" {
		entry.IPAddress = &ip
	}
	if ua := r.UserAgent(); ua != "" {
		entry.UserAgent = &ua
	}
	if err := h.auditor.repo.Append(r.Context(), entry); err != nil {
		slog.Error("audit log write failed", "entity", entityName, "error", err)
	}
}

type fieldDelta struct {
	field string
	from  any
	to    any
}

func auditField(name string, from, to any) fieldDelta {
	return fieldDelta{field: name, from: from, to: to}
}

func auditChanges(fields ...fieldDelta) domain.AuditChanges {
	changes := domain.AuditChanges{}
	for _, f := range fields {
		if reflect.DeepEqual(f.from, f.to) {
			continue
		}
		changes[f.field] = domain.FieldChange{From: f.from, To: f.to}
	}
	return changes
}

func valueStr(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

func valueInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func valueUUID(v *uuid.UUID) any {
	if v == nil {
		return nil
	}
	return v.String()
}
