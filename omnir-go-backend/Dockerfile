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

# Download pre-built goose binary for migrations
RUN wget -qO /build/goose \
    https://github.com/pressly/goose/releases/download/v3.19.2/goose_linux_x86_64 && \
    chmod +x /build/goose

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

# ── Production stage ──────────────────────────────────────────────────────────
FROM alpine:3.19 AS production

RUN apk add --no-cache ca-certificates curl

# Build goose in builder stage and copy here
COPY --from=builder /build/server /server
COPY --from=builder /build/goose /goose
COPY db/migrations /db/migrations

EXPOSE 8080
ENTRYPOINT ["/server"]
