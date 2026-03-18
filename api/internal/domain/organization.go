package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OrgPlan string

const (
	// OrgPlanSingle is the default plan for self-hosted / single-tenant deployments.
	OrgPlanSingle     OrgPlan = "single"
	OrgPlanStarter    OrgPlan = "starter"
	OrgPlanPro        OrgPlan = "pro"
	OrgPlanEnterprise OrgPlan = "enterprise"
)

// DefaultOrgID is the seed organization used for self-hosted deployments.
var DefaultOrgID = uuid.MustParse("00000000-0000-0000-0000-000000000002")

// Organization represents a tenant in the CRM.
type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Plan      OrgPlan   `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
}

type orgContextKey struct{}

// WithOrgID stores an org ID in the context.
func WithOrgID(ctx context.Context, orgID uuid.UUID) context.Context {
	return context.WithValue(ctx, orgContextKey{}, orgID)
}

// OrgIDFromContext retrieves the org ID from the context.
// Returns (orgID, true) if present, (uuid.Nil, false) otherwise.
func OrgIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(orgContextKey{}).(uuid.UUID)
	return id, ok
}
