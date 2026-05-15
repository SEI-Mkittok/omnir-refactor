-- +goose Up
-- +goose StatementBegin
ALTER TYPE custom_field_entity_type ADD VALUE IF NOT EXISTS 'quote';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TYPE custom_field_entity_type ADD VALUE IF NOT EXISTS 'kb_article';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE quotes
    ADD COLUMN IF NOT EXISTS custom_fields JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS custom_fields JSONB NOT NULL DEFAULT '{}'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE articles DROP COLUMN IF EXISTS custom_fields;
ALTER TABLE quotes DROP COLUMN IF EXISTS custom_fields;
-- PostgreSQL does not support dropping enum values; quote and kb_article remain.
-- +goose StatementEnd
