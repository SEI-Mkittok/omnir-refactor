#!/usr/bin/env python3
"""
Omnir CRM fresh-start initialisation script.

Use this instead of migrate.py when setting up a new Omnir deployment
that has NO existing vtiger data to import.

It seeds:
  - One admin user (credentials from env vars or CLI args)
  - Confirms the default org and pipeline seeded by migrations are present

After this runs, the Omnir UI will be ready to use without needing to
complete the /api/setup flow.

Usage:
    python fresh_start.py --admin-name "Alice" --admin-email admin@example.com --admin-password s3cr3t

Environment variables (alternative to CLI args):
    OMNIR_DSN            PostgreSQL DSN
    OMNIR_ADMIN_NAME     Display name for the admin user
    OMNIR_ADMIN_EMAIL    Email for the admin user
    OMNIR_ADMIN_PASSWORD Password for the admin user (min 8 chars)
"""

import argparse
import os
import sys

import bcrypt
import psycopg2
import psycopg2.extras

DEFAULT_ORG_ID = "00000000-0000-0000-0000-000000000002"
DEFAULT_PIPELINE_ID = "00000000-0000-0000-0000-000000000001"


def main():
    parser = argparse.ArgumentParser(description="Seed a fresh Omnir CRM deployment")
    parser.add_argument("--admin-name", default=os.environ.get("OMNIR_ADMIN_NAME", ""))
    parser.add_argument("--admin-email", default=os.environ.get("OMNIR_ADMIN_EMAIL", ""))
    parser.add_argument("--admin-password", default=os.environ.get("OMNIR_ADMIN_PASSWORD", ""))
    args = parser.parse_args()

    if not args.admin_name:
        print("ERROR: --admin-name / OMNIR_ADMIN_NAME is required.", file=sys.stderr)
        sys.exit(1)
    if not args.admin_email:
        print("ERROR: --admin-email / OMNIR_ADMIN_EMAIL is required.", file=sys.stderr)
        sys.exit(1)
    if len(args.admin_password) < 8:
        print("ERROR: --admin-password must be at least 8 characters.", file=sys.stderr)
        sys.exit(1)

    omnir_dsn = os.environ.get("OMNIR_DSN")
    if not omnir_dsn:
        print("ERROR: OMNIR_DSN environment variable is required.", file=sys.stderr)
        sys.exit(1)

    print("Connecting to Omnir PostgreSQL…")
    conn = psycopg2.connect(omnir_dsn)
    psycopg2.extras.register_uuid()
    cur = conn.cursor()

    # Verify migrations have been applied (orgs table + default org must exist).
    cur.execute("SELECT COUNT(*) FROM orgs WHERE id = %s", (DEFAULT_ORG_ID,))
    if cur.fetchone()[0] == 0:
        print(
            "ERROR: Default org not found. Run `goose up` in api/migrations first.",
            file=sys.stderr,
        )
        conn.close()
        sys.exit(1)

    # Verify default pipeline exists.
    cur.execute("SELECT COUNT(*) FROM pipelines WHERE id = %s", (DEFAULT_PIPELINE_ID,))
    if cur.fetchone()[0] == 0:
        print(
            "ERROR: Default pipeline not found. Run `goose up` in api/migrations first.",
            file=sys.stderr,
        )
        conn.close()
        sys.exit(1)

    # Check no users exist yet (guard against re-running on a live instance).
    cur.execute("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL")
    existing = cur.fetchone()[0]
    if existing > 0:
        print(
            f"WARNING: {existing} user(s) already exist. "
            "Skipping admin seed — use the Omnir UI to manage users.",
            file=sys.stderr,
        )
        conn.close()
        sys.exit(0)

    # Hash password with bcrypt.
    pw_hash = bcrypt.hashpw(args.admin_password.encode(), bcrypt.gensalt()).decode()

    cur.execute(
        """
        INSERT INTO users (email, name, role, org_id, password_hash, created_at, updated_at)
        VALUES (%s, %s, 'admin', %s, %s, NOW(), NOW())
        RETURNING id
        """,
        (args.admin_email, args.admin_name, DEFAULT_ORG_ID, pw_hash),
    )
    admin_id = cur.fetchone()[0]
    conn.commit()
    conn.close()

    print(f"\nFresh start complete.")
    print(f"  Admin user created: {args.admin_name} <{args.admin_email}> (id={admin_id})")
    print(f"  Org:      {DEFAULT_ORG_ID}")
    print(f"  Pipeline: {DEFAULT_PIPELINE_ID}")
    print("\nYou can now log in to Omnir CRM with the admin credentials above.")


if __name__ == "__main__":
    main()
