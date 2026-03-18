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
	Update(ctx context.Context, id uuid.UUID, patch domain.ContactPatch) (*domain.Contact, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.ContactFilter) ([]*domain.Contact, int, error)
}

// AccountRepository defines the persistence contract for accounts.
type AccountRepository interface {
	Create(ctx context.Context, a *domain.Account) (*domain.Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
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
