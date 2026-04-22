-- Add account context to quotes and backfill from linked deals

-- +goose Up

ALTER TABLE quotes
    ADD COLUMN account_id UUID REFERENCES accounts(id) ON DELETE SET NULL;

UPDATE quotes q
SET account_id = d.account_id
FROM deals d
WHERE q.deal_id = d.id
  AND q.account_id IS NULL
  AND d.account_id IS NOT NULL;

CREATE INDEX idx_quotes_account_id ON quotes(account_id);

-- +goose Down

DROP INDEX IF EXISTS idx_quotes_account_id;
ALTER TABLE quotes DROP COLUMN IF EXISTS account_id;
