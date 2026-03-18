-- +goose Up
ALTER TABLE sequence_enrollments
    ADD COLUMN IF NOT EXISTS next_step_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_sequence_enrollments_pending
    ON sequence_enrollments (next_step_at)
    WHERE status = 'active' AND next_step_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_sequence_enrollments_pending;
ALTER TABLE sequence_enrollments DROP COLUMN IF EXISTS next_step_at;
