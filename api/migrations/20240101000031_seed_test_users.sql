-- +goose Up
-- Seed test users for QA/staging environments.
-- These users are inserted only if they do not already exist (idempotent).
-- Password hash is bcrypt of "testpassword" (cost=10).

INSERT INTO users (id, org_id, name, email, password_hash, role, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000002',
    'Test Client',
    'client@omnir.test',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'client',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE email = 'client@omnir.test'
);

-- +goose Down
DELETE FROM users WHERE email = 'client@omnir.test';
