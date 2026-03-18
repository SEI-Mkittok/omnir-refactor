package repository

import (
	"context"

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
	// Create inserts a single notification. Duplicates (same activity + type) are silently ignored.
	Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error)
	// MarkRead sets read_at for a notification owned by userID.
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	// ListByUser returns notifications for a user, with optional unread filter.
	ListByUser(ctx context.Context, filter domain.NotificationFilter) ([]*domain.Notification, int, error)
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
}

// UserRepository defines the persistence contract for users.
type UserRepository interface {
	// CountAll returns the total number of non-deleted users across all orgs.
	CountAll(ctx context.Context) (int, error)
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
}

// TicketRepository defines the persistence contract for help-desk tickets.
type TicketRepository interface {
	Create(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	GetByEmailMessageID(ctx context.Context, messageID string) (*domain.Ticket, error)
	Update(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error)
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
	// Search returns up to limit results across tickets, contacts, accounts, and deals,
	// scoped to the org in context. Returns the flat result list and total count.
	Search(ctx context.Context, q string, limit int) ([]domain.SearchResultItem, int, error)
}

// OrgRepository defines the persistence contract for organizations (tenants).
type OrgRepository interface {
	// Create inserts a new organization. Returns ErrConflict when slug is already taken.
	Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	// GetBySlug returns an org by its unique slug.
	GetBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	// SlugExists reports whether a given slug is already in use.
	SlugExists(ctx context.Context, slug string) (bool, error)
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
}
