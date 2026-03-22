package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrgOnboarding tracks which onboarding wizard steps an org has completed.
type OrgOnboarding struct {
	OrgID          uuid.UUID  `json:"org_id"`
	CompletedSteps []string   `json:"completed_steps"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// OrgInvite represents a pending team invite sent via the onboarding wizard.
type OrgInvite struct {
	ID         uuid.UUID  `json:"id"`
	OrgID      uuid.UUID  `json:"org_id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Token      string     `json:"token,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
