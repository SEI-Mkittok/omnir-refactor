# Go dependency vendoring (offline-friendly)

If module downloads are blocked by network/proxy policy, commit vendored dependencies to the repository.

## Where files should go

From the repository root (`/workspace/omnir-refactor`), run vendoring inside `api/`.

Committed files should be:

- `api/vendor/modules.txt`
- `api/vendor/**` (all vendored packages)

Do **not** commit module cache files from `~/go/pkg/mod` or machine-level `go env` config.

## How to generate/update vendor files

```bash
cd api
go mod tidy
go mod vendor
```

Then commit:

```bash
git add go.mod go.sum vendor/
git commit -m "Vendor Go dependencies for offline builds"
```

## CI behavior in this repo

CI is configured to automatically use vendored dependencies when `api/vendor/` exists by setting:

- `GOFLAGS=-mod=vendor`

That allows `go test`, `go build`, and related commands to resolve dependencies from `api/vendor` without downloading modules.
