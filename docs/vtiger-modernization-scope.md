# Omnir CRM — vtiger CE Audit & Modernization Scope

**Author:** Forge (AI Dev Agent)  
**Date:** 2026-03-17  
**Issue:** OMN-2  
**Status:** Draft v1

---

## Executive Summary

vtiger Community Edition (CE) is a mid-2000s PHP/MySQL CRM with ~750k+ lines of code. It is functional but architecturally brittle: global state, no autoloader until late versions, jQuery UI frontend, and a procedurally-structured backend make it expensive to extend safely. This document identifies what to keep, what to replace, what to eliminate, and the shape of the modernization work ahead.

---

## 1. Core Modules: Keep vs. Replace

### Keep (Business Logic Worth Preserving)
These modules encode domain knowledge and user workflows that should be migrated, not discarded:

| Module | Notes |
|--------|-------|
| **Contacts** | Central entity; relationships, custom fields, history |
| **Accounts (Organizations)** | Company hierarchy, multi-contact support |
| **Leads** | Lead capture, conversion flow → Contacts/Accounts |
| **Opportunities (Potentials)** | Sales pipeline, stages, probability |
| **Activities (Calls/Meetings/Tasks)** | Calendar integration, reminders |
| **Cases (Support Tickets)** | Customer support workflow |
| **Products/Services Catalog** | SKU, pricing, inventory linkage |
| **Invoices / Quotes / POs** | Financial document lifecycle |
| **Campaigns** | Email marketing, prospect lists |
| **Reports** | Custom report builder logic (data model worth keeping) |
| **Workflows** | Rule-based automation triggers |
| **Custom Fields / Field Groups** | Highly flexible per-module schema extension |
| **Role / Profile / Group ACL** | Multi-level access control model |

### Replace (Core Infrastructure Modules)
These provide the plumbing that is tightly coupled to legacy patterns and should be rewritten:

| Module | Problem | Replacement Approach |
|--------|---------|---------------------|
| **vtlib** | Monolithic meta-module framework; global procedural API | Module registry + DI container |
| **Settings** | Flat key-value table, no namespacing | Typed config service with env override |
| **Users** | Mixed auth/profile/ACL concerns | Separate AuthService, UserService, RBAC |
| **Install/Upgrade** | PHP wizard scripts, brittle DB migration | Migration runner (e.g., golang-migrate) |
| **Cron / Scheduled Tasks** | PHP CLI cron manager, no observability | Proper job queue (Redis/DB-backed) |
| **Home / Dashlets** | jQuery widget system, unmaintainable | Modern component-based dashboard |
| **Documents** | File system storage with DB metadata; no versioning | Object storage (S3-compatible) + proper asset service |

### Retire (Remove Entirely)
| Module | Reason |
|--------|-------|
| **vtiger Mobile** | Outdated thin-client; replace with responsive web or native app |
| **vtigerCRM Web Services (legacy SOAP)** | Superseded; expose REST/GraphQL instead |
| **PBX/Asterisk Integration** | Niche, unmaintained; reintroduce as optional plugin |
| **Outlook Plugin** | Windows COM-based; dead technology |
| **SalesPlatform Bridge** | Third-party commercial bridge; drop |
| **Migration Tool (SugarCRM)** | One-time tool; not worth carrying forward |

---

## 2. PHP Legacy Patterns to Eliminate

### Critical / High-Impact
- **Global variables and `$_REQUEST` everywhere** — no input abstraction layer; XSS/CSRF surface area is enormous
- **`include`/`require` chains** instead of autoloading — spaghetti dependency graph
- **Mixed HTML in PHP logic** — controllers directly echo HTML; zero separation of concerns
- **Database calls inline in views** — `mysql_query()` / `mysqli_query()` scattered throughout templates
- **No type hints, no return types** — PHP 5.x-era code running on PHP 8.x with suppressed warnings
- **Procedurally-namespaced functions** — `vtlib_*`, `getEntityName()`, etc. scattered across 50+ utility files
- **God objects** — `Vtiger_Module_Model`, `Vtiger_Record_Model` doing far too much
- **`eval()` usage** — found in workflow and template engines; security liability

### Medium-Impact
- **`die()` / `exit()` in library code** — breaks testability
- **Session variables as application state** — e.g., `$_SESSION['authenticated_user_id']` checked everywhere
- **Hardcoded SQL strings** — no query builder, no parameterized queries in older modules
- **Config via PHP includes** — `config.inc.php` set as a global; not injectable
- **Error suppression (`@`)** — masks real failures throughout

