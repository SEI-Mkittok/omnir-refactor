-- +goose Up
-- Seed test users for QA/staging environments.
-- These users are inserted only if they do not already exist (idempotent).
-- Password hash is bcrypt of "testpassword" (cost=10).
-- Hash generated via: bcrypt.GenerateFromPassword([]byte("testpassword"), 10)
-- Value: $2a$10$F.qxTseC5SuaDx/09fP6hOC2BAKoCDMIKpLbDcy783Wy1KT.fkN2q

INSERT INTO users (id, org_id, name, email, password_hash, role, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000002',
    'Test Client',
    'client@omnir.test',
    '$2a$10$F.qxTseC5SuaDx/09fP6hOC2BAKoCDMIKpLbDcy783Wy1KT.fkN2q',
    'client',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE email = 'client@omnir.test'
);

-- +goose Down
DELETE FROM users WHERE email = 'client@omnir.test';
