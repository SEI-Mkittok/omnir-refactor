# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install dependencies (git needed for some modules)
RUN apk add --no-cache git

# Cache module downloads separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /build/server \
    ./cmd/api

# ── Dev stage (hot reload via air) ───────────────────────────────────────────
FROM golang:1.24-alpine AS dev

WORKDIR /app

RUN apk add --no-cache git wget

# Install air for hot reloading
RUN go install github.com/air-verse/air@latest

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

EXPOSE 8080
# Command is set in docker-compose.yml

# ── Migration stage ───────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS migrator
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# ── Production stage ──────────────────────────────────────────────────────────
FROM alpine:3.19 AS production

RUN apk add --no-cache ca-certificates curl

COPY --from=builder /build/server /server
COPY --from=migrator /root/go/bin/goose /goose
COPY db/migrations /db/migrations

EXPOSE 8080
ENTRYPOINT ["/server"]
