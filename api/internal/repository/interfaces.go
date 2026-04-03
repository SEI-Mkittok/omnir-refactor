package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// ContactRepository defines the persistence contract for contacts.
type ContactRepository interface {
	Create(ctx context.Context, c *domain.Contact) (*domain.Contact, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error)
	GetByEmail(ctx context.Context, email string) (*domain.Contact, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.ContactPatch) (*domain.Contact, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.ContactFilter) ([]*domain.Contact, int, error)

	// Lead management (Phase 8)
	UpdateLeadScore(ctx context.Context, id uuid.UUID, patch domain.LeadScorePatch) (*domain.Contact, error)
	ConvertLead(ctx context.Context, id, byUserID uuid.UUID, dealID *uuid.UUID) (*domain.Contact, error)
	ListLeadSources(ctx context.Context) ([]string, error)

	// Email opt-out and bounce tracking (OMN-398)
	SetEmailOptOut(ctx context.Context, contactID uuid.UUID) error
	IncrementBounceCount(ctx context.Context, contactID uuid.UUID) error
}

// AccountRepository defines the persistence contract for accounts.
type AccountRepository interface {
	Create(ctx context.Context, a *domain.Account) (*domain.Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	GetByName(ctx context.Context, name string) (*domain.Account, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.AccountPatch) (*domain.Account, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.AccountFilter) ([]*domain.Account, int, error)
}

// DealRepository defines the persistence contract for deals.
type DealRepository interface {
	Create(ctx context.Context, d *domain.Deal) (*domain.Deal, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Deal, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.DealPatch) (*domain.Deal, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.DealFilter) ([]*domain.Deal, int, error)

	// Many-to-many contact management.
	AddContact(ctx context.Context, dealID, contactID uuid.UUID, role string) error
	ListContacts(ctx context.Context, dealID uuid.UUID) ([]domain.Contact, error)
}

// ActivityRepository defines the persistence contract for activities.
type ActivityRepository interface {
	Create(ctx context.Context, a *domain.Activity) (*domain.Activity, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Activity, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.ActivityPatch) (*domain.Activity, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.ActivityFilter) ([]*domain.Activity, int, error)
}

// NoteRepository defines the persistence contract for notes.
type NoteRepository interface {
	Create(ctx context.Context, n *domain.Note) (*domain.Note, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByEntity(ctx context.Context, filter domain.NoteFilter) ([]*domain.Note, int, error)
}

// NotificationRepository defines the persistence contract for notifications.
type NotificationRepository interface {
	// Create inserts a new notification. Returns it with ID and timestamps set.
	Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error)
	// MarkRead sets read_at for a notification owned by userID.
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	// MarkAllRead sets read_at for all unread notifications for userID in their org.
	MarkAllRead(ctx context.Context, userID, orgID uuid.UUID) error
	// UnreadCount returns the count of unread notifications for a user.
	UnreadCount(ctx context.Context, userID, orgID uuid.UUID) (int, error)
	// ListByUser returns notifications using cursor-based pagination.
	ListByUser(ctx context.Context, filter domain.NotificationFilter) ([]*domain.Notification, error)
	// GenerateReminders scans due activities and inserts missing reminder notifications.
	GenerateReminders(ctx context.Context) error
}

// ReportsRepository defines aggregation queries for the analytics dashboard.
type ReportsRepository interface {
	// DealsByStage returns deal count and total value grouped by stage, org-scoped.
	DealsByStage(ctx context.Context) ([]domain.DealStageMetric, error)
	// ContactsMonthly returns new contact counts per month for the last 12 months.
	ContactsMonthly(ctx context.Context) ([]domain.ContactMonthlyMetric, error)
	// ActivitiesByType returns activity counts grouped by type.
	ActivitiesByType(ctx context.Context) ([]domain.ActivityTypeMetric, error)

	// TicketMetrics returns aggregated ticket metrics for the given date range.
	TicketMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.TicketReport, error)
	// ContactMetrics returns aggregated contact metrics for the given date range.
	ContactMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.ContactReport, error)
	// DealMetrics returns aggregated deal metrics for the given date range.
	DealMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.DealReport, error)
	// LeadMetrics returns aggregated lead metrics for the given date range.
	LeadMetrics(ctx context.Context, filter domain.ReportFilter) (*domain.LeadReport, error)
	// PipelineFunnel returns deal count and value by stage for a pipeline.
	PipelineFunnel(ctx context.Context, pipelineID *uuid.UUID, filter domain.ReportFilter) (*domain.PipelineFunnelReport, error)
	// ConversionRates returns stage-to-stage conversion rate approximations.
	ConversionRates(ctx context.Context, filter domain.ReportFilter) (*domain.ConversionRatesReport, error)
	// RevenueProjection returns projected revenue grouped by expected close month.
	RevenueProjection(ctx context.Context, months int) (*domain.RevenueProjectionReport, error)
	// ActivitySummary returns activity counts by kind and by owner.
	ActivitySummary(ctx context.Context, filter domain.ReportFilter) (*domain.ActivitySummaryReport, error)
}

