-- +goose Up

-- Add submitted_by_user_id to tickets so client-submitted tickets can be
-- filtered and access-controlled per the portal API.
ALTER TABLE tickets
    ADD COLUMN IF NOT EXISTS submitted_by_user_id UUID REFERENCES users(id);

CREATE INDEX idx_tickets_submitted_by ON tickets (submitted_by_user_id)
    WHERE submitted_by_user_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_tickets_submitted_by;
ALTER TABLE tickets DROP COLUMN IF EXISTS submitted_by_user_id;
