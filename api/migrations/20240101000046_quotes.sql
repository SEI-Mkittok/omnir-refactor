-- Phase 11: Products, Price Books, and Quotes
-- OMN-409/OMN-410

-- +goose Up

-- Products catalog
CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    sku         TEXT,
    description TEXT,
    unit_price_cents BIGINT NOT NULL DEFAULT 0,
    currency    TEXT NOT NULL DEFAULT 'USD',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_org_id ON products(org_id);
CREATE INDEX idx_products_org_active ON products(org_id, is_active);

-- Quotes
CREATE TABLE quotes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    deal_id      UUID REFERENCES deals(id) ON DELETE SET NULL,
    contact_id   UUID REFERENCES contacts(id) ON DELETE SET NULL,
    title        TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','sent','approved','rejected','expired')),
    currency     TEXT NOT NULL DEFAULT 'USD',
    valid_until  DATE,
    notes        TEXT,
    sent_at      TIMESTAMPTZ,
    approved_at  TIMESTAMPTZ,
    rejected_at  TIMESTAMPTZ,
    created_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    total_cents  BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quotes_org_id ON quotes(org_id);
CREATE INDEX idx_quotes_deal_id ON quotes(deal_id);
CREATE INDEX idx_quotes_status ON quotes(org_id, status);

-- Quote line items
CREATE TABLE quote_line_items (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_id         UUID NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    product_id       UUID REFERENCES products(id) ON DELETE SET NULL,
    product_name     TEXT NOT NULL,
    description      TEXT,
    quantity         NUMERIC(12,3) NOT NULL DEFAULT 1,
    unit_price_cents BIGINT NOT NULL DEFAULT 0,
    discount_pct     NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (discount_pct >= 0 AND discount_pct <= 100),
    total_cents      BIGINT NOT NULL DEFAULT 0,
    sort_order       INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quote_line_items_quote_id ON quote_line_items(quote_id);

-- +goose Down

DROP INDEX IF EXISTS idx_quote_line_items_quote_id;
DROP TABLE IF EXISTS quote_line_items;

DROP INDEX IF EXISTS idx_quotes_status;
DROP INDEX IF EXISTS idx_quotes_deal_id;
DROP INDEX IF EXISTS idx_quotes_org_id;
DROP TABLE IF EXISTS quotes;

DROP INDEX IF EXISTS idx_products_org_active;
DROP INDEX IF EXISTS idx_products_org_id;
DROP TABLE IF EXISTS products;
