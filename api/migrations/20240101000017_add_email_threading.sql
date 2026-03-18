-- +goose Up

-- Store the inbound email Message-ID on tickets for reply threading.
-- When a reply arrives with In-Reply-To matching a stored message ID, we
-- append a comment to the existing ticket instead of creating a new one.
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS email_message_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tickets_email_message_id
    ON tickets (email_message_id)
    WHERE email_message_id IS NOT NULL AND deleted_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_tickets_email_message_id;
ALTER TABLE tickets DROP COLUMN IF EXISTS email_message_id;
