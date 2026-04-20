# Omnir CRM

Modern CRM built on Go + React. Replaces vtiger CE with a clean, fast, maintainable stack.
[![CI](https://github.com/SEI-Mkittok/omnir-refactor/actions/workflows/ci.yml/badge.svg?branch=develop)](https://github.com/SEI-Mkittok/omnir-refactor/actions/workflows/ci.yml)

---

## Setup & Environment (Local Development)

### Prerequisites

- Docker Desktop (or Docker Engine + Compose plugin)
- Git
- Optional for running outside containers: Go 1.22+ and Node 20+

### 1) Clone and bootstrap env files

```bash
git clone https://github.com/SEI-Mkittok/omnir-refactor.git
cd omnir-refactor
make setup
```

`make setup` creates these files only if they do not already exist:

- `.env`
- `api/.env`
- `web/.env.local`

### 2) Start the development stack

```bash
make up
```

Stack is defined in `infra/docker-compose.yml` and starts:

- `postgres` (PostgreSQL 16)
- `api` (Go API with hot reload via Air)
- `frontend` (React/Vite)
- `minio` (S3-compatible storage)
- `mailhog` (local SMTP + inbox UI)

### 3) Verify services

- Frontend: http://localhost:5173
- API: http://localhost:8080
- API health: http://localhost:8080/health
- Postgres: `localhost:5432`
- MinIO API: http://localhost:9000
- MinIO Console: http://localhost:9001
- MailHog UI: http://localhost:8025
- pgAdmin (optional profile): http://localhost:5050

To include pgAdmin:

```bash
docker compose -f infra/docker-compose.yml --profile tools up
```

---

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

---

## Environment Variables

Primary references:

- Root compose env template: `.env.example`
- API local env template: `api/.env.example`
- Frontend local env template: `web/.env.example`
- Extended ops reference: `docs/ops/env-reference.md`

Never commit real secrets in `.env` files.

---

## Project Structure

```text
omnir-refactor/
├── api/                  # Go REST API
├── web/                  # React + TypeScript + Vite frontend
├── infra/                # Docker Compose + Dockerfiles
├── scripts/              # Utility scripts and DB init
├── docs/                 # Architecture and operations docs
├── .github/workflows/    # CI/CD workflows
└── Makefile              # Local development shortcuts
```

---

## Architecture

- **API:** Go 1.22, Chi router, pgx/v5, JWT auth, goose migrations
- **Frontend:** React 18, TypeScript, Vite, TanStack Query/Router, Zustand
- **Database:** PostgreSQL 16
- **CI:** GitHub Actions (lint → test → build)

See `docs/frontend-architecture.md` and API docs in `api/` for implementation details.

---

## CI/CD

| Workflow | Trigger | Purpose |
|---|---|---|
| `ci.yml` | Push/PR to `main`, `develop` | Lint + test + build validation |
| `deploy-staging.yml` | Push to `develop` | Build images and deploy to staging |

---

## Contributing

1. Branch from `develop`
2. Keep CI green
3. Open PR with clear scope and test evidence
