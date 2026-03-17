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
