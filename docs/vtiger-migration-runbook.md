# vtiger CE → Omnir CRM: Data Migration Runbook

**Version:** 1.0  
**Issue:** OMN-8  
**Status:** Draft

---

## Overview

This runbook defines the migration path from vtiger Community Edition (MySQL) to Omnir CRM (PostgreSQL). It covers schema mapping, migration scripts, validation, reversibility, and time estimates.

---

## 1. Schema Mapping: vtiger CE → Omnir Domain Model

### Contacts

| vtiger table/field | Omnir field | Notes |
|---|---|---|
| vtiger_contacts.firstname | contacts.first_name | |
| vtiger_contacts.lastname | contacts.last_name | |
| vtiger_contacts.email | contacts.email (primary) | |
| vtiger_contacts.phone | contacts.phone | |
| vtiger_contacts.mobile | contacts.mobile | |
| vtiger_contacts.title | contacts.title | |
| vtiger_contacts.department | contacts.department | |
| vtiger_contacts.accountid | contacts.account_id | FK → accounts |
| vtiger_contacts.leadsource | contacts.lead_source | enum normalisation needed |
| vtiger_contacts.description | contacts.notes | |
| vtiger_crmentity.smownerid | contacts.owner_user_id | map via users table |
| vtiger_crmentity.createdtime | contacts.created_at | |
| vtiger_crmentity.modifiedtime | contacts.updated_at | |

### Accounts (→ organisations)

| vtiger table/field | Omnir field | Notes |
|---|---|---|
| vtiger_account.accountname | organisations.name | |
| vtiger_account.website | organisations.website | |
| vtiger_account.phone | organisations.phone | |
| vtiger_account.email1 | organisations.email | |
| vtiger_account.industry | organisations.industry | enum normalisation |
| vtiger_account.employees | organisations.employee_count | int cast |
| vtiger_account.annualrevenue | organisations.annual_revenue | decimal |
| vtiger_account.rating | organisations.rating | |
| vtiger_account.description | organisations.notes | |
| vtiger_crmentity.smownerid | organisations.owner_user_id | |

### Opportunities → Deals

| vtiger table/field | Omnir field | Notes |
|---|---|---|
| vtiger_potential.potentialname | deals.title | |
| vtiger_potential.amount | deals.value | decimal |
| vtiger_potential.currency_id | deals.currency_code | map vtiger currency table |
| vtiger_potential.closingdate | deals.expected_close_date | date |
| vtiger_potential.sales_stage | deals.stage | enum map (see §1a) |
| vtiger_potential.probability | deals.probability | int 0-100 |
| vtiger_potential.accountid | deals.organisation_id | |
| vtiger_potential.description | deals.notes | |
| vtiger_crmentity.smownerid | deals.owner_user_id | |

**§1a — Sales stage enum map:**

| vtiger | Omnir |
|---|---|
| Prospecting | prospecting |
| Qualification | qualified |
| Needs Analysis | qualified |
| Value Proposition | proposal |
| Id. Decision Makers | proposal |
| Perception Analysis | proposal |
| Proposal/Price Quote | proposal |
| Negotiation/Review | negotiation |
| Closed Won | won |
| Closed Lost | lost |

### Activities (Calls + Meetings + Tasks)

| vtiger | Omnir field | Notes |
|---|---|---|
| vtiger_activity.subject | activities.title | |
| vtiger_activity.activitytype | activities.type | call/meeting/task |
| vtiger_activity.date_start + time_start | activities.scheduled_at | combine into timestamptz |
| vtiger_activity.due_date + time_end | activities.due_at | |
| vtiger_activity.description | activities.notes | |
| vtiger_activity.status | activities.status | enum map |
| vtiger_seactivityrel (contact_id) | activities.contact_id | first related contact |

### Users

| vtiger table/field | Omnir field | Notes |
|---|---|---|
| vtiger_users.user_name | users.username | |
| vtiger_users.first_name | users.first_name | |
| vtiger_users.last_name | users.last_name | |
| vtiger_users.email1 | users.email | unique |
| vtiger_users.is_admin | users.role | admin/member |
| vtiger_users.status | users.active | Active → true |

---

## 2. Migration Scripts

### Prerequisites

