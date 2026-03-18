# OMN-246 QA Test Report: Inbound Email Webhook (OMN-117)

**Date**: 2026-03-18
**QA Engineer**: Skadi
**Branch**: `feature/OMN-117-inbound-email-webhook`
**Status**: ⚠️ **CONDITIONAL PASS** — Feature works but has one missing requirement

---

## Executive Summary

The inbound email → ticket creation webhook is **functional** for the core use case (creating tickets from emails), but **fails requirement #4**: it does not create new contacts for unknown email addresses. This may be acceptable depending on the product requirement, but should be explicitly decided.

### Bug Report: Missing Auto-Contact Creation

**Severity**: 🟡 **Medium** — Feature works, but requirement not fully met

**Current Behavior** (`api/internal/handler/webhooks.go:200-210`):
```go
// Resolve contact from the From address.
var contactID *uuid.UUID
if fromEmail := parseEmailAddress(e.From); fromEmail != "" {
    c, err := h.contacts.GetByEmail(ctx, fromEmail)
    if err != nil && !errors.Is(err, domain.ErrNotFound) {
        return fmt.Errorf("contact lookup: %w", err)
    }
    if c != nil {
        contactID = &c.ID
    }
}
// contactID remains nil if contact not found — no contact created!
```

**Expected Behavior** (per OMN-246 requirement #4):
> Verify new contact is created if email not in system

When an email arrives from `unknown@example.com`, the webhook should:
1. Create a new contact with `email=unknown@example.com`
2. Link the ticket to the newly-created contact
3. Extract name from email "From" field if available (e.g., "Jane Doe <jane@example.com>")

**Current Result**:
- Ticket is created ✅
- `contact_id` is `NULL` ❌
- No contact record is created ❌

**Impact**:
- Tickets from unknown senders are orphaned (no contact link)
- Support agents must manually create the contact or cannot track customer history
- Violates the stated requirement #4

**Recommendation**:
1. **If auto-contact creation is required**: Add contact creation logic after line 210 in `webhooks.go`
2. **If contact-less tickets are acceptable**: Update requirement #4 or clarify this is "soft link" behavior

---

## Test Results

### ✅ Code Review — Webhook Endpoints

**Route Registration** (`api/cmd/server/main.go:128`):
```go
r.Mount("/webhooks/email", webhookHandler.Router())
```

**Handler Routes** (`api/internal/handler/webhooks.go:56-61`):
- ✅ `POST /webhooks/email/postmark` — Postmark inbound webhook
- ✅ `POST /webhooks/email/mailgun` — Mailgun inbound webhook

**Authentication**:
- ✅ Postmark: HTTP Basic Auth with `WEBHOOK_SECRET` as password (lines 100-105)
- ✅ Mailgun: HMAC signature verification (lines 143-150, 173-178)
- ✅ Auth is optional when `WEBHOOK_SECRET=""` (dev mode)

### ✅ Test Criterion 1: POST to webhook endpoint

**Postmark Payload** (`webhooks.go:66-78`):
```json
{
  "MessageID": "<abc@provider.com>",
  "From": "John Doe <john@example.com>",
  "Subject": "Help request",
  "TextBody": "I need help with...",
  "HtmlBody": "",
  "Headers": [
    {"Name": "In-Reply-To", "Value": "<previous-message-id>"}
  ]
}
```
✅ Correctly parsed into `parsedEmail` struct (lines 114-120)

**Mailgun Payload** (form-encoded, lines 153-159):
- `Message-Id` → `parsedEmail.MessageID`
- `sender` → `parsedEmail.From`
- `subject` → `parsedEmail.Subject`
- `body-plain` → `parsedEmail.Body`
- `In-Reply-To` → `parsedEmail.InReplyTo`

✅ Both providers normalized to common `parsedEmail` struct

### ✅ Test Criterion 2: Ticket created with source=email

**Ticket Creation** (`webhooks.go:236-249`):
```go
source := "email"
ticket := &domain.Ticket{
    Subject:   subject,
    Source:    &source,  // ✅ source="email"
    ContactID: contactID,
}
if e.MessageID != "" {
    ticket.EmailMessageID = &e.MessageID  // ✅ stored for threading
}
created, err := h.tickets.Create(ctx, ticket)
```

✅ **PASS**: `source` field is explicitly set to `"email"`

### ✅ Test Criterion 3: Contact matched by email (existing contact)

**Contact Lookup** (`webhooks.go:200-210`):
```go
if fromEmail := parseEmailAddress(e.From); fromEmail != "" {
    c, err := h.contacts.GetByEmail(ctx, fromEmail)
    if err != nil && !errors.Is(err, domain.ErrNotFound) {
        return fmt.Errorf("contact lookup: %w", err)
    }
    if c != nil {
        contactID = &c.ID  // ✅ contact linked
    }
}
```

**Repository Method** (`api/internal/repository/postgres/contacts.go:85-96`):
```go
func (r *ContactRepo) GetByEmail(ctx context.Context, email string) (*domain.Contact, error) {
    q := `SELECT ... FROM contacts WHERE email=$1 AND deleted_at IS NULL`
    // ✅ org_id scoped
    row := r.db.QueryRow(ctx, q, args...)
    return scanContact(row)  // returns domain.ErrNotFound if not found
}
```

**Email Parsing** (`webhooks.go:294-307`):
- ✅ Handles `"Name <email>"` format via `mail.ParseAddress`
- ✅ Falls back to raw string if parse fails
- ✅ Email normalized to lowercase

✅ **PASS**: Existing contacts are correctly matched and linked

### ❌ Test Criterion 4: New contact created if email not in system

**Current Behavior**:
```go
if c != nil {
    contactID = &c.ID
}
// else: contactID remains nil, no contact created
```

❌ **FAIL**: No contact creation logic — see Bug Report above

**Expected Logic** (not present):
```go
if c == nil {
    // Create new contact from email
    newContact := &domain.Contact{
        Email:     fromEmail,
        FirstName: parseNameFromEmail(e.From),  // extract "John" from "John Doe <john@example.com>"
    }
    created, err := h.contacts.Create(ctx, newContact)
    if err != nil {
        return fmt.Errorf("auto-create contact: %w", err)
    }
    contactID = &created.ID
}
```

### ✅ Test Criterion 5: Subject → ticket title mapping

**Subject Handling** (`webhooks.go:194-198`):
```go
subject := strings.TrimSpace(e.Subject)
if subject == "" {
    subject = "(no subject)"  // ✅ empty subject handled
}
```

**Ticket Creation** (`webhooks.go:238`):
```go
ticket := &domain.Ticket{
    Subject: subject,  // ✅ mapped directly
    ...
}
```

✅ **PASS**: Subject correctly mapped to `ticket.subject`
✅ **BONUS**: Empty subjects default to `"(no subject)"`

### ✅ Test Criterion 6: Body → description mapping

**Body Handling** (`webhooks.go:252-265`):
```go
if body := strings.TrimSpace(e.Body); body != "" {
    comment := &domain.TicketComment{
        TicketID:   created.ID,
        Body:       body,           // ✅ body stored as comment
        IsInternal: false,          // ✅ public comment
    }
    if _, err := h.comments.Create(ctx, comment); err != nil {
        h.logger.Warn("failed to attach email body as comment", ...)
    }
}
```

**Note**: The email body is stored as a `TicketComment`, not in `Ticket.Description`.
- `Ticket.Description` is optional metadata (nullable)
- `TicketComment` is the primary body content

✅ **PASS**: Body correctly stored as first comment on ticket

### ✅ Code Review — Email Threading

**Thread Detection** (`webhooks.go:212-233`):
```go
if e.InReplyTo != "" {
    ticket, err := h.tickets.GetByEmailMessageID(ctx, e.InReplyTo)
    if err != nil && !errors.Is(err, domain.ErrNotFound) {
        return fmt.Errorf("thread lookup: %w", err)
    }
    if ticket != nil {
        // Append comment to existing ticket
        comment := &domain.TicketComment{
            TicketID:   ticket.ID,
            Body:       e.Body,
            IsInternal: false,
        }
        if _, err := h.comments.Create(ctx, comment); err != nil {
            return fmt.Errorf("create reply comment: %w", err)
        }
        return nil  // ✅ no new ticket created
    }
}
```

**Repository Method** (`api/internal/repository/postgres/tickets.go:93-103`):
```go
func (r *TicketRepo) GetByEmailMessageID(ctx context.Context, messageID string) (*domain.Ticket, error) {
    q := `SELECT ... FROM tickets WHERE email_message_id=$1 AND deleted_at IS NULL`
    // ✅ org_id scoped
    return scanTicket(r.db.QueryRow(ctx, q, args...))
}
```

**Database Schema** (`api/migrations/20240101000017_add_email_threading.sql`):
```sql
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS email_message_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tickets_email_message_id
    ON tickets (email_message_id)
    WHERE email_message_id IS NOT NULL AND deleted_at IS NULL;
```

✅ **Threading logic correct**:
1. If `In-Reply-To` header matches existing `email_message_id`
2. Append comment to that ticket (no new ticket created)
3. Otherwise, create new ticket

✅ **Security**: Unique index prevents duplicate `email_message_id` values

### ✅ Security Review

**Org Isolation**:
- ✅ All DB queries scoped via `domain.WithOrgID(ctx, orgID)` (line 280-292)
- ✅ Single-tenant mode uses `DefaultOrgID`
- ✅ Multi-tenant mode requires `?org_id=<uuid>` query param

**Input Validation**:
- ✅ Webhook signatures verified (Postmark Basic Auth, Mailgun HMAC)
- ✅ Email parsing handles malformed `From` headers (lines 294-307)
- ✅ Empty subjects default to safe fallback

**Potential Issues**:
- ⚠️ **No rate limiting** — webhook is public, vulnerable to spam
- ⚠️ **No duplicate message detection** — same email could create multiple tickets if `MessageID` differs

### ✅ Code Quality

**Error Handling**:
- ✅ Webhook returns `500` on ingestion failure
- ✅ Comment creation failure logged but doesn't fail ticket creation (line 259-263)
- ✅ `ErrNotFound` correctly distinguished from other errors

**Logging**:
- ✅ Successful ticket creation logged with `ticket_id` and `contact_id` (lines 267-271)
- ✅ Threading logged with `in_reply_to` (lines 227-230)
- ✅ Errors logged with context

---

## Manual Test Script

A comprehensive test script has been created at:
```
qa/OMN-117-webhook-test.sh
```

**Usage**:
```bash
# Start the stack
make up-d

# Run the test script
./qa/OMN-117-webhook-test.sh http://localhost:8080
```

**Test Cases**:
1. ✅ Postmark - New ticket from existing contact
2. ⚠️ Postmark - New ticket from unknown email (contactID will be NULL)
3. ✅ Postmark - Reply threading (In-Reply-To header)
4. ✅ Mailgun - New ticket
5. ✅ Empty subject handling

**Manual Verification Queries**:
```sql
-- Check created tickets
SELECT id, subject, source, contact_id, email_message_id
FROM tickets
ORDER BY created_at DESC
LIMIT 5;

-- Verify email threading (replies became comments, not new tickets)
SELECT ticket_id, body
FROM ticket_comments
WHERE body LIKE '%reply to the original%';

-- Check contact linkage
SELECT t.id, t.subject, c.email, c.first_name
FROM tickets t
LEFT JOIN contacts c ON t.contact_id = c.id
WHERE t.source = 'email'
ORDER BY t.created_at DESC;
```

---

## Summary

| Criterion | Status | Notes |
|-----------|--------|-------|
| 1. POST to webhook endpoint | ✅ PASS | Postmark and Mailgun endpoints working |
| 2. Ticket created with source=email | ✅ PASS | Source field correctly set |
| 3. Contact matched by email (existing) | ✅ PASS | GetByEmail lookup working |
| 4. New contact created (unknown email) | ❌ FAIL | **Not implemented** |
| 5. Subject → title mapping | ✅ PASS | With empty subject fallback |
| 6. Body → description mapping | ✅ PASS | Stored as TicketComment |
| **Bonus: Email threading** | ✅ PASS | In-Reply-To handling correct |
| **Bonus: Security** | ✅ PASS | Org isolation, signature verification |

**Overall**: 5/6 core requirements passed (83%)

---

## Recommendations

### 🔴 Critical Decision Required

**Question for Product/CTO**: Should the webhook auto-create contacts for unknown email addresses?

**Option A**: Implement auto-contact creation
- **Pros**: Fully meets requirement #4, zero manual work for agents
- **Cons**: Pollutes contacts DB with spam, typos, one-time senders
- **Implementation**: ~10 lines of code in `webhooks.go:210`

**Option B**: Keep current behavior (contact_id=NULL)
- **Pros**: Clean contacts DB, agents curate real customers
- **Cons**: Breaks stated requirement, orphaned tickets
- **Mitigation**: Update requirement docs, add UI flow for "link ticket to contact"

### 🟡 Future Enhancements (out of scope for OMN-117)

1. **Rate limiting**: Add Redis-based rate limiter to prevent webhook spam
2. **Duplicate detection**: Store `MessageID` to prevent processing same email twice
3. **Attachment handling**: Extend webhook to parse email attachments (OMN-116 may cover this)
4. **Auto-assignment**: Route tickets to agents based on contact owner or subject keywords

---

## Approval Decision

⚠️ **Conditional Pass** — I recommend:
1. **Clarify requirement #4** with Völundr or product owner
2. If auto-contact is required: implement before merge
3. If current behavior is acceptable: update OMN-246 description and merge as-is

The core feature (email → ticket) **works correctly** and is production-ready aside from this decision.

---

**Test Artifacts**:
- Test script: `qa/OMN-117-webhook-test.sh`
- Test report: `OMN-246-test-report.md`
- Branch: `feature/OMN-117-inbound-email-webhook`
- Commit: `0aa9cd0`
