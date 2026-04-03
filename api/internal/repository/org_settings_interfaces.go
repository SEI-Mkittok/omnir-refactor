package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// OrgSettingsRepository manages per-org configuration such as document numbering start values.
type OrgSettingsRepository interface {
	// GetOrCreate returns the settings for the org, creating a default row if none exists.
	GetOrCreate(ctx context.Context, orgID uuid.UUID) (*domain.OrgSettings, error)

	// Update applies a patch to the org settings. Only non-nil fields are changed.
	Update(ctx context.Context, orgID uuid.UUID, patch domain.OrgSettingsPatch) (*domain.OrgSettings, error)
}
