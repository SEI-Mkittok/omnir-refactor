#!/usr/bin/env python3
"""
vtiger CE → Omnir CRM migration script
Reference: docs/vtiger-migration-runbook.md

Run order: users → accounts → contacts → deals → activities

Usage:
    python migrate.py [--dry-run] [--org-id UUID]

Environment variables:
    VTIGER_DSN    MySQL DSN: mysql://user:pass@host/dbname
    OMNIR_DSN     PostgreSQL DSN: postgresql://user:pass@host/dbname

    Or individual vars:
    VTIGER_HOST, VTIGER_PORT, VTIGER_USER, VTIGER_PASSWORD, VTIGER_DATABASE
    OMNIR_DSN (postgres connection string)

Idempotency:
    All inserts use ON CONFLICT (vtiger_legacy_id) DO NOTHING.
    Safe to re-run after partial failures; already-migrated rows are skipped.
"""

import argparse
import os
import re
import sys
from datetime import datetime

import mysql.connector
import psycopg2
import psycopg2.extras
from tqdm import tqdm

# Default org_id for the self-hosted Omnir deployment
DEFAULT_ORG_ID = "00000000-0000-0000-0000-000000000002"
DEFAULT_PIPELINE_ID = "00000000-0000-0000-0000-000000000001"

STAGE_MAP = {
    "Prospecting": "lead",
    "Qualification": "qualified",
    "Needs Analysis": "qualified",
    "Value Proposition": "proposal",
    "Id. Decision Makers": "proposal",
    "Perception Analysis": "proposal",
    "Proposal/Price Quote": "proposal",
    "Negotiation/Review": "negotiation",
    "Closed Won": "closed_won",
    "Closed Lost": "closed_lost",
}

TICKET_STATUS_MAP = {
    "Open": "open",
    "In Progress": "in_progress",
    "Wait For Response": "pending",
    "Hold": "pending",
    "Closed": "closed",
}

TICKET_PRIORITY_MAP = {
    "Low": "low",
    "Medium": "medium",
    "High": "high",
    "Critical": "critical",
}

ACTIVITY_TYPE_MAP = {
    "Call": "call",
    "Meeting": "meeting",
    "Task": "task",
    "Emails": "email",
    "Email": "email",
}

EMPLOYEE_SIZE_MAP = [
    (10, "1-10"),
    (50, "11-50"),
    (200, "51-200"),
    (500, "201-500"),
    (None, "501+"),
]


def parse_mysql_dsn(dsn: str) -> dict:
    """Parse mysql://user:pass@host:port/dbname into connector kwargs."""
    m = re.match(
        r"mysql://(?P<user>[^:]+):(?P<password>[^@]+)@(?P<host>[^:/]+)(?::(?P<port>\d+))?/(?P<database>.+)",
        dsn,
    )
    if not m:
        raise ValueError(f"Invalid VTIGER_DSN: {dsn!r}")
    d = m.groupdict()
    result = {k: v for k, v in d.items() if v is not None}
    if "port" in result:
        result["port"] = int(result["port"])
    return result


def employees_to_size(n) -> str | None:
    if n is None:
        return None
    try:
        n = int(n)
    except (TypeError, ValueError):
        return None
    for threshold, label in EMPLOYEE_SIZE_MAP:
        if threshold is None or n <= threshold:
            return label
    return "501+"


