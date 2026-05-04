-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    CREATE TYPE account_relationship_type AS ENUM ('parent', 'subsidiary', 'partner', 'reseller', 'vendor');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS account_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    parent_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    child_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    relationship_type account_relationship_type NOT NULL,
    ownership_percent NUMERIC(5,2),
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    deleted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT account_relationships_no_self_loop CHECK (parent_account_id <> child_account_id),
    CONSTRAINT account_relationships_ownership_pct_range CHECK (ownership_percent IS NULL OR (ownership_percent >= 0 AND ownership_percent <= 100)),
    CONSTRAINT account_relationships_effective_dates CHECK (effective_to IS NULL OR effective_to >= effective_from)
);
-- +goose StatementEnd

CREATE UNIQUE INDEX IF NOT EXISTS account_relationships_unique_active_idx
    ON account_relationships (org_id, parent_account_id, child_account_id, relationship_type)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS account_relationships_parent_idx ON account_relationships (org_id, parent_account_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS account_relationships_child_idx ON account_relationships (org_id, child_account_id) WHERE deleted_at IS NULL;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_account_relationships_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_account_relationships_updated_at ON account_relationships;
CREATE TRIGGER trg_account_relationships_updated_at
BEFORE UPDATE ON account_relationships
FOR EACH ROW EXECUTE FUNCTION set_account_relationships_updated_at();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION validate_account_relationship_cycle()
RETURNS TRIGGER AS $$
DECLARE
    has_cycle BOOLEAN;
BEGIN
    IF NEW.parent_account_id = NEW.child_account_id THEN
        RAISE EXCEPTION 'account relationship cannot self-reference';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM accounts p
        WHERE p.id = NEW.parent_account_id
          AND p.org_id = NEW.org_id
          AND p.deleted_at IS NULL
    ) IS FALSE THEN
        RAISE EXCEPTION 'parent account does not exist in org';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM accounts c
        WHERE c.id = NEW.child_account_id
          AND c.org_id = NEW.org_id
          AND c.deleted_at IS NULL
    ) IS FALSE THEN
        RAISE EXCEPTION 'child account does not exist in org';
    END IF;

    WITH RECURSIVE descendants(id) AS (
        SELECT NEW.child_account_id
        UNION
        SELECT ar.child_account_id
        FROM account_relationships ar
        JOIN descendants d ON d.id = ar.parent_account_id
        WHERE ar.org_id = NEW.org_id
          AND ar.deleted_at IS NULL
          AND (TG_OP <> 'UPDATE' OR ar.id <> NEW.id)
    )
    SELECT EXISTS (SELECT 1 FROM descendants WHERE id = NEW.parent_account_id) INTO has_cycle;

    IF has_cycle THEN
        RAISE EXCEPTION 'account relationship cycle detected';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_validate_account_relationship_cycle ON account_relationships;
CREATE TRIGGER trg_validate_account_relationship_cycle
BEFORE INSERT OR UPDATE OF parent_account_id, child_account_id, deleted_at ON account_relationships
FOR EACH ROW
WHEN (NEW.deleted_at IS NULL)
EXECUTE FUNCTION validate_account_relationship_cycle();

ALTER TABLE account_relationships ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS org_isolation ON account_relationships;
CREATE POLICY org_isolation ON account_relationships
    USING (org_id = current_org_id());

-- +goose Down
DROP POLICY IF EXISTS org_isolation ON account_relationships;
ALTER TABLE account_relationships DISABLE ROW LEVEL SECURITY;
DROP TRIGGER IF EXISTS trg_validate_account_relationship_cycle ON account_relationships;
DROP FUNCTION IF EXISTS validate_account_relationship_cycle();
DROP TRIGGER IF EXISTS trg_account_relationships_updated_at ON account_relationships;
DROP FUNCTION IF EXISTS set_account_relationships_updated_at();
DROP TABLE IF EXISTS account_relationships;
DROP TYPE IF EXISTS account_relationship_type;