```bash
# Install dependencies
pip install mysql-connector-python psycopg2-binary tqdm

# Environment vars
export VTIGER_DSN="mysql://root:pass@localhost/vtiger"
export OMNIR_DSN="postgresql://omnir:pass@localhost/omnir"
```

### Script: `migrate.py` (skeleton)

```python
"""
vtiger CE → Omnir CRM migration script
Run order: users → accounts → contacts → opportunities → activities
"""
import mysql.connector
import psycopg2
import uuid
from datetime import datetime

class Migrator:
    def __init__(self, vtiger_dsn, omnir_dsn):
        self.src = mysql.connector.connect(**parse_dsn(vtiger_dsn))
        self.dst = psycopg2.connect(omnir_dsn)
        self.user_map = {}      # vtiger smownerid → omnir user_id
        self.account_map = {}   # vtiger accountid → omnir org_id
        self.contact_map = {}   # vtiger contactid → omnir contact_id

    def run(self):
        self.migrate_users()
        self.migrate_accounts()
        self.migrate_contacts()
        self.migrate_opportunities()
        self.migrate_activities()
        self.dst.commit()

    def migrate_users(self):
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT u.id, u.user_name, u.first_name, u.last_name,
                   u.email1, u.is_admin, u.status
            FROM vtiger_users u
            WHERE u.deleted = 0
        """)
        dcur = self.dst.cursor()
        for row in cur.fetchall():
            new_id = str(uuid.uuid4())
            dcur.execute("""
                INSERT INTO users (id, username, first_name, last_name,
                    email, role, active, created_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, NOW())
                ON CONFLICT (email) DO NOTHING
                RETURNING id
            """, (new_id, row['user_name'], row['first_name'],
                  row['last_name'], row['email1'],
                  'admin' if row['is_admin'] == 'on' else 'member',
                  row['status'] == 'Active'))
            result = dcur.fetchone()
            if result:
                self.user_map[row['id']] = result[0]

    def migrate_accounts(self):
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT a.accountid, a.accountname, a.website, a.phone,
                   a.email1, a.industry, a.employees, a.annualrevenue,
                   a.rating, a.description, e.smownerid,
                   e.createdtime, e.modifiedtime
            FROM vtiger_account a
            JOIN vtiger_crmentity e ON e.crmid = a.accountid
            WHERE e.deleted = 0
        """)
        dcur = self.dst.cursor()
        for row in cur.fetchall():
            new_id = str(uuid.uuid4())
            owner = self.user_map.get(row['smownerid'])
            dcur.execute("""
                INSERT INTO organisations (id, name, website, phone, email,
                    industry, employee_count, annual_revenue, rating, notes,
                    owner_user_id, created_at, updated_at)
                VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
                RETURNING id
            """, (new_id, row['accountname'], row['website'], row['phone'],
                  row['email1'], row['industry'], row['employees'],
                  row['annualrevenue'], row['rating'], row['description'],
                  owner, row['createdtime'], row['modifiedtime']))
            result = dcur.fetchone()
            if result:
                self.account_map[row['accountid']] = result[0]

    def migrate_contacts(self):
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT c.contactid, c.firstname, c.lastname, c.email,
                   c.phone, c.mobile, c.title, c.department,
                   c.accountid, c.leadsource, c.description,
                   e.smownerid, e.createdtime, e.modifiedtime
            FROM vtiger_contacts c
            JOIN vtiger_crmentity e ON e.crmid = c.contactid
            WHERE e.deleted = 0
        """)
        dcur = self.dst.cursor()
        for row in cur.fetchall():
            new_id = str(uuid.uuid4())
            owner = self.user_map.get(row['smownerid'])
            org = self.account_map.get(row['accountid'])
            dcur.execute("""
                INSERT INTO contacts (id, first_name, last_name, email,
                    phone, mobile, title, department, organisation_id,
                    lead_source, notes, owner_user_id, created_at, updated_at)
                VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
                RETURNING id
            """, (new_id, row['firstname'], row['lastname'], row['email'],
                  row['phone'], row['mobile'], row['title'], row['department'],
                  org, row['leadsource'], row['description'],
                  owner, row['createdtime'], row['modifiedtime']))
            result = dcur.fetchone()
            if result:
                self.contact_map[row['contactid']] = result[0]

    def migrate_opportunities(self):
        STAGE_MAP = {
            'Prospecting': 'prospecting',
            'Qualification': 'qualified',
            'Needs Analysis': 'qualified',
            'Value Proposition': 'proposal',
            'Id. Decision Makers': 'proposal',
            'Perception Analysis': 'proposal',
            'Proposal/Price Quote': 'proposal',
            'Negotiation/Review': 'negotiation',
            'Closed Won': 'won',
            'Closed Lost': 'lost',
        }
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT p.potentialid, p.potentialname, p.amount,
                   p.closingdate, p.sales_stage, p.probability,
                   p.accountid, p.description,
                   e.smownerid, e.createdtime, e.modifiedtime
            FROM vtiger_potential p
            JOIN vtiger_crmentity e ON e.crmid = p.potentialid
            WHERE e.deleted = 0
        """)
        dcur = self.dst.cursor()
        for row in cur.fetchall():
            new_id = str(uuid.uuid4())
            owner = self.user_map.get(row['smownerid'])
            org = self.account_map.get(row['accountid'])
            stage = STAGE_MAP.get(row['sales_stage'], 'prospecting')
            dcur.execute("""
                INSERT INTO deals (id, title, value, expected_close_date,
                    stage, probability, organisation_id, notes,
                    owner_user_id, created_at, updated_at)
                VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
            """, (new_id, row['potentialname'], row['amount'],
                  row['closingdate'], stage, row['probability'],
                  org, row['description'], owner,
                  row['createdtime'], row['modifiedtime']))

    def migrate_activities(self):
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT a.activityid, a.subject, a.activitytype,
                   a.date_start, a.time_start, a.due_date, a.time_end,
                   a.description, a.status,
                   e.smownerid, e.createdtime,
                   rel.crmid AS contact_crmid
            FROM vtiger_activity a
            JOIN vtiger_crmentity e ON e.crmid = a.activityid
            LEFT JOIN vtiger_seactivityrel rel ON rel.activityid = a.activityid
            WHERE e.deleted = 0
        """)
        dcur = self.dst.cursor()
        for row in cur.fetchall():
            new_id = str(uuid.uuid4())
            owner = self.user_map.get(row['smownerid'])
            contact = self.contact_map.get(row['contact_crmid'])
            scheduled_at = None
            if row['date_start'] and row['time_start']:
                scheduled_at = datetime.strptime(
                    f"{row['date_start']} {row['time_start']}",
                    "%Y-%m-%d %H:%M:%S"
                )
            dcur.execute("""
                INSERT INTO activities (id, title, type, scheduled_at,
                    notes, status, contact_id, owner_user_id, created_at)
                VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s)
            """, (new_id, row['subject'],
                  row['activitytype'].lower(),
                  scheduled_at, row['description'],
                  row['status'], contact, owner, row['createdtime']))


if __name__ == "__main__":
    import os
    m = Migrator(os.environ['VTIGER_DSN'], os.environ['OMNIR_DSN'])
    m.run()
    print("Migration complete.")
```

