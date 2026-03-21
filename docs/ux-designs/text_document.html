# PraestOS Platform Rewrite — Technical Specification

**Strategy:** Full greenfield rewrite · **Phase 1 Scope:** Help Desk + Contacts/Accounts/Leads · **Deployment:** Single-tenant & multi-tenant SaaS

---

## 1. Guiding Principles

- **Org-aware from day one.** Every table carries `org_id`. Tenant isolation is enforced at the DB layer, not the application layer — no leakage risk when multi-tenant.
- **Strangler-proof schema.** Even though this is a full rewrite, design the PostgreSQL schema so a future data migration from vTiger's MySQL is mechanical (column mapping, not restructuring).
- **Contract-first API.** OpenAPI 3.1 spec is the source of truth. Go handlers and the React client are both generated/validated against it.
- **Twelve-factor infra.** All config via environment. No secrets in images. Stateless app containers.

---

## 2. Repository Layout

```
praestos/
├── api/                    # Go backend
│   ├── cmd/server/         # main.go entrypoint
│   ├── internal/
│   │   ├── auth/           # JWT, sessions, RBAC
│   │   ├── helpdesk/       # tickets domain
│   │   ├── crm/            # contacts, accounts, leads
│   │   ├── db/             # pgx pool, query wrappers
│   │   ├── middleware/      # logging, auth, org scoping
│   │   └── config/         # env parsing
│   ├── migrations/         # goose SQL files
│   └── openapi/            # spec.yaml
├── web/                    # React frontend
│   ├── src/
│   │   ├── api/            # TanStack Query hooks (generated from spec)
│   │   ├── components/ui/  # shadcn/ui components
│   │   ├── features/
│   │   │   ├── helpdesk/
│   │   │   └── crm/
│   │   ├── stores/         # Zustand slices
│   │   └── router.tsx      # TanStack Router
│   └── vite.config.ts
├── infra/
│   ├── docker/
│   │   ├── Dockerfile.api
│   │   ├── Dockerfile.web
│   │   └── nginx.conf
│   ├── docker-compose.yml          # local dev
│   └── docker-compose.prod.yml
└── .github/
    └── workflows/
        ├── ci.yml
        └── deploy.yml
```

---

## 3. Database Design (PostgreSQL 16)

### Core Tenancy Pattern

```sql
-- Every domain table follows this pattern
CREATE TABLE tickets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    -- domain columns...
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ          -- soft delete
);

CREATE INDEX ON tickets (org_id) WHERE deleted_at IS NULL;
```

### Deployment Modes

| Mode | Mechanism | Isolation |
|------|-----------|-----------|
| Single-tenant | Single DB, `org_id` = one row | Schema-level |
| Multi-tenant SaaS | Single DB, many `org_id` rows | Row-level security (RLS) |
| Enterprise isolation | Separate DB per org | Database-level |

Enable RLS for SaaS mode:

```sql
ALTER TABLE tickets ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON tickets
    USING (org_id = current_setting('app.current_org_id')::uuid);
```

Set `app.current_org_id` in the Go middleware per request via `SET LOCAL`.

### Phase 1 Schema

```sql
-- orgs & users
CREATE TABLE orgs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    plan        TEXT NOT NULL DEFAULT 'single',  -- 'single' | 'saas'
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id),
    email       TEXT NOT NULL,
    name        TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'agent',   -- 'admin' | 'agent' | 'client'
    pw_hash     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id, email)
);

-- CRM
CREATE TABLE accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id),
    name        TEXT NOT NULL,
    website     TEXT,
    industry    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE contacts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id),
    account_id  UUID REFERENCES accounts(id),
    first_name  TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    email       TEXT,
    phone       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE leads (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id),
    first_name  TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    email       TEXT,
    company     TEXT,
    status      TEXT NOT NULL DEFAULT 'new',    -- 'new' | 'contacted' | 'qualified' | 'converted' | 'dead'
    converted_contact_id UUID REFERENCES contacts(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

-- Help Desk
CREATE TABLE tickets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id),
    contact_id  UUID REFERENCES contacts(id),
    account_id  UUID REFERENCES accounts(id),
    assigned_to UUID REFERENCES users(id),
    subject     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'open',   -- 'open' | 'pending' | 'resolved' | 'closed'
    priority    TEXT NOT NULL DEFAULT 'medium', -- 'low' | 'medium' | 'high' | 'critical'
    source      TEXT,                           -- 'email' | 'portal' | 'phone' | 'api'
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE ticket_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id),
    ticket_id   UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL REFERENCES users(id),
    body        TEXT NOT NULL,
    is_internal BOOL NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ticket_attachments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id),
    ticket_id   UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    filename    TEXT NOT NULL,
    size_bytes  BIGINT,
    storage_key TEXT NOT NULL,   -- S3/object-store path
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Goose Migration Convention

```
migrations/
  00001_create_orgs.sql
  00002_create_users.sql
  00003_create_crm.sql
  00004_create_helpdesk.sql
  00005_add_ticket_rls.sql
