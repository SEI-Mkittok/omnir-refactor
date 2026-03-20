-- +goose Up
-- Add storage_key (opaque backend path) and storage_backend columns.
-- storage_url remains for legacy records; new uploads set it to the API download path.

ALTER TABLE ticket_attachments
    ADD COLUMN IF NOT EXISTS storage_key     TEXT,
    ADD COLUMN IF NOT EXISTS storage_backend TEXT;

-- +goose Down

ALTER TABLE ticket_attachments
    DROP COLUMN IF EXISTS storage_key,
    DROP COLUMN IF EXISTS storage_backend;
