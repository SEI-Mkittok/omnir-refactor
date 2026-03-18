-- +goose Up

-- Performance indexes for leads and activities list queries.
-- Context: migration 000027 added composite indexes for tickets/deals/accounts/contacts
-- but leads and activities were not covered. These fill that gap.
-- Also adds GIN FTS indexes for leads and activities to replace ILIKE '%..%' searches
-- (leading wildcards cannot use btree indexes; expression GIN indexes can).

-- Leads: composite indexes for common filter combinations in LeadRepo.List()
CREATE INDEX idx_leads_org_status   ON leads (org_id, status)         WHERE deleted_at IS NULL;
CREATE INDEX idx_leads_org_owner    ON leads (org_id, owner_id)       WHERE deleted_at IS NULL;
CREATE INDEX idx_leads_org_created  ON leads (org_id, created_at DESC) WHERE deleted_at IS NULL;

-- Leads: GIN FTS index to support to_tsvector @@ plainto_tsquery search.
-- Replaces the ILIKE '%q%' pattern in LeadRepo.List().
CREATE INDEX idx_leads_fts ON leads
    USING GIN (to_tsvector('english',
        first_name || ' ' || last_name || ' ' ||
        coalesce(email, '') || ' ' || coalesce(company, '')))
    WHERE deleted_at IS NULL;

-- Activities: composite indexes for common filter combinations in ActivityRepo.List()
CREATE INDEX idx_activities_org_type    ON activities (org_id, type)          WHERE deleted_at IS NULL;
CREATE INDEX idx_activities_org_owner   ON activities (org_id, owner_id)      WHERE deleted_at IS NULL;
CREATE INDEX idx_activities_org_created ON activities (org_id, created_at DESC) WHERE deleted_at IS NULL;

-- Activities: GIN FTS index to support to_tsvector @@ plainto_tsquery search.
-- Replaces the ILIKE '%q%' pattern in ActivityRepo.List().
CREATE INDEX idx_activities_fts ON activities
    USING GIN (to_tsvector('english', subject || ' ' || coalesce(description, '')))
    WHERE deleted_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_activities_fts;
DROP INDEX IF EXISTS idx_activities_org_created;
DROP INDEX IF EXISTS idx_activities_org_owner;
DROP INDEX IF EXISTS idx_activities_org_type;
DROP INDEX IF EXISTS idx_leads_fts;
DROP INDEX IF EXISTS idx_leads_org_created;
DROP INDEX IF EXISTS idx_leads_org_owner;
DROP INDEX IF EXISTS idx_leads_org_status;