// UserRepository defines the persistence contract for users.
type UserRepository interface {
	// CountAll returns the total number of non-deleted users across all orgs.
	CountAll(ctx context.Context) (int, error)
	// HasAdminUser returns true if any non-deleted admin user exists across all orgs.
	// Used by the setup handler to determine whether initial setup is still required.
	HasAdminUser(ctx context.Context) (bool, error)
	// Create inserts a new user with a bcrypt password hash.
	Create(ctx context.Context, u *domain.User, passwordHash string) (*domain.User, error)
	// FindByEmail returns the user and their bcrypt password hash by email.
	// Returns nil user (not error) when not found.
	FindByEmail(ctx context.Context, email string) (*domain.User, string, error)
	// GetByID returns a single user by ID, scoped to the org in context.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	// Update applies a partial patch to a user.
	Update(ctx context.Context, id uuid.UUID, patch domain.UserPatch) (*domain.User, error)
	// Delete soft-deletes a user.
	Delete(ctx context.Context, id uuid.UUID) error
	// List returns users matching the filter along with the total count.
	List(ctx context.Context, filter domain.UserFilter) ([]*domain.User, int, error)
}

// LeadRepository defines the persistence contract for leads.
type LeadRepository interface {
	Create(ctx context.Context, l *domain.Lead) (*domain.Lead, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.LeadPatch) (*domain.Lead, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.LeadFilter) ([]*domain.Lead, int, error)
	ListSources(ctx context.Context) ([]string, error)
}

