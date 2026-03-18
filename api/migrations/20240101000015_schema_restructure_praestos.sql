-- +goose Up

-- 1. Rename organizations → orgs (FK references follow automatically in Postgres)
ALTER TABLE organizations RENAME TO orgs;

-- Drop the auto-named plan check and replace with updated values + new default.
-- +goose StatementBegin
DO $$
DECLARE
    c TEXT;
BEGIN
    SELECT conname INTO c
    FROM pg_constraint
    WHERE conrelid = 'orgs'::regclass
      AND contype  = 'c'
      AND pg_get_constraintdef(oid) LIKE '%plan%';
    IF c IS NOT NULL THEN
        EXECUTE 'ALTER TABLE orgs DROP CONSTRAINT ' || quote_ident(c);
    END IF;
END $$;
-- +goose StatementEnd

-- Migrate all non-conforming plan values before adding the constraint.
UPDATE orgs SET plan = 'single' WHERE plan NOT IN ('single', 'starter', 'pro', 'enterprise');

-- 'single' is the new default plan for self-hosted/single-tenant deployments.
-- Use NOT VALID to add constraint without scanning existing rows, then validate separately.
ALTER TABLE orgs
    ADD CONSTRAINT orgs_plan_check
        CHECK (plan IN ('single', 'starter', 'pro', 'enterprise'))
        NOT VALID;

ALTER TABLE orgs VALIDATE CONSTRAINT orgs_plan_check;

ALTER TABLE orgs ALTER COLUMN plan SET DEFAULT 'single';

-- 2. Update users.role enum: admin/user/viewer → admin/agent/client
-- +goose StatementBegin
DO $$
DECLARE
    c TEXT;
BEGIN
    SELECT conname INTO c
    FROM pg_constraint
    WHERE conrelid = 'users'::regclass
      AND contype  = 'c'
      AND pg_get_constraintdef(oid) LIKE '%role%';
    IF c IS NOT NULL THEN
        EXECUTE 'ALTER TABLE users DROP CONSTRAINT ' || quote_ident(c);
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE users
    ADD CONSTRAINT users_role_check
        CHECK (role IN ('admin', 'agent', 'client'));

ALTER TABLE users ALTER COLUMN role SET DEFAULT 'agent';

-- Migrate existing role values.
UPDATE users SET role = 'agent'  WHERE role = 'user';
UPDATE users SET role = 'client' WHERE role = 'viewer';

-- 3. Add leads table (separate from contacts, with status workflow).
CREATE TABLE IF NOT EXISTS leads (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id               UUID        NOT NULL REFERENCES orgs(id),
    first_name           TEXT        NOT NULL,
    last_name            TEXT        NOT NULL,
    email                TEXT,
    phone                TEXT,
    company              TEXT,
    lead_source          TEXT,
    status               TEXT        NOT NULL DEFAULT 'new'
                             CHECK (status IN ('new', 'contacted', 'qualified', 'unqualified', 'converted')),
    owner_id             UUID        REFERENCES users(id),
    converted_contact_id UUID        REFERENCES contacts(id),
    custom_fields        JSONB,
    vtiger_legacy_id     TEXT        UNIQUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);

CREATE INDEX idx_leads_org_id   ON leads (org_id)   WHERE deleted_at IS NULL;
CREATE INDEX idx_leads_owner_id ON leads (owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_leads_status   ON leads (status)   WHERE deleted_at IS NULL;
CREATE INDEX idx_leads_email    ON leads (email)    WHERE deleted_at IS NULL;

-- 4. Row-Level Security setup.
-- RLS is enabled on all entity tables with a permissive bypass policy so that
-- existing behaviour is unchanged. A future SaaS-mode migration can add
-- restrictive per-org policies alongside these, replacing or superseding them.

ALTER TABLE orgs          ENABLE ROW LEVEL SECURITY;
ALTER TABLE users         ENABLE ROW LEVEL SECURITY;
ALTER TABLE accounts      ENABLE ROW LEVEL SECURITY;
ALTER TABLE contacts      ENABLE ROW LEVEL SECURITY;
ALTER TABLE deals         ENABLE ROW LEVEL SECURITY;
ALTER TABLE activities    ENABLE ROW LEVEL SECURITY;
ALTER TABLE notes         ENABLE ROW LEVEL SECURITY;
ALTER TABLE pipelines     ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE leads         ENABLE ROW LEVEL SECURITY;

-- Permissive bypass policies (USING true = allow all rows).
-- These must exist before RLS is meaningful; SaaS mode adds org-scoped policies.
CREATE POLICY rls_bypass ON orgs          AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON users         AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON accounts      AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON contacts      AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON deals         AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON activities    AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON notes         AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON pipelines     AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON notifications AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON leads         AS PERMISSIVE FOR ALL USING (true);

-- +goose Down

-- Reverse RLS
DROP POLICY IF EXISTS rls_bypass ON leads;
DROP POLICY IF EXISTS rls_bypass ON notifications;
DROP POLICY IF EXISTS rls_bypass ON pipelines;
DROP POLICY IF EXISTS rls_bypass ON notes;
DROP POLICY IF EXISTS rls_bypass ON activities;
DROP POLICY IF EXISTS rls_bypass ON deals;
DROP POLICY IF EXISTS rls_bypass ON contacts;
DROP POLICY IF EXISTS rls_bypass ON accounts;
DROP POLICY IF EXISTS rls_bypass ON users;
DROP POLICY IF EXISTS rls_bypass ON orgs;

ALTER TABLE leads         DISABLE ROW LEVEL SECURITY;
ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE pipelines     DISABLE ROW LEVEL SECURITY;
ALTER TABLE notes         DISABLE ROW LEVEL SECURITY;
ALTER TABLE activities    DISABLE ROW LEVEL SECURITY;
ALTER TABLE deals         DISABLE ROW LEVEL SECURITY;
ALTER TABLE contacts      DISABLE ROW LEVEL SECURITY;
ALTER TABLE accounts      DISABLE ROW LEVEL SECURITY;
ALTER TABLE users         DISABLE ROW LEVEL SECURITY;
ALTER TABLE orgs          DISABLE ROW LEVEL SECURITY;

-- Drop leads table
DROP TABLE IF EXISTS leads;

-- Reverse users.role to original values (best-effort data migration)
UPDATE users SET role = 'user'   WHERE role = 'agent';
UPDATE users SET role = 'viewer' WHERE role = 'client';
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

-- +goose StatementBegin
DO $$
BEGIN
    ALTER TABLE users
        ADD CONSTRAINT users_role_check
            CHECK (role IN ('admin', 'user', 'viewer'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

ALTER TABLE users ALTER COLUMN role SET DEFAULT 'user';

-- Reverse orgs rename and plan update
UPDATE orgs SET plan = 'self_hosted' WHERE plan = 'single';
ALTER TABLE orgs DROP CONSTRAINT IF EXISTS orgs_plan_check;

-- +goose StatementBegin
DO $$
BEGIN
    ALTER TABLE orgs
        ADD CONSTRAINT organizations_plan_check
            CHECK (plan IN ('self_hosted', 'starter', 'pro', 'enterprise'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

ALTER TABLE orgs ALTER COLUMN plan SET DEFAULT 'self_hosted';
ALTER TABLE orgs RENAME TO organizations;
