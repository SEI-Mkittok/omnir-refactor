-- +goose Up
-- +goose StatementBegin
-- Ensure the default single-tenant org is on the 'pro' plan so that all
-- plan-gated features (KB, quotes, enrichment, etc.) work out of the box.
-- Uses ON CONFLICT to upsert: inserts a new row if none exists, otherwise
-- upgrades the plan to 'pro'.
INSERT INTO org_plans (org_id, plan, status)
VALUES ('00000000-0000-0000-0000-000000000002', 'pro', 'active')
ON CONFLICT (org_id) DO UPDATE
    SET plan       = 'pro',
        status     = 'active',
        updated_at = NOW()
WHERE org_plans.plan <> 'enterprise';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE org_plans
SET plan       = 'free',
    updated_at = NOW()
WHERE org_id = '00000000-0000-0000-0000-000000000002'
  AND plan    = 'pro';
-- +goose StatementEnd
