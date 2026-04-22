-- +goose Up
CREATE TABLE IF NOT EXISTS account_contacts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id            UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id        UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    contact_id        UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL DEFAULT 'champion',
    is_primary        BOOLEAN NOT NULL DEFAULT FALSE,
    title_at_account  TEXT,
    start_date        DATE,
    end_date          DATE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

-- Add org FK when organizations table exists in this migration context.
DO $$
BEGIN
    IF to_regclass('organizations') IS NOT NULL AND NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_account_contacts_org'
    ) THEN
        ALTER TABLE account_contacts
            ADD CONSTRAINT fk_account_contacts_org
            FOREIGN KEY (org_id) REFERENCES organizations(id);
    END IF;
END $$;

-- Ensure only one active primary account relationship for a contact in an org.
CREATE UNIQUE INDEX IF NOT EXISTS uniq_account_contacts_primary_active_contact_org
    ON account_contacts (org_id, contact_id)
    WHERE is_primary = TRUE AND deleted_at IS NULL AND end_date IS NULL;

CREATE INDEX IF NOT EXISTS idx_account_contacts_org_account
    ON account_contacts (org_id, account_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_account_contacts_org_contact
    ON account_contacts (org_id, contact_id)
    WHERE deleted_at IS NULL;

-- Backfill the compatibility field into the new junction table.
INSERT INTO account_contacts (
    id, org_id, account_id, contact_id, relationship_type, is_primary,
    start_date, created_at, updated_at
)
SELECT
    gen_random_uuid(), c.org_id, c.account_id, c.id, 'champion', TRUE,
    CURRENT_DATE, NOW(), NOW()
FROM contacts c
WHERE c.account_id IS NOT NULL
  AND c.deleted_at IS NULL
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS account_contacts;
