-- +goose Up
-- contact_emails stores inbound and outbound emails linked to CRM contacts/deals.
CREATE TABLE contact_emails (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    contact_id  UUID        REFERENCES contacts(id) ON DELETE SET NULL,
    deal_id     UUID        REFERENCES deals(id) ON DELETE SET NULL,
    direction   TEXT        NOT NULL CHECK (direction IN ('inbound', 'outbound')),
    from_addr   TEXT        NOT NULL,
    to_addr     TEXT        NOT NULL,
    subject     TEXT        NOT NULL DEFAULT '',
    body        TEXT        NOT NULL DEFAULT '',
    thread_id   TEXT        NOT NULL DEFAULT '',
    message_id  TEXT,
    sent_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contact_emails_contact_id ON contact_emails(contact_id);
CREATE INDEX idx_contact_emails_org_id     ON contact_emails(org_id);
CREATE INDEX idx_contact_emails_thread_id  ON contact_emails(thread_id);

-- +goose Down

DROP TABLE IF EXISTS contact_emails;
