-- +goose Up
CREATE TABLE IF NOT EXISTS contacts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    email         TEXT UNIQUE,
    phone         TEXT,
    account_id    UUID REFERENCES accounts(id),
    owner_id      UUID NOT NULL REFERENCES users(id),
    lead_source   TEXT,
    stage         TEXT NOT NULL DEFAULT 'lead'
                      CHECK (stage IN ('lead', 'prospect', 'customer', 'churned')),
    tags          TEXT[] NOT NULL DEFAULT '{}',
    custom_fields JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_contacts_owner_id   ON contacts (owner_id)   WHERE deleted_at IS NULL;
CREATE INDEX idx_contacts_account_id ON contacts (account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_contacts_stage      ON contacts (stage)      WHERE deleted_at IS NULL;
CREATE INDEX idx_contacts_email      ON contacts (email)      WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS contacts;
