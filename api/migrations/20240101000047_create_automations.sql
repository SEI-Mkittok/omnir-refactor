-- Phase 11: Workflow Automation Builder
-- OMN-413

-- +goose Up

CREATE TABLE automations (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','paused')),
    trigger_type   TEXT NOT NULL,
    trigger_config JSONB NOT NULL DEFAULT '{}',
    conditions     JSONB NOT NULL DEFAULT '[]',
    actions        JSONB NOT NULL DEFAULT '[]',
    created_by     UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_automations_org_id ON automations(org_id);
CREATE INDEX idx_automations_status  ON automations(org_id, status);

CREATE TABLE automation_runs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id UUID NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    org_id        UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    status        TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','running','succeeded','failed')),
    entity_type   TEXT,
    entity_id     UUID,
    error_message TEXT,
    started_at    TIMESTAMPTZ,
    finished_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_automation_runs_automation_id ON automation_runs(automation_id);
CREATE INDEX idx_automation_runs_org_id        ON automation_runs(org_id);

-- +goose Down

DROP INDEX IF EXISTS idx_automation_runs_org_id;
DROP INDEX IF EXISTS idx_automation_runs_automation_id;
DROP TABLE IF EXISTS automation_runs;

DROP INDEX IF EXISTS idx_automations_status;
DROP INDEX IF EXISTS idx_automations_org_id;
DROP TABLE IF EXISTS automations;
