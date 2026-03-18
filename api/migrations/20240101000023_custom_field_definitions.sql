-- +goose Up

-- custom_field_definitions stores admin-defined field schemas per entity type.
-- The custom_fields JSONB column on each entity (ticket, contact, lead) holds
-- the actual values keyed by field name.
CREATE TYPE custom_field_entity_type AS ENUM ('ticket', 'contact', 'lead', 'deal', 'account');
CREATE TYPE custom_field_type AS ENUM ('text', 'number', 'date', 'checkbox', 'select', 'multiselect', 'url');

CREATE TABLE IF NOT EXISTS custom_field_definitions (
    id           UUID                      PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID                      NOT NULL,
    entity_type  custom_field_entity_type  NOT NULL,
    name         TEXT                      NOT NULL,           -- machine-readable key (snake_case)
    label        TEXT                      NOT NULL,           -- human-readable display name
    field_type   custom_field_type         NOT NULL,
    options      JSONB,                                        -- for select/multi_select: ["opt1","opt2"]
    required     BOOLEAN                   NOT NULL DEFAULT FALSE,
    order_idx    INTEGER                   NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ               NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ               NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE (org_id, entity_type, name)
);

CREATE INDEX IF NOT EXISTS idx_cfd_org_entity ON custom_field_definitions (org_id, entity_type)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS custom_field_definitions;
DROP TYPE IF EXISTS custom_field_type;
DROP TYPE IF EXISTS custom_field_entity_type;
