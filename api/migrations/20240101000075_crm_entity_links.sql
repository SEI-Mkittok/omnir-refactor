-- +goose Up
-- Centralized reusable entity-to-entity relationship graph for CRM records.

CREATE TABLE IF NOT EXISTS crm_entity_links (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    from_entity_type TEXT        NOT NULL CHECK (from_entity_type IN (
        'account', 'contact', 'deal', 'lead', 'ticket', 'quote', 'activity', 'sequence', 'note', 'entity_attachment'
    )),
    from_entity_id   UUID        NOT NULL,
    to_entity_type   TEXT        NOT NULL CHECK (to_entity_type IN (
        'account', 'contact', 'deal', 'lead', 'ticket', 'quote', 'activity', 'sequence', 'note', 'entity_attachment'
    )),
    to_entity_id     UUID        NOT NULL,
    link_type        TEXT        NOT NULL,
    metadata         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT crm_entity_links_not_self CHECK (
        NOT (from_entity_type = to_entity_type AND from_entity_id = to_entity_id)
    )
);

CREATE INDEX IF NOT EXISTS idx_crm_entity_links_org_from
    ON crm_entity_links (org_id, from_entity_type, from_entity_id, link_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_crm_entity_links_org_to
    ON crm_entity_links (org_id, to_entity_type, to_entity_id, link_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_crm_entity_links_org_link_type
    ON crm_entity_links (org_id, link_type, created_at DESC);

ALTER TABLE crm_entity_links ENABLE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON crm_entity_links
    USING (org_id = current_org_id());

-- +goose Down
DROP POLICY IF EXISTS org_isolation ON crm_entity_links;
ALTER TABLE crm_entity_links DISABLE ROW LEVEL SECURITY;
DROP INDEX IF EXISTS idx_crm_entity_links_org_link_type;
DROP INDEX IF EXISTS idx_crm_entity_links_org_to;
DROP INDEX IF EXISTS idx_crm_entity_links_org_from;
DROP TABLE IF EXISTS crm_entity_links;