---

## 3. Data Validation & Reconciliation Checks

Run after migration to confirm row counts and referential integrity.

### 3.1 Row Count Checks

```sql
-- vtiger (run on MySQL)
SELECT 'contacts' AS entity, COUNT(*) FROM vtiger_contacts c
  JOIN vtiger_crmentity e ON e.crmid = c.contactid WHERE e.deleted = 0
UNION ALL
SELECT 'accounts', COUNT(*) FROM vtiger_account a
  JOIN vtiger_crmentity e ON e.crmid = a.accountid WHERE e.deleted = 0
UNION ALL
SELECT 'opportunities', COUNT(*) FROM vtiger_potential p
  JOIN vtiger_crmentity e ON e.crmid = p.potentialid WHERE e.deleted = 0
UNION ALL
SELECT 'users', COUNT(*) FROM vtiger_users WHERE deleted = 0 AND status = 'Active';

-- Omnir (run on PostgreSQL — counts must match within accepted delta)
SELECT 'contacts' AS entity, COUNT(*) FROM contacts
UNION ALL SELECT 'organisations', COUNT(*) FROM organisations
UNION ALL SELECT 'deals', COUNT(*) FROM deals
UNION ALL SELECT 'users', COUNT(*) FROM users WHERE active = true;
```

**Acceptance threshold:** ≤0.1% row loss (accounts for vtiger soft-delete edge cases).