// TicketRepository defines the persistence contract for help-desk tickets.
type TicketRepository interface {
	Create(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	GetDetailByID(ctx context.Context, id uuid.UUID) (*domain.TicketDetail, error)
	GetByEmailMessageID(ctx context.Context, messageID string) (*domain.Ticket, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error)
	UpdateContact(ctx context.Context, id uuid.UUID, contactID *uuid.UUID) (*domain.Ticket, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.TicketFilter) ([]*domain.Ticket, int, error)
}

// TicketCommentRepository defines the persistence contract for ticket comments.
type TicketCommentRepository interface {
	Create(ctx context.Context, c *domain.TicketComment) (*domain.TicketComment, error)
	List(ctx context.Context, filter domain.TicketCommentFilter) ([]*domain.TicketComment, error)
	Delete(ctx context.Context, id, ticketID uuid.UUID) error
}

// TicketAttachmentRepository defines the persistence contract for ticket attachments.
type TicketAttachmentRepository interface {
	Create(ctx context.Context, a *domain.TicketAttachment) (*domain.TicketAttachment, error)
	GetByID(ctx context.Context, id, ticketID uuid.UUID) (*domain.TicketAttachment, error)
	List(ctx context.Context, ticketID uuid.UUID) ([]*domain.TicketAttachment, error)
	Delete(ctx context.Context, id, ticketID uuid.UUID) error
}

// APIKeyRepository manages API keys for external client authentication.
type APIKeyRepository interface {
	// Create inserts a new API key. keyHash must be the SHA-256 hex of the plaintext key.
	Create(ctx context.Context, k *domain.APIKey, keyHash string) (*domain.APIKey, error)
	// GetByHash looks up an API key by its SHA-256 hash. Returns nil (not error) when not found.
	GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
	// List returns all non-revoked keys for the given org.
	List(ctx context.Context, orgID uuid.UUID) ([]*domain.APIKey, error)
	// Revoke sets revoked_at for the given key, scoped to orgID.
	Revoke(ctx context.Context, id, orgID uuid.UUID) error
	// UpdateLastUsed sets last_used_at to now for the given key.
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

// CustomFieldDefinitionRepository manages admin-defined field schemas per entity type.
type CustomFieldDefinitionRepository interface {
	Create(ctx context.Context, def *domain.CustomFieldDefinition) (*domain.CustomFieldDefinition, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomFieldDefinition, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.CustomFieldDefinitionPatch) (*domain.CustomFieldDefinition, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.CustomFieldDefinitionFilter) ([]*domain.CustomFieldDefinition, error)
}

// NotificationPrefRepository manages per-user email notification preferences.
type NotificationPrefRepository interface {
	// GetByUser returns the prefs for a user+org pair, or a default (all-enabled) pref if none exist.
	GetByUser(ctx context.Context, userID, orgID uuid.UUID) (*domain.UserNotificationPref, error)
	// Upsert creates or updates the notification prefs for a user.
	Upsert(ctx context.Context, pref *domain.UserNotificationPref) (*domain.UserNotificationPref, error)
}

// SearchRepository provides full-text search across CRM entities.
type SearchRepository interface {
	// Search returns results grouped by entity type (contacts, accounts, deals,
	// tickets), scoped to the org in context, up to limit per entity type.
	Search(ctx context.Context, q string, limit int) (*domain.SearchGroupedResult, error)
}

// OrgRepository defines the persistence contract for organizations (tenants).
type OrgRepository interface {
	// Create inserts a new organization. Returns ErrConflict when slug is already taken.
	Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	// GetByID returns an org by its UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	// GetBySlug returns an org by its unique slug.
	GetBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	// SlugExists reports whether a given slug is already in use.
	SlugExists(ctx context.Context, slug string) (bool, error)
	// HasAny reports whether at least one organization row exists.
	HasAny(ctx context.Context) (bool, error)
	// List returns all organizations ordered by name. Used by super_admin tenant switcher.
	List(ctx context.Context) ([]*domain.Organization, error)
	// UpdateName sets the display name for the given org.
	UpdateName(ctx context.Context, orgID uuid.UUID, name string) error
}

// EmailRepository defines the persistence contract for contact emails.
type EmailRepository interface {
	// Create inserts a new email record (inbound or outbound).
	Create(ctx context.Context, e *domain.ContactEmail) (*domain.ContactEmail, error)
	// List returns emails matching the filter along with total count.
	List(ctx context.Context, filter domain.EmailFilter) ([]*domain.ContactEmail, int, error)
}

// SLAPolicyRepository defines the persistence contract for SLA policies.
type SLAPolicyRepository interface {
	Create(ctx context.Context, p *domain.SLAPolicy) (*domain.SLAPolicy, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SLAPolicy, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.SLAPolicyPatch) (*domain.SLAPolicy, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID) ([]*domain.SLAPolicy, error)
	// MatchByPriority returns the first SLA policy whose priority_filter includes
	// the given priority, scoped to the org in context. Returns nil, nil when none match.
	MatchByPriority(ctx context.Context, priority domain.TicketPriority) (*domain.SLAPolicy, error)
	// MatchForEntity returns all policies matching the given entity type for the org in context.
	MatchForEntity(ctx context.Context, entityType domain.SLAEntityType) ([]*domain.SLAPolicy, error)
}

// SLAInstanceRepository defines the persistence contract for SLA instances.
type SLAInstanceRepository interface {
	// Create inserts a new SLA instance.
	Create(ctx context.Context, inst *domain.SLAInstance) (*domain.SLAInstance, error)
	// GetByID fetches a single SLA instance by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SLAInstance, error)
	// List returns SLA instances matching the filter.
	List(ctx context.Context, filter domain.SLAInstanceFilter) ([]*domain.SLAInstance, error)
	// MarkResponded sets responded_at for the instance associated with the entity.
	MarkResponded(ctx context.Context, entityID uuid.UUID, entityType domain.SLAEntityType, t time.Time) error
	// MarkResolved sets resolved_at for the instance associated with the entity.
	MarkResolved(ctx context.Context, entityID uuid.UUID, entityType domain.SLAEntityType, t time.Time) error
	// ScanBreaches marks instances as breached where due_at has passed.
	// Returns number of newly breached instances.
	ScanBreaches(ctx context.Context) (int, error)
	// ScanWarnings returns instances that have hit 80% of their SLA window and haven't been warned yet.
	ScanWarnings(ctx context.Context) ([]*domain.SLAInstance, error)
	// MarkWarned records that a warning notification was sent for the given instance.
	MarkWarned(ctx context.Context, id uuid.UUID, t time.Time) error
	// Dashboard returns aggregated SLA health for the org in context.
	Dashboard(ctx context.Context) (*domain.SLADashboard, error)
}

// PortalLinkRepository manages shareable deal portal links for external client access.
type PortalLinkRepository interface {
	// Create inserts a new portal link.
	Create(ctx context.Context, l *domain.PortalLink) (*domain.PortalLink, error)
	// GetByToken returns the portal link for the given token. Returns ErrNotFound when absent.
	// This query does NOT require org context — the token is globally unique.
	GetByToken(ctx context.Context, token string) (*domain.PortalLink, error)
	// ListByDeal returns all non-revoked portal links for a deal, newest first.
	ListByDeal(ctx context.Context, dealID uuid.UUID) ([]*domain.PortalLink, error)
	// Revoke sets revoked_at for the given link, scoped to orgID.
	Revoke(ctx context.Context, id, orgID uuid.UUID) error
	// IncrementView atomically increments view_count and sets last_viewed_at.
	IncrementView(ctx context.Context, id uuid.UUID) error
}

// SavedViewRepository manages named, saved filter views for CRM entity lists.
type SavedViewRepository interface {
	Create(ctx context.Context, v *domain.SavedView) (*domain.SavedView, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SavedView, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.SavedViewPatch) (*domain.SavedView, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.SavedViewFilter) ([]*domain.SavedView, error)
	// Pin toggles the pinned state. If unpinning, pinned_order is cleared.
	// If pinning, pinned_order is set to the next available slot within the org+entityType.
	Pin(ctx context.Context, id uuid.UUID, isPinned bool) (*domain.SavedView, error)
}

// OutboundWebhookRepository manages registered outbound webhook endpoints and their deliveries.
type OutboundWebhookRepository interface {
	Create(ctx context.Context, w *domain.Webhook) (*domain.Webhook, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error)
	List(ctx context.Context) ([]*domain.Webhook, error)
	ListByEvent(ctx context.Context, orgID uuid.UUID, event domain.WebhookEvent) ([]*domain.Webhook, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.WebhookPatch) (*domain.Webhook, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CreateDelivery(ctx context.Context, d *domain.WebhookDelivery) (*domain.WebhookDelivery, error)
	UpdateDelivery(ctx context.Context, d *domain.WebhookDelivery) error
	PendingDeliveries(ctx context.Context) ([]*domain.WebhookDelivery, error)
	ListDeliveries(ctx context.Context, webhookID uuid.UUID, limit int) ([]*domain.WebhookDelivery, error)
}

// ProductRepository defines the persistence contract for products.
type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) (*domain.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.ProductPatch) (*domain.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.ProductFilter) ([]*domain.Product, int, error)
}

// QuoteRepository defines the persistence contract for quotes.
type QuoteRepository interface {
	Create(ctx context.Context, q *domain.Quote) (*domain.Quote, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Quote, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.QuotePatch) (*domain.Quote, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.QuoteFilter) ([]*domain.Quote, int, error)
	// ReplaceLineItems atomically replaces all line items for a quote.
	ReplaceLineItems(ctx context.Context, quoteID uuid.UUID, items []domain.QuoteLineItemInput) ([]domain.QuoteLineItem, error)
	// MarkSent sets status=sent and sent_at=now.
	MarkSent(ctx context.Context, id uuid.UUID) (*domain.Quote, error)
	// MarkApproved sets status=approved and approved_at=now.
	MarkApproved(ctx context.Context, id uuid.UUID) (*domain.Quote, error)
	// MarkRejected sets status=rejected and rejected_at=now.
	MarkRejected(ctx context.Context, id uuid.UUID) (*domain.Quote, error)
}

// EntityAttachmentRepository manages file attachments for contacts, accounts, and deals.
type EntityAttachmentRepository interface {
	// Create inserts a new attachment record.
	Create(ctx context.Context, a *domain.EntityAttachment) (*domain.EntityAttachment, error)
	// List returns all attachments for a given entity, ordered by created_at ASC.
	List(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.EntityAttachment, error)
	// GetByID returns a single attachment by ID (for download authorization).
	GetByID(ctx context.Context, id uuid.UUID) (*domain.EntityAttachment, error)
	// Delete removes an attachment record. Returns ErrNotFound when absent.
	Delete(ctx context.Context, id uuid.UUID, entityType domain.EntityType, entityID uuid.UUID) error
}

// SSOConfigRepository manages per-org OIDC SSO configurations.
type SSOConfigRepository interface {
	// Upsert creates or updates the SSO config for an org.
	Upsert(ctx context.Context, cfg *domain.SSOConfig) (*domain.SSOConfig, error)
	// GetByOrgID returns the SSO config for the given org. Returns nil, nil when not found.
	GetByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.SSOConfig, error)
	// GetByOrgSlug returns the SSO config by org slug. Used during SSO login flow.
	GetByOrgSlug(ctx context.Context, slug string) (*domain.SSOConfig, error)
}

