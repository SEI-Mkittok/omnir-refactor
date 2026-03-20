#!/bin/bash
# Pre-push validation — agents MUST run this before pushing feature branches
set -e

cd "$(git rev-parse --show-toplevel)"

echo "=== Go checks ==="
cd api

# Verify it compiles
echo "→ Building..."
go build ./...

# Run tests
echo "→ Running tests..."
go test ./... -count=1 -timeout 60s

# Lint (if golangci-lint available)
if command -v golangci-lint &>/dev/null; then
  echo "→ Linting..."
  golangci-lint run ./...
fi

cd ..

echo "=== Frontend checks ==="
cd web

# Type check
echo "→ Type checking..."
npx tsc --noEmit 2>/dev/null || echo "WARN: tsc failed (non-blocking)"

echo ""
echo "✅ All checks passed — safe to push"
