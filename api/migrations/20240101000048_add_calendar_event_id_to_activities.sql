-- Phase 11: Calendar Sync
-- OMN-415

-- +goose Up

ALTER TABLE activities ADD COLUMN IF NOT EXISTS calendar_event_id TEXT;

CREATE INDEX IF NOT EXISTS idx_activities_calendar_event_id
    ON activities(calendar_event_id)
    WHERE calendar_event_id IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_activities_calendar_event_id;
ALTER TABLE activities DROP COLUMN IF EXISTS calendar_event_id;
