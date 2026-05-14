# QA Report: OMN-411 - Phase 11 / Quotes Implementation

**Date:** 2026-03-20
**QA Engineer:** Skadi
**Status:** ❌ **BLOCKED - Critical bugs prevent deployment**

## Executive Summary

The quotes implementation has **5 critical bugs** that will cause runtime failures:
1. Duplicate migration files with conflicting schemas
2. Missing org_id column in quote_line_items INSERT
3. Missing total_cents persistence
4. Missing products table migration (in one file)
5. Missing goose markers (in one file)

**Recommendation:** Fix all critical bugs before merge. No runtime testing possible until code compiles.

---

## Critical Bugs (Must Fix)

### 🔴 **BUG-1: Duplicate Migration Files with Same Version**

**Files:**
- `api/migrations/20240101000046_create_quotes.sql`
- `api/migrations/20240101000046_quotes.sql`

**Impact:** Goose will only run ONE of these migrations (whichever sorts first alphabetically). The second file will be ignored, causing schema mismatch.

**Details:**
- Both files have identical version number `20240101000046`
- File 1 (`create_quotes.sql`): Has RLS, missing products table, status='invoiced'
- File 2 (`quotes.sql`): No RLS, has products table, status='expired'

**Fix Required:**
1. Delete one of the files
2. OR rename to different version numbers (e.g., `000046` for quotes, `000047` for products)
3. Merge schemas into single correct migration

---

### 🔴 **BUG-2: Missing org_id Column in Line Items INSERT**

**File:** `api/internal/repository/postgres/quotes.go:289-294`

**Code:**
```go
row := tx.QueryRow(ctx,
    `INSERT INTO quote_line_items
     (id, quote_id, product_id, product_name, description, quantity, unit_price_cents, discount_pct, total_cents, sort_order)
     VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
     RETURNING ...`,
```

**Problem:**
- INSERT lists 10 columns but VALUES provides 10 parameters
- Migration `20240101000046_create_quotes.sql` has `org_id` column in line_items table
- Migration `20240101000046_quotes.sql` does NOT have `org_id` column

**Impact:**
- If using `create_quotes.sql`: SQL error "column org_id does not have a default value"
- If using `quotes.sql`: No RLS protection on line items

**Fix Required:**
- **Option A:** Add `org_id` to quote_line_items table and INSERT statement (recommended for RLS)
- **Option B:** Remove `org_id` from table schema if RLS not needed

---

### 🔴 **BUG-3: total_cents Never Persisted to Database**

**File:** `api/internal/repository/postgres/quotes.go`

**Problem:**
- Domain model has `TotalCents int64` field
- Migration `20240101000046_create_quotes.sql` has `total_cents` column in quotes table
- Migration `20240101000046_quotes.sql` is MISSING `total_cents` column
- Repository NEVER updates total_cents in database (only computed in memory)

**Impact:**
- Quote totals are calculated on every read but never stored
- No ability to query/sort quotes by total amount
- Inconsistent state if line items change outside normal flow

**Fix Required:**
1. Add `total_cents BIGINT NOT NULL DEFAULT 0` to quotes table in migration
2. Update `Create()` to save computed total:
   ```sql
   UPDATE quotes SET total_cents=$1 WHERE id=$2
   ```
3. Update `Update()` to recompute and save total after line item changes

---

### 🔴 **BUG-4: Schema Mismatch - Status Values**

**File:** `api/migrations/20240101000046_create_quotes.sql:8`

**Migration:**
```sql
status TEXT NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft','sent','approved','rejected','invoiced'))
```

**Domain Model (`api/internal/domain/quote.go:12-17`):**
```go
const (
    QuoteStatusDraft    QuoteStatus = "draft"
    QuoteStatusSent     QuoteStatus = "sent"
    QuoteStatusApproved QuoteStatus = "approved"
    QuoteStatusRejected QuoteStatus = "rejected"
    QuoteStatusExpired  QuoteStatus = "expired"  // NOT 'invoiced'!
)
```

**Impact:** Cannot set quote status to "expired" - database will reject with CHECK constraint violation.

**Fix Required:** Change migration to use `'expired'` instead of `'invoiced'`

---

### 🔴 **BUG-5: Missing goose Up/Down Markers**

**File:** `api/migrations/20240101000046_quotes.sql`

**Problem:** Missing `-- +goose Up` and `-- +goose Down` directives

**Impact:** Goose will not recognize this as a valid migration file and will skip it.

**Fix Required:** Add goose markers:
```sql
-- +goose Up

CREATE TABLE products ...

-- +goose Down

DROP TABLE IF EXISTS quote_line_items;
DROP TABLE IF EXISTS quotes;
DROP TABLE IF EXISTS products;
```

---

## Schema Comparison

| Feature | create_quotes.sql | quotes.sql | Domain Model | Repo Code |
|---------|-------------------|------------|--------------|-----------|
| Goose markers | ✅ Has Up/Down | ❌ Missing | N/A | N/A |
| RLS policies | ✅ Enabled | ❌ None | N/A | N/A |
| Products table | ❌ Missing | ✅ Included | ✅ domain/product.go | ✅ postgres/products.go |
| Quote.title | ❌ Missing | ✅ Included | ✅ Required | ✅ Scanned |
| Quote.total_cents | ✅ Included | ❌ Missing | ✅ Computed | ❌ Not persisted |
| Quote.currency | ❌ Missing | ✅ Included | ✅ Default USD | ✅ Scanned |
| Quote.sent_at/approved_at/rejected_at | ❌ Missing | ✅ Included | ✅ Optional | ✅ Scanned |
| Quote.created_by | ❌ Missing | ✅ Included | ✅ Optional | ✅ Scanned |
| Quote status values | ❌ 'invoiced' | ✅ 'expired' | ✅ 'expired' | ✅ 'expired' |
| Line items org_id | ✅ Included | ❌ Missing | ❌ Not in struct | ❌ Not in INSERT |
| Line items discount column | `discount_percent` | `discount_pct` | `discount_pct` | `discount_pct` |
| Soft delete (deleted_at) | ✅ Included | ❌ Hard delete | ❌ Hard delete | ❌ Hard delete |

