-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS crm_sharing_rules (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    module       TEXT NOT NULL,
    source_type  TEXT NOT NULL CHECK (source_type IN ('all', 'user', 'role', 'role_subordinates', 'group')),
    source_id    UUID,
    target_type  TEXT NOT NULL CHECK (target_type IN ('user', 'role', 'role_subordinates', 'group')),
    target_id    UUID NOT NULL,
    access_level TEXT NOT NULL CHECK (access_level IN ('read', 'write')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (source_type = 'all' AND source_id IS NULL)
        OR (source_type <> 'all' AND source_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_crm_sharing_rules_unique_source
    ON crm_sharing_rules (org_id, module, source_type, source_id, target_type, target_id, access_level)
    WHERE source_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_crm_sharing_rules_unique_all_source
    ON crm_sharing_rules (org_id, module, source_type, target_type, target_id, access_level)
    WHERE source_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_crm_sharing_rules_target
    ON crm_sharing_rules (org_id, module, target_type, target_id);

CREATE INDEX IF NOT EXISTS idx_crm_sharing_rules_source
    ON crm_sharing_rules (org_id, module, source_type, source_id);

INSERT INTO crm_sharing_rules
    (org_id, module, source_type, source_id, target_type, target_id, access_level, created_at, updated_at)
SELECT
    org_id,
    module,
    'all',
    NULL,
    grantee_type,
    grantee_id,
    access_level,
    created_at,
    created_at
FROM crm_sharing_grants
ON CONFLICT DO NOTHING;

CREATE OR REPLACE FUNCTION mirror_crm_sharing_grant_to_rule()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO crm_sharing_rules
        (org_id, module, source_type, source_id, target_type, target_id, access_level, created_at, updated_at)
    VALUES
        (NEW.org_id, NEW.module, 'all', NULL, NEW.grantee_type, NEW.grantee_id, NEW.access_level, NEW.created_at, NEW.created_at)
    ON CONFLICT DO NOTHING;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_mirror_crm_sharing_grants_to_rules ON crm_sharing_grants;
CREATE TRIGGER trg_mirror_crm_sharing_grants_to_rules
AFTER INSERT ON crm_sharing_grants
FOR EACH ROW
EXECUTE FUNCTION mirror_crm_sharing_grant_to_rule();

ALTER TABLE crm_sharing_rules ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS org_isolation ON crm_sharing_rules;
CREATE POLICY org_isolation ON crm_sharing_rules USING (org_id = current_org_id());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_mirror_crm_sharing_grants_to_rules ON crm_sharing_grants;
DROP FUNCTION IF EXISTS mirror_crm_sharing_grant_to_rule();
DROP POLICY IF EXISTS org_isolation ON crm_sharing_rules;
DROP TABLE IF EXISTS crm_sharing_rules;
-- +goose StatementEnd
