-- +goose Up

-- Migration 20240101000016 created ticket_attachments with a permissive
-- rls_bypass policy (USING true) — that policy allows every row for every org,
-- which means FORCE ROW LEVEL SECURITY would not provide tenant isolation.
--
-- This migration replaces the bypass policy with the same org_isolation policy
-- used by all other domain tables, and adds ticket_attachments to the tables
-- that receive FORCE ROW LEVEL SECURITY in multi-tenant modes.

DROP POLICY IF EXISTS rls_bypass ON ticket_attachments;

CREATE POLICY org_isolation ON ticket_attachments
    USING (org_id = current_org_id());

-- +goose Down

DROP POLICY IF EXISTS org_isolation ON ticket_attachments;

CREATE POLICY rls_bypass ON ticket_attachments AS PERMISSIVE FOR ALL USING (true);
