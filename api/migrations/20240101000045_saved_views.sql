-- +goose Up
CREATE TABLE saved_views (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id        UUID NOT NULL REFERENCES orgs(id),
  created_by    UUID NOT NULL REFERENCES users(id),
  entity_type   TEXT NOT NULL,
  name          TEXT NOT NULL,
  filters       JSONB NOT NULL DEFAULT '[]',
  sort_by       TEXT,
  sort_dir      TEXT CHECK (sort_dir IN ('asc', 'desc')),
  is_shared     BOOLEAN NOT NULL DEFAULT false,
  is_pinned     BOOLEAN NOT NULL DEFAULT false,
  pinned_order  INTEGER,
  deleted_at    TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON saved_views(org_id, entity_type) WHERE deleted_at IS NULL;
CREATE INDEX ON saved_views(org_id, is_pinned) WHERE deleted_at IS NULL AND is_pinned = true;

-- +goose Down
DROP TABLE IF EXISTS saved_views;
