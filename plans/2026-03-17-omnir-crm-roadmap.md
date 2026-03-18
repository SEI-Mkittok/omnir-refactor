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

## Phase 2: Core CRM Entities 🚧 IN PROGRESS

Ship the minimum viable CRM: contacts, accounts, deals.

| Task | Owner | Status | Issue |
|------|-------|--------|-------|
| Go API: Contacts, Accounts, Deals CRUD | Backend | 🔄 Todo | OMN-6 |
| React UI: Contacts, Accounts, Deals | Frontend | 🔄 Todo | OMN-7 |
| Activities table + API | Backend | 🔄 Todo | OMN-11 |
| Notes on contacts/deals | Backend | 🔄 Todo | OMN-12 |
| Deal-contact many-to-many | Backend | 🔄 Todo | OMN-13 |
| Multi-tenancy: org_id across all entities | Backend | 🔄 Todo | OMN-14 |

**Exit criteria:** A user can create/read/update/delete contacts, accounts, and deals. Pipeline kanban works.

---

## Phase 3: Auth + User Management

| Task | Notes |
|------|-------|
| Login / JWT auth flow | users table exists, auth endpoints needed |
| Role-based access (admin/user/viewer) | Schema has role column, enforcement needed |
| User management UI | |

---

## Phase 4: Activities + Timeline

| Task | Notes |
|------|-------|
| Activities schema (calls, emails, meetings, tasks) | Biggest current gap |
| Activity CRUD API | |
| Contact/deal timeline UI | |
| Activity reminders | |

---

## Phase 5: Polish + Migration

| Task | Notes |
|------|-------|
| vtiger data migration scripts | Runbook in docs/vtiger-migration-runbook.md |
| Search across all entities | |
| Reports / basic analytics | |
| Mobile-responsive audit | |

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
