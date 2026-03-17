# CRM Domain Knowledge

_Last updated: 2026-03-17_

## Core CRM Entities
- **Contacts** — Individual people (leads, prospects, customers)
- **Accounts** — Companies/organizations
- **Deals/Opportunities** — Sales pipeline items with stages and values
- **Activities** — Calls, emails, meetings, tasks (NOT yet in Omnir schema)
- **Pipelines** — Configurable stage workflows for deals

## vtiger CE Notes
- PHP/jQuery legacy stack — high replacement priority
- MySQL database — migrating to PostgreSQL
- Lead tracking via `potentials` module (= deals in Omnir)
- Activities are critical: vtiger's activity module is heavily used
- Custom fields everywhere — handled via JSONB in Omnir

## Key Design Decisions (Omnir)
- Money stored as `value_cents BIGINT` to avoid float issues
- Soft deletes via `deleted_at` on all entities
- Flexible extension via `custom_fields JSONB`
- Single default pipeline seeded at migration
- Leads are contacts with `stage = 'lead'` (no separate leads table)
