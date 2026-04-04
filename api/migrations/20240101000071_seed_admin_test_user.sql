-- +goose Up
-- Seed the admin test user for QA/staging environments.
-- The Playwright e2e suite expects admin@omnir.test / testpassword to exist.
-- Password hash is bcrypt of "testpassword" (cost=10).
-- Hash generated via: bcrypt.GenerateFromPassword([]byte("testpassword"), 10)
-- Value: $2a$10$ir1cSGkQOTGDBE5Q0/4CruvovW9GavYvgivCW0NZ1GsY/dz.gVnVa

INSERT INTO users (id, org_id, name, email, password_hash, role, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    'Admin',
    'admin@omnir.test',
    '$2a$10$ir1cSGkQOTGDBE5Q0/4CruvovW9GavYvgivCW0NZ1GsY/dz.gVnVa',
    'admin',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE email = 'admin@omnir.test'
);

-- +goose Down
DELETE FROM users WHERE email = 'admin@omnir.test';
