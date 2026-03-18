-- +goose Up

-- Enable RLS and add org-isolation policies for tables added after the
-- initial RLS migration (20240101000014): tickets, ticket_comments, leads.

ALTER TABLE tickets         ENABLE ROW LEVEL SECURITY;
ALTER TABLE ticket_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE leads           ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_isolation ON tickets
    USING (org_id = current_org_id());

CREATE POLICY org_isolation ON ticket_comments
    USING (org_id = current_org_id());

CREATE POLICY org_isolation ON leads
    USING (org_id = current_org_id());

-- +goose Down

DROP POLICY IF EXISTS org_isolation ON leads;
DROP POLICY IF EXISTS org_isolation ON ticket_comments;
DROP POLICY IF EXISTS org_isolation ON tickets;

ALTER TABLE leads           DISABLE ROW LEVEL SECURITY;
ALTER TABLE ticket_comments DISABLE ROW LEVEL SECURITY;
ALTER TABLE tickets         DISABLE ROW LEVEL SECURITY;
