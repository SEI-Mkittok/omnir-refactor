# Omnir CRM

Modern CRM built on Go + React. Replaces vtiger CE with a clean, fast, maintainable stack.
[![CI](https://github.com/SEI-Mkittok/omnir-refactor/actions/workflows/ci.yml/badge.svg?branch=develop)](https://github.com/SEI-Mkittok/omnir-refactor/actions/workflows/ci.yml)
---

## Quickstart (< 5 minutes)

**Prerequisites:** Docker Desktop (or Docker + Compose), Git.

```bash
# 1. Clone
git clone https://github.com/omnir/omnir-crm.git
cd omnir-crm

# 2. Copy env files (safe defaults for local dev)
make setup

# 3. Start the stack
make up
```

That's it. Services:

| Service  | URL                        | Notes                   |
|----------|----------------------------|-------------------------|
| Frontend | http://localhost:5173      | React/Vite (hot reload) |
| API      | http://localhost:8080      | Go/Chi (hot reload)     |
| Health   | http://localhost:8080/health | `{"status":"ok"}`     |
| DB       | localhost:5432             | PostgreSQL 16           |
| pgAdmin  | http://localhost:5050      | `make up -- --profile tools` |

---

## Common Commands

```bash
make up          # Start all services
make down        # Stop + remove volumes
make logs        # Tail all logs
make shell-api   # Shell into API container
make shell-db    # psql into the database
make migrate     # Run pending DB migrations
make test        # Run all tests (API + frontend)
make lint        # Lint all code
```

---

## Project Structure

```
omnir-crm/
├── omnir-go-backend/    # Go REST API (Chi + pgx + JWT)
├── omnir-frontend/      # React + TypeScript + Vite
├── .github/workflows/   # CI/CD (GitHub Actions)
├── scripts/             # DB init, seed scripts
├── docker-compose.yml   # Local dev stack
├── docker-compose.prod.yml  # Production deploy
├── Makefile             # Dev shortcuts
└── .env.example         # Env config template
```

---

## Architecture

- **API:** Go 1.22, Chi router, pgx/v5 (raw SQL, no ORM), JWT auth, goose migrations
- **Frontend:** React 18, TypeScript, Vite, shadcn/ui, TanStack Query + Router, Zustand
- **Database:** PostgreSQL 16
- **CI:** GitHub Actions (lint → test → build → Docker → staging deploy)

See `omnir-go-backend/ARCHITECTURE.md` and `docs/frontend-architecture.md` for full details.

---

## Environment Variables

Copy `.env.example` → `.env` (root) for Docker Compose.  
See `omnir-go-backend/.env.example` for API vars.  
See `omnir-frontend/.env.example` for frontend vars.

**Never commit `.env` files with real secrets.**  
In CI/staging, inject secrets via GitHub Secrets + environment vars.

---

## CI/CD

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | Push/PR to main, develop | Lint + test + build check |
| `deploy-staging.yml` | Push to develop | Build → push images → deploy to staging |

---

## Contributing

1. Branch from `develop`
2. PRs require CI green
3. Squash merge to `develop`, rebase-merge to `main` for releases
