-- +goose Up

-- Help Desk tickets
CREATE TABLE IF NOT EXISTS tickets (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL REFERENCES orgs(id),
    subject      TEXT        NOT NULL,
    description  TEXT,
    status       TEXT        NOT NULL DEFAULT 'open'
                     CHECK (status IN ('open', 'pending', 'resolved', 'closed')),
    priority     TEXT        NOT NULL DEFAULT 'medium'
                     CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    assignee_id  UUID        REFERENCES users(id),
    contact_id   UUID        REFERENCES contacts(id),
    account_id   UUID        REFERENCES accounts(id),
    source       TEXT,
    tags         TEXT[]      NOT NULL DEFAULT '{}',
    custom_fields JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_tickets_org_id     ON tickets (org_id)     WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_status     ON tickets (status)     WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_priority   ON tickets (priority)   WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_assignee   ON tickets (assignee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_contact    ON tickets (contact_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_account    ON tickets (account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tickets_created_at ON tickets (created_at DESC) WHERE deleted_at IS NULL;

-- Ticket comments (public replies + internal notes)
CREATE TABLE IF NOT EXISTS ticket_comments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id  UUID        NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    org_id     UUID        NOT NULL REFERENCES orgs(id),
    author_id  UUID        REFERENCES users(id),
    body       TEXT        NOT NULL,
    is_internal BOOLEAN    NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_ticket_comments_ticket ON ticket_comments (ticket_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ticket_comments_org    ON ticket_comments (org_id)    WHERE deleted_at IS NULL;

-- Ticket file attachments
CREATE TABLE IF NOT EXISTS ticket_attachments (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id    UUID        NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    org_id       UUID        NOT NULL REFERENCES orgs(id),
    uploaded_by  UUID        REFERENCES users(id),
    filename     TEXT        NOT NULL,
    content_type TEXT        NOT NULL DEFAULT 'application/octet-stream',
    size_bytes   BIGINT,
    storage_url  TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticket_attachments_ticket ON ticket_attachments (ticket_id);
CREATE INDEX idx_ticket_attachments_org    ON ticket_attachments (org_id);

-- RLS: same permissive pattern as other tables
ALTER TABLE tickets             ENABLE ROW LEVEL SECURITY;
ALTER TABLE ticket_comments     ENABLE ROW LEVEL SECURITY;
ALTER TABLE ticket_attachments  ENABLE ROW LEVEL SECURITY;

CREATE POLICY rls_bypass ON tickets            AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON ticket_comments    AS PERMISSIVE FOR ALL USING (true);
CREATE POLICY rls_bypass ON ticket_attachments AS PERMISSIVE FOR ALL USING (true);

-- +goose Down

DROP POLICY IF EXISTS rls_bypass ON ticket_attachments;
DROP POLICY IF EXISTS rls_bypass ON ticket_comments;
DROP POLICY IF EXISTS rls_bypass ON tickets;

ALTER TABLE ticket_attachments DISABLE ROW LEVEL SECURITY;
ALTER TABLE ticket_comments    DISABLE ROW LEVEL SECURITY;
ALTER TABLE tickets            DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS ticket_attachments;
DROP TABLE IF EXISTS ticket_comments;
DROP TABLE IF EXISTS tickets;
