package config

import (
	"strings"
	"testing"
)

func TestValidate_DevelopmentAlwaysPasses(t *testing.T) {
	cfg := &Config{
		Env: "development",
		// all secrets intentionally empty — dev mode must not reject them
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error in development, got: %v", err)
	}
}

func TestValidate_ProductionRequiresSecrets(t *testing.T) {
	base := func() *Config {
		return &Config{
			Env:                          "production",
			JWTSecret:                    "strong-jwt-secret-abc123",
			SequenceTokenSecret:          "strong-seq-secret-xyz789",
			SSOEncryptionKey:             "strong-sso-enc-key-32bytes!!",
			IntegrationCredentialsEncKey: "strong-integ-creds-key-32bytes!",
			EmailInbox: EmailInboxConfig{
				EncryptionKey: "strong-email-enc-key-32bytes!!",
			},
		}
	}

	t.Run("all secrets set — passes", func(t *testing.T) {
		if err := base().Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing JWT_SECRET — fails", func(t *testing.T) {
		cfg := base()
		cfg.JWTSecret = ""
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected error for empty JWT_SECRET")
		}
		if !strings.Contains(err.Error(), "JWT_SECRET") {
			t.Errorf("error should mention JWT_SECRET: %v", err)
		}
	})

	t.Run("dev-default JWT_SECRET — fails", func(t *testing.T) {
		cfg := base()
		cfg.JWTSecret = "dev-secret-change-in-production"
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected error for dev-default JWT_SECRET")
		}
		if !strings.Contains(err.Error(), "JWT_SECRET") {
			t.Errorf("error should mention JWT_SECRET: %v", err)
		}
	})

	t.Run("dev-default SEQUENCE_TOKEN_SECRET — fails", func(t *testing.T) {
		cfg := base()
		cfg.SequenceTokenSecret = "sequence-dev-secret"
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected error for dev-default SEQUENCE_TOKEN_SECRET")
		}
	})

	t.Run("dev-default EMAIL_INBOX_ENCRYPTION_KEY — fails", func(t *testing.T) {
		cfg := base()
		cfg.EmailInbox.EncryptionKey = "dev-email-inbox-key-change-in-prod"
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected error for dev-default EMAIL_INBOX_ENCRYPTION_KEY")
		}
	})

	t.Run("dev-default SSO_ENCRYPTION_KEY — fails", func(t *testing.T) {
		cfg := base()
		cfg.SSOEncryptionKey = "dev-sso-encryption-key-change-in-prod"
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected error for dev-default SSO_ENCRYPTION_KEY")
		}
	})

	t.Run("dev-default INTEGRATION_CREDENTIALS_ENCRYPTION_KEY — fails", func(t *testing.T) {
		cfg := base()
		cfg.IntegrationCredentialsEncKey = "dev-integration-creds-key-change-in-prod"
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected error for dev-default INTEGRATION_CREDENTIALS_ENCRYPTION_KEY")
		}
	})

	t.Run("multiple missing — all reported", func(t *testing.T) {
		cfg := &Config{
			Env:         "production",
			JWTSecret:   "",
			SSOEncryptionKey: "",
		}
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected error")
		}
		for _, key := range []string{"JWT_SECRET", "SEQUENCE_TOKEN_SECRET", "SSO_ENCRYPTION_KEY"} {
			if !strings.Contains(err.Error(), key) {
				t.Errorf("expected error to mention %s: %v", key, err)
			}
		}
	})
}

func TestValidate_StagingEnv(t *testing.T) {
	cfg := &Config{
		Env:                          "staging",
		JWTSecret:                    "strong-jwt-secret-abc123",
		SequenceTokenSecret:          "strong-seq-secret-xyz789",
		SSOEncryptionKey:             "strong-sso-enc-key-32bytes!!",
		IntegrationCredentialsEncKey: "strong-integ-creds-key-32bytes!",
		EmailInbox: EmailInboxConfig{
			EncryptionKey: "strong-email-enc-key-32bytes!!",
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error for staging with strong secrets: %v", err)
	}
}
