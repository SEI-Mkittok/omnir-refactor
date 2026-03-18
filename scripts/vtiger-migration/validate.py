#!/usr/bin/env python3
"""
Post-migration validation script for vtiger → Omnir CRM migration.

Checks:
  1. Row counts (vtiger source vs Omnir destination)
  2. Referential integrity (dangling FKs)
  3. Spot-checks a random sample of each entity by vtiger_legacy_id

Usage:
    python validate.py [--sample N] [--fail-on-error]

Environment variables (same as migrate.py):
    VTIGER_DSN   mysql://user:pass@host/dbname
    OMNIR_DSN    postgresql://user:pass@host/dbname

Exit code:
    0 = all checks passed
    1 = one or more checks failed
"""

import argparse
import os
import re
import sys

import mysql.connector
import psycopg2
import psycopg2.extras


def parse_mysql_dsn(dsn: str) -> dict:
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


class Validator:
    ACCEPTANCE_THRESHOLD = 0.001  # 0.1% max row loss

    def __init__(self, vtiger_config: dict, omnir_dsn: str, sample: int = 20):
        self.sample = sample
        self.failures: list[str] = []

        print("Connecting to vtiger MySQL…")
        self.src = mysql.connector.connect(**vtiger_config)

        print("Connecting to Omnir PostgreSQL…")
        self.dst = psycopg2.connect(omnir_dsn)
        psycopg2.extras.register_uuid()

    def run(self) -> bool:
        self._check_row_counts()
        self._check_referential_integrity()
        self._spot_check_users()
        self._spot_check_contacts()
        self._spot_check_accounts()
        self._spot_check_leads()
        self._spot_check_deals()
        self._spot_check_tickets()

        print(f"\n{'='*50}")
        if self.failures:
            print(f"VALIDATION FAILED — {len(self.failures)} issue(s):")
            for f in self.failures:
                print(f"  ✗ {f}")
            return False
        else:
            print("VALIDATION PASSED — all checks OK")
            return True

    # ------------------------------------------------------------------
    # 1. Row counts
    # ------------------------------------------------------------------

    def _check_row_counts(self):
        print("\n--- Row count checks ---")

        src_cur = self.src.cursor(dictionary=True)
        dst_cur = self.dst.cursor()

        checks = [
            (
                "users",
                "SELECT COUNT(*) FROM vtiger_users WHERE deleted = 0 AND status = 'Active'",
                "SELECT COUNT(*) FROM users WHERE vtiger_legacy_id IS NOT NULL AND deleted_at IS NULL",
            ),
            (
                "accounts",
                """SELECT COUNT(*) FROM vtiger_account a
                   JOIN vtiger_crmentity e ON e.crmid = a.accountid WHERE e.deleted = 0""",
                "SELECT COUNT(*) FROM accounts WHERE vtiger_legacy_id IS NOT NULL AND deleted_at IS NULL",
            ),
            (
                "contacts",
                """SELECT COUNT(*) FROM vtiger_contacts c
                   JOIN vtiger_crmentity e ON e.crmid = c.contactid WHERE e.deleted = 0""",
                "SELECT COUNT(*) FROM contacts WHERE vtiger_legacy_id IS NOT NULL AND deleted_at IS NULL",
            ),
            (
                "leads",
                """SELECT COUNT(*) FROM vtiger_leaddetails l
                   JOIN vtiger_crmentity e ON e.crmid = l.leadid WHERE e.deleted = 0""",
                "SELECT COUNT(*) FROM leads WHERE vtiger_legacy_id IS NOT NULL AND deleted_at IS NULL",
            ),
            (
                "deals",
                """SELECT COUNT(*) FROM vtiger_potential p
                   JOIN vtiger_crmentity e ON e.crmid = p.potentialid WHERE e.deleted = 0""",
                "SELECT COUNT(*) FROM deals WHERE vtiger_legacy_id IS NOT NULL AND deleted_at IS NULL",
            ),
            (
                "tickets",
                """SELECT COUNT(*) FROM vtiger_troubletickets t
                   JOIN vtiger_crmentity e ON e.crmid = t.ticketid WHERE e.deleted = 0""",
                "SELECT COUNT(*) FROM tickets WHERE vtiger_legacy_id IS NOT NULL AND deleted_at IS NULL",
            ),
            (
                "activities",
                """SELECT COUNT(*) FROM vtiger_activity a
                   JOIN vtiger_crmentity e ON e.crmid = a.activityid
                   WHERE e.deleted = 0
                   AND a.activitytype IN ('Call','Meeting','Task','Emails','Email')""",
                "SELECT COUNT(*) FROM activities WHERE vtiger_legacy_id IS NOT NULL AND deleted_at IS NULL",
            ),
        ]

        for entity, src_sql, dst_sql in checks:
            src_cur.execute(src_sql)
            src_count = src_cur.fetchone()
            src_count = list(src_count.values())[0] if isinstance(src_count, dict) else src_count[0]

            dst_cur.execute(dst_sql)
            dst_count = dst_cur.fetchone()[0]

            delta = src_count - dst_count
            pct = (delta / src_count) if src_count > 0 else 0.0
            status = "OK" if pct <= self.ACCEPTANCE_THRESHOLD else "FAIL"
            print(f"  {entity:<12} src={src_count:<6} dst={dst_count:<6} delta={delta:<6} ({pct:.3%}) [{status}]")

            if pct > self.ACCEPTANCE_THRESHOLD:
                self.failures.append(
                    f"{entity}: {delta} rows missing ({pct:.3%} > {self.ACCEPTANCE_THRESHOLD:.1%} threshold)"
                )

    # ------------------------------------------------------------------
    # 2. Referential integrity
    # ------------------------------------------------------------------

    def _check_referential_integrity(self):
        print("\n--- Referential integrity checks ---")
        dst_cur = self.dst.cursor()

        ri_checks = [
            (
                "contacts.account_id dangling",
                """SELECT COUNT(*) FROM contacts
                   WHERE account_id IS NOT NULL
                     AND account_id NOT IN (SELECT id FROM accounts)
                     AND deleted_at IS NULL""",
            ),
            (
                "deals.account_id dangling",
                """SELECT COUNT(*) FROM deals
                   WHERE account_id IS NOT NULL
                     AND account_id NOT IN (SELECT id FROM accounts)
                     AND deleted_at IS NULL""",
            ),
            (
                "activities.contact_id dangling",
                """SELECT COUNT(*) FROM activities
                   WHERE contact_id IS NOT NULL
                     AND contact_id NOT IN (SELECT id FROM contacts)
                     AND deleted_at IS NULL""",
            ),
            (
                "leads.owner_id dangling",
                """SELECT COUNT(*) FROM leads
                   WHERE owner_id NOT IN (SELECT id FROM users)
                     AND deleted_at IS NULL""",
            ),
            (
                "deals.owner_id dangling",
                """SELECT COUNT(*) FROM deals
                   WHERE owner_id NOT IN (SELECT id FROM users)
                     AND deleted_at IS NULL""",
            ),
            (
                "tickets.contact_id dangling",
                """SELECT COUNT(*) FROM tickets
                   WHERE contact_id IS NOT NULL
                     AND contact_id NOT IN (SELECT id FROM contacts)
                     AND deleted_at IS NULL""",
            ),
            (
                "tickets.account_id dangling",
                """SELECT COUNT(*) FROM tickets
                   WHERE account_id IS NOT NULL
                     AND account_id NOT IN (SELECT id FROM accounts)
                     AND deleted_at IS NULL""",
            ),
        ]

        for label, sql in ri_checks:
            dst_cur.execute(sql)
            count = dst_cur.fetchone()[0]
            status = "OK" if count == 0 else "FAIL"
            print(f"  {label:<40} count={count} [{status}]")
            if count > 0:
                self.failures.append(f"{label}: {count} dangling references")

    # ------------------------------------------------------------------
    # 3. Spot-checks
    # ------------------------------------------------------------------

    def _spot_check_users(self):
        print(f"\n--- Spot-checking {self.sample} users ---")
        src_cur = self.src.cursor(dictionary=True)
        dst_cur = self.dst.cursor(row_factory=None)

        src_cur.execute(
            f"""SELECT id, first_name, last_name, email1
                FROM vtiger_users
                WHERE deleted = 0 AND status = 'Active'
                ORDER BY RAND() LIMIT {self.sample}"""
        )
        rows = src_cur.fetchall()

        mismatches = 0
        for row in rows:
            dst_cur.execute(
                "SELECT email FROM users WHERE vtiger_legacy_id = %s",
                (str(row["id"]),),
            )
            result = dst_cur.fetchone()
            if not result:
                print(f"  MISSING user vtiger_id={row['id']} email={row['email1']}")
                mismatches += 1
            elif result[0] != row["email1"]:
                print(f"  EMAIL MISMATCH vtiger_id={row['id']} src={row['email1']} dst={result[0]}")
                mismatches += 1

        print(f"  Users spot-check: {len(rows) - mismatches}/{len(rows)} OK")
        if mismatches:
            self.failures.append(f"users spot-check: {mismatches} mismatches")

    def _spot_check_contacts(self):
        print(f"\n--- Spot-checking {self.sample} contacts ---")
        src_cur = self.src.cursor(dictionary=True)
        dst_cur = self.dst.cursor()

        src_cur.execute(
            f"""SELECT c.contactid, c.firstname, c.lastname, c.email
                FROM vtiger_contacts c
                JOIN vtiger_crmentity e ON e.crmid = c.contactid
                WHERE e.deleted = 0
                ORDER BY RAND() LIMIT {self.sample}"""
        )
        rows = src_cur.fetchall()

        mismatches = 0
        for row in rows:
            dst_cur.execute(
                "SELECT first_name, last_name, email FROM contacts WHERE vtiger_legacy_id = %s",
                (str(row["contactid"]),),
            )
            result = dst_cur.fetchone()
            if not result:
                print(f"  MISSING contact vtiger_id={row['contactid']}")
                mismatches += 1
            else:
                if result[0] != (row["firstname"] or "") or result[1] != (row["lastname"] or ""):
                    print(
                        f"  NAME MISMATCH vtiger_id={row['contactid']} "
                        f"src={row['firstname']} {row['lastname']} "
                        f"dst={result[0]} {result[1]}"
                    )
                    mismatches += 1

        print(f"  Contacts spot-check: {len(rows) - mismatches}/{len(rows)} OK")
        if mismatches:
            self.failures.append(f"contacts spot-check: {mismatches} mismatches")

    def _spot_check_accounts(self):
        print(f"\n--- Spot-checking {self.sample} accounts ---")
        src_cur = self.src.cursor(dictionary=True)
        dst_cur = self.dst.cursor()

        src_cur.execute(
            f"""SELECT a.accountid, a.accountname
                FROM vtiger_account a
                JOIN vtiger_crmentity e ON e.crmid = a.accountid
                WHERE e.deleted = 0
                ORDER BY RAND() LIMIT {self.sample}"""
        )
        rows = src_cur.fetchall()

        mismatches = 0
        for row in rows:
            dst_cur.execute(
                "SELECT name FROM accounts WHERE vtiger_legacy_id = %s",
                (str(row["accountid"]),),
            )
            result = dst_cur.fetchone()
            if not result:
                print(f"  MISSING account vtiger_id={row['accountid']}")
                mismatches += 1
            elif result[0] != row["accountname"]:
                print(
                    f"  NAME MISMATCH vtiger_id={row['accountid']} "
                    f"src={row['accountname']!r} dst={result[0]!r}"
                )
                mismatches += 1

        print(f"  Accounts spot-check: {len(rows) - mismatches}/{len(rows)} OK")
        if mismatches:
            self.failures.append(f"accounts spot-check: {mismatches} mismatches")

    def _spot_check_leads(self):
        print(f"\n--- Spot-checking {self.sample} leads ---")
        src_cur = self.src.cursor(dictionary=True)
        dst_cur = self.dst.cursor()

        src_cur.execute(
            f"""SELECT l.leadid, l.firstname, l.lastname, l.email
                FROM vtiger_leaddetails l
                JOIN vtiger_crmentity e ON e.crmid = l.leadid
                WHERE e.deleted = 0
                ORDER BY RAND() LIMIT {self.sample}"""
        )
        rows = src_cur.fetchall()

        mismatches = 0
        for row in rows:
            dst_cur.execute(
                "SELECT first_name, last_name, email FROM leads WHERE vtiger_legacy_id = %s",
                (str(row["leadid"]),),
            )
            result = dst_cur.fetchone()
            if not result:
                print(f"  MISSING lead vtiger_id={row['leadid']}")
                mismatches += 1
            else:
                if result[0] != (row["firstname"] or "") or result[1] != (row["lastname"] or ""):
                    print(
                        f"  NAME MISMATCH vtiger_id={row['leadid']} "
                        f"src={row['firstname']} {row['lastname']} "
                        f"dst={result[0]} {result[1]}"
                    )
                    mismatches += 1

        print(f"  Leads spot-check: {len(rows) - mismatches}/{len(rows)} OK")
        if mismatches:
            self.failures.append(f"leads spot-check: {mismatches} mismatches")

    def _spot_check_deals(self):
        print(f"\n--- Spot-checking {self.sample} deals ---")
        src_cur = self.src.cursor(dictionary=True)
        dst_cur = self.dst.cursor()

        src_cur.execute(
            f"""SELECT p.potentialid, p.potentialname, p.amount
                FROM vtiger_potential p
                JOIN vtiger_crmentity e ON e.crmid = p.potentialid
                WHERE e.deleted = 0
                ORDER BY RAND() LIMIT {self.sample}"""
        )
        rows = src_cur.fetchall()

        mismatches = 0
        for row in rows:
            dst_cur.execute(
                "SELECT title, value_cents FROM deals WHERE vtiger_legacy_id = %s",
                (str(row["potentialid"]),),
            )
            result = dst_cur.fetchone()
            if not result:
                print(f"  MISSING deal vtiger_id={row['potentialid']}")
                mismatches += 1
            else:
                expected_cents = int(float(row["amount"] or 0) * 100)
                if result[0] != row["potentialname"] or abs(result[1] - expected_cents) > 1:
                    print(
                        f"  MISMATCH vtiger_id={row['potentialid']} "
                        f"src_title={row['potentialname']!r} dst_title={result[0]!r} "
                        f"src_cents={expected_cents} dst_cents={result[1]}"
                    )
                    mismatches += 1

        print(f"  Deals spot-check: {len(rows) - mismatches}/{len(rows)} OK")
        if mismatches:
            self.failures.append(f"deals spot-check: {mismatches} mismatches")

    def _spot_check_tickets(self):
        print(f"\n--- Spot-checking {self.sample} tickets ---")
        src_cur = self.src.cursor(dictionary=True)
        dst_cur = self.dst.cursor()

        src_cur.execute(
            f"""SELECT t.ticketid, t.title, t.status, t.priority
                FROM vtiger_troubletickets t
                JOIN vtiger_crmentity e ON e.crmid = t.ticketid
                WHERE e.deleted = 0
                ORDER BY RAND() LIMIT {self.sample}"""
        )
        rows = src_cur.fetchall()

        mismatches = 0
        for row in rows:
            dst_cur.execute(
                "SELECT subject FROM tickets WHERE vtiger_legacy_id = %s",
                (str(row["ticketid"]),),
            )
            result = dst_cur.fetchone()
            if not result:
                print(f"  MISSING ticket vtiger_id={row['ticketid']}")
                mismatches += 1
            elif result[0] != (row["title"] or "(no subject)"):
                print(
                    f"  SUBJECT MISMATCH vtiger_id={row['ticketid']} "
                    f"src={row['title']!r} dst={result[0]!r}"
                )
                mismatches += 1

        print(f"  Tickets spot-check: {len(rows) - mismatches}/{len(rows)} OK")
        if mismatches:
            self.failures.append(f"tickets spot-check: {mismatches} mismatches")

    def __del__(self):
        try:
            self.src.close()
            self.dst.close()
        except Exception:
            pass


# ------------------------------------------------------------------
# Entry point
# ------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(description="Validate vtiger → Omnir migration")
    parser.add_argument("--sample", type=int, default=20, help="Rows to spot-check per entity")
    parser.add_argument("--fail-on-error", action="store_true", help="Exit 1 if any check fails")
    args = parser.parse_args()

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

    validator = Validator(vtiger_config=vtiger_config, omnir_dsn=omnir_dsn, sample=args.sample)
    passed = validator.run()

    if not passed and args.fail_on_error:
        sys.exit(1)


if __name__ == "__main__":
    main()
