# Original CRM Feature Implementation Status

Source index: [original-crm-screenshot-index.md](original-crm-screenshot-index.md)

Review date: 2026-05-09

Last updated: 2026-05-11 after Bundle 4 RBAC/sharing implementation.

Method: initial static review of frontend routes/pages, backend route mounts, repositories, migrations, workers, and API clients. Bundle 4 status reflects the 2026-05-11 implementation pass and validation.

## Definition of FULL

A feature is counted as `FULL` only when it has:

- A user-facing page, portal, or first-class workflow.
- Backend API and persistence support.
- Frontend/backend wiring through API clients, hooks, or workers.
- The main create/read/update/action behavior expected by the current product.

Backend-only tables, orphaned routes, static navigation links, or partial settings screens are not counted as `FULL`.

## Fully Implemented

These are the CRM functions that are fully implemented end to end in the current refactor.

| Feature | Status | Evidence | Scope notes |
|---|---|---|---|
| App shell and implemented-module navigation | FULL | `web/src/main.tsx`, `web/src/components/layout/Sidebar.tsx` | Modern navigation is complete for implemented modules, but it is not the legacy configurable Marketing/Sales/Inventory menu system. |
| Dashboard / executive overview | FULL | `/dashboard`, `web/src/pages/DashboardPage.tsx`, `/api/v1/reports/*` | Modern dashboard and reporting summary are implemented. |
| Contacts | FULL | `/contacts`, `/contacts/:id`, `/api/v1/contacts` | Includes list/detail, CRUD, custom fields, notes, activities, email/timeline integration, attachments, linked entities, import/export, and saved views. |
| Organizations / Accounts | FULL | `/accounts`, `/accounts/:id`, `/api/v1/accounts` | Includes CRUD, hierarchy/relationships, linked contacts/deals/tickets/quotes/sequences/inbox, import/export, and saved views. |
| Leads | FULL | `/leads`, `/api/v1/leads` | Includes CRUD, scoring/status/source fields, custom fields, import, saved views, and lead conversion workflow. Admin lead-conversion field mapping is not implemented. |
| Opportunities / Deals | FULL | `/deals`, `/api/v1/deals` | Includes list/kanban, CRUD, stage changes, account/contact relationships, custom fields, activities, notes, and quote links. Opportunity-to-project mapping is not implemented. |
| Tickets / service desk | FULL | `/tickets`, `/tickets/:id`, `/api/v1/tickets` | Includes CRUD, priority/status filtering, comments, attachments, contact/account assignment, SLA support, and portal ticket integration. |
| Customer support portal runtime | FULL | `/portal/login`, `/portal/tickets`, `/api/v1/portal` | Portal ticket login/list/create/detail/comment flow exists. The legacy portal configuration screen is not implemented. |
| Knowledge Base / FAQ replacement | FULL | `/kb`, `/help/:orgSlug/*`, `/api/v1/kb`, `/api/portal/help` | Modern KB and public help center cover the legacy FAQ/support knowledge use case. |
| Reports | FULL | `/reports`, `/api/v1/reports/*` | Includes summary, tickets, contacts, deals, leads, funnel, conversion, revenue projection, activity, and manager reporting endpoints. |
| Custom dashboards and scheduled reports | FULL | `/dashboards`, `/api/v1/dashboards`, `/api/v1/reports/schedules` | Dashboard CRUD/run and scheduled report persistence/worker are present. |
| Saved views and list filtering | FULL | `/api/v1/views`, `ViewPinBar` usage in core lists | Implemented for major CRM list workflows. |
| Custom fields on core CRM modules | FULL | `/settings/custom-fields`, `/api/v1/custom-fields` | Dynamic fields are available on contacts, accounts, leads, deals, and tickets. This is not a full legacy layout/block editor. |
| Core quote management | FULL | `/quotes`, `/api/v1/quotes`, `/api/v1/deals/{dealId}/quotes` | Quote list/builder/create/update/delete/send are wired. Backend approve/reject exists, but no dedicated UI was found for those actions. |
| Email sequences | FULL | `/sequences`, `/api/v1/sequences`, tracking/unsubscribe/bounce routes | Modern drip/sequence workflow is implemented. It does not cover legacy campaigns, submissions, or full email marketing administration. |
| Email inbox / Mail Manager baseline | FULL | `/inbox`, `/api/v1/emails/*`, `/api/integrations/email/*` | Connected inbox, Gmail/Outlook OAuth, sync worker, threads, read state, compose/reply, and templates picker exist. Legacy Mail Converter rule setup is not implemented. |
| Calendar activity view and external sync | FULL | `/calendar`, `/api/v1/calendar`, calendar sync worker | Calendar view, activity interaction, Google/Microsoft connection, and sync exist. Legacy personal calendar preference settings are not implemented. |
| Notifications | FULL | `/notifications`, notification bell/components, `/api/v1/notifications` | Notification list/preferences and worker-backed notifications are present. |
| Basic user administration | FULL | `/users`, `/api/v1/users` | User CRUD, platform role compatibility, and org role/profile assignment are implemented. |
| Org RBAC: roles, profiles, sharing rules, groups, login history | FULL | `/settings/roles`, `/settings/profiles`, `/settings/sharing-rules`, `/settings/groups`, `/api/v1/settings/*`, `/admin/audit` | Includes role hierarchy/closure, profile permissions, field write controls, sharing defaults and role/group grants, group memberships, user role/profile assignment, and login history through the Audit Log `action=login` filter. |
| API keys | FULL | `/api-keys`, `/api/v1/api-keys` | Admin API key management is implemented. |
| Audit log | FULL | `/admin/audit`, `/api/v1/admin/audit-log` | Admin audit log browsing is implemented, including login history entries for successful password and SSO login. |
| Outbound webhooks | FULL | `/settings/webhooks`, `/api/v1/webhooks` | Webhook CRUD, test delivery, and delivery logging are implemented. |
| SLA policies | FULL | `/settings/sla`, `/api/v1/sla-policies`, `/api/v1/sla-instances` | SLA policies, instances, and breach worker are implemented for support. |
| Security controls | FULL | `/settings/security`, `/settings/security/2fa/enroll`, SSO/2FA APIs | 2FA and SSO support are implemented as modern security surfaces. |

