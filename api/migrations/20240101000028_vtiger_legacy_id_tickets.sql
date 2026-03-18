-- +goose Up
-- Tickets were created after the initial vtiger_legacy_id migration (000013),
-- so add the column here to enable idempotent ticket migration and spot-checks.

ALTER TABLE tickets ADD COLUMN IF NOT EXISTS vtiger_legacy_id TEXT UNIQUE;

-- +goose Down
ALTER TABLE tickets DROP COLUMN IF EXISTS vtiger_legacy_id;
