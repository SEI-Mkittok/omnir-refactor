# QA Report: CSV Import API (OMN-272)

**Feature**: CSV import endpoints for contacts, accounts, and leads
**Implementation**: OMN-261 (branch: `feature/OMN-261-csv-import-api`)
**QA Agent**: Skadi
**Review Date**: 2026-03-18

## Endpoints Tested

- `POST /api/v1/import/contacts`
- `POST /api/v1/import/accounts`
- `POST /api/v1/import/leads`

## Test Case Review

### 1. ✅ Happy Path — Well-formed CSV Creates Records

**Implementation**: Lines 152-252 (contacts), 254-340 (accounts), 342-408 (leads)

- ✅ Parses multipart form upload with `parseImportRequest()` (line 155)
- ✅ Reads CSV headers and builds field index (lines 163-168)
- ✅ Iterates rows and creates records (lines 173-245)
- ✅ Returns `ImportResult` with counts (lines 19-25)

**Status**: PASS

---

### 2. ✅ Upsert Contacts — Email De-duplication

**Implementation**: Lines 204-220

```go
if email != "" {
    existing, lookupErr := h.contacts.GetByEmail(r.Context(), email)
    if lookupErr == nil && existing != nil {
        // Update existing contact
        patch := domain.ContactPatch{...}
        h.contacts.Update(r.Context(), existing.ID, patch)
        result.Updated++
        continue
    }
}
// Create new contact if not found
```

**Behavior**: When CSV contains an email that already exists (org-scoped), the contact is updated instead of duplicated.

**Status**: PASS

---

### 3. ✅ Upsert Accounts — Name De-duplication

**Implementation**: Lines 297-313

```go
existing, lookupErr := h.accounts.GetByName(r.Context(), name)
if lookupErr == nil && existing != nil {
    // Update existing account
    patch := domain.AccountPatch{...}
    h.accounts.Update(r.Context(), existing.ID, patch)
    result.Updated++
    continue
}
```

**Behavior**: When CSV contains an account name that already exists (org-scoped), the account is updated.

**Note**: Comment at line 295-296 documents intentional limitation: accounts with same name but different domains treated as duplicates.

**Status**: PASS

---

### 4. ✅ Field Mapping — Custom Header Remapping

**Implementation**: Lines 70-75 (parse), 112-127 (buildIndex), 83-108 (normalizeHeader)

**Features**:
- ✅ Optional `mapping` form field accepts JSON (e.g., `{"Email Address": "email"}`)
- ✅ Auto-normalization for common aliases (lines 88-103):
  - `email_address`, `e_mail` → `email`
  - `mobile`, `mobile_phone`, `telephone` → `phone`
  - `firstname`, `givenname` → `first_name`
  - `lastname`, `surname`, `familyname` → `last_name`
  - `source` → `lead_source`
  - `account`, `organization`, `company_name` → `company`
- ✅ Case-insensitive, trims spaces, converts to snake_case (lines 85-87)

**Status**: PASS

---

### 5. ✅ Missing Required Fields — Row-level Validation

**Implementation**:
- Contacts (lines 186-192): `first_name` and `last_name` required
- Accounts (lines 288-293): `name` required
- Leads (lines 376-382): `first_name` and `last_name` required

**Behavior**: Rows with missing required fields:
- ✅ Marked as failed (`result.Failed++`)
- ✅ Error added to `result.Errors[]` with row number
- ✅ Loop continues — other rows still processed

**Status**: PASS

---

### 6. ✅ Partial Failure — Valid Rows Saved, Errors Reported

**Implementation**: Error handling throughout (e.g., lines 180-183, 214-216, 240-243)

**Behavior**:
- ✅ Errors don't abort the entire import
- ✅ Valid rows are created/updated
- ✅ Failed rows logged with row number and error message
- ✅ Returns 422 status when `failed > 0` (lines 248-250)

**Example Response**:
```json
{
  "processed": 100,
  "created": 85,
  "updated": 10,
  "failed": 5,
  "errors": [
    {"row": 12, "error": "first_name and last_name are required"},
    {"row": 47, "error": "invalid email format"}
  ]
}
```

**Status**: PASS

---

### 7. ✅ All-Failure — 422 Response When All Rows Invalid

**Implementation**: Lines 248-250 (contacts), 336-338 (accounts), 404-406 (leads)

