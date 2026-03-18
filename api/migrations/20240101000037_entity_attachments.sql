-- +goose Up
-- Phase 10: File attachments for contacts, accounts, and deals

CREATE TABLE IF NOT EXISTS entity_attachments (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type  TEXT        NOT NULL CHECK (entity_type IN ('contact', 'account', 'deal')),
    entity_id    UUID        NOT NULL,
    org_id       UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    uploaded_by  UUID        REFERENCES users(id) ON DELETE SET NULL,
    filename     TEXT        NOT NULL,
    content_type TEXT        NOT NULL DEFAULT 'application/octet-stream',
    size_bytes   BIGINT,
    storage_path TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entity_attachments_entity ON entity_attachments (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_attachments_org    ON entity_attachments (org_id);

ALTER TABLE entity_attachments ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_bypass ON entity_attachments AS PERMISSIVE FOR ALL USING (true);

-- +goose Down
DROP POLICY IF EXISTS rls_bypass ON entity_attachments;
ALTER TABLE entity_attachments DISABLE ROW LEVEL SECURITY;
DROP INDEX IF EXISTS idx_entity_attachments_org;
DROP INDEX IF EXISTS idx_entity_attachments_entity;
DROP TABLE IF EXISTS entity_attachments;
