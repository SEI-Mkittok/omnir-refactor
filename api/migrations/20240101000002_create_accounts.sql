-- +goose Up
CREATE TABLE IF NOT EXISTS accounts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    domain        TEXT,
    industry      TEXT,
    size          TEXT CHECK (size IN ('1-10', '11-50', '51-200', '201-500', '501+')),
    owner_id      UUID NOT NULL REFERENCES users(id),
    tags          TEXT[] NOT NULL DEFAULT '{}',
    custom_fields JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_accounts_owner_id ON accounts (owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_name ON accounts (name) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS accounts;
