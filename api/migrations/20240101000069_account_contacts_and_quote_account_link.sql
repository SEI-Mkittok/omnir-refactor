-- OMN: account-contact link model rollout (Phase 1)

-- +goose Up

CREATE TABLE IF NOT EXISTS account_contacts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id  UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    contact_id  UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
    role        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, account_id, contact_id)
);

-- At most one primary account per contact per org.
CREATE UNIQUE INDEX IF NOT EXISTS idx_account_contacts_primary_per_contact
    ON account_contacts (org_id, contact_id)
    WHERE is_primary = TRUE;

-- Read/index consistency for account- and contact-oriented lookups.
CREATE INDEX IF NOT EXISTS idx_account_contacts_org_account
    ON account_contacts (org_id, account_id, contact_id);
CREATE INDEX IF NOT EXISTS idx_account_contacts_org_contact
    ON account_contacts (org_id, contact_id, account_id);
CREATE INDEX IF NOT EXISTS idx_account_contacts_org_primary
    ON account_contacts (org_id, is_primary, contact_id);

-- Optional quote->account association used during relationship migration.
ALTER TABLE quotes
    ADD COLUMN IF NOT EXISTS account_id UUID REFERENCES accounts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_quotes_org_account_id
    ON quotes (org_id, account_id)
    WHERE account_id IS NOT NULL;

-- Backfill primary links from legacy contacts.account_id.
INSERT INTO account_contacts (org_id, account_id, contact_id, is_primary, created_at, updated_at)
SELECT c.org_id, c.account_id, c.id, TRUE, NOW(), NOW()
FROM contacts c
WHERE c.account_id IS NOT NULL
ON CONFLICT (org_id, account_id, contact_id)
DO UPDATE SET
    is_primary = EXCLUDED.is_primary,
    updated_at = NOW();

-- +goose Down

DROP INDEX IF EXISTS idx_quotes_org_account_id;
ALTER TABLE quotes DROP COLUMN IF EXISTS account_id;

DROP INDEX IF EXISTS idx_account_contacts_org_primary;
DROP INDEX IF EXISTS idx_account_contacts_org_contact;
DROP INDEX IF EXISTS idx_account_contacts_org_account;
DROP INDEX IF EXISTS idx_account_contacts_primary_per_contact;
DROP TABLE IF EXISTS account_contacts;
