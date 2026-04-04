-- +goose Up
ALTER TABLE org_onboarding
    ADD COLUMN IF NOT EXISTS step_status JSONB NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE org_onboarding
    DROP COLUMN IF EXISTS step_status;
