-- +goose Up
CREATE TABLE IF NOT EXISTS organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE,
    plan       TEXT NOT NULL DEFAULT 'self_hosted' CHECK (plan IN ('self_hosted', 'starter', 'pro', 'enterprise')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed a default org for self-hosted deployments.
INSERT INTO organizations (id, name, slug, plan)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'Default',
    'default',
    'self_hosted'
);

-- +goose Down
DROP TABLE IF EXISTS organizations;
