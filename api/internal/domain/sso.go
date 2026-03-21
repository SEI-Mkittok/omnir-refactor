package domain

import (
	"time"

	"github.com/google/uuid"
)

// SSOConfig holds the OIDC provider configuration for an org.
type SSOConfig struct {
	ID               uuid.UUID         `json:"id"`
	OrgID            uuid.UUID         `json:"org_id"`
	Provider         string            `json:"provider"` // google | oidc
	ClientID         string            `json:"client_id"`
	ClientSecret     string            `json:"-"` // never serialised
	IssuerURL        string            `json:"issuer_url"`
	AttributeMapping map[string]string `json:"attribute_mapping,omitempty"`
	Enabled          bool              `json:"enabled"`
	CreatedAt        time.Time         `json:"created_at"`
}

// TOTPBackupCode is a single-use bcrypt-hashed recovery code.
type TOTPBackupCode struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	CodeHash  string     `json:"-"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// TOTPSetupResult is returned from the 2FA setup endpoint.
type TOTPSetupResult struct {
	OTPAuthURI  string   `json:"otp_auth_uri"`
	QRDataURL   string   `json:"qr_data_url"`
	BackupCodes []string `json:"backup_codes"` // plain-text, shown once
}
