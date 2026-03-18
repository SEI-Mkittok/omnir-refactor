-- +goose Up

-- api_keys stores hashed API keys for external client authentication.
-- The plaintext key is only returned on creation; only the SHA-256 hash is stored.
CREATE TABLE api_keys (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID         REFERENCES organizations(id),
    created_by   UUID         REFERENCES users(id),
    name         TEXT         NOT NULL,
    key_hash     TEXT         NOT NULL UNIQUE,
    key_prefix   TEXT         NOT NULL,
    scopes       TEXT[]       NOT NULL DEFAULT '{}',
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_org_id   ON api_keys(org_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);

-- +goose Down
DROP TABLE IF EXISTS api_keys;
