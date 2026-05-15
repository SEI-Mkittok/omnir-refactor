# Bundle 6 CRM Admin Configuration Closeout Task List

Updated: 2026-05-15

Legend: `[x]` implemented or covered on `codex/bundle-6-crm-admin-custom-fields`; `[ ]` remains for QA/manual smoke or issue closure.

Smoke note: route-level browser smoke on 2026-05-15 used Vite plus a local mock API to verify Bundle 6 screens render without redirects, blank pages, or error boundaries. Items that require real seeded backend mutations remain explicitly unchecked.

## #231 CRM QA: Dashboard / Executive Overview

- [x] Verify dashboard KPIs map to current report API contracts.
- [x] Fix broken quick-action routing to implemented CRM routes.
- [x] Add frontend regression coverage for dashboard quick actions and task shortcuts.
- [x] Local browser route smoke `/dashboard`.

## #237 CRM QA: Customer Support Portal Runtime

- [x] Verify portal ticket list/create/detail/comment ownership checks.
- [x] Add portal public-comment list route and tests for owned vs foreign tickets.
- [x] Local browser route smoke `/portal` ticket list/create/detail routes, including reload session restore.

## #238 CRM QA: Knowledge Base / FAQ Replacement

- [x] Validate admin KB article custom fields on create/update.
- [x] Return expanded custom fields on authenticated KB article list/detail/search.
- [x] Keep public help-center/list/search article payloads sanitized.
- [x] Local browser route smoke `/kb` and public help routes.

## #239 CRM QA: Reports

- [x] Extend frontend report clients for conversion, revenue projection, and manager dashboard contracts.
- [x] Add report API contract coverage.
- [x] Harden newer report repository endpoints to require tenant scope.
- [x] Local browser route smoke `/reports`.

## #240 CRM QA: Custom Dashboards And Scheduled Reports

- [x] Normalize object-shaped widget data for custom-dashboard charts.
- [x] Add frontend widget normalization coverage.
- [x] Verify scheduled report dashboard references are scoped to the current org.
- [x] Local browser route smoke `/dashboards`.
- [x] Automated QA coverage for scheduled-report create/update/delete.

## #241 CRM QA: Saved Views And List Filtering

- [x] Harden saved-view pin ordering and explicit pin-order restore.
- [x] Preserve deletion/default behavior through existing repository boundaries.
- [x] Automated QA coverage for saved-view restore on contacts, accounts, leads, deals, and quotes.

## #242 CRM QA: Custom Fields On Core CRM Modules

- [x] Preserve existing contacts/accounts/leads/deals/tickets custom-field validation.
- [x] Add quote and KB article custom-field entity support without removing existing types.
- [x] Scope picklist usage/remap/delete to the field definition org.
- [x] Local browser route smoke `/settings/custom-fields`.

## #247 CRM QA: Notifications

- [x] Cover unread count and mark-read behavior.
- [x] Add SLA notification kinds to backend constraint and frontend UI mapping.
- [x] Fix notification unread indicator styling.
- [x] Local browser route smoke `/notifications`.

## #248 CRM QA: Basic User Administration

- [x] Trim create-user name/email payloads.
- [x] Enforce backend and frontend password minimum alignment.
- [x] Preserve admin-only user management route coverage.
- [x] Local browser route smoke `/users`.

## #249 CRM QA: API Keys

- [x] Preserve reveal-once create response with no hash exposure.
- [x] Reject revoked/expired/read-only API keys in auth middleware tests.
- [x] Audit create/revoke actions without logging plaintext or hashes.
- [x] Local browser route smoke `/api-keys`.

## #250 CRM QA: Audit Log

- [x] Keep editable audit search and actor display regressions fixed.
- [x] Include API key entity type in backend/frontend audit filters.
- [x] Cover API key audit writes.
- [x] Local browser route smoke `/admin/audit`.

## #252 CRM QA: SLA Policies

- [x] Preserve fractional SLA values regression fix.
- [x] Make breach scanning return only newly transitioned breaches.
- [x] Send SLA warning/breach notifications to real org-scoped recipients.
- [x] Local browser route smoke `/settings/sla`.

## #253 CRM QA: Security Controls

- [x] Add authenticated 2FA status endpoint and frontend loading/status wiring.
- [x] Add backend test coverage for 2FA status.
- [x] Local browser route smoke `/settings/security` and 2FA enrollment.

## #307 Feature: Extend Custom Fields Support To Quotes

- [x] Add `quote` custom-field entity type.
- [x] Add `quotes.custom_fields` migration, domain, repository, handler, API, and frontend types.
- [x] Validate, persist, expand, render, and edit quote custom fields.
- [x] Scope quote read/update/delete/status actions to org context.
- [x] Local browser route smoke `/quotes`.

## #308 Feature: Extend Custom Fields Support To Knowledge Base Articles

- [x] Add `kb_article` custom-field entity type.
- [x] Add `articles.custom_fields` migration, domain, repository, handler, API, and frontend types.
- [x] Validate, persist, expand, render, and edit admin KB article custom fields.
- [x] Keep public KB/help-center responses free of custom fields.
- [x] Local browser route smoke `/kb`.
