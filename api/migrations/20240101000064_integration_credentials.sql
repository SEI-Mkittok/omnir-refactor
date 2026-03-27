-- +goose Up
-- +goose StatementBegin
CREATE TABLE integration_credentials (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id            UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    provider          TEXT        NOT NULL,
    client_id         TEXT,
    client_secret_enc TEXT,
    api_key_enc       TEXT,
    webhook_secret    TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, provider)
);

CREATE INDEX idx_integration_credentials_org_id ON integration_credentials(org_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS integration_credentials;
-- +goose StatementEnd
