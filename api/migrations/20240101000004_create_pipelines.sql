-- +goose Up
CREATE TABLE IF NOT EXISTS pipelines (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    stages     JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed a default pipeline so deals can reference it immediately.
INSERT INTO pipelines (id, name, stages)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Default',
    '[
        {"id":"lead","name":"Lead","order":1},
        {"id":"qualified","name":"Qualified","order":2},
        {"id":"proposal","name":"Proposal","order":3},
        {"id":"negotiation","name":"Negotiation","order":4},
        {"id":"closed_won","name":"Closed Won","order":5},
        {"id":"closed_lost","name":"Closed Lost","order":6}
    ]'::jsonb
);

-- +goose Down
DROP TABLE IF EXISTS pipelines;