## Not Fully Implemented

These areas appear in the original screenshots but are only partially implemented, backend-only, or absent.

| Area from screenshots | Current status | Why it is not FULL |
|---|---|---|
| Legacy Marketing suite | PARTIAL | Leads, contacts, accounts, inbox, and sequences exist. Campaigns, submissions, email actions, and full email marketing administration were not found. |
| Legacy Sales suite | PARTIAL | Deals, quotes, contacts, and accounts exist. Products are backend/picker only, and services are not a full module. |
| Inventory suite | NOT FULL | Products API exists and quotes exist. Price books, sales orders, purchase orders, vendors, product/service templates, and full inventory UI were not found. |
| Products module | PARTIAL | `/api/v1/products` and product picker hooks exist, but there is no first-class `/products` page. |
| Services module | NOT FULL | No full service catalog UI/API was found. Service contracts exist only in ops/finance backend shape. |
| Service contracts and assets | BACKEND ONLY | Ops/finance tables and endpoints exist, but no user-facing CRM module pages were found. |
| Projects and project tasks | BACKEND ONLY | Ops/finance backend has projects/tasks, but no `/projects` UI workflow. |
| Project milestones | NOT FOUND | No full milestone workflow was found. |
| Invoices, payments, time entries, expenses | BACKEND ONLY | Ops/finance endpoints/tables exist, but no full frontend workflow was found. |
| Vendors, sales orders, purchase orders, price books | NOT FOUND | No meaningful implementation was found. |
| Documents, recycle bin, document designer, email designer, timesheets, product/service templates | NOT FOUND | No full user-facing workflows were found. |
| Settings home parity | PARTIAL | `SettingsHubPage` exists, but it does not cover the full legacy settings summary/counts and shortcuts. |
| Extended user profile/preferences | PARTIAL | Account/security settings exist. Legacy per-user currency/number defaults, landing page, photo, service settings, tags, and detailed preferences are not full. |
| Module layouts and fields | PARTIAL | Custom fields are full, but drag/drop layout blocks, hidden fields, mandatory/quick-create/mass-edit flags, relationship layout editor, and duplicate prevention are not full. |
| Module relationships editor | PARTIAL | Some entity relationships are implemented, but there is no generic relationship editor UI. |
| Module numbering | PARTIAL | Backend `/api/v1/settings/numbering` exists and SettingsHub links to `/settings/numbering`, but no frontend route was found for that page. |
| Webforms | NOT FOUND | No webform list/builder/public capture workflow was found. |
| Scheduler console | PARTIAL | Many workers exist, but no admin page shows scheduled jobs, frequencies, status, or scan times. |
| Workflow automation | PARTIAL | `/automations` UI, backend, and worker exist. Some triggers/actions are wired, but legacy parity is missing and `ticket_created` appears in the automation domain without ticket handler event wiring. |
| Company details | PARTIAL | Organization/setup/onboarding data exists, but no full company profile/logo settings screen was found. |
| Customer portal configuration | PARTIAL | Portal runtime exists, but default assignee, portal menu, home announcement, shortcuts, and recent-record widget configuration were not found. |
| Outgoing server settings | NOT FULL | SMTP is environment/config driven; no admin SMTP settings screen was found. |
| Configuration editor | PARTIAL | Some org settings exist, but no broad global configuration editor for support identity, upload limits, page sizes, and similar values was found. |
| Currencies | NOT FOUND | No full currency management workflow was found. |
| Picklist field values | PARTIAL | Select/multiselect custom-field options exist, but no generic module/field picklist value editor was found. |
| Picklist dependency | NOT FOUND | No dependency mapping workflow was found. |
| Main menu configuration | NOT FOUND | Sidebar is static; no drag/drop menu configuration workflow was found. |
| Lead conversion data mapping | PARTIAL | Lead conversion exists, but admin field mapping from leads to accounts/contacts/deals was not found. |
| Opportunity to project mapping | NOT FOUND | No project conversion/mapping workflow was found. |
| Tax calculations | NOT FOUND | No tax management workflow was found. |
| Terms and conditions | NOT FOUND | No module-specific terms editor was found. |
| Calendar settings | PARTIAL | Calendar usage/sync exists, but personal preferences from the screenshot are not implemented. |
| Mail Converter | PARTIAL | Inbox sync and email threading exist, but mailbox rule setup and automatic CRM record creation/update rules were not found. |
| Extension store / extension packs | NOT FOUND | No extension marketplace/pack workflow was found. |
| Google integration settings | PARTIAL | Google calendar and Gmail connection flows exist, but no extension-store style Google settings screen exists. |
| Integrations settings | PARTIAL | Integration pages exist, but Gmail/Outlook settings OAuth paths do not match the backend email OAuth routes; the Inbox page uses the correct OAuth paths. |

