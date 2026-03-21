-- +goose Up

CREATE TABLE article_categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL,
    sort_order INT  NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id, slug)
);

CREATE TABLE articles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    body        TEXT NOT NULL DEFAULT '',
    category_id UUID REFERENCES article_categories(id) ON DELETE SET NULL,
    tags        TEXT[] NOT NULL DEFAULT '{}',
    status      TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published')),
    author_id   UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    view_count  INT  NOT NULL DEFAULT 0,
    search_vec  TSVECTOR,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_articles_org_id      ON articles (org_id)      WHERE deleted_at IS NULL;
CREATE INDEX idx_articles_category_id ON articles (category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_articles_search_vec  ON articles USING GIN (search_vec);
CREATE INDEX idx_article_categories_org ON article_categories (org_id);

-- Function + trigger to keep search_vec up to date.
CREATE OR REPLACE FUNCTION articles_search_vec_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vec :=
        setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.body,  '')), 'B') ||
        setweight(to_tsvector('english', array_to_string(NEW.tags, ' ')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER articles_search_vec_trigger
BEFORE INSERT OR UPDATE ON articles
FOR EACH ROW EXECUTE FUNCTION articles_search_vec_update();

-- +goose Down

DROP TRIGGER IF EXISTS articles_search_vec_trigger ON articles;
DROP FUNCTION IF EXISTS articles_search_vec_update();
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS article_categories;