class Migrator:
    def __init__(self, vtiger_config: dict, omnir_dsn: str, org_id: str, dry_run: bool = False):
        self.org_id = org_id
        self.dry_run = dry_run

        print("Connecting to vtiger MySQL…")
        self.src = mysql.connector.connect(**vtiger_config)

        print("Connecting to Omnir PostgreSQL…")
        self.dst = psycopg2.connect(omnir_dsn)
        psycopg2.extras.register_uuid()

        # ID remapping tables
        self.user_map: dict[int, str] = {}     # vtiger user id   → omnir user UUID
        self.account_map: dict[int, str] = {}  # vtiger accountid → omnir account UUID
        self.contact_map: dict[int, str] = {}  # vtiger contactid → omnir contact UUID

        # Counters
        self.counts = {
            "users": {"migrated": 0, "skipped": 0},
            "accounts": {"migrated": 0, "skipped": 0},
            "contacts": {"migrated": 0, "skipped": 0},
            "deals": {"migrated": 0, "skipped": 0},
            "tickets": {"migrated": 0, "skipped": 0},
            "activities": {"migrated": 0, "skipped": 0},
        }

    def run(self):
        try:
            self.migrate_users()
            self.migrate_accounts()
            self.migrate_contacts()
            self.migrate_deals()
            self.migrate_tickets()
            self.migrate_activities()

            if self.dry_run:
                print("\n[DRY RUN] Rolling back — no changes written.")
                self.dst.rollback()
            else:
                self.dst.commit()
                print("\nMigration committed successfully.")

            self._print_summary()
        except Exception:
            self.dst.rollback()
            raise
        finally:
            self.src.close()
            self.dst.close()

    # ------------------------------------------------------------------
    # Users
    # ------------------------------------------------------------------

    def migrate_users(self):
        print("\n--- Migrating users ---")
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT id, user_name, first_name, last_name, email1, is_admin, status
            FROM vtiger_users
            WHERE deleted = 0
              AND status IN ('Active', 'Inactive')
        """)
        rows = cur.fetchall()

        dcur = self.dst.cursor()
        for row in tqdm(rows, desc="users"):
            name = f"{(row['first_name'] or '').strip()} {(row['last_name'] or '').strip()}".strip()
            if not name:
                name = row["user_name"] or f"User {row['id']}"
            role = "admin" if row["is_admin"] == "on" else "user"
            legacy_id = str(row["id"])

            dcur.execute(
                """
                INSERT INTO users (email, name, role, org_id, vtiger_legacy_id, created_at, updated_at)
                VALUES (%s, %s, %s, %s, %s, NOW(), NOW())
                ON CONFLICT (vtiger_legacy_id) DO NOTHING
                RETURNING id
                """,
                (row["email1"], name, role, self.org_id, legacy_id),
            )
            result = dcur.fetchone()
            if result:
                self.user_map[row["id"]] = result[0]
                self.counts["users"]["migrated"] += 1
            else:
                # Already exists; fetch the mapped id
                dcur.execute(
                    "SELECT id FROM users WHERE vtiger_legacy_id = %s", (legacy_id,)
                )
                existing = dcur.fetchone()
                if existing:
                    self.user_map[row["id"]] = existing[0]
                self.counts["users"]["skipped"] += 1

        # Resolve a fallback owner for rows whose vtiger owner wasn't migrated
        dcur.execute("SELECT id FROM users WHERE org_id = %s LIMIT 1", (self.org_id,))
        fallback = dcur.fetchone()
        self._fallback_owner = str(fallback[0]) if fallback else None

    def _owner(self, smownerid) -> str | None:
        return self.user_map.get(smownerid, self._fallback_owner)

    # ------------------------------------------------------------------
    # Accounts
    # ------------------------------------------------------------------

    def migrate_accounts(self):
        print("\n--- Migrating accounts ---")
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT a.accountid, a.accountname, a.website, a.phone,
                   a.email1, a.industry, a.employees,
                   a.description,
                   e.smownerid, e.createdtime, e.modifiedtime
            FROM vtiger_account a
            JOIN vtiger_crmentity e ON e.crmid = a.accountid
            WHERE e.deleted = 0
        """)
        rows = cur.fetchall()

        dcur = self.dst.cursor()
        for row in tqdm(rows, desc="accounts"):
            owner = self._owner(row["smownerid"])
            if not owner:
                self.counts["accounts"]["skipped"] += 1
                continue

            legacy_id = str(row["accountid"])
            size = employees_to_size(row["employees"])

            dcur.execute(
                """
                INSERT INTO accounts (
                    name, domain, industry, size, owner_id, org_id,
                    vtiger_legacy_id, custom_fields, created_at, updated_at
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (vtiger_legacy_id) DO NOTHING
                RETURNING id
                """,
                (
                    row["accountname"],
                    row["website"],
                    row["industry"],
                    size,
                    owner,
                    self.org_id,
                    legacy_id,
                    psycopg2.extras.Json(
                        {"vtiger_phone": row["phone"], "vtiger_notes": row["description"]}
                        if row["phone"] or row["description"]
                        else None
                    ),
                    row["createdtime"],
                    row["modifiedtime"],
                ),
            )
            result = dcur.fetchone()
            if result:
                self.account_map[row["accountid"]] = result[0]
                self.counts["accounts"]["migrated"] += 1
            else:
                dcur.execute(
                    "SELECT id FROM accounts WHERE vtiger_legacy_id = %s", (legacy_id,)
                )
                existing = dcur.fetchone()
                if existing:
                    self.account_map[row["accountid"]] = existing[0]
                self.counts["accounts"]["skipped"] += 1

    # ------------------------------------------------------------------
    # Contacts
    # ------------------------------------------------------------------

    def migrate_contacts(self):
        print("\n--- Migrating contacts ---")
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT c.contactid, c.firstname, c.lastname, c.email,
                   c.phone, c.accountid, c.leadsource,
                   e.smownerid, e.createdtime, e.modifiedtime
            FROM vtiger_contacts c
            JOIN vtiger_crmentity e ON e.crmid = c.contactid
            WHERE e.deleted = 0
        """)
        rows = cur.fetchall()

        dcur = self.dst.cursor()
        for row in tqdm(rows, desc="contacts"):
            owner = self._owner(row["smownerid"])
            if not owner:
                self.counts["contacts"]["skipped"] += 1
                continue

            legacy_id = str(row["contactid"])
            account_id = self.account_map.get(row["accountid"])

            dcur.execute(
                """
                INSERT INTO contacts (
                    first_name, last_name, email, phone, account_id,
                    owner_id, lead_source, stage, org_id,
                    vtiger_legacy_id, created_at, updated_at
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, 'lead', %s, %s, %s, %s)
                ON CONFLICT (vtiger_legacy_id) DO NOTHING
                RETURNING id
                """,
                (
                    row["firstname"] or "",
                    row["lastname"] or "",
                    row["email"] or None,
                    row["phone"],
                    account_id,
                    owner,
                    row["leadsource"],
                    self.org_id,
                    legacy_id,
                    row["createdtime"],
                    row["modifiedtime"],
                ),
            )
            result = dcur.fetchone()
            if result:
                self.contact_map[row["contactid"]] = result[0]
                self.counts["contacts"]["migrated"] += 1
            else:
                dcur.execute(
                    "SELECT id FROM contacts WHERE vtiger_legacy_id = %s", (legacy_id,)
                )
                existing = dcur.fetchone()
                if existing:
                    self.contact_map[row["contactid"]] = existing[0]
                self.counts["contacts"]["skipped"] += 1

    # ------------------------------------------------------------------
    # Deals (vtiger Opportunities / Potentials)
    # ------------------------------------------------------------------

    def migrate_deals(self):
        print("\n--- Migrating deals ---")
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT p.potentialid, p.potentialname, p.amount,
                   p.currency_id, p.closingdate, p.sales_stage,
                   p.probability, p.accountid, p.description,
                   e.smownerid, e.createdtime, e.modifiedtime
            FROM vtiger_potential p
            JOIN vtiger_crmentity e ON e.crmid = p.potentialid
            WHERE e.deleted = 0
        """)
        rows = cur.fetchall()

        # Build a currency map (vtiger currency_id → ISO code)
        ccur = self.src.cursor(dictionary=True)
        try:
            ccur.execute("SELECT id, currency_code FROM vtiger_currency_info")
            currency_map = {r["id"]: r["currency_code"] for r in ccur.fetchall()}
        except Exception:
            currency_map = {}

        dcur = self.dst.cursor()
        for row in tqdm(rows, desc="deals"):
            owner = self._owner(row["smownerid"])
            if not owner:
                self.counts["deals"]["skipped"] += 1
                continue

            legacy_id = str(row["potentialid"])
            account_id = self.account_map.get(row["accountid"])
            stage = STAGE_MAP.get(row["sales_stage"], "lead")

            # amount → value_cents (store as integer cents)
            amount = row["amount"]
            value_cents = int(float(amount) * 100) if amount else 0

            currency = currency_map.get(row["currency_id"], "USD")
            if not currency:
                currency = "USD"

            dcur.execute(
                """
                INSERT INTO deals (
                    title, value_cents, currency, stage, probability,
                    expected_close_date, account_id, owner_id, pipeline_id,
                    org_id, vtiger_legacy_id, created_at, updated_at
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (vtiger_legacy_id) DO NOTHING
                """,
                (
                    row["potentialname"],
                    value_cents,
                    currency,
                    stage,
                    row["probability"] or 0,
                    row["closingdate"],
                    account_id,
                    owner,
                    DEFAULT_PIPELINE_ID,
                    self.org_id,
                    legacy_id,
                    row["createdtime"],
                    row["modifiedtime"],
                ),
            )
            if dcur.rowcount:
                self.counts["deals"]["migrated"] += 1
            else:
                self.counts["deals"]["skipped"] += 1

    # ------------------------------------------------------------------
    # Tickets (vtiger HelpDesk → Omnir tickets)
    # ------------------------------------------------------------------

    def migrate_tickets(self):
        print("\n--- Migrating tickets ---")
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT t.ticketid, t.title, t.solution,
                   t.status, t.priority,
                   t.parent_id,
                   e.smownerid, e.createdtime, e.modifiedtime,
                   e.description
            FROM vtiger_troubletickets t
            JOIN vtiger_crmentity e ON e.crmid = t.ticketid
            WHERE e.deleted = 0
        """)
        rows = cur.fetchall()

        # Build a set of known contact and account vtiger crmids so we can
        # correctly classify the parent_id FK.
        contact_ids = set(self.contact_map.keys())
        account_ids = set(self.account_map.keys())

        dcur = self.dst.cursor()
        for row in tqdm(rows, desc="tickets"):
            legacy_id = str(row["ticketid"])
            assignee_id = self._owner(row["smownerid"])

            # parent_id in vtiger can point to either an account or a contact.
            contact_id = None
            account_id = None
            parent = row.get("parent_id")
            if parent:
                try:
                    parent = int(parent)
                except (TypeError, ValueError):
                    parent = None
            if parent:
                if parent in contact_ids:
                    contact_id = self.contact_map.get(parent)
                elif parent in account_ids:
                    account_id = self.account_map.get(parent)

            status = TICKET_STATUS_MAP.get(row["status"] or "", "open")
            priority = TICKET_PRIORITY_MAP.get(row["priority"] or "", "medium")

            # Combine vtiger crmentity.description + solution as the ticket body.
            description_parts = [p for p in [row.get("description"), row.get("solution")] if p]
            description = "\n\n".join(description_parts) or None

            dcur.execute(
                """
                INSERT INTO tickets (
                    subject, description, status, priority,
                    assignee_id, contact_id, account_id,
                    org_id, vtiger_legacy_id, created_at, updated_at
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (vtiger_legacy_id) DO NOTHING
                """,
                (
                    row["title"] or "(no subject)",
                    description,
                    status,
                    priority,
                    assignee_id,
                    contact_id,
                    account_id,
                    self.org_id,
                    legacy_id,
                    row["createdtime"],
                    row["modifiedtime"],
                ),
            )
            if dcur.rowcount:
                self.counts["tickets"]["migrated"] += 1
            else:
                self.counts["tickets"]["skipped"] += 1

    # ------------------------------------------------------------------
    # Activities
    # ------------------------------------------------------------------

    def migrate_activities(self):
        print("\n--- Migrating activities ---")
        cur = self.src.cursor(dictionary=True)
        cur.execute("""
            SELECT a.activityid, a.subject, a.activitytype,
                   a.date_start, a.time_start, a.due_date, a.time_end,
                   a.description, a.status,
                   e.smownerid, e.createdtime,
                   MAX(rel.crmid) AS contact_crmid
            FROM vtiger_activity a
            JOIN vtiger_crmentity e ON e.crmid = a.activityid
            LEFT JOIN vtiger_seactivityrel rel
                   ON rel.activityid = a.activityid
            WHERE e.deleted = 0
              AND a.activitytype IN ('Call', 'Meeting', 'Task', 'Emails', 'Email')
            GROUP BY a.activityid, a.subject, a.activitytype,
                     a.date_start, a.time_start, a.due_date, a.time_end,
                     a.description, a.status,
                     e.smownerid, e.createdtime
        """)
        rows = cur.fetchall()

        dcur = self.dst.cursor()
        for row in tqdm(rows, desc="activities"):
            owner = self._owner(row["smownerid"])
            if not owner:
                self.counts["activities"]["skipped"] += 1
                continue

            legacy_id = str(row["activityid"])
            activity_type = ACTIVITY_TYPE_MAP.get(row["activitytype"], "task")
            contact_id = self.contact_map.get(row["contact_crmid"])

            # Combine date + time into a single timestamptz for due_date
            due_date = None
            date_field = row.get("due_date") or row.get("date_start")
            time_field = row.get("time_end") or row.get("time_start")
            if date_field:
                if time_field:
                    try:
                        due_date = datetime.strptime(
                            f"{date_field} {time_field}", "%Y-%m-%d %H:%M:%S"
                        )
                    except (ValueError, TypeError):
                        due_date = date_field
                else:
                    due_date = date_field

            dcur.execute(
                """
                INSERT INTO activities (
                    type, subject, description, due_date,
                    contact_id, owner_id, org_id,
                    vtiger_legacy_id, created_at, updated_at
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, NOW())
                ON CONFLICT (vtiger_legacy_id) DO NOTHING
                """,
                (
                    activity_type,
                    row["subject"] or "(no subject)",
                    row["description"],
                    due_date,
                    contact_id,
                    owner,
                    self.org_id,
                    legacy_id,
                    row["createdtime"],
                ),
            )
            if dcur.rowcount:
                self.counts["activities"]["migrated"] += 1
            else:
                self.counts["activities"]["skipped"] += 1

    # ------------------------------------------------------------------
    # Summary
    # ------------------------------------------------------------------

    def _print_summary(self):
        mode = "[DRY RUN]" if self.dry_run else "[LIVE]"
        print(f"\n{'='*50}")
        print(f"Migration summary {mode}")
        print(f"{'='*50}")
        for entity, c in self.counts.items():
            total = c["migrated"] + c["skipped"]
            print(f"  {entity:<12} migrated={c['migrated']:<6} skipped={c['skipped']:<6} total_src={total}")
        print(f"{'='*50}")


# ------------------------------------------------------------------
# Entry point
# ------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(description="Migrate vtiger CE data to Omnir CRM")
    parser.add_argument("--dry-run", action="store_true", help="Run in a transaction and roll back — no data written")
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID, help="Omnir org UUID to assign all records to")
    args = parser.parse_args()

    # Build vtiger connection config
    vtiger_dsn = os.environ.get("VTIGER_DSN")
    if vtiger_dsn:
        vtiger_config = parse_mysql_dsn(vtiger_dsn)
    else:
        vtiger_config = {
            "host": os.environ.get("VTIGER_HOST", "localhost"),
            "port": int(os.environ.get("VTIGER_PORT", 3306)),
            "user": os.environ.get("VTIGER_USER", "root"),
            "password": os.environ.get("VTIGER_PASSWORD", ""),
            "database": os.environ.get("VTIGER_DATABASE", "vtiger"),
        }

    omnir_dsn = os.environ.get("OMNIR_DSN")
    if not omnir_dsn:
        print("ERROR: OMNIR_DSN environment variable is required.", file=sys.stderr)
        sys.exit(1)

    migrator = Migrator(
        vtiger_config=vtiger_config,
        omnir_dsn=omnir_dsn,
        org_id=args.org_id,
        dry_run=args.dry_run,
    )
    migrator.run()


if __name__ == "__main__":
    main()
