-- +goose Up
ALTER TABLE entity_attachments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_entity_attachments_deleted_at ON entity_attachments (deleted_at) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_entity_attachments_deleted_at;
ALTER TABLE entity_attachments DROP COLUMN IF EXISTS deleted_at;
