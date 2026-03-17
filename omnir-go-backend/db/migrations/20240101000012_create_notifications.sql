-- +goose Up
CREATE TYPE notification_type AS ENUM ('upcoming_15m', 'upcoming_1h', 'upcoming_1d', 'overdue');

CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    type        notification_type NOT NULL,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Prevent duplicate notifications for the same activity + reminder type.
CREATE UNIQUE INDEX notifications_activity_type_uidx ON notifications(activity_id, type);
CREATE INDEX notifications_user_id_idx ON notifications(user_id);
CREATE INDEX notifications_org_id_idx  ON notifications(org_id);

-- +goose Down
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_type;
