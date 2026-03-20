-- +goose Up
CREATE TABLE audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id),
    user_id     UUID REFERENCES users(id),
    agent_id    TEXT,
    action      TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   UUID,
    entity_name TEXT,
    changes     JSONB,
    ip_address  TEXT,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON audit_log(org_id, created_at DESC);
CREATE INDEX ON audit_log(org_id, entity_type, entity_id);

-- +goose Down
DROP TABLE IF EXISTS audit_log;