```

Each file uses `-- +goose Up` / `-- +goose Down` tags.

---

## 4. Backend — Go 1.24

### Stack

| Layer | Choice | Why |
|-------|--------|-----|
| Router | `go-chi/chi v5` | Lightweight, stdlib-compatible, middleware composable |
| DB driver | `jackc/pgx v5` | Native PG protocol, typed scans, pipeline support |
| Migrations | `pressly/goose v3` | SQL-first, embeddable, rollback support |
| Auth | JWT via `golang-jwt/jwt v5` | Stateless, fits both tenancy models |
| Validation | `go-playground/validator v10` | Struct tags, custom rules |
| Config | `caarlos0/env v11` | Zero-dep env parsing |
| Logging | `rs/zerolog` | Structured JSON, zero-alloc |

### Project Structure (internal/)

```
internal/
├── auth/
│   ├── middleware.go      # JWT validation, sets ctx org_id + user_id
│   ├── rbac.go            # role permission map
│   └── tokens.go          # issue / refresh / revoke
├── db/
│   ├── pool.go            # pgx pool init, SET LOCAL org middleware
│   └── tx.go              # transaction helper
├── helpdesk/
│   ├── handler.go         # HTTP handlers (Chi sub-router)
│   ├── service.go         # business logic
│   ├── repository.go      # pgx queries
│   └── model.go           # Go structs
└── crm/
    ├── contacts/
    ├── accounts/
    └── leads/
```

### Org Scoping Middleware

```go
func OrgScope(pool *pgxpool.Pool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            orgID := auth.OrgIDFromCtx(r.Context())
            conn, _ := pool.Acquire(r.Context())
            defer conn.Release()
            conn.Exec(r.Context(),
                "SET LOCAL app.current_org_id = $1", orgID)
            next.ServeHTTP(w, r)
        })
    }
}
```

### API Route Design

```
/api/v1
  /auth
    POST /login
    POST /refresh
    POST /logout
  /tickets
    GET    /              list (filterable: status, priority, assignee, contact)
    POST   /              create
    GET    /:id            detail
    PATCH  /:id            update
    DELETE /:id            soft delete
    GET    /:id/comments
    POST   /:id/comments
    POST   /:id/attachments
  /contacts
    GET / POST / GET/:id / PATCH/:id / DELETE/:id
  /accounts
    GET / POST / GET/:id / PATCH/:id / DELETE/:id
  /leads
    GET / POST / GET/:id / PATCH/:id / DELETE/:id
    POST /:id/convert      convert lead → contact
  /users
    GET / POST / GET/:id / PATCH/:id
```

### Repository Pattern (example: tickets)

```go
type TicketRepository struct { pool *pgxpool.Pool }