// TOTPRepository manages TOTP backup codes and user 2FA state.
type TOTPRepository interface {
	// SetSecret stores the (encrypted) TOTP secret for a user.
	SetSecret(ctx context.Context, userID uuid.UUID, encryptedSecret string) error
	// GetSecret returns the encrypted TOTP secret for a user.
	GetSecret(ctx context.Context, userID uuid.UUID) (string, error)
	// Activate marks TOTP as enabled for the user and replaces backup codes.
	Activate(ctx context.Context, userID uuid.UUID, hashedCodes []string) error
	// Disable clears the TOTP secret and backup codes for a user.
	Disable(ctx context.Context, userID uuid.UUID) error
	// IsEnabled returns whether TOTP is enabled for a user.
	IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error)
	// FindUnusedBackupCode returns all unused backup code rows for the user,
	// or (nil, nil) when not found.
	FindUnusedBackupCode(ctx context.Context, userID uuid.UUID) ([]*domain.TOTPBackupCode, error)
	// MarkBackupCodeUsed marks a backup code as used.
	MarkBackupCodeUsed(ctx context.Context, codeID uuid.UUID) error
}

// EnrichmentCacheRepository manages domain enrichment cache entries.
type EnrichmentCacheRepository interface {
	// GetByDomain returns the cached enrichment for a domain, or nil if not found.
	// TTL enforcement (30-day expiry) is handled by the caller.
	GetByDomain(ctx context.Context, d string) (*domain.EnrichmentCache, error)
	// Upsert inserts or replaces the enrichment cache entry for the given domain.
	Upsert(ctx context.Context, entry *domain.EnrichmentCache) (*domain.EnrichmentCache, error)
}

