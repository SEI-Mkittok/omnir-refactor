-- +goose Up

-- SSO provider configuration per org.
CREATE TABLE sso_configs (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID        NOT NULL UNIQUE REFERENCES orgs(id) ON DELETE CASCADE,
    provider        TEXT        NOT NULL CHECK (provider IN ('google','oidc')),
    client_id       TEXT        NOT NULL,
    client_secret   TEXT        NOT NULL,   -- AES-256-GCM encrypted
    issuer_url      TEXT        NOT NULL,
    attribute_mapping JSONB,                -- map OIDC claims → role
    enabled         BOOL        NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sso_configs_org_id ON sso_configs (org_id);

ALTER TABLE sso_configs ENABLE ROW LEVEL SECURITY;
CREATE POLICY sso_configs_org_isolation ON sso_configs
    USING (org_id = current_setting('app.current_org_id', true)::uuid);

-- TOTP backup codes: single-use, bcrypt-hashed.
CREATE TABLE totp_backup_codes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash  TEXT        NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_totp_backup_codes_user_id ON totp_backup_codes (user_id);

-- Add 2FA columns to users.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS totp_secret  TEXT,
    ADD COLUMN IF NOT EXISTS totp_enabled BOOL NOT NULL DEFAULT false;

-- Add SSO/2FA policy columns to orgs.
ALTER TABLE orgs
    ADD COLUMN IF NOT EXISTS require_2fa    BOOL NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS sso_required   BOOL NOT NULL DEFAULT false;

-- +goose Down

ALTER TABLE orgs
    DROP COLUMN IF EXISTS sso_required,
    DROP COLUMN IF EXISTS require_2fa;

ALTER TABLE users
    DROP COLUMN IF EXISTS totp_enabled,
    DROP COLUMN IF EXISTS totp_secret;

DROP TABLE IF EXISTS totp_backup_codes;
DROP TABLE IF EXISTS sso_configs;