func (r *TicketRepository) List(ctx context.Context, orgID uuid.UUID, f ListFilter) ([]Ticket, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT id, subject, status, priority, assigned_to, created_at
        FROM tickets
        WHERE org_id = $1
          AND ($2::text IS NULL OR status = $2)
          AND deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT $3 OFFSET $4
    `, orgID, f.Status, f.Limit, f.Offset)
    // pgx.CollectRows → []Ticket
}
```

---

## 5. Frontend — React 18 + TypeScript

### Stack

| Concern | Library |
|---------|---------|
| Build | Vite 5 |
| Routing | TanStack Router v1 (file-based) |
| Server state | TanStack Query v5 |
| Client state | Zustand v4 |
| UI components | shadcn/ui + Tailwind CSS v3 |
| Forms | React Hook Form + Zod |
| Tables | TanStack Table v8 |
| Auth | JWT stored in `httpOnly` cookie (no localStorage) |

### Feature Folder Structure

```
features/helpdesk/
├── components/
│   ├── TicketList.tsx
│   ├── TicketDetail.tsx
│   ├── TicketForm.tsx
│   └── CommentThread.tsx
├── hooks/
│   ├── useTickets.ts      # TanStack Query wrappers
│   └── useTicketMutations.ts
├── store/
│   └── ticketFilters.ts   # Zustand slice (local UI state only)
└── types.ts               # mirrors OpenAPI models
```

### TanStack Query Pattern

```typescript
// hooks/useTickets.ts
export const ticketKeys = {
  all: ['tickets'] as const,
  list: (filters: TicketFilters) => [...ticketKeys.all, 'list', filters] as const,
  detail: (id: string) => [...ticketKeys.all, 'detail', id] as const,
};

export function useTickets(filters: TicketFilters) {
  return useQuery({
    queryKey: ticketKeys.list(filters),
    queryFn: () => api.tickets.list(filters),
    staleTime: 30_000,
  });
}

export function useCreateTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.tickets.create,
    onSuccess: () => qc.invalidateQueries({ queryKey: ticketKeys.all }),
  });
}
```

### Zustand (UI-only state)

```typescript
// stores/ticketFilters.ts
interface TicketFilterStore {
  status: string | null;
  priority: string | null;
  setStatus: (s: string | null) => void;
  setPriority: (p: string | null) => void;
}

export const useTicketFilterStore = create<TicketFilterStore>((set) => ({
  status: null,
  priority: null,
  setStatus: (status) => set({ status }),
  setPriority: (priority) => set({ priority }),
}));
```

> **Rule:** Zustand holds only ephemeral UI state (filters, modal open state, sidebar collapsed). All server state lives in TanStack Query. No duplication.

---

## 6. Infrastructure

### Docker

**`Dockerfile.api`**
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
```

**`Dockerfile.web`**
```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN npm i -g pnpm && pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM nginx:1.27-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY infra/docker/nginx.conf /etc/nginx/conf.d/default.conf
```

**`nginx.conf`**
```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location / {
        try_files $uri $uri/ /index.html;  # SPA fallback
    }
}
```

**`docker-compose.yml` (local dev)**
```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: praestos
      POSTGRES_USER: praestos
      POSTGRES_PASSWORD: dev_only
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports: ["5432:5432"]

  api:
    build: { context: ./api, dockerfile: ../infra/docker/Dockerfile.api }
    environment:
      DATABASE_URL: postgres://praestos:dev_only@db:5432/praestos
      JWT_SECRET: dev_secret_change_me
      ORG_MODE: single          # 'single' | 'saas'
    depends_on: [db]
    ports: ["8080:8080"]

  web:
    build: { context: ./web, dockerfile: ../infra/docker/Dockerfile.web }
    depends_on: [api]
    ports: ["3000:80"]

volumes:
  pgdata:
```

---

## 7. CI/CD — GitHub Actions

### `ci.yml` (on every PR)

```yaml
jobs:
  backend:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env: { POSTGRES_PASSWORD: test }
        options: --health-cmd pg_isready
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - run: go vet ./...
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - run: pnpm install --frozen-lockfile
      - run: pnpm typecheck
      - run: pnpm lint
      - run: pnpm test --run

  build-images:
    needs: [backend, frontend]
    runs-on: ubuntu-latest
    steps:
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v6
        with:
          context: ./api
          push: true
          tags: ghcr.io/${{ github.repository }}/api:${{ github.sha }}
      - uses: docker/build-push-action@v6
        with:
          context: ./web
          push: true
          tags: ghcr.io/${{ github.repository }}/web:${{ github.sha }}
```

### `deploy.yml` (on merge to `main`)

```yaml
jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          go install github.com/pressly/goose/v3/cmd/goose@latest
          goose -dir ./api/migrations postgres "$DATABASE_URL" up
        env:
          DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}

  deploy:
    needs: migrate
    runs-on: ubuntu-latest
    steps:
      - name: Pull & restart on host
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.DEPLOY_HOST }}
          key: ${{ secrets.DEPLOY_KEY }}
          script: |
            cd /opt/praestos
            echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u _ --password-stdin
            docker compose pull
            docker compose up -d --remove-orphans