// KBCategoryRepository defines persistence for knowledge base categories.
type KBCategoryRepository interface {
	Create(ctx context.Context, c *domain.KBCategory) (*domain.KBCategory, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.KBCategory, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.KBCategoryPatch) (*domain.KBCategory, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.KBCategoryFilter) ([]*domain.KBCategory, int, error)
}

// KBArticleRepository defines persistence for knowledge base articles.
type KBArticleRepository interface {
	Create(ctx context.Context, a *domain.KBArticle) (*domain.KBArticle, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.KBArticle, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.KBArticlePatch) (*domain.KBArticle, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.KBArticleFilter) ([]*domain.KBArticle, int, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	Suggest(ctx context.Context, orgID uuid.UUID, subject string, limit int) ([]*domain.KBSuggestResult, error)
}

// TeamsConnectionRepository manages Microsoft Teams Incoming Webhook connections per org.
type TeamsConnectionRepository interface {
	// Upsert creates or replaces the Teams connection for an org (one connection per org).
	Upsert(ctx context.Context, c *domain.TeamsConnection) (*domain.TeamsConnection, error)
	// GetByOrgID returns the Teams connection for the given org, or nil if not configured.
	GetByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.TeamsConnection, error)
	// Delete removes the Teams connection for an org.
	Delete(ctx context.Context, orgID uuid.UUID) error
}

// BillingRepository manages org billing plans and invoice records.
type BillingRepository interface {
	// GetOrCreatePlan returns the org's billing plan, inserting a free-tier record if absent.
	GetOrCreatePlan(ctx context.Context, orgID uuid.UUID) (*domain.OrgPlanRecord, error)
	// UpsertPlan applies a patch to the org's billing plan (webhook-driven updates).
	UpsertPlan(ctx context.Context, orgID uuid.UUID, patch domain.OrgPlanPatch) (*domain.OrgPlanRecord, error)
	// GetPlanByStripeSubscriptionID looks up a plan by Stripe subscription ID.
	GetPlanByStripeSubscriptionID(ctx context.Context, subID string) (*domain.OrgPlanRecord, error)
	// GetPlanByStripeCustomerID looks up a plan by Stripe customer ID.
	GetPlanByStripeCustomerID(ctx context.Context, customerID string) (*domain.OrgPlanRecord, error)
	// UpsertInvoice inserts or updates an invoice by stripe_invoice_id.
	UpsertInvoice(ctx context.Context, inv *domain.Invoice) (*domain.Invoice, error)
	// ListInvoices returns invoices for an org ordered by created_at desc.
	ListInvoices(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.Invoice, int, error)
	// GetUsageStats returns current user and contact counts for an org.
	GetUsageStats(ctx context.Context, orgID uuid.UUID) (userCount, contactCount int, err error)
}

