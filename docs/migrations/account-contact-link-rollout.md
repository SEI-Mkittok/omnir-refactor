# Account–Contact Link Migration Rollout (Phases 1–4)

This rollout replaces the legacy single-link model (`contacts.account_id`) with an explicit junction model (`account_contacts`) while preserving compatibility during migration.

## Goals

- Support many-to-many account/contact relationships.
- Keep API and frontend behavior stable during migration.
- Prevent cross-org data leakage during dual-write and read-switch periods.
- Gate every phase with repeatable data-quality checks.

## Phase 1 — Schema additions + backfill

### Scope

1. Add `account_contacts` table:
   - `id UUID PK`
   - `org_id UUID NOT NULL`
   - `account_id UUID NOT NULL`
   - `contact_id UUID NOT NULL`
   - `is_primary BOOLEAN NOT NULL DEFAULT false`
   - `role TEXT NULL`
   - `created_at`, `updated_at`
2. Add optional `quotes.account_id` column (nullable FK).
3. Add consistency indexes for join/read patterns and org scoping.
4. Backfill from `contacts.account_id` into `account_contacts` with `is_primary=true`.

### Gate criteria

Run Phase 1 data-quality script(s). Gate passes only if:
- No orphaned links (`account_contacts.account_id` / `contact_id` pointing to missing rows).
- No cross-org leaks (link `org_id` mismatches account/contact orgs).
- No backfill mismatches (`contacts.account_id` not represented in `account_contacts`).

## Phase 2 — Dual-write in repositories/handlers

### Scope

1. On contact/account link mutation paths, write to both:
   - Legacy field path (`contacts.account_id`, where still used).
   - New link model (`account_contacts`).
2. Add idempotent upsert semantics for `account_contacts` writes.
3. Add parity logging/metrics for dual-write failures.

### Gate criteria

Run Phase 2 data-quality script(s). Gate passes only if:
- No orphaned links.
- No cross-org leaks.
- No mismatches between legacy field and primary link projection.

## Phase 3 — Read switch behind feature flags

### Scope

1. Introduce feature flags for read paths:
   - `ACCOUNT_CONTACT_LINK_READS`
   - `ACCOUNT_CONTACT_LINK_LISTS`
2. Flip repository/handler reads to source account relationships from `account_contacts` when flags are enabled.
3. Keep fallback to legacy reads while flags are off.

### Gate criteria

Run Phase 3 data-quality script(s). Gate passes only if:
- No orphaned links.
- No cross-org leaks.
- No read parity mismatches between legacy and new model for sampled production-like datasets.

## Phase 4 — Remove deprecated fields/paths

### Scope

1. After frontend migration + verification, remove:
   - Legacy read/write code paths using `contacts.account_id` as source of truth.
2. Optionally retain `contacts.account_id` as derived/compatibility field, or drop in a separate migration window.
3. Remove migration flags after stable rollout window.

### Gate criteria

Run Phase 4 data-quality script(s). Gate passes only if:
- No orphaned links.
- No cross-org leaks.
- No contact/account mismatches in post-cutover reads.

## Data-Quality Scripts (required before each phase gate)

Use SQL checks in `scripts/data-quality/`:

- `phase-1-precheck.sql`
- `phase-2-precheck.sql`
- `phase-3-precheck.sql`
- `phase-4-precheck.sql`

Each script emits `check_name`, `issue_count`, and a compact sample payload. Any non-zero `issue_count` blocks the next phase.
