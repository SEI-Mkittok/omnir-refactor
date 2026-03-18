-- +goose Up
-- Phase 8 (OMN-313): lead_score on leads table + conversion back-reference on contacts

-- Add lead_score to the leads table
ALTER TABLE leads
    ADD COLUMN IF NOT EXISTS lead_score SMALLINT NOT NULL DEFAULT 0
        CHECK (lead_score >= 0 AND lead_score <= 100);

-- Allow contacts to record which lead they were converted from
ALTER TABLE contacts
    ADD COLUMN IF NOT EXISTS converted_from_lead_id UUID REFERENCES leads(id);

-- Support score-range filter queries on the leads list
CREATE INDEX IF NOT EXISTS idx_leads_lead_score
    ON leads (org_id, lead_score)
    WHERE deleted_at IS NULL;

-- Support lead-source filter queries on the leads list
CREATE INDEX IF NOT EXISTS idx_leads_lead_source
    ON leads (org_id, lead_source)
    WHERE deleted_at IS NULL AND lead_source IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_leads_lead_source;
DROP INDEX IF EXISTS idx_leads_lead_score;
ALTER TABLE contacts DROP COLUMN IF EXISTS converted_from_lead_id;
ALTER TABLE leads    DROP COLUMN IF EXISTS lead_score;
