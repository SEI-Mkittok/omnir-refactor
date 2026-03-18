-- +goose Up
-- +goose StatementBegin

-- Drop the old activity-reminder-specific notifications table and its type
DROP TABLE IF EXISTS notifications CASCADE;
DROP TYPE IF EXISTS notification_type;

-- Create the new general CRM notifications table
CREATE TABLE notifications (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID        NOT NULL,
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id      UUID        REFERENCES users(id) ON DELETE SET NULL,
    kind          TEXT        NOT NULL CHECK (kind IN (
                                'activity_reminder',
                                'deal_stage_changed',
                                'mention',
                                'assignment'
                              )),
    entity_type   TEXT,
    entity_id     UUID,
    title         TEXT        NOT NULL,
    body          TEXT,
    read_at       TIMESTAMPTZ,
    emailed_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX notifications_user_read_created ON notifications (user_id, read_at, created_at DESC);
CREATE INDEX notifications_org_id ON notifications (org_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd
