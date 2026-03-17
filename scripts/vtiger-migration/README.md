# vtiger CE → Omnir CRM Migration Scripts

Migrates data from vtiger Community Edition (MySQL) to Omnir CRM (PostgreSQL).

See `docs/vtiger-migration-runbook.md` for the full runbook, schema map, and rollback procedure.

## Prerequisites

```bash
pip install -r requirements.txt
```

## Environment variables

| Variable       | Description                                             |
|----------------|---------------------------------------------------------|
| `VTIGER_DSN`   | `mysql://user:pass@host:port/dbname`                    |
| `OMNIR_DSN`    | `postgresql://user:pass@host:port/dbname`               |

Alternatively, set individual vtiger vars: `VTIGER_HOST`, `VTIGER_PORT`, `VTIGER_USER`, `VTIGER_PASSWORD`, `VTIGER_DATABASE`.

## Pre-migration: apply the vtiger_legacy_id migration

```bash
cd omnir-go-backend
goose -dir db/migrations postgres "$OMNIR_DSN" up
```

This adds `vtiger_legacy_id` columns to all entity tables, enabling idempotent re-runs.

## Usage

### Step 1 — Dry run (recommended first)

```bash
python migrate.py --dry-run
```

Runs the full migration inside a transaction and rolls back. Prints row counts and any errors — zero writes to the database.

### Step 2 — Live migration

```bash
python migrate.py
```

### Step 3 — Validate

```bash
python validate.py --sample 50 --fail-on-error
```

Checks row counts (≤0.1% loss threshold), referential integrity, and spot-checks a random sample of each entity.

### Optional: target a specific org

```bash
python migrate.py --org-id "YOUR-ORG-UUID"
python validate.py  # (uses same tables, org_id is not checked in validation)
```

## Migration order

`users → accounts → contacts → deals → activities`

Each entity depends on the previous step's ID remapping table (user_map, account_map, contact_map). Do not change the order.

## Idempotency

All inserts use `ON CONFLICT (vtiger_legacy_id) DO NOTHING`. Re-running after a partial failure will skip already-migrated rows and continue from where it left off.

## Exit criteria

- `validate.py` exits 0
- Row counts within 0.1% of source
- No referential integrity violations
- Spot-checks pass
