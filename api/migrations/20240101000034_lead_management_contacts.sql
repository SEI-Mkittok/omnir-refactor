-- +goose Up
-- Phase 8: lead management fields on contacts
ALTER TABLE contacts
    ADD COLUMN IF NOT EXISTS lead_score      INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS converted_at    TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS converted_by    UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS converted_deal_id UUID REFERENCES deals(id);

-- Index to support score-range filtering for lead scoring views
CREATE INDEX IF NOT EXISTS idx_contacts_lead_score
    ON contacts (org_id, lead_score)
    WHERE deleted_at IS NULL;

-- Index to support source-filter queries on the lead list
CREATE INDEX IF NOT EXISTS idx_contacts_lead_source
    ON contacts (org_id, lead_source)
    WHERE deleted_at IS NULL AND lead_source IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_contacts_lead_source;
DROP INDEX IF EXISTS idx_contacts_lead_score;
ALTER TABLE contacts
    DROP COLUMN IF EXISTS converted_deal_id,
    DROP COLUMN IF EXISTS converted_by,
    DROP COLUMN IF EXISTS converted_at,
    DROP COLUMN IF EXISTS lead_score;
