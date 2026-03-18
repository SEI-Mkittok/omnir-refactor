-- +goose Up

-- Per-user email notification preferences.
-- A missing row means "all enabled" (default-on behaviour).
CREATE TABLE IF NOT EXISTS user_notification_prefs (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id              UUID        NOT NULL,
    email_on_assigned   BOOLEAN     NOT NULL DEFAULT TRUE,
    email_on_resolved   BOOLEAN     NOT NULL DEFAULT TRUE,
    email_on_closed     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, org_id)
);

CREATE INDEX IF NOT EXISTS idx_user_notification_prefs_user_org ON user_notification_prefs (user_id, org_id);

-- Ensure submitted_by_user_id exists on tickets (may already exist from OMN-145 branch).
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS submitted_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_tickets_submitted_by ON tickets (submitted_by_user_id) WHERE submitted_by_user_id IS NOT NULL;

-- +goose Down
ALTER TABLE tickets DROP COLUMN IF EXISTS submitted_by_user_id;
DROP TABLE IF EXISTS user_notification_prefs;
