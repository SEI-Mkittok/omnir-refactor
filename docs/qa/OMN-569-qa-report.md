# QA Report: OMN-569 - Email Mark-as-Read Feature

**QA Engineer:** Skadi  
**Date:** 2026-03-27  
**Status:** ❌ BLOCKED  
**Staging URL:** http://100.73.134.90  
**Staging Commit:** eff59ed0

## Summary

Cannot complete E2E testing for email mark-as-read feature due to missing valid test credentials on staging environment.

## Environment Verification

✅ **Staging Deployment:**
- Commit: `eff59ed0 feat(backend): integrations API`
- Branch: `develop` (335 commits ahead, includes OMN-552 frontend+backend)
- Frontend: Running (HTTP 200)
- API: Running (HTTP 200 on health)
- Database: PostgreSQL 16, database `omnir_crm`
- Migration `20240101000062_email_inbox_read_at`: Applied ✅

## Blocker: Invalid Test Credentials

**Issue:** All attempts to authenticate with documented test credentials fail with `401 Unauthorized`.

**Users found in staging database:**
- `admin@omnir.test` (role: admin)
- `client@omnir.test` (role: client)
- `qatest@omnir.test` (role: agent)
- `matthias@omnir.com` (role: admin)

**All users share the same password hash:**  
`$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi`

**Attempted passwords (all failed):**
- `testpassword` (documented in migration comment line 4 of `20240101000031_seed_test_users.sql`)
- `password`

**API Response:**
```bash
$ curl -X POST http://100.73.134.90/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@omnir.test","password":"testpassword"}'
{"error":"unauthorized"}
```

**Root Cause Hypothesis:**
The bcrypt hash in staging database does not match the documented password ("testpassword"). Either:
1. The migration comment is incorrect
2. The hash was manually changed after migration
3. The seed migration was never run, and users were created differently

## Test Artifacts Created

**E2E Test Suite:** `/home/omnirdev/omnir-refactor/web/e2e/inbox.spec.ts`
- Covers all 7 acceptance criteria from OMN-569
- Uses Playwright for automated browser testing
- Ready to run once credentials are resolved

**Staging Config:** `/home/omnirdev/omnir-refactor/web/playwright.staging.config.ts`
- Configured to test against live staging (no local docker required)

## Acceptance Criteria (NOT TESTED)

Cannot verify the following until auth is resolved:

1. ❓ Open inbox and see unread thread indicator
2. ❓ Click thread to open it
3. ❓ Unread indicator disappears after opening
4. ❓ Unread count badge in header decrements
5. ❓ Network call `PATCH /api/v1/emails/{threadId}/read` returns 204
6. ❓ Page reload preserves read status (read_at persisted)
7. ❓ Opening already-read thread is a no-op (204, no error)

## Resolution Required

**To unblock QA testing, one of the following is needed:**

1. **Option A:** Provide correct staging credentials for test users
2. **Option B:** Reset test user password on staging to known value
3. **Option C:** Add seed script to staging deployment that creates/updates test users with documented passwords

**Recommended:** Option C - Add test user seeding to staging deployment process to prevent this issue in future QA cycles.

## Next Steps

1. Escalate credential issue to DevOps/deployment team (Völundr)
2. Once resolved, execute E2E test suite: `npx playwright test inbox.spec.ts --config=playwright.staging.config.ts`
3. Document test results and file bugs for any failures
