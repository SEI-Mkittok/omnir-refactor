# Omnir CRM — Dev Makefile
# Usage: make <target>
# Requires: docker, docker compose, go 1.22+, node 20+

.PHONY: help up down build logs shell-api shell-db migrate seed test test-api test-frontend lint fmt

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
	docker compose up

up-d:
	docker compose up -d

down:
	docker compose down -v

build:
	docker compose build --no-cache

logs:
	docker compose logs -f

# Shell access
shell-api:
	docker compose exec api sh

shell-db:
	docker compose exec postgres psql -U omnir -d omnir_crm

# Migrations
migrate:
	docker compose exec api goose -dir db/migrations postgres "$$DATABASE_URL" up

migrate-down:
	docker compose exec api goose -dir db/migrations postgres "$$DATABASE_URL" down

migrate-status:
	docker compose exec api goose -dir db/migrations postgres "$$DATABASE_URL" status

# Testing
test: test-api test-frontend

test-api:
	cd omnir-go-backend && go test ./... -v -count=1 -race

test-integration:
	cd omnir-go-backend && TEST_DATABASE_URL=$${TEST_DATABASE_URL:-postgres://localhost/omnir_crm_test?sslmode=disable} \
		go test -tags integration ./internal/repository/postgres/... -v -count=1

test-frontend:
	cd omnir-frontend && npm run test -- --run

# Code quality
lint:
	cd omnir-go-backend && golangci-lint run ./...
	cd omnir-frontend && npm run lint

fmt:
	cd omnir-go-backend && gofmt -w .
	cd omnir-frontend && npm run fmt

# Setup for first-time clone
setup:
	@echo "→ Copying .env files..."
	@cp -n .env.example .env 2>/dev/null && echo "  Created .env" || echo "  .env already exists, skipping"
	@cp -n omnir-go-backend/.env.example omnir-go-backend/.env 2>/dev/null && echo "  Created omnir-go-backend/.env" || echo "  omnir-go-backend/.env already exists"
	@cp -n omnir-frontend/.env.example omnir-frontend/.env.local 2>/dev/null && echo "  Created omnir-frontend/.env.local" || echo "  omnir-frontend/.env.local already exists"
	@echo "→ Done. Run 'make up' to start the stack."
