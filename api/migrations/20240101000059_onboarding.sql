-- +goose Up
CREATE TABLE IF NOT EXISTS org_onboarding (
    org_id          UUID PRIMARY KEY REFERENCES orgs(id) ON DELETE CASCADE,
    completed_steps JSONB        NOT NULL DEFAULT '[]',
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

ALTER TABLE org_onboarding ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_onboarding_org_isolation ON org_onboarding
    USING (org_id = current_setting('app.current_org_id', true)::uuid);

CREATE TABLE org_invites (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    email       TEXT        NOT NULL,
    role        TEXT        NOT NULL DEFAULT 'agent',
    token       TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE org_invites ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_invites_org_isolation ON org_invites
    USING (org_id = current_setting('app.current_org_id', true)::uuid);

CREATE INDEX org_invites_token_idx ON org_invites (token);
CREATE INDEX org_invites_org_id_idx ON org_invites (org_id);

-- +goose Down
DROP TABLE IF EXISTS org_invites;
DROP TABLE IF EXISTS org_onboarding;
