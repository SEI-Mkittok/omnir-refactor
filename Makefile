# Omnir CRM — Dev Makefile
# Usage: make <target>
# Requires: docker, docker compose, go 1.22+, node 20+

.PHONY: help up down build logs shell-api shell-db migrate seed test test-api test-frontend lint fmt

COMPOSE := docker compose -f infra/docker-compose.yml

# Default target
help:
	@echo ""
	@echo "  Omnir CRM Dev Commands"
	@echo "  ────────────────────────────────────────"
	@echo "  make up            Start all services (build if needed)"
	@echo "  make down          Stop and remove containers"
	@echo "  make build         Rebuild all Docker images"
	@echo "  make logs          Tail logs from all services"
	@echo "  make shell-api     Shell into the API container"
	@echo "  make shell-db      psql into the DB"
	@echo "  make migrate       Run pending DB migrations"
	@echo "  make seed          Seed DB with sample data"
	@echo "  make test          Run all tests"
	@echo "  make test-api      Run Go API tests"
	@echo "  make test-frontend Run frontend tests"
	@echo "  make lint          Run all linters"
	@echo "  make fmt           Format all code"
	@echo ""

up:
	$(COMPOSE) up

up-d:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down -v

build:
	$(COMPOSE) build --no-cache

logs:
	$(COMPOSE) logs -f

# Shell access
shell-api:
	$(COMPOSE) exec api sh

shell-db:
	$(COMPOSE) exec postgres psql -U omnir -d omnir_crm

# Migrations
migrate:
	$(COMPOSE) exec api goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	$(COMPOSE) exec api goose -dir migrations postgres "$$DATABASE_URL" down

migrate-status:
	$(COMPOSE) exec api goose -dir migrations postgres "$$DATABASE_URL" status

# Testing
test: test-api test-frontend

test-api:
	cd api && go test ./... -v -count=1 -race

test-integration:
	cd api && TEST_DATABASE_URL=$${TEST_DATABASE_URL:-postgres://localhost/omnir_crm_test?sslmode=disable} \
		go test -tags integration ./internal/repository/postgres/... -v -count=1

test-frontend:
	cd web && npm run test -- --run

# Code quality
lint:
	cd api && golangci-lint run ./...
	cd web && npm run lint

fmt:
	cd api && gofmt -w .
	cd web && npm run fmt

# Setup for first-time clone
setup:
	@echo "→ Copying .env files..."
	@cp -n .env.example .env 2>/dev/null && echo "  Created .env" || echo "  .env already exists, skipping"
	@cp -n api/.env.example api/.env 2>/dev/null && echo "  Created api/.env" || echo "  api/.env already exists"
	@cp -n web/.env.example web/.env.local 2>/dev/null && echo "  Created web/.env.local" || echo "  web/.env.local already exists"
	@echo "→ Done. Run 'make up' to start the stack."
