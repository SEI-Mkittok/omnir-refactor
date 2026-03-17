# Omnir CRM — Frontend Architecture

> Status: Draft v1.0  
> Issue: OMN-4  
> Author: Forge (AI Dev Agent)  
> Date: 2026-03-17

---

## 1. Tech Stack

| Layer | Choice | Rationale |
|---|---|---|
| Framework | **React 18** | Mature ecosystem, concurrent features, wide hiring pool |
| Language | **TypeScript 5.x** (strict mode) | Type safety across the full module graph |
| Build tool | **Vite 5** | Fast HMR, native ESM, lean config |
| UI primitives | **shadcn/ui** (Radix UI + Tailwind) | Copy-own components, accessible, unstyled primitives |
| Styling | **Tailwind CSS v3** + CSS variables for theming | Utility-first, easy design tokens, minimal bundle |
| Icons | **Lucide React** | Consistent, tree-shakable, pairs with shadcn |

---

## 2. State Management Strategy

### Server state → **TanStack Query (React Query v5)**
- All API reads: contacts, accounts, deals, activities, reports
- Automatic caching, background refetch, optimistic updates
- Prefetch on hover/navigate for snappy UX
- Mutation + invalidation pattern for writes

### Client / UI state → **Zustand**
- Lightweight (no boilerplate)
- Slices: `ui` (sidebar open, active filters, selected rows), `auth` (user session), `notifications`
- Persist to `localStorage` with `zustand/middleware/persist` for layout prefs

### Form state → **React Hook Form + Zod**
- Zod schemas shared between frontend validation and API response parsing
- Avoids putting form state in Zustand

### Rule of thumb
> If it comes from the server, use React Query. If it's UI-only, use Zustand. If it's a form, use RHF.

---

## 3. Component Library & Design System Tokens

### Token Architecture (CSS custom properties)

```css
/* tokens.css */
:root {
  /* Color — semantic */
  --color-bg:          #0f1117;
  --color-surface:     #1a1d27;
  --color-border:      #2c2f3e;
  --color-text:        #e2e8f0;
  --color-text-muted:  #64748b;

  /* Brand */
  --color-primary:     #6366f1;   /* indigo */
  --color-primary-fg:  #ffffff;
  --color-success:     #22c55e;
  --color-warning:     #f59e0b;
  --color-danger:      #ef4444;

  /* Spacing scale (4px base) */
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-6: 24px;
  --space-8: 32px;

  /* Typography */
  --font-sans:  'Inter', system-ui, sans-serif;
  --font-mono:  'JetBrains Mono', monospace;
  --text-xs:    0.75rem;
  --text-sm:    0.875rem;
  --text-base:  1rem;
  --text-lg:    1.125rem;
  --text-xl:    1.25rem;

  /* Radius */
  --radius-sm:  4px;
  --radius-md:  8px;
  --radius-lg:  12px;
}
```

### shadcn/ui Component Customization
- Override shadcn default tokens via `globals.css` (they use CSS vars internally)
- Extend shadcn with Omnir-specific components under `src/components/omnir/`
- Never fork shadcn primitives — patch via composition

### Component Tiers

| Tier | Location | Examples |
|---|---|---|
| **Primitives** (shadcn) | `src/components/ui/` | Button, Input, Badge, Dialog, Table, Select |
| **Omnir domain** | `src/components/omnir/` | ContactCard, DealKanbanCard, ActivityFeed, PipelineStage |
| **Layout** | `src/components/layout/` | AppShell, Sidebar, TopBar, PageHeader |
| **Pages** | `src/pages/` | one component per route |

---

## 4. Routing Structure

Router: **TanStack Router v1** (type-safe routes, file-based optional)

```
/                           → redirect → /contacts
/contacts                   → ContactsListPage
/contacts/:id               → ContactDetailPage
/contacts/new               → ContactFormPage
/accounts                   → AccountsListPage
/accounts/:id               → AccountDetailPage
/accounts/new               → AccountFormPage
/deals                      → DealsListPage
/deals/:id                  → DealDetailPage
/deals/new                  → DealFormPage
/pipeline                   → PipelineBoardPage   (Kanban)
/pipeline/:pipelineId       → PipelineBoardPage   (specific pipeline)
/activities                 → ActivitiesListPage
/activities/:id             → ActivityDetailPage
/reports                    → ReportsDashboardPage
/reports/:reportId          → ReportViewPage
/settings                   → SettingsPage
/settings/integrations      → IntegrationsPage
/login                      → LoginPage
```

### Layout nesting

```
RootLayout (auth guard, error boundary, query client)
  AppShell (sidebar + topbar)
    ModuleLayout (per-module breadcrumbs, page header)
      <Page />
```

### Code splitting
- Each top-level module (`contacts`, `accounts`, `deals`, `pipeline`, `activities`, `reports`) is a lazy-loaded chunk via `React.lazy` + `Suspense`

