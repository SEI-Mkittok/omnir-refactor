-- +goose Up
-- +goose StatementBegin

-- Per-org, per-type atomic counters for document numbering.
CREATE TABLE document_sequences (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    doc_type     TEXT        NOT NULL CHECK (doc_type IN ('quote', 'ticket', 'kb_article', 'invoice')),
    next_number  BIGINT      NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, doc_type)
);
CREATE INDEX idx_document_sequences_org_id ON document_sequences(org_id);

-- Per-org configurable starting numbers for each document type.
CREATE TABLE org_settings (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID        NOT NULL UNIQUE REFERENCES orgs(id) ON DELETE CASCADE,
    quote_number_start      BIGINT      NOT NULL DEFAULT 1,
    ticket_number_start     BIGINT      NOT NULL DEFAULT 1,
    kb_article_number_start BIGINT      NOT NULL DEFAULT 1,
    invoice_number_start    BIGINT      NOT NULL DEFAULT 1,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_org_settings_org_id ON org_settings(org_id);

-- Add document number columns to quotes, tickets, and kb articles.
ALTER TABLE quotes   ADD COLUMN number         BIGINT;
ALTER TABLE quotes   ADD COLUMN number_prefix  TEXT NOT NULL DEFAULT 'QUO';
CREATE INDEX idx_quotes_org_number ON quotes(org_id, number);

ALTER TABLE tickets  ADD COLUMN number         BIGINT;
ALTER TABLE tickets  ADD COLUMN number_prefix  TEXT NOT NULL DEFAULT 'TKT';
CREATE INDEX idx_tickets_org_number ON tickets(org_id, number);

ALTER TABLE articles ADD COLUMN number         BIGINT;
ALTER TABLE articles ADD COLUMN number_prefix  TEXT NOT NULL DEFAULT 'KB';
CREATE INDEX idx_articles_org_number ON articles(org_id, number);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_articles_org_number;
ALTER TABLE articles DROP COLUMN IF EXISTS number_prefix;
ALTER TABLE articles DROP COLUMN IF EXISTS number;

DROP INDEX IF EXISTS idx_tickets_org_number;
ALTER TABLE tickets  DROP COLUMN IF EXISTS number_prefix;
ALTER TABLE tickets  DROP COLUMN IF EXISTS number;

DROP INDEX IF EXISTS idx_quotes_org_number;
ALTER TABLE quotes   DROP COLUMN IF EXISTS number_prefix;
ALTER TABLE quotes   DROP COLUMN IF EXISTS number;

DROP TABLE IF EXISTS org_settings;
DROP TABLE IF EXISTS document_sequences;
-- +goose StatementEnd
