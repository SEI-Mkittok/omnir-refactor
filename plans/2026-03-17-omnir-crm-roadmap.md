# Omnir CRM — Rolling Roadmap

_Maintained by Völundr (CTO). Updated each phase scoping._

---

## Phase 0–11: Shipped ✅

| Phase | Features |
|-------|----------|
| 0 | Foundation: repo, CI, Docker, auth, org-scoping |
| 1a | Help Desk: tickets, comments, attachments, email inbound |
| 1b | CRM: contacts, accounts, leads, lead conversion |
| 2 | Multi-tenancy hardening (RLS, org onboarding) |
| 3 | Client portal, email notifications |
| 4 | SLA timers, custom fields, reporting, API keys |
| 5 | vTiger data migration scripts + cutover runbook |
| 6 | Deals / pipeline (Kanban + list), activities, notes |
| 7 | Help desk enhancements: SLA v2, canned responses, tags |
| 8 | Client portal v2, integrations hub |
| 9 | Products catalog, invoicing, payment tracking |
| 10 | Documents, products v2, email sequences, advanced reporting |
| 11 | Quotes/proposals (PDF), workflow automation, calendar integration |

---

## Phase 12 — Mobile PWA · Knowledge Base · SSO/2FA

_Scoped: 2026-03-21 | Scoped by: Völundr_

### Rationale

Phase 11 shipped Quotes, Workflow Automation, and Calendar Integration. With the core CRM workflow loop closed, Phase 12 addresses three high-value capability gaps:

1. **Mobile PWA** — most-requested feature, explicitly deferred from Phase 10. Enables field sales and support agents to use Omnir on mobile without a native app.
2. **Knowledge Base / Help Center** — natural extension of the client portal. Deflects repeat tickets, reduces support load, increases portal value.
3. **SSO / 2FA** — enterprise unlock. OIDC/SAML login and TOTP remove the last blockers for enterprise sales.

### Value/Effort Ranking

| Area | Value | Effort | Priority |
|------|-------|--------|----------|
| Mobile PWA | ⭐⭐⭐⭐⭐ | High | P1 — most-requested, deferred twice |
| Knowledge Base | ⭐⭐⭐⭐ | Medium | P2 — natural portal complement, ticket deflection ROI |
| SSO / 2FA | ⭐⭐⭐⭐⭐ | Medium | P3 — enterprise gate, unlocks commercial tier |

**Deferred to Phase 13:** Advanced email (inbox integration), Contact enrichment. These are valuable but secondary once mobile + auth are solid.

---

### Area 1: Mobile PWA

**Owner:** Freya (frontend), Saga (UX design)  
**Issues:** OMN-436, OMN-437, OMN-438

**Scope:**
- Service worker + app manifest (offline cache, installable)
- Offline read mode for tickets, contacts, deals
- Mobile-optimized layouts: bottom nav, responsive list/detail views
- Saga owns mobile design system: touch targets, gestures, navigation patterns

**Key decisions:**
- No React Native — PWA is sufficient for the use case and avoids a second codebase
- Offline writes: queue mutations, sync on reconnect (IndexedDB + service worker background sync)
- Push notifications via Web Push API (VAPID keys)

---

### Area 2: Knowledge Base / Self-Service Help Center

**Owner:** Tyr (backend), Freya (frontend)  
**Issues:** OMN-439, OMN-440

**Scope:**
- `articles` table: title, body (markdown), category, tags, org_id, published/draft
- Full-text search via PostgreSQL `tsvector`
- Admin UI: article editor (markdown), category management, publish/draft toggle
- Public help center portal: browsable + searchable, no login required
- Ticket deflection: surface article suggestions when ticket subject is typed

**Key decisions:**
- Markdown storage, rendered client-side (no WYSIWYG server-side)
- Public portal served under `/help` — same domain, no subdomain complexity
- Article views tracked (anonymous count) for popularity ranking

---

### Area 3: SSO / 2FA

**Owner:** Tyr (backend), Freya (frontend), Heimdall (ops)  
**Issues:** OMN-441, OMN-442, OMN-443

**Scope:**
- OIDC login: Google Workspace + generic OIDC provider (Okta, Entra ID)
- SAML 2.0 SP-initiated login for enterprise orgs
- TOTP 2FA: TOTP secret per user, QR code enrollment, backup codes
- Org-level SSO config: admin enables SSO, maps OIDC claims → roles
- Heimdall: secrets management for OIDC client credentials, staging + prod env setup

**Key decisions:**
- OIDC first, SAML behind a feature flag (SAML is complex, low short-term demand)
- TOTP uses `pquerna/otp` library (Go) — no external dependency
- Backup codes: 8 single-use codes generated on 2FA enrollment
- 2FA enforcement: per-org policy (optional vs required)

---

## Phase 13 — Advanced Email · Integrations Marketplace · Contact Enrichment

_Scoped: 2026-03-21 | Scoped by: Völundr_

### Phase 12 Delivery Review

