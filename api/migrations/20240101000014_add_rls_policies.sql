-- +goose Up

-- Helper function: returns the current org_id from the session variable set
-- by the application before every query. Returns NULL when not set (e.g. for
-- superuser maintenance sessions), causing RLS policies to match nothing.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION current_org_id() RETURNS UUID AS $$
  SELECT NULLIF(current_setting('app.current_org_id', true), '')::UUID;
$$ LANGUAGE SQL STABLE SECURITY DEFINER;
-- +goose StatementEnd

-- Enable Row-Level Security on every org-scoped table.
-- RLS is enabled but not yet FORCED — enforcement is activated at runtime by
-- postgres.EnableRLS() when ORG_MODE is 'multitenant' or 'enterprise'.
ALTER TABLE users          ENABLE ROW LEVEL SECURITY;
ALTER TABLE accounts       ENABLE ROW LEVEL SECURITY;
ALTER TABLE contacts       ENABLE ROW LEVEL SECURITY;
ALTER TABLE pipelines      ENABLE ROW LEVEL SECURITY;
ALTER TABLE deals          ENABLE ROW LEVEL SECURITY;
ALTER TABLE activities     ENABLE ROW LEVEL SECURITY;
ALTER TABLE notes          ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications  ENABLE ROW LEVEL SECURITY;

-- PERMISSIVE policies: a row is visible / writable only when it belongs to the
-- current org. These policies apply regardless of the DB role, once FORCE RLS
-- is active.
CREATE POLICY org_isolation ON users
    USING (org_id = current_org_id());

CREATE POLICY org_isolation ON accounts
    USING (org_id = current_org_id());

CREATE POLICY org_isolation ON contacts
    USING (org_id = current_org_id());

CREATE POLICY org_isolation ON pipelines
    USING (org_id = current_org_id());

CREATE POLICY org_isolation ON deals
    USING (org_id = current_org_id());

CREATE POLICY org_isolation ON activities
    USING (org_id = current_org_id());

CREATE POLICY org_isolation ON notes
    USING (org_id = current_org_id());

CREATE POLICY org_isolation ON notifications
    USING (org_id = current_org_id());

-- +goose Down

DROP POLICY IF EXISTS org_isolation ON notifications;
DROP POLICY IF EXISTS org_isolation ON notes;
DROP POLICY IF EXISTS org_isolation ON activities;
DROP POLICY IF EXISTS org_isolation ON deals;
DROP POLICY IF EXISTS org_isolation ON pipelines;
DROP POLICY IF EXISTS org_isolation ON contacts;
DROP POLICY IF EXISTS org_isolation ON accounts;
DROP POLICY IF EXISTS org_isolation ON users;

ALTER TABLE notifications  DISABLE ROW LEVEL SECURITY;
ALTER TABLE notes          DISABLE ROW LEVEL SECURITY;
ALTER TABLE activities     DISABLE ROW LEVEL SECURITY;
ALTER TABLE deals          DISABLE ROW LEVEL SECURITY;
ALTER TABLE pipelines      DISABLE ROW LEVEL SECURITY;
ALTER TABLE contacts       DISABLE ROW LEVEL SECURITY;
ALTER TABLE accounts       DISABLE ROW LEVEL SECURITY;
ALTER TABLE users          DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS current_org_id();