### Pattern Replacements
| Legacy Pattern | Modern Replacement |
|---------------|-------------------|
| Raw `mysqli_*` calls | Repository pattern over PDO/ORM |
| `include 'header.php'` layout system | Template engine (Twig, or Go html/template) |
| Global `$adb` database object | Injected `DatabaseConnection` |
| `CRMEntity` base class with AR pattern | Domain models + separate persistence layer |
| Procedural ACL checks | Middleware / policy objects |

---

## 3. Database Schema Quality & Migration Needs

### Current State
- **~200+ tables** in a flat MySQL schema
- Table naming: `vtiger_*` prefix convention (consistent, good)
- Heavy use of **Entity-Value (EAV) pattern** for custom fields → kills query performance at scale
- `vtiger_crmentity` as the central entity registry (polymorphic parent table) — a necessary but painful design
- Soft deletes via `deleted` column on most entity tables (inconsistently applied)
- No foreign key constraints enforced — data integrity relies entirely on application logic
- Mixed character set: some tables latin1, some utf8, some utf8mb4 — causes corruption with emoji/non-ASCII

### High-Priority Schema Issues
| Issue | Tables Affected | Action |
|-------|----------------|--------|
| EAV custom fields (`vtiger_*cf` tables) | All entity modules | Evaluate JSONB column for custom fields (PostgreSQL preferred) |
| Missing indexes on foreign key columns | ~40+ tables | Audit + add indexes |
| `vtiger_crmentity` join on every query | All entities | Materialize critical fields, cache entity type lookups |
| Mixed charset | ~30 tables | Migrate to utf8mb4 uniformly |
| No FK constraints | All relations | Add FK constraints in migration or enforce in app layer |
| Audit log in main tables (`modifiedtime`) | All | Separate audit/history table |

### Migration Strategy
1. **Keep MySQL/MariaDB as initial target** (lowest friction for existing deployments)
2. **Add PostgreSQL support** in the new stack (better JSON, better concurrency, FK enforcement)
3. **Write all migrations as version-numbered files** using a migration runner (e.g., `golang-migrate`)
4. **Custom fields**: Move from EAV tables to JSONB columns with indexed paths
5. **Phase 1**: Schema cleanup (charset, indexes, constraints) while keeping vtiger-compatible structure
6. **Phase 2**: Schema reshape for new data model (post-Go rewrite)

---

## 4. API Surface Area

### Existing APIs in vtiger CE
| API | Type | Notes |
|-----|------|-------|
| `/webservice.php` | Legacy REST-ish | Custom challenge/response auth; no OpenAPI spec |
| SOAP endpoint | SOAP/WSDL | Dead; no clients use it |
| Module-specific AJAX | Ad-hoc JSON | `index.php?module=X&action=Save&...` pattern |
| Import/Export CSV | File-based | No programmatic control |

### Problems with Current API
- Auth uses MD5 challenge-response; no JWT/OAuth support
- No versioning
- No pagination standards
- Responses inconsistently structured (sometimes `{"success":true,"result":{...}}`, sometimes raw arrays)
- No rate limiting
- CSRF protection is minimal

### Target API Design (New Stack)
- **REST API** — versioned (`/api/v1/`), OpenAPI 3.1 spec
- **Auth**: JWT (short-lived) + refresh tokens; API key support for integrations
- **Pagination**: cursor-based for large collections
- **Webhooks**: outbound event hooks for integrations (Zapier, Make, etc.)
- **GraphQL** (phase 2): for flexible querying by frontend and third-party integrations

---

## 5. jQuery / Legacy Frontend Inventory

### Current Frontend Stack
| Technology | Version (approx.) | Usage |
|-----------|-----------------|-------|
| jQuery | 1.x–2.x | Core DOM/AJAX everywhere |
| jQuery UI | 1.x | Dialogs, datepickers, sortables |
| Smarty | 2.x | PHP template engine (partial use) |
| Handlebars.js | 1.x | Some dynamic templates |
| FullCalendar | 2.x | Activities / calendar view |
| DataTables | 1.x | List views |
| Select2 | 3.x | Dropdowns |
| Bootstrap | 2.x–3.x | Layout in some views |
| Custom CSS | None | Flat CSS files, no preprocessor |

### Frontend Assessment
- **No build system** — JS files concatenated manually or loaded individually; 200+ `<script>` tags in some views
- **No module system** — global namespace pollution via `window.*` variables
- **No TypeScript** — no type safety at all
- **No component architecture** — UI logic mixed with AJAX calls mixed with business logic
- **Accessibility**: essentially zero ARIA, no keyboard nav patterns
- **Mobile**: not truly responsive; vtiger Mobile was a separate app

