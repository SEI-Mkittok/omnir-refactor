#!/usr/bin/env bash
# seed-staging.sh — Idempotent upsert of test users for staging/QA.
# Runs after goose migrations on every deploy so test users survive
# regardless of whether migrations were already applied or DB was reset.
#
# Requires DATABASE_URL environment variable.

set -euo pipefail

if [ -z "${DATABASE_URL:-}" ]; then
  echo "ERROR: DATABASE_URL is not set" >&2
  exit 1
fi

echo "Seeding staging test users…"

psql "${DATABASE_URL}" <<'SQL'
-- Ensure default test org exists
INSERT INTO orgs (id, name, plan, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000002',
  'Test Organization',
  'pro',
  NOW(), NOW()
)
ON CONFLICT (id) DO NOTHING;

-- admin@omnir.test  (password: testpassword)
INSERT INTO users (id, org_id, name, email, password_hash, role, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000002',
  'Admin',
  'admin@omnir.test',
  '$2a$10$ir1cSGkQOTGDBE5Q0/4CruvovW9GavYvgivCW0NZ1GsY/dz.gVnVa',
  'admin',
  NOW(), NOW()
)
ON CONFLICT (id) DO UPDATE SET
  email       = EXCLUDED.email,
  password_hash = EXCLUDED.password_hash,
  role        = EXCLUDED.role,
  updated_at  = NOW();

-- client@omnir.test  (password: testpassword)
INSERT INTO users (id, org_id, name, email, password_hash, role, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000010',
  '00000000-0000-0000-0000-000000000002',
  'Test Client',
  'client@omnir.test',
  '$2a$10$F.qxTseC5SuaDx/09fP6hOC2BAKoCDMIKpLbDcy783Wy1KT.fkN2q',
  'client',
  NOW(), NOW()
)
ON CONFLICT (id) DO UPDATE SET
  email       = EXCLUDED.email,
  password_hash = EXCLUDED.password_hash,
  role        = EXCLUDED.role,
  updated_at  = NOW();

SQL

echo "Staging test users seeded ✓"
