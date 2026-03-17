-- +goose Up
-- +goose StatementBegin
CREATE TABLE notes (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    content     TEXT        NOT NULL,
    entity_type TEXT        NOT NULL CHECK (entity_type IN ('contact', 'account', 'deal')),
    entity_id   UUID        NOT NULL,
    author_id   UUID        NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX notes_entity_idx ON notes (entity_type, entity_id) WHERE deleted_at IS NULL;
CREATE INDEX notes_author_idx  ON notes (author_id)              WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notes;
-- +goose StatementEnd
