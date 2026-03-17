-- +goose Up
CREATE TYPE activity_type AS ENUM ('call', 'email', 'meeting', 'task');

CREATE TABLE activities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type        activity_type NOT NULL,
    subject     TEXT NOT NULL,
    description TEXT,
    due_date    TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    contact_id  UUID REFERENCES contacts(id) ON DELETE SET NULL,
    account_id  UUID REFERENCES accounts(id) ON DELETE SET NULL,
    deal_id     UUID REFERENCES deals(id) ON DELETE SET NULL,
    owner_id    UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX activities_contact_id_idx ON activities(contact_id) WHERE deleted_at IS NULL;
CREATE INDEX activities_account_id_idx ON activities(account_id) WHERE deleted_at IS NULL;
CREATE INDEX activities_deal_id_idx    ON activities(deal_id)    WHERE deleted_at IS NULL;
CREATE INDEX activities_owner_id_idx   ON activities(owner_id)   WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS activities;
DROP TYPE IF EXISTS activity_type;
