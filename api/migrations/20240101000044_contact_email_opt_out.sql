-- Add email opt-out and bounce tracking to contacts
-- Required by sequence unsubscribe and bounce webhook endpoints (OMN-398)

-- +goose Up

ALTER TABLE contacts
    ADD COLUMN email_opt_out  BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN bounce_count   INT     NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE contacts
    DROP COLUMN IF EXISTS email_opt_out,
    DROP COLUMN IF EXISTS bounce_count;
