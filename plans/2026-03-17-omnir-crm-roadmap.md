# Omnir CRM Roadmap

_Created: 2026-03-17 | Author: CEO + Völundr (CTO)_

---

## Phase 1: Foundation ✅ COMPLETE

Everything needed to start building features.

| Task | Owner | Status |
|------|-------|--------|
| vtiger codebase audit (OMN-2) | Völundr (CTO) | ✅ Done |
| Go API architecture design (OMN-3) | Backend | ✅ Done |
| React frontend architecture (OMN-4) | Frontend | ✅ Done |
| Docker + CI/CD pipeline (OMN-5) | DevOps | ✅ Done |
| Data migration strategy (OMN-8) | Backend | ✅ Done |
| QA testing strategy (OMN-9) | QA | ✅ Done |
| Data model mapping (OMN-10) | Völundr (CTO) | ✅ Done |

**Output:** Architecture docs, migration runbook, QA strategy, data model map, working dev environment.

---

## Phase 2: Core CRM Entities ✅ COMPLETE

Ship the minimum viable CRM: contacts, accounts, deals.

| Task | Owner | Status | Issue |
|------|-------|--------|-------|
| Go API: Contacts, Accounts, Deals CRUD | Backend | ✅ Done | OMN-6 |
| React UI: Contacts, Accounts, Deals | Frontend | ✅ Done | OMN-7 |
| Activities table + API | Backend | ✅ Done | OMN-11 |
| Notes on contacts/deals | Backend | ✅ Done | OMN-12 |
| Deal-contact many-to-many | Backend | ✅ Done | OMN-13 |
| Multi-tenancy: org_id across all entities | Backend | ✅ Done | OMN-14 |

**Exit criteria:** A user can create/read/update/delete contacts, accounts, and deals. Pipeline kanban works.

---

## Phase 3: Auth + User Management ✅ COMPLETE

| Task | Notes |
|------|-------|
| Login / JWT auth flow | ✅ Done — OMN-60 |
| Role-based access (admin/user/viewer) | ✅ Done — OMN-63 |
| User management UI | ✅ Done — OMN-64 |

---

## Phase 4: Activities + Timeline ✅ COMPLETE

| Task | Notes |
|------|-------|
| Activities schema (calls, emails, meetings, tasks) | ✅ Done — OMN-11 |
| Activity CRUD API | ✅ Done |
| Contact/deal timeline UI | ✅ Done — OMN-77 |
| Activity reminders | ✅ Done — OMN-78 |

---

## Phase 5: Polish + Migration ✅ COMPLETE

| Task | Notes |
|------|-------|
| vtiger data migration scripts | ✅ Done — OMN-123 |
| Search across all entities | ✅ Done |
| Reports / basic analytics | ✅ Done |
| Mobile-responsive audit | 🔄 In Progress — OMN-259 (Freya) |
| DB performance review | ✅ Done — OMN-260 (Tyr, in_review) |

---

---

## Phase 6: Integrations & Data Portability 🔜 PLANNED

_Defined: 2026-03-18 | Scoped by: Völundr (CTO) | Parent issue: OMN-258_

**Selected areas (ranked by value/cost):**
1. **Bulk import/export** — critical for vtiger migration completion, medium cost, high immediate value
2. **Email integration** — core CRM differentiator, high cost, high strategic value
3. **Webhooks + integrations** — ecosystem enabler, medium cost, enables Zapier/automation use cases

_Deferred: Advanced analytics (high cost, low immediate need), Mobile-responsive audit (lower strategic ROI)_

### Area 1: Bulk Import/Export

| Task | Owner | Issue |
|------|-------|-------|
| CSV import API (contacts, accounts, leads) | Tyr | OMN-261 |
| CSV export API (contacts, accounts, deals, reports) | Tyr | OMN-262 |
| Bulk import/export UI | Freya | OMN-263 |

### Area 2: Email Integration

| Task | Owner | Issue |
|------|-------|-------|
| Email data model + SMTP outbound API | Tyr | OMN-264 |
| Inbound email parsing + storage | Tyr | OMN-265 |
| Email compose + timeline UI | Freya | OMN-266 |

### Area 3: Webhooks + Integrations

| Task | Owner | Issue |
|------|-------|-------|
| Webhook registration + delivery API | Tyr | OMN-267 |
| Webhook event emission on CRM mutations | Tyr | OMN-268 |
| Webhook management UI | Freya | OMN-269 |

### QA

| Task | Owner | Issue |
|------|-------|-------|
| Phase 6 QA: import/export, email, webhooks test coverage | Skadi | OMN-271 |

### DevOps

| Task | Owner | Issue |
|------|-------|-------|
| Staging: add client test user + configure MailHog SMTP | Heimdall | OMN-249 |

**Exit criteria:** Users can import/export data via CSV, send/receive emails from contact pages, and configure outbound webhooks for automation triggers.

---

## Open Questions / Decisions Needed

1. **Multi-tenancy:** Add `org_id` now or wait? If targeting SaaS, add it in Phase 2.
2. **Email integration:** Send/receive emails from contacts? (Big scope, Phase 4+)
3. **Custom fields UI:** How does a user define custom fields? (Admin panel needed)
4. **Self-hosted vs SaaS:** Affects auth design, multi-tenancy, billing.

---

## Next Actions

- [x] Created OMN-11: Activities table + API
- [x] Created OMN-12: Notes on contacts/deals
- [x] Created OMN-13: Deal-contact many-to-many
- [x] Created OMN-14: Multi-tenancy (org_id) — SaaS + self-hosted
- [ ] OMN-14 needs frontend coordination before merge (OMN-7 in-flight)
