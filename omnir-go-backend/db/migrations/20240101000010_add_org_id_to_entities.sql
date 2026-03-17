-- +goose Up

-- Add org_id to users
ALTER TABLE users
    ADD COLUMN org_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000002'
        REFERENCES organizations(id);

-- Add org_id to accounts
ALTER TABLE accounts
    ADD COLUMN org_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000002'
        REFERENCES organizations(id);

-- Add org_id to contacts
ALTER TABLE contacts
    ADD COLUMN org_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000002'
        REFERENCES organizations(id);

-- Add org_id to pipelines
ALTER TABLE pipelines
    ADD COLUMN org_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000002'
        REFERENCES organizations(id);

-- Add org_id to deals
ALTER TABLE deals
    ADD COLUMN org_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000002'
        REFERENCES organizations(id);

-- Add org_id to activities
ALTER TABLE activities
    ADD COLUMN org_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000002'
        REFERENCES organizations(id);

-- Add org_id to notes
ALTER TABLE notes
    ADD COLUMN org_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000002'
        REFERENCES organizations(id);

-- Indexes for org_id lookups
CREATE INDEX idx_users_org_id     ON users     (org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_org_id  ON accounts  (org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_contacts_org_id  ON contacts  (org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_pipelines_org_id ON pipelines (org_id);
CREATE INDEX idx_deals_org_id     ON deals     (org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_activities_org_id ON activities (org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_notes_org_id     ON notes     (org_id) WHERE deleted_at IS NULL;

-- +goose Down
ALTER TABLE notes      DROP COLUMN IF EXISTS org_id;
ALTER TABLE activities DROP COLUMN IF EXISTS org_id;
ALTER TABLE deals      DROP COLUMN IF EXISTS org_id;
ALTER TABLE pipelines  DROP COLUMN IF EXISTS org_id;
ALTER TABLE contacts   DROP COLUMN IF EXISTS org_id;
ALTER TABLE accounts   DROP COLUMN IF EXISTS org_id;
ALTER TABLE users      DROP COLUMN IF EXISTS org_id;