```

---

## 8. Phased Delivery Roadmap

### Phase 0 — Foundation (Weeks 1–2)
- Repo scaffold, CI green, Docker Compose local stack running
- Goose migration pipeline, `orgs` + `users` tables, JWT auth endpoints
- Vite project, TanStack Router shell, shadcn/ui baseline, authenticated layout
- **Exit criteria:** `POST /auth/login` returns JWT; React app renders authenticated shell

### Phase 1a — Help Desk (Weeks 3–5)
- Ticket CRUD + comments + attachments
- Ticket list with status/priority filters, sort, pagination (TanStack Table)
- Ticket detail view with comment thread, internal notes toggle
- **Exit criteria:** Agent can open, update, comment, and close a ticket end-to-end

### Phase 1b — CRM (Weeks 6–8)
- Contacts, Accounts, Leads CRUD
- Lead → Contact conversion flow
- Contact/Account linked to tickets (sidebar relationship panel)
- **Exit criteria:** Full CRM entity lifecycle; ticket contact panel resolves live

### Phase 2 — Multi-tenancy Hardening (Weeks 9–10)
- RLS policies enabled and integration-tested with `SET LOCAL`
- Org onboarding flow (signup, slug generation, admin user seed)
- Tenant switcher in UI (for internal super-admin use)
- **Exit criteria:** Two orgs in same DB cannot see each other's data under any query path

### Phase 3 — Email & Portal (Weeks 11–14)
- Inbound email → ticket via SMTP/IMAP polling or webhook (Postmark/Mailgun)
- Client portal: limited-role login, ticket submit + view own tickets
- Email notifications on ticket state changes

### Phase 4 — Advanced Features (Weeks 15–20)
- SLA timers, escalation rules
- Custom fields (JSONB-backed, schema stored per org)
- Reporting dashboard (TanStack Table + Recharts)
- API key authentication for integrations (SuperOps, ConnectWise webhooks)

### Phase 5 — Data Migration (Weeks 21–24)
- vTiger MySQL → PostgreSQL migration scripts (per-module)
- Dry-run validation: row counts, field mapping audit
- Cutover runbook: maintenance window, migrate, smoke test, rollback plan

---

## 9. Key Decisions & Rationale

**Why not keep vTiger's MySQL schema?** PostgreSQL's RLS, native UUIDs, `TIMESTAMPTZ`, JSONB for custom fields, and `pgx/v5`'s typed scan API are all meaningfully better for a multi-tenant SaaS. The migration cost is one-time; the ops and correctness benefits are permanent.

**Why Chi over Gin/Echo?** Chi is `net/http`-compatible with zero interface wrapping. Middleware chains compose cleanly. Every stdlib tool (httptest, etc.) works without adapters. For a long-lived internal platform, this matters.

**Why TanStack Router over React Router?** Type-safe route params and search params are a major DX win with TypeScript. No more `useParams()` returning `string | undefined`. Search param state replaces most filter-related Zustand slices entirely.

**Why `httpOnly` cookies for JWT?** PraestOS handles client ticket data. No XSS vector to localStorage tokens. Refresh token rotation on every request handled by the Go auth middleware.

**Custom fields strategy:** Store schema in a `custom_field_definitions` table (per org, per entity type). Store values in a `custom_field_values` JSONB column on the entity row. Avoids ALTER TABLE per org while keeping queries fast with a GIN index.

---

## 10. Open Questions to Resolve Before Phase 0

1. **File storage:** local volume (single-tenant only) vs. S3-compatible (Minio self-hosted or AWS S3)? Recommendation: S3-compatible from day one — pluggable presigned URL approach.
2. **Email inbound:** will clients submit tickets via email, or portal-only in Phase 1? Drives whether Postmark/Mailgun webhook is in Phase 1 or Phase 3.
3. **Existing vTiger data:** is there live client data in the current MySQL vTiger instance that needs migration, or is this net-new? Determines whether Phase 5 is a real cutover or a fresh start.
4. **Auth provider:** self-managed JWT (as designed) vs. delegating to an OIDC provider (e.g. Keycloak, Auth0)? OIDC simplifies SSO for enterprise clients but adds a dependency.
