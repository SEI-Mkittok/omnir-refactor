-- +goose Up

-- Extend sso_configs to support Microsoft (Entra ID) and Google Workspace providers.
ALTER TABLE sso_configs
    DROP CONSTRAINT IF EXISTS sso_configs_provider_check;

ALTER TABLE sso_configs
    ADD CONSTRAINT sso_configs_provider_check
        CHECK (provider IN ('google', 'oidc', 'microsoft')),
    ADD COLUMN IF NOT EXISTS tenant_id  TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS hd         TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- +goose Down

ALTER TABLE sso_configs
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS hd,
    DROP COLUMN IF EXISTS tenant_id;

ALTER TABLE sso_configs
    DROP CONSTRAINT IF EXISTS sso_configs_provider_check;

ALTER TABLE sso_configs
    ADD CONSTRAINT sso_configs_provider_check
        CHECK (provider IN ('google', 'oidc'));
