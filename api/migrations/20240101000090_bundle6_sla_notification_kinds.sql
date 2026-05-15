-- +goose Up

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_kind_check;
ALTER TABLE notifications
    ADD CONSTRAINT notifications_kind_check CHECK (kind IN (
        'activity_reminder',
        'deal_stage_changed',
        'mention',
        'assignment',
        'sla_warning',
        'sla_breached'
    ));

-- +goose Down

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_kind_check;
ALTER TABLE notifications
    ADD CONSTRAINT notifications_kind_check CHECK (kind IN (
        'activity_reminder',
        'deal_stage_changed',
        'mention',
        'assignment'
    ));
