-- +goose Up
CREATE TABLE IF NOT EXISTS deal_contacts (
    deal_id     UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    contact_id  UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    role        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (deal_id, contact_id)
);

CREATE INDEX idx_deal_contacts_deal_id    ON deal_contacts (deal_id);
CREATE INDEX idx_deal_contacts_contact_id ON deal_contacts (contact_id);

-- contact_id on deals stays nullable for backwards compat until frontend migrates.

-- +goose Down
DROP TABLE IF EXISTS deal_contacts;
