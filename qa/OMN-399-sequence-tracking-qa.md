# OMN-399: QA Report - Sequence Tracking Endpoints

**Date:** 2026-03-20
**QA Agent:** Skadi
**Parent Issue:** OMN-398 (Email Sequences: tracking pixel, click redirect, bounce webhook, unsubscribe)
**Status:** Code Review ✅ Complete | Runtime Testing ⚠️ Blocked

---

## Code Review Findings: ✅ PASS

### 1. Open Pixel Tracking ✅
**Implementation:** `api/internal/handler/sequence_tracking.go:76-86`
- Route: `GET /track/open/{token}`
- Returns: 1×1 transparent GIF (43 bytes)
- Content-Type: `image/gif`
- Records: `sequence_events` with `kind=opened`
- Graceful handling: Always returns pixel even if token invalid (for email client retries)

**Email Injection:** `api/internal/email/mailer.go:63-69`
```go
pixel := fmt.Sprintf(`<img src="%s/track/open/%s" width="1" height="1" style="display:none" alt="">`, baseURL, tokens.OpenToken)
```
- Injected before `</body>` tag or appended to HTML

### 2. Click Redirect ✅
**Implementation:** `api/internal/handler/sequence_tracking.go:92-105`
- Route: `GET /track/click/{token}?url={destination}`
- Validation: Requires http:// or https:// prefix
- Returns: 302 redirect to destination
- Records: `sequence_events` with `kind=clicked`

**Link Rewriting:** `api/internal/email/mailer.go:96-106`
```go
hrefRe := regexp.MustCompile(`(?i)<a\s[^>]*href="(https?://[^"]+)"`)
wrapped := fmt.Sprintf(`%s/track/click/%s?url=%s`, baseURL, clickToken, original)
```
- All `<a href>` links in email HTML are rewritten with click tracking

### 3. Unsubscribe Endpoint ✅
**Implementation:** `api/internal/handler/sequence_tracking.go:111-129`
- Routes: `GET|POST /unsubscribe/{token}`
- Sets enrollment `status = unsubscribed`
- Sets contact `email_opt_out = true`
- Records: `sequence_events` with `kind=unsubscribed`
- Returns: HTML confirmation page

**Footer Injection:** `api/internal/email/mailer.go:74-83`
```html
<p style="font-size:11px;color:#999;margin-top:24px">
  Don't want these emails? <a href="%s/unsubscribe/%s">Unsubscribe</a>
</p>
```
- Injected before `</body>` tag (CAN-SPAM compliant)

### 4. Bounce Webhook ✅
**Implementation:** `api/internal/handler/sequence_tracking.go:141-188`
- Route: `POST /api/emails/bounce`
- Format: SendGrid-compatible (array of events with `event=bounce`)
- Actions per bounced email:
  - Fetch contact by email
  - Increment contact `bounce_count`
  - Mark all active enrollments as `bounced`
  - Record `sequence_events` with `kind=bounced`
- Returns: JSON with `{"processed": N}` count

### 5. Token Security ✅
**Implementation:** `api/internal/seqtoken/token.go`
- Algorithm: HMAC-SHA256 (signed tokens)
- Format: `base64url(json_payload) + "." + base64url(hmac_signature)`
- TTL: 90 days (`TokenTTL = 90 * 24 * time.Hour`)
- Claims: `{orgID, sequenceID, enrollmentID, stepID, expiresAt}`
- Verification:
  - `Verify()` checks HMAC signature match
  - Returns `ErrExpiredToken` if `time.Now() > expiresAt`
  - Returns `ErrInvalidToken` if signature mismatch or malformed

### 6. Database Schema ✅
**Migration:** `api/migrations/20240101000044_contact_email_opt_out.sql`
```sql
ALTER TABLE contacts
    ADD COLUMN email_opt_out  BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN bounce_count   INT     NOT NULL DEFAULT 0;
```

### 7. Route Registration ✅
**Main Router:** `api/cmd/server/main.go:209-212`
```go
r.Mount("/track", sequenceTrackingHandler.TrackRouter())
r.Mount("/unsubscribe", sequenceTrackingHandler.UnsubscribeRouter())
r.Mount("/api/emails/bounce", sequenceTrackingHandler.BounceRouter())
```
- Public routes (no JWT auth required)
- Token-based authentication via HMAC signatures

### 8. Sequence Worker Integration ✅
**Worker:** `api/internal/worker/sequence_worker.go:119`
```go
err := w.mailer.SendSequenceEmail(enrollment.ContactEmail, step.Subject, htmlBody, tokens)
```
- Worker generates tokens per enrollment/step
- Passes tokens to mailer for injection

---

## Runtime Testing: ⚠️ BLOCKED

**Blocker:** No access to staging environment from this local machine.

### Required Runtime Tests (from checklist):
- [ ] Enroll a contact in a sequence with an email step
- [ ] Verify outbound email contains `<img src=".../track/open/{token}">` tag
- [ ] GET `/track/open/{token}` returns 200 with `Content-Type: image/gif`
- [ ] `sequence_events` row created with `kind=opened`
- [ ] Verify outbound email links are rewritten to `/track/click/{token}?url={original}`
- [ ] GET `/track/click/{token}?url=https://example.com` returns 302 to `https://example.com`
- [ ] `sequence_events` row created with `kind=clicked`
- [ ] Verify outbound email contains unsubscribe footer link
- [ ] GET `/unsubscribe/{token}` returns 200 HTML confirmation page
- [ ] POST `/unsubscribe/{token}` returns 200 HTML confirmation page
- [ ] Enrollment status set to `unsubscribed`
- [ ] Contact `email_opt_out = true`
- [ ] `sequence_events` row created with `kind=unsubscribed`
- [ ] POST `/api/emails/bounce` with SendGrid bounce payload (array with `event=bounce`)
- [ ] Contact `bounce_count` incremented
- [ ] All active enrollments for that contact set to `bounced`
- [ ] `sequence_events` rows created with `kind=bounced`
- [ ] Tampered token rejected (400/silent fail)
- [ ] Expired token (>90 days) rejected

### Next Steps:
1. **Option A:** Provide SSH access to omnir-dev-2 staging server for runtime testing
2. **Option B:** Spin up local Docker environment with migrations applied
3. **Option C:** Mark code review as sufficient and close this QA task

---

## Conclusion

**Code Quality:** ✅ PASS
**Implementation Completeness:** ✅ PASS
**Security:** ✅ PASS (HMAC-signed tokens, expiry validation)
**Compliance:** ✅ PASS (CAN-SPAM unsubscribe footer)

**Recommendation:** Implementation is production-ready from a code perspective. Runtime testing should be performed on staging to verify end-to-end flow, but no code issues were found during review.

---

**QA Engineer:** Skadi
**Reviewed Files:**
- `api/internal/handler/sequence_tracking.go` (223 lines)
- `api/internal/email/mailer.go` (lines 57-106)
- `api/internal/seqtoken/token.go` (93 lines)
- `api/internal/worker/sequence_worker.go` (line 119)
- `api/migrations/20240101000044_contact_email_opt_out.sql`
- `api/cmd/server/main.go` (lines 209-212)