### Frontend Modernization Plan
| Phase | Scope |
|-------|-------|
| **Phase 1 (Quick Win)** | Replace jQuery UI with modern CSS + minimal Alpine.js or HTMX for dynamic behavior on top of a new clean HTML/CSS base |
| **Phase 2 (Foundation)** | Introduce a proper build system (Vite); migrate to TypeScript; build a component library |
| **Phase 3 (Full Rewrite)** | React or SvelteKit SPA consuming the new REST/GraphQL API |

**Recommended stack for Phase 3:**
- **Framework**: React 18+ or SvelteKit (SvelteKit preferred for lower bundle size and simpler mental model)
- **UI library**: shadcn/ui (Radix primitives) or Melt UI (Svelte)
- **State**: Zustand (React) or Svelte stores
- **Data fetching**: TanStack Query
- **Forms**: React Hook Form / Superforms
- **Testing**: Vitest + Playwright

---

## 6. Recommended Modernization Roadmap

### Option A: Incremental (Lower Risk, Longer Runway)
1. **Wrap**: Add a proper routing layer (Slim or Laravel) in front of existing vtiger, modernize auth
2. **Strangle**: Replace modules one at a time via the strangler fig pattern
3. **Frontend**: Introduce Vite + React alongside existing jQuery
4. **Timeline**: 12–18 months to reach 80% modern

### Option B: Full Rewrite in Go + Modern Frontend (Recommended)
1. **Backend**: Go (net/http + chi router, or Echo), PostgreSQL, repository pattern, OpenAPI spec first
2. **Frontend**: SvelteKit SPA
3. **Data**: Migrate existing MySQL data via migration scripts; reshape schema
4. **Deploy**: Docker + docker-compose first; Kubernetes-ready
5. **Timeline**: 6–9 months for MVP (core modules: Contacts, Accounts, Leads, Opportunities, Activities, basic Reports)

**Why Go:**
- Strong typing catches class of bugs that PHP silently ignores
- Excellent concurrency model for background jobs, webhooks, sync
- Single binary deployment
- Fast compile times → good DX
- Strong ecosystem for HTTP, DB, testing

### Immediate Next Steps (Sprint 1)
1. Stand up Go skeleton with chi router, PostgreSQL, JWT auth, and OpenAPI scaffolding
2. Define core domain models (Contact, Account, Lead, Opportunity, Activity)
3. Write initial DB migration set (Phase 1 schema)
4. Build Contacts CRUD as the first vertical slice (end-to-end: API → DB → UI)
5. Set up CI/CD pipeline (GitHub Actions, Docker builds)

---

## 7. Risks & Open Questions

| Risk | Severity | Mitigation |
|------|---------|-----------|
| Data migration from existing vtiger installs | High | Early investment in migration tooling; offer managed migration service |
| Missing feature parity (user expectations) | High | Prioritize most-used modules first; use feature flags |
| Custom fields (EAV) migration complexity | Medium | Design JSONB schema early; migration script before Phase 2 |
| Third-party integrations (email, calendar, etc.) | Medium | Design webhook/event system early to unblock integrations |
| Team ramp-up on Go | Low | Go is approachable; strong stdlib; worth the investment |

---

## Appendix: vtiger CE File/Directory Structure

```
vtiger_root/
├── index.php               # Main entry point (all requests route through here)
├── config.inc.php          # Global config (DB creds, site URL, etc.)
├── vtigercrm.php           # Bootstrap
├── modules/                # ~80+ modules, each with own MVC-ish structure
│   ├── Contacts/
│   ├── Accounts/
│   └── ...
├── vtlib/                  # Core framework library
│   ├── Vtiger/             # Module, Field, Relationship meta-APIs
│   └── ...
├── include/                # Shared utilities, DB layer, auth, etc.
│   ├── database/           # ADOdb wrapper
│   ├── utils/              # Utility functions
│   └── ...
├── layouts/                # Smarty templates + CSS/JS assets
│   └── v7/
│       ├── modules/        # Per-module view templates
│       ├── resources/      # JS, CSS
│       └── ...
├── cron/                   # Scheduled job scripts
├── tests/                  # Minimal test coverage (mostly integration)
└── storage/                # Uploaded files, logs, temp
```

---

*Produced by Forge for Omnir CRM. This is a living document — update as implementation progresses.*
