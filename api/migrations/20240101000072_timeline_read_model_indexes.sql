-- +goose Up
-- Timeline query-layer indexes (org/account/contact/date window).

CREATE INDEX IF NOT EXISTS idx_activities_org_account_created_at
    ON activities (org_id, account_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_activities_org_contact_created_at
    ON activities (org_id, contact_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tickets_org_account_created_at
    ON tickets (org_id, account_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tickets_org_contact_created_at
    ON tickets (org_id, contact_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ticket_comments_org_ticket_created_at
    ON ticket_comments (org_id, ticket_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_notes_org_entity_created_at
    ON notes (org_id, entity_type, entity_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_contact_emails_org_contact_sent_at
    ON contact_emails (org_id, contact_id, sent_at DESC);

CREATE INDEX IF NOT EXISTS idx_contact_emails_org_deal_sent_at
    ON contact_emails (org_id, deal_id, sent_at DESC);

CREATE INDEX IF NOT EXISTS idx_email_inbox_messages_org_contact_sent_at
    ON email_inbox_messages (org_id, contact_id, sent_at DESC);

CREATE INDEX IF NOT EXISTS idx_sequence_events_org_contact_occurred_at
    ON sequence_events (org_id, contact_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_quotes_org_contact_created_at
    ON quotes (org_id, contact_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_quotes_org_deal_created_at
    ON quotes (org_id, deal_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_deals_org_account_id
    ON deals (org_id, account_id, id)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_deals_org_account_id;
DROP INDEX IF EXISTS idx_quotes_org_deal_created_at;
DROP INDEX IF EXISTS idx_quotes_org_contact_created_at;
DROP INDEX IF EXISTS idx_sequence_events_org_contact_occurred_at;
DROP INDEX IF EXISTS idx_email_inbox_messages_org_contact_sent_at;
DROP INDEX IF EXISTS idx_contact_emails_org_deal_sent_at;
DROP INDEX IF EXISTS idx_contact_emails_org_contact_sent_at;
DROP INDEX IF EXISTS idx_notes_org_entity_created_at;
DROP INDEX IF EXISTS idx_ticket_comments_org_ticket_created_at;
DROP INDEX IF EXISTS idx_tickets_org_contact_created_at;
DROP INDEX IF EXISTS idx_tickets_org_account_created_at;
DROP INDEX IF EXISTS idx_activities_org_contact_created_at;
DROP INDEX IF EXISTS idx_activities_org_account_created_at;
