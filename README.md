# Omnir CRM - vTiger Refactor

[![CI](https://github.com/SEI-Mkittok/omnir-refactor/actions/workflows/ci.yml/badge.svg?branch=develop)](https://github.com/SEI-Mkittok/omnir-refactor/actions/workflows/ci.yml)

Omnir CRM is a ground-up modernization of vTiger Community Edition. The goal is not to skin old PHP screens. It is to preserve the CRM workflows that make vTiger useful, then rebuild them on a faster, safer, and more maintainable Go + React platform.

The refactor keeps the important business concepts: contacts, organizations, leads, opportunities, activities, tickets, quotes, reports, custom fields, workflows, roles, profiles, sharing rules, groups, and admin configuration. It replaces the legacy architecture: PHP globals, jQuery-era UI, ad-hoc AJAX endpoints, weak API contracts, brittle schema conventions, and hard-to-test procedural code.

## Why This Exists

vTiger CE carries a lot of valuable CRM domain knowledge, but the implementation makes modern product work expensive:

- The PHP backend mixes controllers, models, SQL, view rendering, global state, and request parsing.
- The UI is built around legacy jQuery patterns, scattered scripts, and page-specific behavior.
- API responses and permissions are inconsistent, which makes integrations and frontend work fragile.
- The database depends heavily on application-enforced relationships and legacy extension patterns.
- Operational concerns like migrations, CI, deployments, secrets, and background jobs are difficult to reason about.

Omnir turns that into a modern product foundation: typed backend services, explicit persistence, tested APIs, a component frontend, clean deployment, and a path for vTiger data/workflow migration.

## Benefits

**Faster product development**

- Clear Go domain models, handlers, repositories, and migrations.
- React + TypeScript UI with reusable components and typed API clients.
- Contract-oriented APIs that reduce frontend/backend drift.
- Focused tests around handlers, repositories, contracts, and core UI flows.

**Better security and tenant isolation**

- JWT auth, API keys, SSO/2FA surfaces, and admin audit logs.
- Org-aware data model with tenant scoping designed into every domain table.
- Role, profile, group, and sharing-rule enforcement for CRM records.
- Secrets kept in environment files and operational references, not code.

**Modern user experience**

- Responsive CRM shell instead of legacy desktop-only screens.
- First-class workflows for contacts, accounts, leads, deals, tickets, quotes, inbox, calendar, KB, reports, dashboards, and settings.
- Saved views, custom fields, timelines, linked entities, and admin configuration built as product surfaces rather than one-off pages.

**Operational clarity**

- Docker-based local stack with Postgres, API, frontend, MinIO, and MailHog.
- GitHub Actions CI and staging deployment workflows.
- Versioned SQL migrations with validation checks.
- Local and production environment references under `docs/ops/`.

**Migration-friendly architecture**

- vTiger features are audited and mapped before implementation.
- Legacy screenshots and feature indexes are preserved as reference material.
- The new schema favors explicit relationships, JSONB where useful, indexes, and predictable API contracts.

## Current Scope

The refactor already has a functional modern CRM core:

- Contacts, accounts/organizations, leads, deals/opportunities, tickets, and quotes.
- Customer portal runtime, knowledge base/help center, reports, dashboards, inbox, calendar, notifications, saved views, and custom fields.
- Admin/security surfaces for users, roles, profiles, sharing rules, groups, audit log, API keys, SLA policies, webhooks, SSO, and 2FA.
- CRM admin configuration bundles for settings, picklists, currencies, module layouts, relationships, numbering, and related settings work.

Remaining parity work is tracked in the docs and issues, especially deeper legacy admin/configuration areas, inventory/finance expansion, automation parity, webforms, scheduler visibility, tax/terms, and migration hardening.

Key references:

- `docs/vtiger-modernization-scope.md` - original vTiger audit and modernization rationale.
- `docs/original-crm-screenshot-index.md` - indexed legacy UI/function reference.
- `docs/original-crm-feature-implementation-status.md` - current parity status.
- `docs/plans/praestos-rewrite-plan.md` - technical rewrite plan.

## Architecture

- **API:** Go 1.24, Chi, pgx/v5, JWT auth, goose migrations.
- **Frontend:** React 18, TypeScript, Vite, TanStack Query/Router, Zustand, Radix/shadcn-style components.
- **Database:** PostgreSQL 16.
- **Storage/dev services:** MinIO, MailHog, Docker Compose.
- **CI/CD:** GitHub Actions for lint, tests, builds, and staging deployment.

The system is organized around explicit backend boundaries:

- `api/internal/domain/` for domain types and validation.
- `api/internal/handler/` for HTTP handlers.
- `api/internal/repository/` for interfaces.
- `api/internal/repository/postgres/` for persistence.
- `api/migrations/` for database changes.
- `api/openapi/` for API contract documentation.

The frontend keeps API clients, hooks, reusable UI, CRM components, settings pages, and route pages under `web/src/`.

## Repository Layout

```text
omnir-refactor/
├── .github/workflows/    # CI/CD workflows
├── api/                  # Go REST API
├── web/                  # React + TypeScript + Vite app
├── infra/                # Docker Compose, Dockerfiles, nginx
├── scripts/              # Utility scripts and validation helpers
├── docs/                 # Architecture, plans, QA, agent notes, workspace docs
├── AGENTS.md             # Pointer to docs/agents/current/AGENTS.md
├── HEARTBEAT.md          # Pointer to docs/agents/current/HEARTBEAT.md
└── Makefile              # Local development shortcuts
```

Root-level files are intentionally sparse. New notes, reports, plans, agent configuration, memories, and workspace state belong under `docs/`, not the repository root.

## Docs Map

```text
docs/
├── agents/               # Agent instructions, identity, keys, memory, PARA notes
├── memory/               # Main-session daily notes
├── plans/                # Roadmaps and implementation plans
├── qa/                   # QA reports and manual test scripts
├── ops/                  # Environment and operational references
├── migrations/           # Migration planning and rollout notes
├── integrations/         # External integration notes
├── reports/              # Diagnostic reports and logs worth preserving
├── ux-designs/           # UX explorations and visual references
└── workspace/            # Preserved local workspace metadata
```

Product/app functionality belongs in `api/`, `web/`, `infra/`, `scripts/`, and `.github/`. Everything else should be documented under `docs/`.

## Local Development

### Prerequisites

- Docker Desktop, or Docker Engine with the Compose plugin.
- Git.
- Optional for running outside containers: Go 1.24+ and Node 20+.

### Bootstrap

```bash
git clone https://github.com/SEI-Mkittok/omnir-refactor.git
cd omnir-refactor
make setup
```

`make setup` creates these files if they do not already exist:

- `.env`
- `api/.env`
- `web/.env.local`

Never commit real secrets in `.env` files.

### Start the Stack

```bash
make up
```

The local stack is defined in `infra/docker-compose.yml` and starts:

- `postgres` - PostgreSQL 16.
- `api` - Go API with hot reload via Air.
- `frontend` - React/Vite app.
- `minio` - S3-compatible object storage.
- `mailhog` - local SMTP and inbox UI.

Service URLs:

- Frontend: http://localhost:5173
- API: http://localhost:8080
- API health: http://localhost:8080/health
- Postgres: `localhost:5432`
- MinIO API: http://localhost:9000
- MinIO Console: http://localhost:9001
- MailHog UI: http://localhost:8025
- pgAdmin optional profile: http://localhost:5050

To include pgAdmin:

```bash
docker compose -f infra/docker-compose.yml --profile tools up
```

## Common Commands

```bash
make up              # Start all services
make up-d            # Start all services in detached mode
make down            # Stop and remove containers + volumes
make logs            # Tail logs from all services
make shell-api       # Shell into API container
make shell-db        # psql into PostgreSQL
make migrate         # Run pending DB migrations
make migrate-status  # Show migration status
make test            # Run all tests (API + frontend)
make lint            # Run linters
make fmt             # Format API + frontend code
```

Before pushing any branch:

```bash
# Linux/macOS/WSL/Git Bash
bash scripts/pre-push-check.sh

# Windows PowerShell/CMD
scripts\pre-push-check.cmd
```

## Environment References

- Root compose env template: `.env.example`
- API local env template: `api/.env.example`
- Frontend local env template: `web/.env.example`
- Extended ops reference: `docs/ops/env-reference.md`

## CI/CD

| Workflow | Trigger | Purpose |
|---|---|---|
| `ci.yml` | Push/PR to `main`, `develop` | Lint, test, and build validation |
| `deploy-staging.yml` | Push to `develop` | Build images and deploy to staging |

## Contributing

1. Branch from `develop`.
2. Keep changes scoped to the feature or cleanup at hand.
3. Run the mandatory pre-push validation script.
4. Open a PR with clear scope, test evidence, and any migration/API notes.
