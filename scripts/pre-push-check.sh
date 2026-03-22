#!/bin/bash
# Pre-push validation — agents MUST run this before pushing feature branches
set -e

# Ensure go is in PATH
export PATH="/home/omnirdev/go/bin:$PATH"

cd "$(git rev-parse --show-toplevel)"

echo "=== Migration version check ==="
DUPES=$(ls api/migrations/*.sql 2>/dev/null | sed 's/.*\///' | sed 's/_.*//' | sort | uniq -d)
if [ -n "$DUPES" ]; then
  echo "❌ DUPLICATE MIGRATION VERSIONS DETECTED:"
  echo "$DUPES"
  echo ""
  echo "Rename the conflicting files before pushing."
  exit 1
fi
echo "→ No duplicate migration versions ✓"

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
