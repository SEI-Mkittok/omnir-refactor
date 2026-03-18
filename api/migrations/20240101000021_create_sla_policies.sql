-- +goose Up

-- SLA policy definitions per org
CREATE TABLE sla_policies (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name                  TEXT        NOT NULL,
    response_time_hours   NUMERIC(10,2) NOT NULL CHECK (response_time_hours > 0),
    resolution_time_hours NUMERIC(10,2) NOT NULL CHECK (resolution_time_hours > 0),
    priority_filter       JSONB       NOT NULL DEFAULT '[]',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sla_policies_org_id ON sla_policies (org_id);

-- Add SLA tracking columns to tickets
ALTER TABLE tickets
    ADD COLUMN sla_policy_id     UUID        REFERENCES sla_policies(id) ON DELETE SET NULL,
    ADD COLUMN first_responded_at TIMESTAMPTZ;

CREATE INDEX idx_tickets_sla_policy ON tickets (sla_policy_id) WHERE deleted_at IS NULL;

-- RLS: org isolation for sla_policies
ALTER TABLE sla_policies ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_isolation ON sla_policies
    USING (org_id = current_org_id());

-- +goose Down

DROP POLICY IF EXISTS org_isolation ON sla_policies;
ALTER TABLE sla_policies DISABLE ROW LEVEL SECURITY;

DROP INDEX IF EXISTS idx_tickets_sla_policy;
ALTER TABLE tickets
    DROP COLUMN IF EXISTS first_responded_at,
    DROP COLUMN IF EXISTS sla_policy_id;

DROP INDEX IF EXISTS idx_sla_policies_org_id;
DROP TABLE IF EXISTS sla_policies;
