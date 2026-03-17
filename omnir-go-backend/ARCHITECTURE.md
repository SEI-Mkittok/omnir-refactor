# Omnir CRM — Go Backend Architecture

**Issue:** OMN-3  
**Status:** Drafted  
**Author:** Forge (AI Dev Agent)

---

## 1. Overview

Replace the vtiger PHP monolith with a clean Go REST API backend. Stack:

- **Language:** Go 1.22+
- **HTTP Framework:** [Chi](https://github.com/go-chi/chi) (lightweight, idiomatic)
- **Database:** PostgreSQL 16 via [pgx/v5](https://github.com/jackc/pgx)
- **Migrations:** [goose](https://github.com/pressly/goose)
- **Auth:** JWT (access + refresh tokens) via [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- **Config:** Environment variables via [godotenv](https://github.com/joho/godotenv)
- **API Contract:** OpenAPI 3.1 (spec-first)
- **Code generation:** [oapi-codegen](https://github.com/deepmap/oapi-codegen) (optional, for type-safe handlers)

---

## 2. Domain Model

### Core Entities

```
User
├── id (UUID)
├── email (unique)
├── name
├── role (admin | user | viewer)
├── avatar_url
├── created_at / updated_at / deleted_at

Contact
├── id (UUID)
├── first_name, last_name
├── email (unique nullable)
├── phone
├── account_id (FK → Account, nullable)
├── owner_id (FK → User)
├── lead_source
├── stage (lead | prospect | customer | churned)
├── tags []string
├── custom_fields JSONB
├── created_at / updated_at / deleted_at

Account
├── id (UUID)
├── name
├── domain
├── industry
├── size (employees enum)
├── owner_id (FK → User)
├── tags []string
├── custom_fields JSONB
├── created_at / updated_at / deleted_at

Deal
├── id (UUID)
├── title
├── value_cents (int64)
├── currency (ISO 4217)
├── stage (lead | qualified | proposal | negotiation | closed_won | closed_lost)
├── probability (0-100)
├── expected_close_date
├── contact_id (FK → Contact, nullable)
├── account_id (FK → Account, nullable)
├── owner_id (FK → User)
├── pipeline_id (FK → Pipeline)
├── custom_fields JSONB
├── created_at / updated_at / deleted_at

Activity
├── id (UUID)
├── type (call | meeting | task | email | note)
├── title
├── description
├── status (pending | completed | cancelled)
├── due_at (nullable)
├── completed_at (nullable)
├── contact_id (FK → Contact, nullable)
├── account_id (FK → Account, nullable)
├── deal_id (FK → Deal, nullable)
├── owner_id (FK → User)
├── created_at / updated_at / deleted_at

Pipeline
├── id (UUID)
├── name
├── stages []PipelineStage (JSONB ordered)
├── created_at / updated_at

Note (unified comment/note system)
├── id (UUID)
├── body (text)
├── author_id (FK → User)
├── entity_type (contact | account | deal | activity)
├── entity_id (UUID)
├── created_at / updated_at
```

### Relationship Rules
- Contact → Account: many-to-one (contact belongs to one company)
- Deal → Contact + Account: optional many-to-one for each
- Activity → Contact, Account, Deal: polymorphic association via entity_type/entity_id
- Note → any entity: same polymorphic pattern

---

## 3. REST API Contract

Base path: `/api/v1`

### Auth
```
POST   /auth/login           → { access_token, refresh_token, expires_in }
POST   /auth/refresh         → { access_token, expires_in }
POST   /auth/logout
GET    /auth/me              → User
```

### Users
```
GET    /users                → paginated list
GET    /users/{id}
PATCH  /users/{id}
DELETE /users/{id}
```

### Contacts
```
GET    /contacts             → paginated, filterable
POST   /contacts
GET    /contacts/{id}
PATCH  /contacts/{id}
DELETE /contacts/{id}
GET    /contacts/{id}/activities
GET    /contacts/{id}/deals
GET    /contacts/{id}/notes
POST   /contacts/{id}/notes
```

### Accounts
```
GET    /accounts
POST   /accounts
GET    /accounts/{id}
PATCH  /accounts/{id}
DELETE /accounts/{id}
GET    /accounts/{id}/contacts
GET    /accounts/{id}/deals
GET    /accounts/{id}/notes
POST   /accounts/{id}/notes
```

### Deals
```
GET    /deals
POST   /deals
GET    /deals/{id}
PATCH  /deals/{id}
DELETE /deals/{id}
GET    /deals/{id}/activities
GET    /deals/{id}/notes
POST   /deals/{id}/notes
```

### Activities
```
GET    /activities
POST   /activities
GET    /activities/{id}
PATCH  /activities/{id}
DELETE /activities/{id}
```

### Pipelines
```
GET    /pipelines
POST   /pipelines
GET    /pipelines/{id}
PATCH  /pipelines/{id}
DELETE /pipelines/{id}
GET    /pipelines/{id}/deals   → deals grouped by stage
```

### Search
```
GET    /search?q=...&types=contacts,accounts,deals   → unified search
```

### Pagination & Filtering
All list endpoints support:
```
?page=1&limit=50
?sort=created_at&order=desc
?owner_id=...
?q=...     (full-text search on key fields)
```

---

## 4. Database Layer Strategy

### Connection
- Pool via `pgxpool` (pgx v5)
- Connection string from env: `DATABASE_URL=postgres://...`
- Max conns: 25, idle timeout: 30s

### Pattern: Repository Interface
```go
type ContactRepository interface {
    Create(ctx context.Context, c *Contact) (*Contact, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Contact, error)
    Update(ctx context.Context, id uuid.UUID, patch ContactPatch) (*Contact, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filter ContactFilter) ([]*Contact, int, error)
}
```

Each entity gets a `postgres` package implementing the interface. This lets us unit-test with mocks or swap to a test DB.

### Migrations (goose)
Directory: `db/migrations/`  
Format: `YYYYMMDDHHMMSS_description.sql`

Initial migrations:
1. `create_users`
2. `create_contacts`
3. `create_accounts`
4. `create_deals`
5. `create_activities`
6. `create_pipelines`
7. `create_notes`
8. `add_indexes_and_fks`

Run via: `goose -dir db/migrations postgres $DATABASE_URL up`

### Soft Deletes
All major entities have `deleted_at TIMESTAMPTZ NULL`. Queries always include `WHERE deleted_at IS NULL` unless explicitly fetching deleted records.

### Custom Fields
`custom_fields JSONB` on contacts, accounts, deals. Module system can define schemas for validation.

---

## 5. Auth Model

### JWT Strategy
- **Access token:** Short-lived (15 min), signed HS256 or RS256
- **Refresh token:** Long-lived (30 days), stored in DB (refresh_tokens table), rotated on use
- **Claims:** `sub` (user_id), `role`, `company_id`, `exp`, `iat`

### Middleware Stack (Chi)
```
Router
 └── /api/v1
      ├── /auth/*        (public)
      └── /*             → AuthMiddleware
                              └── RoleMiddleware (route-level)
```

### AuthMiddleware
1. Extract `Authorization: Bearer <token>`
2. Validate JWT signature + expiry
3. Load user from DB (or cache)
4. Inject `*User` into context

### Role Model
```
admin  → full CRUD + user management + settings
user   → full CRUD on owned/assigned records
viewer → read-only
```

Enforced via middleware or explicit checks in handlers.

---

## 6. Module System

### Goal
Allow CRM features to be packaged as modules (think: Email Integration, Billing, Custom Fields, etc.) without forking the core.

### Module Interface
```go
type Module interface {
    Name() string
    Version() string
    Register(r chi.Router, db *pgxpool.Pool, cfg *Config)
    Migrate(db *pgxpool.Pool) error
}
```

### Module Registry
```go
var registry []Module

func RegisterModule(m Module) {
    registry = append(registry, m)
}

func BootModules(r chi.Router, db *pgxpool.Pool, cfg *Config) {
    for _, m := range registry {
        m.Migrate(db)
        m.Register(r, db, cfg)
    }
}
```

### Module Convention
Each module lives in `modules/<name>/`:
```
modules/
  email/
    module.go       ← implements Module interface
    handler.go
    repository.go
    migrations/
  billing/
    ...
  custom_fields/
    ...
```

Modules register their own routes under `/api/v1/modules/<name>/` to avoid conflicts.

---

## 7. Project Structure

```
omnir-crm-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── jwt.go
│   │   ├── middleware.go
│   │   └── refresh.go
│   ├── domain/
│   │   ├── contact.go
│   │   ├── account.go
│   │   ├── deal.go
│   │   ├── activity.go
│   │   ├── user.go
│   │   └── pipeline.go
│   ├── handler/
│   │   ├── contacts.go
│   │   ├── accounts.go
│   │   ├── deals.go
│   │   ├── activities.go
│   │   ├── auth.go
│   │   └── search.go
│   ├── repository/
│   │   ├── interface.go
│   │   └── postgres/
│   │       ├── contacts.go
│   │       ├── accounts.go
│   │       └── ...
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── logger.go
│   │   └── cors.go
│   └── config/
│       └── config.go
├── modules/
│   └── (pluggable modules here)
├── db/
│   └── migrations/
├── api/
│   └── openapi.yaml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 8. Key Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Framework | Chi | Stdlib-compatible, composable, no magic |
| ORM | Raw pgx queries | Performance + control; no ORM overhead |
| Migrations | goose | SQL-first, simple CLI |
| Auth | JWT + refresh tokens | Stateless API, mobile-friendly |
| Soft deletes | deleted_at column | Audit trail, recoverable records |
| Custom fields | JSONB | Flexible without schema churn |
| Module system | Interface + registry | Clean extension without forking |
| Pagination | cursor or page-based | Start page-based, migrate to cursor if needed |
| Error format | RFC 7807 (Problem Details) | Standard, machine-readable errors |

---

## Next Steps (Suggested Issues)

1. **OMN-4:** Scaffold Go project, set up Chi router, DB pool, basic health endpoint
2. **OMN-5:** Implement auth (JWT login/refresh/logout + middleware)
3. **OMN-6:** Contacts CRUD with repository pattern
4. **OMN-7:** Accounts + Deals CRUD
5. **OMN-8:** Activities CRUD
6. **OMN-9:** Unified search endpoint
7. **OMN-10:** Module system scaffold + custom fields module
