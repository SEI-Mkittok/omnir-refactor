package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// IntegrationCredentialRepository manages admin-supplied OAuth app credentials and API keys.
type IntegrationCredentialRepository interface {
	// Upsert creates or replaces the credential record for (org_id, provider).
	// client_secret_enc and api_key_enc must already be encrypted before calling.
	Upsert(ctx context.Context, cred *domain.IntegrationCredential) (*domain.IntegrationCredential, error)

	// GetByProvider returns the credential record for the given org+provider, or ErrNotFound.
	GetByProvider(ctx context.Context, orgID uuid.UUID, provider domain.IntegrationProvider) (*domain.IntegrationCredential, error)

	// ListByOrg returns all credential records for an org.
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*domain.IntegrationCredential, error)

	// Delete removes the credential record for the given org+provider.
	Delete(ctx context.Context, orgID uuid.UUID, provider domain.IntegrationProvider) error
}
