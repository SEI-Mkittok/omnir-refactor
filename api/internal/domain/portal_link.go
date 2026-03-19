package domain

import (
	"time"

	"github.com/google/uuid"
)

// PortalLink is a shareable link that gives external clients read-only access
// to a deal snapshot without requiring a CRM account. The token IS the
// credential — revoked or expired links return 404.
type PortalLink struct {
	ID              uuid.UUID  `json:"id"`
	OrgID           uuid.UUID  `json:"org_id"`
	DealID          uuid.UUID  `json:"deal_id"`
	CreatedByUserID *uuid.UUID `json:"created_by_user_id,omitempty"`
	Token           string     `json:"token"`
	Label           *string    `json:"label,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	ViewCount       int        `json:"view_count"`
	LastViewedAt    *time.Time `json:"last_viewed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// IsActive returns false when the link has been revoked or has expired.
func (l *PortalLink) IsActive() bool {
	if l.RevokedAt != nil {
		return false
	}
	if l.ExpiresAt != nil && l.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// DealPortalSnapshot is the public, curated view of a deal returned to
// unauthenticated portal visitors. Internal IDs and org data are omitted.
type DealPortalSnapshot struct {
	Title             string       `json:"title"`
	Stage             DealStage    `json:"stage"`
	ValueCents        int64        `json:"value_cents"`
	Currency          string       `json:"currency"`
	ExpectedCloseDate *time.Time   `json:"expected_close_date,omitempty"`
	OrgName           string       `json:"org_name"`
	LinkLabel         *string      `json:"link_label,omitempty"`
	Notes             []PortalNote `json:"notes"`
}

// PortalNote is a redacted note body suitable for external exposure.
type PortalNote struct {
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}
