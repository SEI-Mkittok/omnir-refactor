-- +goose Up
CREATE TABLE IF NOT EXISTS deals (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title                TEXT NOT NULL,
    value_cents          BIGINT NOT NULL DEFAULT 0,
    currency             TEXT NOT NULL DEFAULT 'USD',
    stage                TEXT NOT NULL DEFAULT 'lead'
                             CHECK (stage IN ('lead', 'qualified', 'proposal', 'negotiation', 'closed_won', 'closed_lost')),
    probability          INT NOT NULL DEFAULT 0 CHECK (probability BETWEEN 0 AND 100),
    expected_close_date  DATE,
    contact_id           UUID REFERENCES contacts(id),
    account_id           UUID REFERENCES accounts(id),
    owner_id             UUID NOT NULL REFERENCES users(id),
    pipeline_id          UUID NOT NULL REFERENCES pipelines(id),
    custom_fields        JSONB,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);

CREATE INDEX idx_deals_owner_id    ON deals (owner_id)    WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_account_id  ON deals (account_id)  WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_contact_id  ON deals (contact_id)  WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_pipeline_id ON deals (pipeline_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_stage       ON deals (stage)       WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS deals;
