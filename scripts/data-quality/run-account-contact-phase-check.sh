#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <phase(1|2|3|4)> <database_url>"
  exit 1
fi

PHASE="$1"
DATABASE_URL="$2"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="${SCRIPT_DIR}/phase-${PHASE}-precheck.sql"

if [[ ! -f "${SQL_FILE}" ]]; then
  echo "Unknown phase '${PHASE}'. Expected 1, 2, 3, or 4."
  exit 1
fi

echo "Running phase ${PHASE} data-quality checks: ${SQL_FILE}"
psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${SQL_FILE}"
