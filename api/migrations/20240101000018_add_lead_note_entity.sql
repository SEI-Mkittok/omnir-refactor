-- +goose Up
-- +goose StatementBegin
ALTER TABLE notes
    DROP CONSTRAINT IF EXISTS notes_entity_type_check;

ALTER TABLE notes
    ADD CONSTRAINT notes_entity_type_check
        CHECK (entity_type IN ('contact', 'account', 'deal', 'lead'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE notes
    DROP CONSTRAINT IF EXISTS notes_entity_type_check;

ALTER TABLE notes
    ADD CONSTRAINT notes_entity_type_check
        CHECK (entity_type IN ('contact', 'account', 'deal'));

DELETE FROM notes WHERE entity_type = 'lead';
-- +goose StatementEnd
