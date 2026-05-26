-- CRM automation and intake milestone substrate.

-- +goose Up
-- +goose StatementBegin
CREATE TABLE webforms (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    public_id       TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'inactive' CHECK (status IN ('active', 'inactive')),
    target_module   TEXT NOT NULL CHECK (target_module IN ('lead', 'contact', 'ticket')),
    campaign_id     UUID,
    return_url      TEXT,
    success_message TEXT NOT NULL DEFAULT 'Thanks. Your submission has been received.',
    spam_trap_field TEXT NOT NULL DEFAULT 'website',
    fields          JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_webforms_org_status ON webforms(org_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_webforms_campaign ON webforms(org_id, campaign_id) WHERE campaign_id IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE webform_submissions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    webform_id          UUID NOT NULL REFERENCES webforms(id) ON DELETE CASCADE,
    campaign_id         UUID,
    target_module       TEXT NOT NULL CHECK (target_module IN ('lead', 'contact', 'ticket')),
    payload             JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_record_type TEXT,
    created_record_id   UUID,
    ip_address          TEXT,
    user_agent          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webform_submissions_org_created ON webform_submissions(org_id, created_at DESC);
CREATE INDEX idx_webform_submissions_webform ON webform_submissions(org_id, webform_id, created_at DESC);
CREATE INDEX idx_webform_submissions_campaign ON webform_submissions(org_id, campaign_id, created_at DESC) WHERE campaign_id IS NOT NULL;

CREATE TABLE mail_converter_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'inactive' CHECK (status IN ('active', 'inactive')),
    conditions  JSONB NOT NULL DEFAULT '[]'::jsonb,
    actions     JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    last_run_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_mail_converter_rules_org_status ON mail_converter_rules(org_id, status) WHERE deleted_at IS NULL;

CREATE TABLE mail_converter_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    rule_id         UUID NOT NULL REFERENCES mail_converter_rules(id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'succeeded', 'failed')),
    matched_count   INTEGER NOT NULL DEFAULT 0,
    processed_count INTEGER NOT NULL DEFAULT 0,
    skipped_count   INTEGER NOT NULL DEFAULT 0,
    error_text      TEXT,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ
);

CREATE INDEX idx_mail_converter_runs_rule_started ON mail_converter_runs(org_id, rule_id, started_at DESC);

CREATE TABLE mail_converter_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    rule_id             UUID NOT NULL REFERENCES mail_converter_rules(id) ON DELETE CASCADE,
    run_id              UUID NOT NULL REFERENCES mail_converter_runs(id) ON DELETE CASCADE,
    message_id          UUID NOT NULL REFERENCES email_inbox_messages(id) ON DELETE CASCADE,
    status              TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'succeeded', 'failed', 'skipped')),
    created_record_type TEXT,
    created_record_id   UUID,
    error_text          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (rule_id, message_id)
);

CREATE INDEX idx_mail_converter_logs_run ON mail_converter_logs(org_id, run_id);
CREATE INDEX idx_mail_converter_logs_rule ON mail_converter_logs(org_id, rule_id, created_at DESC);

CREATE TABLE campaigns (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    type          TEXT NOT NULL DEFAULT 'marketing',
    status        TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'completed', 'archived')),
    description   TEXT NOT NULL DEFAULT '',
    sequence_id   UUID REFERENCES email_sequences(id) ON DELETE SET NULL,
    automation_id UUID REFERENCES automations(id) ON DELETE SET NULL,
    metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by    UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_campaigns_org_status ON campaigns(org_id, status) WHERE deleted_at IS NULL;

ALTER TABLE webforms
    ADD CONSTRAINT webforms_campaign_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE SET NULL;

ALTER TABLE webform_submissions
    ADD CONSTRAINT webform_submissions_campaign_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE SET NULL;

CREATE TABLE campaign_members (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    member_type TEXT NOT NULL CHECK (member_type IN ('lead', 'contact', 'account')),
    member_id   UUID NOT NULL,
    source      TEXT NOT NULL DEFAULT 'manual',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (campaign_id, member_type, member_id)
);

CREATE INDEX idx_campaign_members_campaign ON campaign_members(org_id, campaign_id, created_at DESC);

ALTER TABLE webforms ENABLE ROW LEVEL SECURITY;
ALTER TABLE webform_submissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE mail_converter_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE mail_converter_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE mail_converter_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE campaigns ENABLE ROW LEVEL SECURITY;
ALTER TABLE campaign_members ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_isolation ON webforms USING (org_id = current_org_id());
CREATE POLICY org_isolation ON webform_submissions USING (org_id = current_org_id());
CREATE POLICY org_isolation ON mail_converter_rules USING (org_id = current_org_id());
CREATE POLICY org_isolation ON mail_converter_runs USING (org_id = current_org_id());
CREATE POLICY org_isolation ON mail_converter_logs USING (org_id = current_org_id());
CREATE POLICY org_isolation ON campaigns USING (org_id = current_org_id());
CREATE POLICY org_isolation ON campaign_members USING (org_id = current_org_id());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS org_isolation ON campaign_members;
DROP POLICY IF EXISTS org_isolation ON campaigns;
DROP POLICY IF EXISTS org_isolation ON mail_converter_logs;
DROP POLICY IF EXISTS org_isolation ON mail_converter_runs;
DROP POLICY IF EXISTS org_isolation ON mail_converter_rules;
DROP POLICY IF EXISTS org_isolation ON webform_submissions;
DROP POLICY IF EXISTS org_isolation ON webforms;

ALTER TABLE campaign_members DISABLE ROW LEVEL SECURITY;
ALTER TABLE campaigns DISABLE ROW LEVEL SECURITY;
ALTER TABLE mail_converter_logs DISABLE ROW LEVEL SECURITY;
ALTER TABLE mail_converter_runs DISABLE ROW LEVEL SECURITY;
ALTER TABLE mail_converter_rules DISABLE ROW LEVEL SECURITY;
ALTER TABLE webform_submissions DISABLE ROW LEVEL SECURITY;
ALTER TABLE webforms DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS campaign_members;
ALTER TABLE webform_submissions DROP CONSTRAINT IF EXISTS webform_submissions_campaign_fkey;
ALTER TABLE webforms DROP CONSTRAINT IF EXISTS webforms_campaign_fkey;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS mail_converter_logs;
DROP TABLE IF EXISTS mail_converter_runs;
DROP TABLE IF EXISTS mail_converter_rules;
DROP TABLE IF EXISTS webform_submissions;
DROP TABLE IF EXISTS webforms;
-- +goose StatementEnd
