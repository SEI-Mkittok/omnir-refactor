-- +goose Up

CREATE TABLE product_help_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    sort_order  INT  NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE product_help_articles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID REFERENCES product_help_categories(id) ON DELETE SET NULL,
    title       TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    body        TEXT NOT NULL DEFAULT '',
    excerpt     TEXT NOT NULL DEFAULT '',
    tags        TEXT[] NOT NULL DEFAULT '{}',
    status      TEXT NOT NULL DEFAULT 'published' CHECK (status IN ('draft','published')),
    source_path TEXT NOT NULL,
    wiki_url    TEXT NOT NULL DEFAULT '',
    edit_url    TEXT NOT NULL DEFAULT '',
    sort_order  INT  NOT NULL DEFAULT 0,
    view_count  INT  NOT NULL DEFAULT 0,
    search_vec  TSVECTOR,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE product_help_sync_runs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status           TEXT NOT NULL CHECK (status IN ('succeeded','failed')),
    message          TEXT NOT NULL DEFAULT '',
    categories_count INT  NOT NULL DEFAULT 0,
    articles_count   INT  NOT NULL DEFAULT 0,
    started_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at      TIMESTAMPTZ
);

CREATE INDEX idx_product_help_articles_category ON product_help_articles (category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_help_articles_status   ON product_help_articles (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_help_articles_search   ON product_help_articles USING GIN (search_vec);
CREATE INDEX idx_product_help_categories_active ON product_help_categories (sort_order, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_help_sync_runs_started ON product_help_sync_runs (started_at DESC);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION product_help_articles_search_vec_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vec :=
        setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.excerpt, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.body,  '')), 'C') ||
        setweight(to_tsvector('english', array_to_string(NEW.tags, ' ')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER product_help_articles_search_vec_trigger
BEFORE INSERT OR UPDATE ON product_help_articles
FOR EACH ROW EXECUTE FUNCTION product_help_articles_search_vec_update();

-- +goose Down

DROP TRIGGER IF EXISTS product_help_articles_search_vec_trigger ON product_help_articles;
DROP FUNCTION IF EXISTS product_help_articles_search_vec_update();
DROP TABLE IF EXISTS product_help_sync_runs;
DROP TABLE IF EXISTS product_help_articles;
DROP TABLE IF EXISTS product_help_categories;
