-- +goose Up
-- Add vtiger_legacy_id to each entity table for idempotent re-migration
-- and to support post-migration spot-checks by original vtiger record ID.

ALTER TABLE users       ADD COLUMN IF NOT EXISTS vtiger_legacy_id TEXT UNIQUE;
ALTER TABLE accounts    ADD COLUMN IF NOT EXISTS vtiger_legacy_id TEXT UNIQUE;
ALTER TABLE contacts    ADD COLUMN IF NOT EXISTS vtiger_legacy_id TEXT UNIQUE;
ALTER TABLE deals       ADD COLUMN IF NOT EXISTS vtiger_legacy_id TEXT UNIQUE;
ALTER TABLE activities  ADD COLUMN IF NOT EXISTS vtiger_legacy_id TEXT UNIQUE;

-- +goose Down
ALTER TABLE activities  DROP COLUMN IF EXISTS vtiger_legacy_id;
ALTER TABLE deals       DROP COLUMN IF EXISTS vtiger_legacy_id;
ALTER TABLE contacts    DROP COLUMN IF EXISTS vtiger_legacy_id;
ALTER TABLE accounts    DROP COLUMN IF EXISTS vtiger_legacy_id;
ALTER TABLE users       DROP COLUMN IF EXISTS vtiger_legacy_id;