| Issue | Area | Status |
|-------|------|--------|
| OMN-436 | Mobile PWA (service worker, offline) | in_review |
| OMN-437 | Mobile PWA UX (design system) | done ✅ |
| OMN-438 | Mobile PWA (push notifications) | in_review |
| OMN-439 | Knowledge Base backend | todo |
| OMN-440 | Knowledge Base frontend | in_progress |
| OMN-441 | SSO backend (OIDC + TOTP) | todo |
| OMN-442 | SSO frontend | todo |
| OMN-443 | DevOps (VAPID keys, SSO secrets) | done ✅ |

Phase 12 is active — Mobile PWA areas are in final CTO review, KB and SSO are still being built. Phase 13 scoping proceeds now so engineers can transition immediately once Phase 12 closes.

---

### Rationale

Phase 12 closes the enterprise auth and mobile gaps. Phase 13 targets the next commercial tier unlock: deep email integration (turns Omnir into a daily-driver communication hub), an integrations marketplace (unlocks Zapier/Slack/n8n workflows the sales team already uses), and contact enrichment (reduces data entry friction on every new contact).

### Value/Effort Ranking

| Area | Value | Effort | Priority |
|------|-------|--------|----------|
| Advanced email inbox integration | ⭐⭐⭐⭐⭐ | High | P1 — daily-driver unlock, Gmail/Outlook are table stakes |
| Integrations marketplace | ⭐⭐⭐⭐⭐ | Medium | P2 — Zapier/Slack are blocking enterprise adoption |
| Contact enrichment | ⭐⭐⭐⭐ | Medium | P3 — reduces friction, complements email integration |

**Deferred:** Advanced mobile (background sync v2, native share targets) — core PWA ships in Phase 12; v2 enhancements are low urgency post-launch.

---

### Area 1: Advanced Email Inbox Integration

**Owner:** Tyr (backend), Freya (frontend)
**Issues:** OMN-446, OMN-447

**Scope:**
- `email_connections` table: user_id, provider (gmail/outlook), oauth_tokens, sync_cursor, last_synced_at
- Gmail OAuth2 (Google APIs) + Outlook OAuth (Microsoft Graph) inbox sync
- Background sync job: poll every 5min, thread emails to contacts by address match
- Reply tracking: detect when a contact replies to a sent email (In-Reply-To header matching)
- Template library: `email_templates` table (name, subject_template, body_html_template, org_id), CRUD API + UI
- Compose from template in email modal, variable substitution ({{contact.name}}, {{sender.name}})

**Key decisions:**
- OAuth tokens encrypted at rest (AES-256, same key as SSO_ENCRYPTION_KEY)
- Inbox is read-only pull model — no SMTP send path changes
- Thread grouping: same algorithm as existing inbound email (In-Reply-To + References headers)
- Template variables: mustache-style, rendered server-side before send

---

### Area 2: Integrations Marketplace (Zapier/n8n + Slack)

**Owner:** Tyr (backend), Freya (frontend), Heimdall (ops)
**Issues:** OMN-448, OMN-449

**Scope:**
- Enhanced webhook triggers (Phase 6 extension): add ticket.*, note.*, activity.* events; add retry logic with exponential backoff + dead-letter log
- Zapier app: publish public Zapier integration pointing at Omnir webhook + REST API (contact read, deal read, create) — no extra backend work, just Zapier developer portal setup
- n8n: document Omnir nodes using existing REST API — create `docs/integrations/n8n.md`
- Slack integration: `slack_connections` table (org_id, team_id, webhook_url, bot_token); `POST /api/integrations/slack/connect` (OAuth); notification routing: deal.stage_changed → Slack channel; mention/DM on deal assigned
- UI: Settings > Integrations hub — list available integrations (Zapier, n8n, Slack, Webhooks), connect/disconnect, configure

**Key decisions:**
- Zapier + n8n don't require backend changes beyond what exists — they use the REST API + webhooks
- Slack is the only new OAuth integration needing backend work
- Webhook retry: 3 attempts with exponential backoff, store delivery log in `webhook_deliveries` table

---

### Area 3: Contact Enrichment

**Owner:** Tyr (backend), Freya (frontend)
**Issues:** OMN-450, OMN-451

**Scope:**
- `enrichment_cache` table: domain, data JSONB (company name, industry, size, logo_url, linkedin_url), fetched_at (TTL 30 days)
- On contact create/edit with email: extract domain, check cache, if miss → query Clearbit Enrichment API (or fallback: basic whois/DNS lookup for self-hosted)
- Auto-fill suggestion bar: when email is typed in contact form, show "We found info for acme.com — fill in?" banner
- Manual trigger: "Enrich contact" button on contact detail page
- Enrichment sources: Clearbit (SaaS, API key in env), fallback to Hunter.io suggest API, final fallback to domain DNS lookup for basic info
- UI: enrichment status badge on contact (enriched/not enriched/failed), one-click re-enrich

**Key decisions:**
- Enrichment is non-blocking (async): contact saves immediately, enrichment runs in background
- All external API keys optional — feature degrades gracefully if no key configured
- Cache at domain level, not per-contact (50 contacts @acme.com = 1 lookup)
- Self-hosted: basic enrichment from public DNS/WHOIS (no external API required)
