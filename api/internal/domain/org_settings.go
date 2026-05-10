package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DocType identifies document types that have auto-incrementing numbers.
type DocType string

const (
	DocTypeQuote     DocType = "quote"
	DocTypeTicket    DocType = "ticket"
	DocTypeKBArticle DocType = "kb_article"
	DocTypeInvoice   DocType = "invoice"
)

// OrgSettings holds per-org configuration, including document numbering start values.
type OrgSettings struct {
	ID                      uuid.UUID       `json:"id"`
	OrgID                   uuid.UUID       `json:"org_id"`
	QuoteNumberStart        int64           `json:"quote_number_start"`
	TicketNumberStart       int64           `json:"ticket_number_start"`
	KBArticleNumberStart    int64           `json:"kb_article_number_start"`
	InvoiceNumberStart      int64           `json:"invoice_number_start"`
	QuoteNumberPrefix       string          `json:"quote_number_prefix"`
	TicketNumberPrefix      string          `json:"ticket_number_prefix"`
	KBArticleNumberPrefix   string          `json:"kb_article_number_prefix"`
	InvoiceNumberPrefix     string          `json:"invoice_number_prefix"`
	CompanyName             *string         `json:"company_name,omitempty"`
	CompanyLogoURL          *string         `json:"company_logo_url,omitempty"`
	CompanyWebsite          *string         `json:"company_website,omitempty"`
	CompanyEmail            *string         `json:"company_email,omitempty"`
	CompanyPhone            *string         `json:"company_phone,omitempty"`
	CompanyAddressLine1     *string         `json:"company_address_line1,omitempty"`
	CompanyAddressLine2     *string         `json:"company_address_line2,omitempty"`
	CompanyCity             *string         `json:"company_city,omitempty"`
	CompanyState            *string         `json:"company_state,omitempty"`
	CompanyPostalCode       *string         `json:"company_postal_code,omitempty"`
	CompanyCountry          *string         `json:"company_country,omitempty"`
	PortalEnabled           bool            `json:"portal_enabled"`
	PortalDisplayName       *string         `json:"portal_display_name,omitempty"`
	PortalAnnouncement      *string         `json:"portal_announcement,omitempty"`
	PortalDefaultAssigneeID *uuid.UUID      `json:"portal_default_assignee_id,omitempty"`
	PortalMenu              []string        `json:"portal_menu"`
	PortalShortcuts         []string        `json:"portal_shortcuts"`
	PortalRecentWidgetLimit int             `json:"portal_recent_widget_limit"`
	SMTPHost                *string         `json:"smtp_host,omitempty"`
	SMTPPort                *int            `json:"smtp_port,omitempty"`
	SMTPUsername            *string         `json:"smtp_username,omitempty"`
	SMTPFromEmail           *string         `json:"smtp_from_email,omitempty"`
	SMTPFromName            *string         `json:"smtp_from_name,omitempty"`
	SMTPSecurity            *string         `json:"smtp_security,omitempty"`
	SMTPAuthType            *string         `json:"smtp_auth_type,omitempty"`
	SMTPPasswordEnc         *string         `json:"-"`
	SMTPPasswordSet         bool            `json:"smtp_password_set"`
	ConfigSupportEmail      *string         `json:"config_support_email,omitempty"`
	ConfigUploadMaxMB       int             `json:"config_upload_max_mb"`
	ConfigDefaultPageSize   int             `json:"config_default_page_size"`
	ConfigListPreviewChars  int             `json:"config_list_preview_chars"`
	MenuConfig              map[string]bool `json:"menu_config"`
	CreatedAt               time.Time       `json:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at"`
}

// OrgSettingsPatch holds the updatable fields for org settings.
type OrgSettingsPatch struct {
	QuoteNumberStart        *int64           `json:"quote_number_start"`
	TicketNumberStart       *int64           `json:"ticket_number_start"`
	KBArticleNumberStart    *int64           `json:"kb_article_number_start"`
	InvoiceNumberStart      *int64           `json:"invoice_number_start"`
	QuoteNumberPrefix       *string          `json:"quote_number_prefix"`
	TicketNumberPrefix      *string          `json:"ticket_number_prefix"`
	KBArticleNumberPrefix   *string          `json:"kb_article_number_prefix"`
	InvoiceNumberPrefix     *string          `json:"invoice_number_prefix"`
	CompanyName             *string          `json:"company_name"`
	CompanyLogoURL          *string          `json:"company_logo_url"`
	CompanyWebsite          *string          `json:"company_website"`
	CompanyEmail            *string          `json:"company_email"`
	CompanyPhone            *string          `json:"company_phone"`
	CompanyAddressLine1     *string          `json:"company_address_line1"`
	CompanyAddressLine2     *string          `json:"company_address_line2"`
	CompanyCity             *string          `json:"company_city"`
	CompanyState            *string          `json:"company_state"`
	CompanyPostalCode       *string          `json:"company_postal_code"`
	CompanyCountry          *string          `json:"company_country"`
	PortalEnabled           *bool            `json:"portal_enabled"`
	PortalDisplayName       *string          `json:"portal_display_name"`
	PortalAnnouncement      *string          `json:"portal_announcement"`
	PortalDefaultAssigneeID *uuid.UUID       `json:"portal_default_assignee_id"`
	PortalMenu              *[]string        `json:"portal_menu"`
	PortalShortcuts         *[]string        `json:"portal_shortcuts"`
	PortalRecentWidgetLimit *int             `json:"portal_recent_widget_limit"`
	SMTPHost                *string          `json:"smtp_host"`
	SMTPPort                *int             `json:"smtp_port"`
	SMTPUsername            *string          `json:"smtp_username"`
	SMTPFromEmail           *string          `json:"smtp_from_email"`
	SMTPFromName            *string          `json:"smtp_from_name"`
	SMTPSecurity            *string          `json:"smtp_security"`
	SMTPAuthType            *string          `json:"smtp_auth_type"`
	SMTPPasswordEnc         *string          `json:"-"`
	ConfigSupportEmail      *string          `json:"config_support_email"`
	ConfigUploadMaxMB       *int             `json:"config_upload_max_mb"`
	ConfigDefaultPageSize   *int             `json:"config_default_page_size"`
	ConfigListPreviewChars  *int             `json:"config_list_preview_chars"`
	MenuConfig              *map[string]bool `json:"menu_config"`
}

// FormattedDocNumber formats a number with the given prefix, e.g. QUO-00001.
func FormattedDocNumber(prefix string, number int64) string {
	return fmt.Sprintf("%s-%05d", prefix, number)
}
