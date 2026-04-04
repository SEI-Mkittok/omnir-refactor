-- +goose Up
ALTER TABLE org_onboarding
    ADD COLUMN IF NOT EXISTS dismissed BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE org_onboarding
    DROP COLUMN IF EXISTS dismissed;
