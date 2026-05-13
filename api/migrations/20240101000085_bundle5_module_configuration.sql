-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS module_layouts (
    id          UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID                    NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    entity_type custom_field_entity_type NOT NULL,
    blocks      JSONB                   NOT NULL DEFAULT '[]'::jsonb,
    created_at  TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, entity_type)
);

CREATE TABLE IF NOT EXISTS module_relationship_definitions (
    id               UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID                    NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    relationship_key TEXT                    NOT NULL,
    from_entity_type custom_field_entity_type NOT NULL,
    to_entity_type   custom_field_entity_type NOT NULL,
    label            TEXT                    NOT NULL,
    cardinality      TEXT                    NOT NULL CHECK (cardinality IN ('one_to_one', 'many_to_one', 'one_to_many', 'many_to_many')),
    storage_strategy TEXT                    NOT NULL CHECK (storage_strategy IN ('native', 'crm_entity_links')),
    is_enabled       BOOLEAN                 NOT NULL DEFAULT TRUE,
    system_locked    BOOLEAN                 NOT NULL DEFAULT FALSE,
    order_idx        INTEGER                 NOT NULL DEFAULT 0,
    metadata         JSONB                   NOT NULL DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, relationship_key)
);

ALTER TABLE crm_entity_links
    ADD COLUMN IF NOT EXISTS relationship_definition_id UUID;

ALTER TABLE crm_entity_links
    ADD COLUMN IF NOT EXISTS relationship_cardinality TEXT CHECK (relationship_cardinality IS NULL OR relationship_cardinality IN ('one_to_one', 'many_to_one', 'one_to_many', 'many_to_many'));

UPDATE crm_entity_links link
SET relationship_cardinality = def.cardinality
FROM module_relationship_definitions def
WHERE link.relationship_definition_id = def.id
  AND link.relationship_cardinality IS NULL;

ALTER TABLE crm_entity_links
    DROP CONSTRAINT IF EXISTS crm_entity_links_relationship_definition_id_fkey;

CREATE UNIQUE INDEX IF NOT EXISTS idx_module_relationship_definitions_org_id
    ON module_relationship_definitions(org_id, id);

ALTER TABLE crm_entity_links
    ADD CONSTRAINT crm_entity_links_relationship_definition_id_fkey
    FOREIGN KEY (org_id, relationship_definition_id) REFERENCES module_relationship_definitions(org_id, id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_module_layouts_org_entity
    ON module_layouts(org_id, entity_type);
CREATE INDEX IF NOT EXISTS idx_module_relationship_definitions_org_from
    ON module_relationship_definitions(org_id, from_entity_type, is_enabled, order_idx);
CREATE INDEX IF NOT EXISTS idx_module_relationship_definitions_org_to
    ON module_relationship_definitions(org_id, to_entity_type, is_enabled, order_idx);
CREATE INDEX IF NOT EXISTS idx_crm_entity_links_relationship_definition
    ON crm_entity_links(org_id, relationship_definition_id, from_entity_type, from_entity_id, created_at DESC)
    WHERE relationship_definition_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_crm_entity_links_relationship_pair
    ON crm_entity_links(org_id, relationship_definition_id, from_entity_type, from_entity_id, to_entity_type, to_entity_id)
    WHERE relationship_definition_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_crm_entity_links_relationship_unique_from
    ON crm_entity_links(org_id, relationship_definition_id, from_entity_type, from_entity_id)
    WHERE relationship_definition_id IS NOT NULL AND relationship_cardinality IN ('one_to_one', 'many_to_one');
CREATE UNIQUE INDEX IF NOT EXISTS idx_crm_entity_links_relationship_unique_to
    ON crm_entity_links(org_id, relationship_definition_id, to_entity_type, to_entity_id)
    WHERE relationship_definition_id IS NOT NULL AND relationship_cardinality IN ('one_to_one', 'one_to_many');

ALTER TABLE module_layouts ENABLE ROW LEVEL SECURITY;
ALTER TABLE module_relationship_definitions ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS org_isolation ON module_layouts;
DROP POLICY IF EXISTS org_isolation ON module_relationship_definitions;
CREATE POLICY org_isolation ON module_layouts USING (org_id = current_org_id());
CREATE POLICY org_isolation ON module_relationship_definitions USING (org_id = current_org_id());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS org_isolation ON module_relationship_definitions;
DROP POLICY IF EXISTS org_isolation ON module_layouts;
ALTER TABLE module_relationship_definitions DISABLE ROW LEVEL SECURITY;
ALTER TABLE module_layouts DISABLE ROW LEVEL SECURITY;
ALTER TABLE crm_entity_links DROP CONSTRAINT IF EXISTS crm_entity_links_relationship_definition_id_fkey;
DROP INDEX IF EXISTS idx_crm_entity_links_relationship_pair;
DROP INDEX IF EXISTS idx_crm_entity_links_relationship_unique_to;
DROP INDEX IF EXISTS idx_crm_entity_links_relationship_unique_from;
DROP INDEX IF EXISTS idx_crm_entity_links_relationship_definition;
DROP INDEX IF EXISTS idx_module_relationship_definitions_org_to;
DROP INDEX IF EXISTS idx_module_relationship_definitions_org_from;
DROP INDEX IF EXISTS idx_module_relationship_definitions_org_id;
DROP INDEX IF EXISTS idx_module_layouts_org_entity;
ALTER TABLE crm_entity_links DROP COLUMN IF EXISTS relationship_cardinality;
ALTER TABLE crm_entity_links DROP COLUMN IF EXISTS relationship_definition_id;
DROP TABLE IF EXISTS module_relationship_definitions;
DROP TABLE IF EXISTS module_layouts;
-- +goose StatementEnd
