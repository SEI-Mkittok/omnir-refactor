-- +goose Up

-- Products catalog
CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    sku         TEXT,
    price       NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency    TEXT NOT NULL DEFAULT 'USD',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_products_org_id ON products (org_id);
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
CREATE POLICY products_org ON products USING (true) WITH CHECK (true);

-- Price books
CREATE TABLE IF NOT EXISTS price_books (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_price_books_org_id ON price_books (org_id);
ALTER TABLE price_books ENABLE ROW LEVEL SECURITY;
CREATE POLICY price_books_org ON price_books USING (true) WITH CHECK (true);

-- Price book entries (product price overrides per book)
CREATE TABLE IF NOT EXISTS price_book_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    price_book_id   UUID NOT NULL REFERENCES price_books(id) ON DELETE CASCADE,
    product_id      UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price_override  NUMERIC(12,2),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (price_book_id, product_id)
);
ALTER TABLE price_book_entries ENABLE ROW LEVEL SECURITY;
CREATE POLICY price_book_entries_org ON price_book_entries USING (true) WITH CHECK (true);

-- Deal line items
CREATE TABLE IF NOT EXISTS deal_line_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id      UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    product_id   UUID REFERENCES products(id) ON DELETE SET NULL,
    name         TEXT NOT NULL,
    quantity     NUMERIC(12,4) NOT NULL DEFAULT 1,
    unit_price   NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_pct NUMERIC(5,2)  NOT NULL DEFAULT 0 CHECK (discount_pct >= 0 AND discount_pct <= 100),
    subtotal     NUMERIC(12,2) NOT NULL DEFAULT 0,
    position     INT           NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_deal_line_items_deal_id ON deal_line_items (deal_id);
ALTER TABLE deal_line_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY deal_line_items_org ON deal_line_items USING (true) WITH CHECK (true);

-- +goose Down
DROP TABLE IF EXISTS deal_line_items;
DROP TABLE IF EXISTS price_book_entries;
DROP TABLE IF EXISTS price_books;
DROP TABLE IF EXISTS products;
