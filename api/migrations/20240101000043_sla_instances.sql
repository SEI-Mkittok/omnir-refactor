-- +goose Up

-- Extend sla_policies to support entity types beyond tickets
ALTER TABLE sla_policies
    ADD COLUMN IF NOT EXISTS entity_type TEXT NOT NULL DEFAULT 'ticket',
    ADD COLUMN IF NOT EXISTS conditions  JSONB NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_sla_policies_entity_type ON sla_policies (org_id, entity_type);

-- sla_instances tracks per-entity SLA state
CREATE TABLE sla_instances (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    policy_id           UUID        NOT NULL REFERENCES sla_policies(id) ON DELETE CASCADE,
    entity_id           UUID        NOT NULL,
    entity_type         TEXT        NOT NULL,  -- 'deal' | 'contact' | 'activity' | 'ticket'
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    response_due_at     TIMESTAMPTZ NOT NULL,
    resolution_due_at   TIMESTAMPTZ NOT NULL,
    responded_at        TIMESTAMPTZ,
    resolved_at         TIMESTAMPTZ,
    breached            BOOLEAN     NOT NULL DEFAULT FALSE,
    breach_type         TEXT        NOT NULL DEFAULT 'none', -- 'none' | 'response' | 'resolution'
    warned_at           TIMESTAMPTZ,  -- when 80% warning was emitted
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sla_instances_org_entity ON sla_instances (org_id, entity_type, entity_id);
CREATE INDEX idx_sla_instances_open       ON sla_instances (org_id, breached, resolved_at)
    WHERE resolved_at IS NULL;

-- RLS: org isolation for sla_instances
ALTER TABLE sla_instances ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_isolation ON sla_instances
    USING (org_id = current_org_id());

-- +goose Down

DROP POLICY IF EXISTS org_isolation ON sla_instances;
ALTER TABLE sla_instances DISABLE ROW LEVEL SECURITY;

DROP INDEX IF EXISTS idx_sla_instances_open;
DROP INDEX IF EXISTS idx_sla_instances_org_entity;
DROP TABLE IF EXISTS sla_instances;

DROP INDEX IF EXISTS idx_sla_policies_entity_type;
ALTER TABLE sla_policies
    DROP COLUMN IF EXISTS conditions,
    DROP COLUMN IF EXISTS entity_type;