---

## Correct Migration Schema (Recommended)

Merge both files into a single migration with:
- ✅ Goose Up/Down markers
- ✅ Products table
- ✅ Quotes table with ALL fields (title, currency, timestamps, created_by, total_cents)
- ✅ Status CHECK using 'expired' not 'invoiced'
- ✅ Quote line items with org_id for RLS
- ✅ RLS policies for multi-tenant isolation
- ✅ Consistent column names (`discount_pct` not `discount_percent`)

---

## Code Review Findings (Non-Critical)

### ⚠️ **ISSUE-1: Missing Error Handling for Email Send**

**File:** `api/internal/handler/quotes.go:227`

```go
_ = h.mailer.SendDirect(req.To, subject, req.Message)
```

**Issue:** Email errors are silently ignored. If SMTP fails, quote is marked as "sent" but no email was delivered.

**Recommendation:** Either log the error or return 500 if email is critical to the send flow.

---

### ⚠️ **ISSUE-2: N+1 Query Problem in List()**

**File:** `api/internal/repository/postgres/quotes.go:251-258`

**Code:**
```go
for _, q := range quotes {
    items, err := r.listLineItems(ctx, q.ID)
    // ...
}
```

**Issue:** For 50 quotes, this makes 51 database queries (1 for quotes + 50 for line items).

**Recommendation:** Use a single JOIN or IN clause to fetch all line items at once:
```sql
SELECT ... FROM quote_line_items WHERE quote_id = ANY($1)
```

---

### ℹ️ **INFO-1: Missing PDF Generation**

**QA Requirement:** "Test PDF generation"

**Status:** No PDF generation code found in handlers or repository.

**Impact:** Cannot test PDF generation feature. May be planned for future implementation.

---

### ℹ️ **INFO-2: Missing Invoice Linkage**

**QA Requirement:** "Test invoice linkage"

**Status:** No invoice table or foreign key relationship found.

**Impact:** Cannot test invoice linkage. May be planned for future phase.

---

## Test Coverage

### ✅ Code Review: PASS (with fixes required)
- Domain models: Well-structured, validation present
- Repository: Standard CRUD pattern, transactions used correctly
- Handlers: Proper error handling, JWT claims integration
- API routes: Proper role-based access control

### ❌ Compilation: BLOCKED
- Cannot compile due to missing Go binary
- Pre-push validation script failed

### ❌ Unit Tests: NOT RUN
- Blocked on compilation issues

### ❌ Integration Tests: NOT RUN
- Need staging database access
- Blocked on migration conflicts

### ❌ Edge Cases: NOT TESTED
- Empty quote: Cannot test without runtime
- Discount > 100%: CHECK constraint should prevent (in quotes.sql only)
- Expired quote: Status value mismatch prevents testing

---

## Recommendations

### Immediate Actions (Before Merge)
1. **Fix BUG-1:** Delete duplicate migration, create single correct version
2. **Fix BUG-2:** Add org_id to line items table + INSERT statement
3. **Fix BUG-3:** Persist total_cents to quotes table
4. **Fix BUG-4:** Change 'invoiced' to 'expired' in status CHECK
5. **Fix BUG-5:** Add goose markers to migration file

### Before Deployment
1. Run pre-push validation to confirm compilation
2. Run unit tests for quote domain logic
3. Test migration rollback (goose Down)
4. Manual API testing on staging

### Future Improvements
1. Implement PDF generation endpoint
2. Add invoice table and linkage
3. Optimize N+1 query in List()
4. Add error handling for email failures

---

## Files Reviewed

### Domain Models
- ✅ `api/internal/domain/quote.go` (135 lines)
- ✅ `api/internal/domain/product.go` (52 lines)

### Repository
- ✅ `api/internal/repository/interfaces.go` (diff: +22 lines)
- ✅ `api/internal/repository/postgres/quotes.go` (355 lines)
- ⚠️ `api/internal/repository/postgres/products.go` (not reviewed yet)

### Handlers
- ✅ `api/internal/handler/quotes.go` (241 lines)
- ⚠️ `api/internal/handler/products.go` (not reviewed yet)

### Migrations
- ❌ `api/migrations/20240101000046_create_quotes.sql` (62 lines, conflicts)
- ❌ `api/migrations/20240101000046_quotes.sql` (60 lines, conflicts)

---

## QA Sign-Off

**Status:** ❌ **REJECT - Critical bugs must be fixed**

**Blocker:** Cannot proceed with runtime testing until:
1. Migration conflicts resolved
2. Schema matches domain model
3. Code compiles successfully

**Next Steps:**
1. Implementation team fixes 5 critical bugs
2. Re-run QA validation
3. Manual testing on staging environment

**Estimated Fix Time:** 1-2 hours for experienced developer

---

**QA Engineer:** Skadi
**Date:** 2026-03-20T23:10:00Z
**Issue:** [OMN-411](/OMN/issues/OMN-411)
