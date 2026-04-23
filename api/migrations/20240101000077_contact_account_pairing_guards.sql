-- +goose Up
-- Optional safeguards for contact/account consistency across core entities.

CREATE TABLE IF NOT EXISTS account_contacts (
    account_id  UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    contact_id  UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    is_primary  BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (account_id, contact_id)
);

CREATE INDEX IF NOT EXISTS idx_account_contacts_contact_id ON account_contacts(contact_id);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION crm_contact_related_to_account(p_contact_id UUID, p_account_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql
AS $$
BEGIN
    IF p_contact_id IS NULL OR p_account_id IS NULL THEN
        RETURN TRUE;
    END IF;

    RETURN EXISTS (
        SELECT 1
        FROM contacts c
        WHERE c.id = p_contact_id
          AND c.deleted_at IS NULL
          AND (
            c.account_id = p_account_id
            OR EXISTS (
                SELECT 1
                FROM account_contacts ac
                WHERE ac.contact_id = p_contact_id
                  AND ac.account_id = p_account_id
            )
          )
    );
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION crm_validate_entity_contact_account_pair()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.contact_id IS NULL OR NEW.account_id IS NULL THEN
        RETURN NEW;
    END IF;

    IF NOT crm_contact_related_to_account(NEW.contact_id, NEW.account_id) THEN
        RAISE EXCEPTION 'invalid contact/account pairing: contact % is not related to account %',
            NEW.contact_id, NEW.account_id
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_tickets_validate_contact_account_pair ON tickets;
CREATE TRIGGER trg_tickets_validate_contact_account_pair
BEFORE INSERT OR UPDATE OF contact_id, account_id ON tickets
FOR EACH ROW EXECUTE FUNCTION crm_validate_entity_contact_account_pair();

DROP TRIGGER IF EXISTS trg_deals_validate_contact_account_pair ON deals;
CREATE TRIGGER trg_deals_validate_contact_account_pair
BEFORE INSERT OR UPDATE OF contact_id, account_id ON deals
FOR EACH ROW EXECUTE FUNCTION crm_validate_entity_contact_account_pair();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION crm_validate_quote_contact_pair()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_deal_account_id UUID;
BEGIN
    IF NEW.contact_id IS NULL OR NEW.deal_id IS NULL THEN
        RETURN NEW;
    END IF;

    SELECT d.account_id
      INTO v_deal_account_id
      FROM deals d
     WHERE d.id = NEW.deal_id
       AND d.deleted_at IS NULL;

    IF v_deal_account_id IS NOT NULL THEN
        IF NOT crm_contact_related_to_account(NEW.contact_id, v_deal_account_id) THEN
            RAISE EXCEPTION 'invalid quote contact/deal pairing: contact % is not related to deal account %',
                NEW.contact_id, v_deal_account_id
                USING ERRCODE = '23514';
        END IF;
        RETURN NEW;
    END IF;

    IF NOT EXISTS (
        SELECT 1
          FROM deal_contacts dc
         WHERE dc.deal_id = NEW.deal_id
           AND dc.contact_id = NEW.contact_id
    ) THEN
        RAISE EXCEPTION 'invalid quote contact/deal pairing: contact % is not linked to deal %',
            NEW.contact_id, NEW.deal_id
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_quotes_validate_contact_pair ON quotes;
CREATE TRIGGER trg_quotes_validate_contact_pair
BEFORE INSERT OR UPDATE OF contact_id, deal_id ON quotes
FOR EACH ROW EXECUTE FUNCTION crm_validate_quote_contact_pair();

-- +goose Down
DROP TRIGGER IF EXISTS trg_quotes_validate_contact_pair ON quotes;
DROP TRIGGER IF EXISTS trg_deals_validate_contact_account_pair ON deals;
DROP TRIGGER IF EXISTS trg_tickets_validate_contact_account_pair ON tickets;

DROP FUNCTION IF EXISTS crm_validate_quote_contact_pair();
DROP FUNCTION IF EXISTS crm_validate_entity_contact_account_pair();
DROP FUNCTION IF EXISTS crm_contact_related_to_account(UUID, UUID);

DROP INDEX IF EXISTS idx_account_contacts_contact_id;
DROP TABLE IF EXISTS account_contacts;
