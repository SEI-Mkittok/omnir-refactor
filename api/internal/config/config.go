package config

import (
	"os"
	"strings"
)

// StorageBackend selects the file storage implementation.
type StorageBackend string

const (
	StorageBackendLocal StorageBackend = "local"
	StorageBackendS3    StorageBackend = "s3"
)

// StorageConfig holds configuration for the file storage backend.
type StorageConfig struct {
	Backend StorageBackend

	// S3 / S3-compatible (MinIO) settings
	S3Endpoint  string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3UseSSL    bool

	// Local filesystem settings
	LocalBasePath string
}

// OrgMode controls multi-tenancy behaviour.
type OrgMode string

const (
	OrgModeSingle      OrgMode = "single"      // one org, no tenant isolation
	OrgModeSaaS        OrgMode = "saas"        // SaaS multi-tenant — org_id required in JWT, RLS enforced
	OrgModeMultitenant OrgMode = "multitenant" // alias for saas (legacy name)
	OrgModeEnterprise  OrgMode = "enterprise"  // enterprise — same as saas with stricter isolation
)

// CalendarConfig holds OAuth 2.0 credentials for Google Calendar and Microsoft Graph.
type CalendarConfig struct {
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURL     string
	MicrosoftClientID     string
	MicrosoftClientSecret string
	MicrosoftRedirectURL  string
	MicrosoftTenantID     string // "common" for multi-tenant
}

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	Env                 string
	Port                string
	DatabaseURL         string
	JWTSecret           string
	CORSOrigins         []string
	LogLevel            string
	OrgMode             OrgMode
	SMTP                SMTPConfig
	WebhookSecret       string
	SequenceTokenSecret string
	Storage             StorageConfig
	Calendar            CalendarConfig
	SSOEncryptionKey    string
	SSOCallbackURL      string
}

// SMTPConfig holds SMTP connection and sender settings.
type SMTPConfig struct {
	Enabled  bool
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Load reads configuration from environment variables.
// Defaults are set for local development.
func Load() *Config {
	return &Config{
		Env:         getEnv("ENV", "development"),
		Port:        getEnv("SERVER_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://omnir:omnir_dev@localhost:5432/omnir_crm?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		CORSOrigins: splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173")),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		OrgMode:     OrgMode(getEnv("ORG_MODE", string(OrgModeSingle))),
		SMTP: SMTPConfig{
			Enabled:  getEnv("SMTP_ENABLED", "false") == "true",
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@omnir.io"),
		},
		WebhookSecret:       getEnv("WEBHOOK_SECRET", ""),
		SequenceTokenSecret: getEnv("SEQUENCE_TOKEN_SECRET", "sequence-dev-secret"),
		Storage: StorageConfig{
			Backend:       StorageBackend(getEnv("STORAGE_BACKEND", string(StorageBackendLocal))),
			S3Endpoint:    getEnv("S3_ENDPOINT", ""),
			S3Bucket:      getEnv("S3_BUCKET", "omnir-attachments"),
			S3AccessKey:   getEnv("S3_ACCESS_KEY", ""),
			S3SecretKey:   getEnv("S3_SECRET_KEY", ""),
			S3UseSSL:      getEnv("S3_USE_SSL", "true") == "true",
			LocalBasePath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
		},
		Calendar: CalendarConfig{
			GoogleClientID:        getEnv("GOOGLE_CALENDAR_CLIENT_ID", ""),
			GoogleClientSecret:    getEnv("GOOGLE_CALENDAR_CLIENT_SECRET", ""),
			GoogleRedirectURL:     getEnv("GOOGLE_CALENDAR_REDIRECT_URL", "http://localhost:8080/api/v1/calendar/auth/google/callback"),
			MicrosoftClientID:     getEnv("MICROSOFT_CALENDAR_CLIENT_ID", ""),
			MicrosoftClientSecret: getEnv("MICROSOFT_CALENDAR_CLIENT_SECRET", ""),
			MicrosoftRedirectURL:  getEnv("MICROSOFT_CALENDAR_REDIRECT_URL", "http://localhost:8080/api/v1/calendar/auth/microsoft/callback"),
			MicrosoftTenantID:     getEnv("MICROSOFT_TENANT_ID", "common"),
		},
		SSOEncryptionKey: getEnv("SSO_ENCRYPTION_KEY", "dev-sso-encryption-key-change-in-prod"),
		SSOCallbackURL:   getEnv("SSO_CALLBACK_URL", "http://localhost:8080/auth/sso/callback"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
