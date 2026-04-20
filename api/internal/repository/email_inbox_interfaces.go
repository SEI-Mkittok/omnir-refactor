package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// EmailConnectionRepository defines the persistence contract for Gmail/Outlook OAuth connections.
type EmailConnectionRepository interface {
	// Upsert creates or replaces a connection for (org_id, user_id, provider).
	Upsert(ctx context.Context, c *domain.EmailConnection) (*domain.EmailConnection, error)

	// GetByID fetches a single connection by its UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailConnection, error)

	// GetByUserAndProvider returns the connection for a specific user/provider pair.
	GetByUserAndProvider(ctx context.Context, orgID, userID uuid.UUID, provider domain.EmailProvider) (*domain.EmailConnection, error)

	// List returns all connections matching the filter.
	List(ctx context.Context, filter domain.EmailConnectionFilter) ([]*domain.EmailConnection, error)

	// Update patches mutable fields (tokens, cursor, last_synced_at) on a connection.
	Update(ctx context.Context, id uuid.UUID, patch domain.EmailConnectionPatch) (*domain.EmailConnection, error)

	// Delete removes a connection by ID (disconnect flow).
	Delete(ctx context.Context, id uuid.UUID) error

	// ListAllActive returns every connection across all orgs for the sync worker.
	ListAllActive(ctx context.Context) ([]*domain.EmailConnection, error)
}

// EmailInboxRepository defines the persistence contract for synced inbox messages.
type EmailInboxRepository interface {
	// Upsert inserts a message or ignores if (connection_id, message_id) already exists.
	Upsert(ctx context.Context, msg *domain.EmailInboxMessage) (*domain.EmailInboxMessage, error)

	// List returns messages matching the filter, newest first.
	List(ctx context.Context, filter domain.EmailInboxFilter) ([]*domain.EmailInboxMessage, int, error)

	// ListThreads returns thread summaries matching the filter, newest thread first.
	ListThreads(ctx context.Context, filter domain.EmailInboxFilter) ([]*domain.EmailInboxThreadSummary, int, error)

	// GetThread returns all messages with the given thread_id, ordered by sent_at asc.
	GetThread(ctx context.Context, orgID uuid.UUID, threadID string) ([]*domain.EmailInboxMessage, error)

	// LinkContact sets contact_id on all messages matching from_addr within the org.
	LinkContact(ctx context.Context, orgID uuid.UUID, addr string, contactID uuid.UUID) error

	// MarkThreadRead sets read_at = NOW() on all unread messages in the given thread.
	MarkThreadRead(ctx context.Context, orgID uuid.UUID, threadID string) error
}
