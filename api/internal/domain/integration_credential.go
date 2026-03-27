package domain

import (
	"time"

	"github.com/google/uuid"
)

// IntegrationProvider identifies a supported third-party integration.
type IntegrationProvider string

const (
	IntegrationProviderGmail        IntegrationProvider = "gmail"
	IntegrationProviderOutlook      IntegrationProvider = "outlook"
	IntegrationProviderGoogleCal    IntegrationProvider = "google_calendar"
	IntegrationProviderOutlookCal   IntegrationProvider = "outlook_calendar"
	IntegrationProviderConfluence   IntegrationProvider = "confluence"
	IntegrationProviderSlack        IntegrationProvider = "slack"
	IntegrationProviderTeams        IntegrationProvider = "microsoft_teams"
	IntegrationProviderStripe       IntegrationProvider = "stripe"
	IntegrationProviderSendGrid     IntegrationProvider = "sendgrid"
	IntegrationProviderTwilio       IntegrationProvider = "twilio"
	IntegrationProviderZapier       IntegrationProvider = "zapier"
)

// IntegrationConnectionStatus describes the current state of an integration.
type IntegrationConnectionStatus string

const (
	IntegrationStatusConnected    IntegrationConnectionStatus = "connected"
	IntegrationStatusDisconnected IntegrationConnectionStatus = "disconnected"
	IntegrationStatusComingSoon   IntegrationConnectionStatus = "coming_soon"
)

// IntegrationCredential stores admin-supplied OAuth app credentials or API keys for a provider.
type IntegrationCredential struct {
	ID               uuid.UUID           `json:"id"`
	OrgID            uuid.UUID           `json:"org_id"`
	Provider         IntegrationProvider `json:"provider"`
	ClientID         *string             `json:"client_id,omitempty"`
	ClientSecretEnc  *string             `json:"-"`
	APIKeyEnc        *string             `json:"-"`
	WebhookSecret    *string             `json:"webhook_secret,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

// IntegrationStatus is the aggregated view returned by GET /integrations.
type IntegrationStatus struct {
	Provider         IntegrationProvider         `json:"provider"`
	Status           IntegrationConnectionStatus `json:"status"`
	EmailAddress     *string                     `json:"email_address,omitempty"`
	LastSyncedAt     *time.Time                  `json:"last_synced_at,omitempty"`
	HasCustomCreds   bool                        `json:"has_custom_creds"`
}

// UpsertIntegrationCredentialParams holds the data for creating or updating credentials.
type UpsertIntegrationCredentialParams struct {
	OrgID         uuid.UUID
	Provider      IntegrationProvider
	ClientID      *string
	ClientSecret  *string // plaintext — handler encrypts before calling repo
	APIKey        *string // plaintext — handler encrypts before calling repo
	WebhookSecret *string
}