## Screenshot Row Verdicts

| Screenshot rows | Legacy function | Verdict |
|---|---|---|
| 01 | Global navigation | PARTIAL - modern nav exists, legacy module group parity/configuration does not. |
| 02 | Marketing navigation | PARTIAL - leads/contacts/accounts/sequences exist; campaigns/submissions/email actions do not. |
| 03 | Sales navigation | PARTIAL - deals/quotes/contacts/accounts exist; products/services are not full modules. |
| 04 | Inventory navigation | NOT FULL - quote/product pieces exist, but inventory suite is absent. |
| 05 | Support navigation | PARTIAL - tickets and KB are full; service contracts/assets/feedback are not. |
| 06 | Projects navigation | BACKEND ONLY - project/task backend exists without UI. |
| 07 | Tools navigation | PARTIAL - inbox, notifications, reports, dashboards exist; several designer/recycle/template tools do not. |
| 08 | Settings home | PARTIAL. |
| 09 | User management navigation | FULL - users, Roles, Profiles, Sharing Rules, Groups, and Audit Log login filtering are implemented. |
| 10 | Module management | PARTIAL. |
| 11 | Automation settings | PARTIAL. |
| 12 | Configuration settings | PARTIAL. |
| 13 | Marketing/sales mapping tools | PARTIAL - conversion exists, mapping screens do not. |
| 14 | Inventory settings | NOT FOUND. |
| 15 | User preferences | PARTIAL. |
| 16 | Extensions | PARTIAL. |
| 17 | Other settings / Mail Converter | PARTIAL. |
| 18-19 | User detail/profile | PARTIAL. |
| 20 | Roles | FULL - org roles, system-role compatibility, hierarchy closure, CRUD, and parent changes are implemented. |
| 21 | Profiles | FULL - default profiles, module action permissions, field write controls, CRUD, and permissions update are implemented. |
| 22 | Sharing rules | FULL - private/public defaults plus role/group grants are implemented and enforced on CRM records. |
| 23 | Groups | FULL - groups, memberships, settings UI/API, and sharing grant participation are implemented. |
| 24 | Module layouts and fields | PARTIAL - custom fields are full, full layout editing is not. |
| 25 | Module relationships | PARTIAL. |
| 26 | Module numbering | PARTIAL - backend exists, UI route is missing. |
| 27-28 | Webforms | NOT FOUND. |
| 29 | Scheduler | PARTIAL - workers exist, admin console does not. |
| 30 | Workflows | PARTIAL. |
| 31 | Company details | PARTIAL. |
| 32 | Customer portal configuration | PARTIAL - runtime is full, configuration UI is not. |
| 33 | Outgoing server | NOT FULL. |
| 34 | Transitional crop | N/A. |
| 35 | Configuration editor | PARTIAL. |
| 36 | Picklist field values | PARTIAL. |
| 37 | Picklist dependency | NOT FOUND. |
| 38 | Main menu configuration | NOT FOUND. |
| 39 | Lead conversion mapping | PARTIAL. |
| 40 | Opportunity to project mapping | NOT FOUND. |
| 41 | Tax calculations | NOT FOUND. |
| 42 | Terms and conditions | NOT FOUND. |
| 43 | Calendar settings | PARTIAL. |
| 44 | Mail converter | PARTIAL. |

## Bottom Line

The refactor has a fully functional modern CRM core: contacts, accounts, leads, deals, tickets, KB/help center, reports, dashboards, quotes, inbox, sequences, saved views, custom fields, notifications, and key admin/security surfaces.

The largest unfinished surface is now the remaining legacy administration/configuration layer outside Bundle 4: module layout editors, webforms, scheduler visibility, workflow parity, company/portal/SMTP/global settings, menu configuration, inventory settings, tax, terms, and deeper mapping screens.
