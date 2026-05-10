-- +goose Up
ALTER TABLE activities
    ADD COLUMN IF NOT EXISTS start_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS end_at TIMESTAMPTZ;

UPDATE activities
SET start_at = due_date
WHERE start_at IS NULL
  AND due_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS activities_start_at_idx
    ON activities(start_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS activities_end_at_idx
    ON activities(end_at)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS activities_end_at_idx;
DROP INDEX IF EXISTS activities_start_at_idx;

ALTER TABLE activities
    DROP COLUMN IF EXISTS end_at,
    DROP COLUMN IF EXISTS start_at;
