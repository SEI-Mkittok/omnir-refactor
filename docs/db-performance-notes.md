# DB Performance Notes

**Issue:** OMN-260
**Date:** 2026-03-18
**Scope:** Post Phase 1–5 performance audit — indexes, N+1, pagination

---

## 1. Index Audit

### Status: ✅ All tables covered after migration 000029

All entity tables (`contacts`, `accounts`, `deals`, `tickets`, `users`, `orgs`, `leads`, `activities`) have:

- Single-column `org_id` partial indexes (migration 000010, 000016, 000020)
- Composite `(org_id, <filter_col>)` indexes for common List() filter patterns (migrations 000027, 000029)
- GIN FTS expression indexes for full-text search (migrations 000026, 000029)

**Gaps found and fixed in migration 000029:**

| Table | Gap | Fix |
|---|---|---|
| `leads` | No composite indexes on `(org_id, status)`, `(org_id, owner_id)`, `(org_id, created_at DESC)` | Added in 000029 |
| `leads` | No GIN FTS index; search used `ILIKE '%..%'` (seq scan) | Added `idx_leads_fts` in 000029 |
| `activities` | No composite indexes on `(org_id, type)`, `(org_id, owner_id)`, `(org_id, created_at DESC)` | Added in 000029 |
| `activities` | No GIN FTS index; search used `ILIKE '%..%'` (seq scan) | Added `idx_activities_fts` in 000029 |

### Index inventory (key tables)

**contacts**
- `idx_contacts_org_id`, `idx_contacts_owner_id`, `idx_contacts_account_id`, `idx_contacts_stage`, `idx_contacts_email`
- Composites: `idx_contacts_org_owner`, `idx_contacts_org_created`
- FTS: `idx_contacts_fts` — GIN on `to_tsvector(first_name, last_name, email, phone)`

**accounts**
- `idx_accounts_org_id`, `idx_accounts_owner_id`, `idx_accounts_name`
- Composites: `idx_accounts_org_owner`, `idx_accounts_org_created`
- FTS: `idx_accounts_fts` — GIN on `to_tsvector(name, domain, industry)`

**deals**
- `idx_deals_org_id`, `idx_deals_owner_id`, `idx_deals_stage`, `idx_deals_pipeline_id`, `idx_deals_account_id`, `idx_deals_contact_id`
- Composites: `idx_deals_org_stage`, `idx_deals_org_owner`, `idx_deals_org_created`
- FTS: `idx_deals_fts` — GIN on `to_tsvector(title)`

**leads** *(fixed in 000029)*
- `idx_leads_org_id`, `idx_leads_owner_id`, `idx_leads_status`, `idx_leads_email`
- Composites: `idx_leads_org_status`, `idx_leads_org_owner`, `idx_leads_org_created`
- FTS: `idx_leads_fts` — GIN on `to_tsvector(first_name, last_name, email, company)`

**tickets**
- `idx_tickets_org_id`, `idx_tickets_assignee`, `idx_tickets_contact`, `idx_tickets_account`, `idx_tickets_status`, `idx_tickets_priority`
- Composites: `idx_tickets_org_status`, `idx_tickets_org_priority`, `idx_tickets_org_assignee`, `idx_tickets_org_created`
- FTS: `idx_tickets_fts` — GIN on `to_tsvector(subject, description)`

**activities** *(fixed in 000029)*
- `idx_activities_org_id`, `activities_owner_id_idx`, `activities_contact_id_idx`, `activities_account_id_idx`, `activities_deal_id_idx`
- Composites: `idx_activities_org_type`, `idx_activities_org_owner`, `idx_activities_org_created`
- FTS: `idx_activities_fts` — GIN on `to_tsvector(subject, description)`

---

## 2. N+1 Query Check

### Status: ✅ No N+1 issues found

All `List()` repository methods use a single `COUNT(*) + SELECT` pattern per request — no per-row queries. Reviewed:

- `ContactRepo.List()` — single query ✅
- `AccountRepo.List()` — single query ✅
- `DealRepo.List()` — single query ✅
- `LeadRepo.List()` — single query ✅
- `TicketRepo.List()` — single query ✅
- `ActivityRepo.List()` — single query ✅
- `NoteRepo.ListByEntity()` — single query ✅
- `NotificationRepo.ListByUser()` — single query ✅

`DealHandler.AddContact` performs 2 queries (insert link + fetch updated deal) — expected write-path pattern, not an N+1.

Reports queries use `GROUP BY` aggregations — appropriate, no pagination needed.

---

## 3. Search Query Fix (ILIKE → FTS)

`LeadRepo.List()` and `ActivityRepo.List()` previously used:
```sql
(first_name ILIKE '%query%' OR last_name ILIKE '%query%' ...)
```

Leading-wildcard `ILIKE` cannot use btree indexes and forces a sequential scan. Fixed in both repos to:
```sql
to_tsvector('english', ...) @@ plainto_tsquery('english', $N)
```

This matches the expression GIN indexes added in migration 000029 and is consistent with how `contacts`, `accounts`, `deals`, and `tickets` already handle search.

**Trade-off:** FTS does not match partial word stems (e.g. searching "acme" won't match "acmecorp"). If partial-match search is needed in future, a `pg_trgm` GIN index can be layered on top. The `pg_trgm` extension is already enabled (`db-init/00_extensions.sql`).

---

## 4. Pagination Review

### Status: ✅ All list endpoints paginated

Every list handler enforces `LIMIT/OFFSET` with:
- Default page size: **50**
- Maximum page size: **200** (hard cap enforced in handler layer)

Endpoints audited:
- `GET /contacts` ✅
- `GET /accounts` ✅
- `GET /deals` ✅
- `GET /leads` ✅
- `GET /tickets` ✅
- `GET /activities` ✅
- `GET /notes` (entity-scoped) ✅
- `GET /notifications` ✅

No unbounded result set paths found.

---

## 5. Slow Query Log Recommendation

To capture real slow queries in staging, enable temporarily:

```sql
-- In postgresql.conf or via ALTER SYSTEM:
ALTER SYSTEM SET log_min_duration_statement = '100ms';
SELECT pg_reload_conf();
```

Run the QA test suite and check `pg_stat_statements` or the Postgres log for queries > 100ms. Reset with:

```sql
ALTER SYSTEM RESET log_min_duration_statement;
SELECT pg_reload_conf();
```

---

## 6. Future Considerations

- **Cursor pagination**: `LIMIT/OFFSET` degrades at high page numbers (page 100+ of 50 = offset 5000). For high-volume tables (tickets, activities) consider keyset/cursor pagination in a follow-up.
- **Partial-match search on leads**: If UX requires "starts with" or substring matching (vs FTS token matching), add a `pg_trgm` GIN index on `lower(first_name || ' ' || last_name)`.
- **`EXPLAIN ANALYZE` baseline**: Before shipping to production load, capture `EXPLAIN ANALYZE` plans for the 5 most common list queries and store them in this file as a regression baseline.
