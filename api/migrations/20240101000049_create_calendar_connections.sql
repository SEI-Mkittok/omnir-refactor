-- Phase 11: Calendar Sync
-- OMN-415

-- +goose Up

CREATE TYPE calendar_provider AS ENUM ('google', 'microsoft');

CREATE TABLE calendar_connections (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider      calendar_provider NOT NULL,
    access_token  TEXT        NOT NULL,
    refresh_token TEXT,
    token_expiry  TIMESTAMPTZ,
    sync_cursor   TEXT,
    calendar_id   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, user_id, provider)
);

CREATE INDEX idx_calendar_connections_org_id  ON calendar_connections(org_id);
CREATE INDEX idx_calendar_connections_user_id ON calendar_connections(user_id);

-- +goose Down

DROP INDEX IF EXISTS idx_calendar_connections_user_id;
DROP INDEX IF EXISTS idx_calendar_connections_org_id;
DROP TABLE IF EXISTS calendar_connections;
DROP TYPE IF EXISTS calendar_provider;
