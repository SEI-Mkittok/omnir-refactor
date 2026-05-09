-- +goose Up
UPDATE saved_views
SET filters = '{}'::jsonb
WHERE filters IS NULL
   OR jsonb_typeof(filters) <> 'object';

ALTER TABLE saved_views
  ALTER COLUMN filters SET DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE saved_views
  ALTER COLUMN filters SET DEFAULT '[]'::jsonb;