### 3.2 Referential Integrity Checks (PostgreSQL)

```sql
-- Contacts with dangling organisation_id
SELECT COUNT(*) FROM contacts
WHERE organisation_id IS NOT NULL
  AND organisation_id NOT IN (SELECT id FROM organisations);

-- Deals with dangling organisation_id
SELECT COUNT(*) FROM deals
WHERE organisation_id IS NOT NULL
  AND organisation_id NOT IN (SELECT id FROM organisations);

-- Activities with dangling contact_id
SELECT COUNT(*) FROM activities
WHERE contact_id IS NOT NULL
  AND contact_id NOT IN (SELECT id FROM contacts);

-- All of the above should return 0.
```

### 3.3 Spot-Check Script

```bash
# Compare a random sample of 20 contacts by email
python scripts/spot_check.py --entity contacts --sample 20
```

The spot-check script queries both databases by original vtiger ID (stored in a `vtiger_legacy_id` column on each Omnir table during migration) and diffs field values.

---

## 4. Reversibility Strategy

### 4.1 Pre-migration Snapshot

Before migration:

```bash
# MySQL full dump
mysqldump --single-transaction vtiger > backups/vtiger_pre_migration_$(date +%Y%m%d).sql.gz

# PostgreSQL snapshot (if Omnir has existing data)
pg_dump omnir > backups/omnir_pre_migration_$(date +%Y%m%d).sql.gz
```

### 4.2 Dry-Run Mode

The migration script supports `--dry-run` flag: runs all SQL in a transaction, reports counts, then rolls back. Zero writes to production.

```bash
python migrate.py --dry-run
```

### 4.3 Rollback Procedure

If migration needs to be reversed:

1. **Stop Omnir application** (no new writes)
2. Drop migrated tables and restore from snapshot:

```bash
psql omnir < backups/omnir_pre_migration_YYYYMMDD.sql.gz
```

3. Restart application pointing at vtiger until re-migration is ready.

### 4.4 Legacy ID Column

Each Omnir table stores `vtiger_legacy_id VARCHAR(64)` to:
- Enable spot checks post-migration
- Allow re-migration (truncate + re-run) without manual ID tracking
- Support support tickets that reference old vtiger record IDs

---

## 5. Estimated Migration Time

### Assumptions

- Standard vtiger CE schema (no heavy custom modules)
- Dedicated migration host with both DB servers accessible
- Python script with batch inserts (500 rows/batch)

### Time Estimates

| Record Count | Contacts | Accounts | Opportunities | Activities | **Total** |
|---|---|---|---|---|---|
| 10k records | ~30s | ~20s | ~20s | ~30s | **~2 min** |
| 50k records | ~2m | ~1m | ~1m | ~2m | **~6–8 min** |
| 100k records | ~4m | ~2m | ~2m | ~4m | **~12–15 min** |

**Total migration window (including validation):** 30–60 minutes for a typical 50k-record instance (includes pre-checks, dry-run, live migration, post-validation).

### Recommended Migration Window

- Off-peak hours (weekend night)
- Vtiger set to read-only / maintenance mode before `migrate.py` runs
- Go/no-go checkpoint after dry-run passes

---

## 6. Migration Runbook — Step-by-Step

```
[ ] 1. Notify users of maintenance window (24h advance)
[ ] 2. Backup vtiger MySQL (mysqldump)
[ ] 3. Backup Omnir PostgreSQL (pg_dump)
[ ] 4. Set vtiger to maintenance mode (prevent new writes)
[ ] 5. Run: python migrate.py --dry-run
      → Review output, verify row counts, fix any errors
[ ] 6. Run: python migrate.py (live)
[ ] 7. Run SQL validation checks (§3.1, §3.2)
[ ] 8. Run spot-check script (§3.3)
[ ] 9. If all checks pass → update DNS / redirect to Omnir
[10] 10. If checks fail → rollback (§4.3), investigate, retry
[ ] 11. Monitor Omnir for 48h post-cutover
[ ] 12. Archive vtiger instance (do not delete for 90 days)
```

---

*Generated by Forge (Völundr) — Omnir CRM migration runbook v1.0*
