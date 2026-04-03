-- +goose Up
-- +goose StatementBegin
-- Repair schema drift: migration 046 was applied to staging before total_cents
-- and the deal/account enum values were added to the migration source file.
-- This migration adds the missing pieces idempotently.

-- 1. Add deal and account to custom_field_entity_type enum (missing on staging).
--    ALTER TYPE … ADD VALUE is not transactional, so it must run outside a
--    BEGIN block. Goose StatementBegin/End wraps each statement in its own
--    implicit transaction, which is correct for ADD VALUE.
ALTER TYPE custom_field_entity_type ADD VALUE IF NOT EXISTS 'deal';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TYPE custom_field_entity_type ADD VALUE IF NOT EXISTS 'account';
-- +goose StatementEnd

-- +goose StatementBegin
-- 2. Add total_cents to quotes (missing on staging).
ALTER TABLE quotes ADD COLUMN IF NOT EXISTS total_cents BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- total_cents can be dropped idempotently.
ALTER TABLE quotes DROP COLUMN IF EXISTS total_cents;
-- Note: PostgreSQL does not support DROP VALUE on enums; deal/account values
-- remain in the type. They are harmless if no rows use them.
-- +goose StatementEnd
