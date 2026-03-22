-- +goose Up
CREATE TABLE teams_connections (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id            UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    tenant_id         TEXT NOT NULL DEFAULT '',
    bot_token         TEXT NOT NULL,
    default_channel_id TEXT NOT NULL DEFAULT '',
    channel_name      TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id)
);

ALTER TABLE teams_connections ENABLE ROW LEVEL SECURITY;

CREATE POLICY teams_connections_org_isolation ON teams_connections
    USING (org_id = current_setting('app.current_org_id', true)::uuid);

-- +goose Down
DROP TABLE IF EXISTS teams_connections;
