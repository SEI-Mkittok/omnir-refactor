-- OMN: account-contact link model rollout (Phase 1)

-- +goose Up

ALTER TABLE account_contacts
    ADD COLUMN IF NOT EXISTS id UUID,
    ADD COLUMN IF NOT EXISTS org_id UUID,
    ADD COLUMN IF NOT EXISTS role TEXT,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE account_contacts ac
SET id = gen_random_uuid()
WHERE ac.id IS NULL;

UPDATE account_contacts ac
SET org_id = a.org_id
FROM accounts a
WHERE a.id = ac.account_id
  AND ac.org_id IS NULL;

ALTER TABLE account_contacts
    ALTER COLUMN id SET DEFAULT gen_random_uuid(),
    ALTER COLUMN id SET NOT NULL,
    ALTER COLUMN org_id SET NOT NULL;

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_account_contacts_org_id'
    ) THEN
        ALTER TABLE account_contacts
            ADD CONSTRAINT fk_account_contacts_org_id
            FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE;
    END IF;
END $$;
-- +goose StatementEnd

CREATE UNIQUE INDEX IF NOT EXISTS idx_account_contacts_id
    ON account_contacts (id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_account_contacts_org_account_contact
    ON account_contacts (org_id, account_id, contact_id);

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
DROP INDEX IF EXISTS idx_account_contacts_org_account_contact;
DROP INDEX IF EXISTS idx_account_contacts_id;

DROP INDEX IF EXISTS idx_account_contacts_org_primary;
DROP INDEX IF EXISTS idx_account_contacts_org_contact;
DROP INDEX IF EXISTS idx_account_contacts_org_account;
DROP INDEX IF EXISTS idx_account_contacts_primary_per_contact;