// DashboardRepository defines persistence for custom dashboards.
type DashboardRepository interface {
	// CreateDashboard inserts a new dashboard.
	CreateDashboard(ctx context.Context, d *domain.CustomDashboard) (*domain.CustomDashboard, error)
	// GetDashboardByID returns a dashboard by ID, scoped to the org in context.
	GetDashboardByID(ctx context.Context, id uuid.UUID) (*domain.CustomDashboard, error)
	// UpdateDashboard applies a partial update to a dashboard.
	UpdateDashboard(ctx context.Context, id uuid.UUID, patch domain.CustomDashboardPatch) (*domain.CustomDashboard, error)
	// DeleteDashboard removes a dashboard (and cascades to scheduled_reports).
	DeleteDashboard(ctx context.Context, id uuid.UUID) error
	// ListDashboards returns all dashboards for the org in context.
	ListDashboards(ctx context.Context) ([]*domain.CustomDashboard, error)

	// CreateSchedule inserts a new scheduled report.
	CreateSchedule(ctx context.Context, s *domain.ScheduledReport) (*domain.ScheduledReport, error)
	// GetScheduleByID returns a scheduled report by ID, scoped to org.
	GetScheduleByID(ctx context.Context, id uuid.UUID) (*domain.ScheduledReport, error)
	// UpdateSchedule applies a partial update to a scheduled report.
	UpdateSchedule(ctx context.Context, id uuid.UUID, patch domain.ScheduledReportPatch) (*domain.ScheduledReport, error)
	// DeleteSchedule removes a scheduled report.
	DeleteSchedule(ctx context.Context, id uuid.UUID) error
	// ListSchedules returns all scheduled reports for the org in context.
	ListSchedules(ctx context.Context) ([]*domain.ScheduledReport, error)
	// ListAllDueSchedules returns all scheduled reports across all orgs that are due to fire at t.
	// Used by the background scheduler worker.
	ListAllDueSchedules(ctx context.Context, t time.Time) ([]*domain.ScheduledReport, error)
	// MarkScheduleSent updates last_sent_at for a schedule after delivery.
	MarkScheduleSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error
}

// OnboardingRepository manages wizard state and team invites per org.
type OnboardingRepository interface {
	// GetOrCreate returns the onboarding record for the org, creating it if absent.
	GetOrCreate(ctx context.Context, orgID uuid.UUID) (*domain.OrgOnboarding, error)
	// UpdateSteps persists the updated completed_steps list and completed_at timestamp.
	UpdateSteps(ctx context.Context, orgID uuid.UUID, steps []string, completedAt *time.Time) (*domain.OrgOnboarding, error)
	// CreateInvite inserts a new org invite.
	CreateInvite(ctx context.Context, invite *domain.OrgInvite) (*domain.OrgInvite, error)
	// GetInviteByToken returns an invite by its token, or nil if not found.
	GetInviteByToken(ctx context.Context, token string) (*domain.OrgInvite, error)
	// AcceptInvite marks an invite as accepted.
	AcceptInvite(ctx context.Context, inviteID uuid.UUID) error
	// ListInvites returns all invites for an org.
	ListInvites(ctx context.Context, orgID uuid.UUID) ([]*domain.OrgInvite, error)
}

// PushSubscriptionRepository manages Web Push subscriptions for users.
type PushSubscriptionRepository interface {
	// Upsert stores a push subscription for a user, keyed on (user_id, endpoint).
	Upsert(ctx context.Context, s *domain.PushSubscription) (*domain.PushSubscription, error)
	// DeleteByEndpoint removes the subscription with the given endpoint for a user.
	DeleteByEndpoint(ctx context.Context, userID uuid.UUID, endpoint string) error
	// ListByUser returns all active push subscriptions for a user within an org.
	ListByUser(ctx context.Context, userID, orgID uuid.UUID) ([]*domain.PushSubscription, error)
}
