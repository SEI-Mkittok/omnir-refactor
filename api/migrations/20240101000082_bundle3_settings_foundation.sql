-- +goose Up
-- +goose StatementBegin
ALTER TABLE org_settings
    ADD COLUMN IF NOT EXISTS quote_number_prefix TEXT NOT NULL DEFAULT 'QUO',
    ADD COLUMN IF NOT EXISTS ticket_number_prefix TEXT NOT NULL DEFAULT 'TKT',
    ADD COLUMN IF NOT EXISTS kb_article_number_prefix TEXT NOT NULL DEFAULT 'KB',
    ADD COLUMN IF NOT EXISTS invoice_number_prefix TEXT NOT NULL DEFAULT 'INV';
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_preferences (
    id                              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id                          UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    default_currency                TEXT,
    number_format                   TEXT,
    default_record_view             TEXT,
    landing_page                    TEXT,
    address_line1                   TEXT,
    address_line2                   TEXT,
    city                            TEXT,
    state                           TEXT,
    postal_code                     TEXT,
    country                         TEXT,
    photo_url                       TEXT,
    tags                            JSONB       NOT NULL DEFAULT '[]'::jsonb,
    service_preferences             JSONB       NOT NULL DEFAULT '{}'::jsonb,
    calendar_start_day              TEXT,
    calendar_date_format            TEXT,
    calendar_time_zone              TEXT,
    calendar_default_activity_status TEXT,
    calendar_default_duration_minutes INTEGER,
    calendar_reminder_interval_minutes INTEGER,
    calendar_default_view           TEXT,
    calendar_day_start_hour         INTEGER,
    calendar_hour_format            TEXT,
    calendar_default_activity_type  TEXT,
    calendar_show_completed_events  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, org_id)
);
CREATE INDEX IF NOT EXISTS idx_user_preferences_user_org ON user_preferences (user_id, org_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS org_currencies (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    code            TEXT        NOT NULL,
    display_name    TEXT        NOT NULL,
    symbol          TEXT        NOT NULL,
    decimal_places  INTEGER     NOT NULL DEFAULT 2,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    is_default      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, code)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_org_currencies_default
    ON org_currencies(org_id)
    WHERE is_default = TRUE;
CREATE INDEX IF NOT EXISTS idx_org_currencies_org_active
    ON org_currencies(org_id, is_active);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS picklist_values (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    custom_field_id UUID        NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
    value           TEXT        NOT NULL,
    display_label   TEXT        NOT NULL,
    order_idx       INTEGER     NOT NULL DEFAULT 0,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, custom_field_id, value)
);
CREATE INDEX IF NOT EXISTS idx_picklist_values_field
    ON picklist_values(org_id, custom_field_id, order_idx, created_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS picklist_dependencies (
    id              UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID                    NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    entity_type     custom_field_entity_type NOT NULL,
    source_field_id UUID                    NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
    target_field_id UUID                    NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
    mapping         JSONB                   NOT NULL DEFAULT '{}'::jsonb,
    is_active       BOOLEAN                 NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, source_field_id, target_field_id)
);
CREATE INDEX IF NOT EXISTS idx_picklist_dependencies_org_entity
    ON picklist_dependencies(org_id, entity_type, is_active);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS lead_conversion_mappings (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    lead_field      TEXT        NOT NULL,
    target_entity   TEXT        NOT NULL CHECK (target_entity IN ('contact', 'account', 'deal')),
    target_field    TEXT        NOT NULL,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, lead_field, target_entity, target_field)
);
CREATE INDEX IF NOT EXISTS idx_lead_conversion_mappings_org
    ON lead_conversion_mappings(org_id, is_active);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO org_currencies (org_id, code, display_name, symbol, decimal_places, is_active, is_default)
SELECT os.org_id, 'USD', 'US Dollar', '$', 2, TRUE, TRUE
FROM org_settings os
WHERE NOT EXISTS (
    SELECT 1 FROM org_currencies oc WHERE oc.org_id = os.org_id
)
ON CONFLICT (org_id, code) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_lead_conversion_mappings_org;
DROP TABLE IF EXISTS lead_conversion_mappings;
DROP INDEX IF EXISTS idx_picklist_dependencies_org_entity;
DROP TABLE IF EXISTS picklist_dependencies;
DROP INDEX IF EXISTS idx_picklist_values_field;
DROP TABLE IF EXISTS picklist_values;
DROP INDEX IF EXISTS idx_org_currencies_org_active;
DROP INDEX IF EXISTS idx_org_currencies_default;
DROP TABLE IF EXISTS org_currencies;
DROP INDEX IF EXISTS idx_user_preferences_user_org;
DROP TABLE IF EXISTS user_preferences;
ALTER TABLE org_settings
    DROP COLUMN IF EXISTS invoice_number_prefix,
    DROP COLUMN IF EXISTS kb_article_number_prefix,
    DROP COLUMN IF EXISTS ticket_number_prefix,
    DROP COLUMN IF EXISTS quote_number_prefix;
-- +goose StatementEnd