---

## 5. API Client Layer

### Stack: **Axios** + **TanStack Query** + **Zod** response validation

```
src/
  api/
    client.ts          ← Axios instance (base URL, auth interceptor, error handling)
    contacts.ts        ← query/mutation fns for contacts
    accounts.ts
    deals.ts
    activities.ts
    pipeline.ts
    reports.ts
  schemas/
    contact.schema.ts  ← Zod schema + inferred TS types
    account.schema.ts
    deal.schema.ts
    ...
```

### `client.ts` pattern

```ts
import axios from 'axios';

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? '/api',
  headers: { 'Content-Type': 'application/json' },
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('omnir_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

apiClient.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);
```

### Query factory pattern

```ts
// src/api/contacts.ts
import { apiClient } from './client';
import { ContactSchema, ContactListSchema } from '../schemas/contact.schema';

export const contactsQueryKeys = {
  all: ['contacts'] as const,
  list: (params: ContactListParams) => ['contacts', 'list', params] as const,
  detail: (id: string) => ['contacts', 'detail', id] as const,
};

export async function fetchContacts(params: ContactListParams) {
  const res = await apiClient.get('/contacts', { params });
  return ContactListSchema.parse(res.data);
}

export async function fetchContact(id: string) {
  const res = await apiClient.get(`/contacts/${id}`);
  return ContactSchema.parse(res.data);
}
```

---

## 6. Project Structure

```
omnir-frontend/
  src/
    api/               ← API functions per module
    schemas/           ← Zod schemas + inferred types
    components/
      ui/              ← shadcn primitives (generated, don't edit)
      omnir/           ← domain-specific components
      layout/          ← AppShell, Sidebar, etc.
    pages/             ← one file per route
    hooks/             ← shared custom hooks
    stores/            ← Zustand slices
    lib/               ← utils, formatters, cn()
    styles/            ← globals.css, tokens.css
    router.ts          ← route definitions
    main.tsx           ← app entry
  public/
  index.html
  vite.config.ts
  tailwind.config.ts
  tsconfig.json
  package.json
```

---

## 7. Component Tree Sketch

```
<App>
  <QueryClientProvider>
    <RouterProvider>
      <RootLayout>                        # auth guard, global error boundary
        <AppShell>
          <Sidebar>
            <NavItem icon={Users}     to="/contacts" />
            <NavItem icon={Building2} to="/accounts" />
            <NavItem icon={Briefcase} to="/deals" />
            <NavItem icon={Kanban}    to="/pipeline" />
            <NavItem icon={Calendar}  to="/activities" />
            <NavItem icon={BarChart}  to="/reports" />
          </Sidebar>
          <TopBar>
            <GlobalSearch />
            <NotificationBell />
            <UserMenu />
          </TopBar>
          <main>
            {/* lazy-loaded module chunks */}
            <ContactsModule />   → ContactsListPage | ContactDetailPage | ContactFormPage
            <AccountsModule />   → AccountsListPage | AccountDetailPage | AccountFormPage
            <DealsModule />      → DealsListPage   | DealDetailPage   | DealFormPage
            <PipelineModule />   → PipelineBoardPage (Kanban columns + DealCard drag)
            <ActivitiesModule /> → ActivitiesListPage | ActivityDetailPage
            <ReportsModule />    → ReportsDashboardPage | ReportViewPage
          </main>
        </AppShell>
      </RootLayout>
    </RouterProvider>
  </QueryClientProvider>
</App>
```

### Key domain components

**ContactCard** — avatar, name, company, last activity, quick-action buttons  
**DealKanbanCard** — deal name, value badge, owner avatar, probability bar, drag handle  
**PipelineStage** — column header (stage name + total value), `<DragDropContext>` drop zone  
**ActivityFeed** — chronological list of calls/emails/meetings, grouped by date  
**DataTable** — generic sortable/filterable table wrapping TanStack Table v8  
**QuickCreateModal** — slide-over panel for fast contact/deal creation without leaving context  

---

## 8. Key Decisions Summary

| Decision | Choice | Alternatives considered |
|---|---|---|
| UI primitives | shadcn/ui | MUI, Ant Design, Chakra |
| State: server | TanStack Query | SWR, Redux Toolkit Query |
| State: client | Zustand | Redux, Jotai, Context |
| Router | TanStack Router | React Router v6 |
| Forms | React Hook Form + Zod | Formik |
| HTTP | Axios | fetch, ky |
| Build | Vite | CRA, Next.js (not needed — pure SPA) |

---

## 9. Next Steps (follow-on issues)

1. Scaffold repo with Vite + shadcn + Zustand + TanStack Query + Router
2. Implement AppShell (sidebar, topbar, routing)
3. Implement Contacts module (list + detail + form)
4. Design vtiger PHP API adapter (REST bridge for legacy data)
5. Implement Pipeline Kanban board
