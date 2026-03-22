-- +goose Up
-- +goose StatementBegin
ALTER TABLE email_inbox_messages ADD COLUMN read_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE email_inbox_messages DROP COLUMN read_at;
-- +goose StatementEnd