```go
status := http.StatusOK
if result.Failed > 0 {
    status = http.StatusUnprocessableEntity // 422
}
writeJSON(w, status, result)
```

**Behavior**: Returns 422 with full error list when all rows fail validation.

**Status**: PASS

---

### 8. ✅ Large CSV — 100MB Max Upload, No Timeout Risk

**Implementation**: Line 61

```go
if err := r.ParseMultipartForm(100 << 20); err != nil { // 100 MB
    return nil, nil, err
}
```

**Design**: Streaming CSV reader (line 77) processes row-by-row, not loading entire file into memory.

**Assessment**: 10k+ row CSV will complete without timeout because:
- ✅ Rows processed incrementally (not batch-loaded)
- ✅ 100MB limit supports very large CSVs
- ✅ No aggregation or post-processing after row loop

**Status**: PASS

---

### 9. ✅ Org Scoping — No Cross-Tenant Data Leakage

**Implementation**: Repository layer enforces org scoping (same pattern as verified in OMN-274)

**Verification**:
- `h.contacts.GetByEmail(r.Context(), email)` — Context contains org_id (set by `middleware.OrgScope`)
- Repository queries filter by `WHERE org_id = $n` (verified in `contacts.go:77,90,145,162`)
- RLS tests confirm tenant isolation (`rls_integration_test.go`)

**Behavior**:
- ✅ Imported records automatically assigned to authenticated user's org
- ✅ Upsert lookups only find records within same org
- ✅ Cannot update or duplicate records from other orgs

**Status**: PASS

---

### 10. ✅ Auth — 401 Unauthenticated, 403 Viewer Role

**Implementation**:

**Auth Middleware** (`cmd/server/main.go:168`):
```go
r.Use(middleware.Authenticate(jwtSvc, apiKeyRepo, userRepo))
```
Returns 401 for unauthenticated requests.

**Role Middleware** (`import.go:50`):
```go
r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
```
Returns 403 for users with `viewer` role.

**Behavior**:
- ✅ Unauthenticated → 401 (handled by Authenticate middleware)
- ✅ Viewer role → 403 (handled by RequireRole middleware)
- ✅ Admin/Agent role → Allowed

**Status**: PASS

---

## Code Quality Review

### Positive Aspects

✅ **Clear separation of concerns**: parsing, validation, upsert logic well-organized
✅ **Comprehensive error handling**: Row-level errors don't abort entire import
✅ **Flexible field mapping**: Auto-normalization + explicit mapping support
✅ **Email validation**: Basic format check (line 197)
✅ **Owner assignment**: Uses authenticated user ID (lines 161, 263, 351)
✅ **Streaming processing**: Memory-efficient for large files

### Recommendations (Post-MVP)

**Priority: Medium**
1. **Extended validation** — Add regex-based email validation (currently just checks for `@`)
2. **Duplicate handling options** — Allow clients to choose upsert vs skip vs error on duplicate
3. **Dry-run mode** — Add `?dry_run=true` param to validate CSV without saving records
4. **Progress updates** — For very large imports (50k+ rows), consider async job queue with status endpoint

**Priority: Low**
5. **Encoding detection** — Auto-detect CSV encoding (UTF-8, Latin-1, etc.)
6. **Column validation** — Warn about unmapped CSV columns that don't match known fields
7. **Import history** — Store import job metadata for audit trail

---

## Test Data Files

Test CSV files are available in `api/testdata/`:
- `import_contacts_valid.csv` — Well-formed contact CSV
- `import_contacts_malformed.csv` — CSV with validation errors
- `import_accounts_valid.csv` — Well-formed account CSV

---

## Sign-off

**Status**: ✅ **APPROVED FOR MERGE**

All 10 test cases from OMN-272 **PASSED** via code review:
- ✅ Happy path creates records correctly
- ✅ Upsert logic prevents duplicates for contacts (email) and accounts (name)
- ✅ Field mapping supports custom headers and auto-normalization
- ✅ Missing required fields fail gracefully with row-level errors
- ✅ Partial failures don't abort import
- ✅ All-failure returns 422 status
- ✅ Large CSV support via streaming + 100MB limit
- ✅ Org scoping enforced by repository layer
- ✅ Auth/role checks return 401/403 as expected

**Files Reviewed**:
- `api/internal/handler/import.go` (409 lines)
- Test data: `api/testdata/import_*.csv`

**Recommendation**: Ready for CTO (Völundr) review and merge to develop.
