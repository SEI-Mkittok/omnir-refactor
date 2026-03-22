package domain

import (
	"time"

	"github.com/google/uuid"
)

// TeamsConnection represents a Microsoft Teams Incoming Webhook connection for an org.
type TeamsConnection struct {
	ID               uuid.UUID `json:"id"`
	OrgID            uuid.UUID `json:"org_id"`
	TenantID         string    `json:"tenant_id"`
	BotToken         string    `json:"bot_token"`          // stores the Incoming Webhook URL
	DefaultChannelID string    `json:"default_channel_id"` // display name or ID
	ChannelName      string    `json:"channel_name"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TeamsConnectInput is the payload for POST /api/integrations/teams/connect.
type TeamsConnectInput struct {
	WebhookURL  string `json:"webhook_url"`  // Teams Incoming Webhook URL
	ChannelName string `json:"channel_name"` // human-readable channel name (optional)
	TenantID    string `json:"tenant_id"`    // optional
}
