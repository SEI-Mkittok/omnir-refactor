# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

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
FROM golang:1.23-alpine AS dev

WORKDIR /app

RUN apk add --no-cache git wget

# Install air for hot reloading
RUN go install github.com/air-verse/air@latest

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

EXPOSE 8080
# Command is set in docker-compose.yml

# ── Production stage ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12 AS production

COPY --from=builder /build/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
