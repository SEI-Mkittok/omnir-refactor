package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// CalendarConnectionRepository defines the persistence contract for calendar OAuth connections.
type CalendarConnectionRepository interface {
	// Upsert creates or replaces a connection for (org_id, user_id, provider).
	Upsert(ctx context.Context, c *domain.CalendarConnection) (*domain.CalendarConnection, error)

	// GetByID fetches a single connection by its UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CalendarConnection, error)

	// GetByUserAndProvider returns the connection for a specific user/provider pair.
	GetByUserAndProvider(ctx context.Context, orgID, userID uuid.UUID, provider domain.CalendarProvider) (*domain.CalendarConnection, error)

	// List returns all connections matching the filter.
	List(ctx context.Context, filter domain.CalendarConnectionFilter) ([]*domain.CalendarConnection, error)

	// Update patches mutable fields on an existing connection (tokens, cursor, etc.).
	Update(ctx context.Context, id uuid.UUID, patch domain.CalendarConnectionPatch) (*domain.CalendarConnection, error)

	// Delete removes a connection by ID (disconnect flow).
	Delete(ctx context.Context, id uuid.UUID) error

	// ListAllActive returns every connection across all orgs for the sync worker.
	ListAllActive(ctx context.Context) ([]*domain.CalendarConnection, error)
}
