-- Phase 1 gate: schema+backfill validation

-- 1) Orphaned links (missing account/contact)
WITH violations AS (
  SELECT ac.id, ac.org_id, ac.account_id, ac.contact_id
  FROM account_contacts ac
  LEFT JOIN accounts a ON a.id = ac.account_id
  LEFT JOIN contacts c ON c.id = ac.contact_id
  WHERE a.id IS NULL OR c.id IS NULL
)
SELECT
  'orphaned_account_contacts' AS check_name,
  (SELECT COUNT(*)::BIGINT FROM violations) AS issue_count,
  COALESCE(
    (
      SELECT JSON_AGG(JSON_BUILD_OBJECT('link_id', s.id, 'org_id', s.org_id, 'account_id', s.account_id, 'contact_id', s.contact_id))
      FROM (SELECT * FROM violations LIMIT 25) s
    ),
    '[]'::JSON
  ) AS sample;

-- 2) Cross-org leaks
WITH violations AS (
  SELECT ac.id, ac.org_id, a.org_id AS account_org_id, c.org_id AS contact_org_id
  FROM account_contacts ac
  JOIN accounts a ON a.id = ac.account_id
  JOIN contacts c ON c.id = ac.contact_id
  WHERE ac.org_id <> a.org_id OR ac.org_id <> c.org_id
)
SELECT
  'cross_org_account_contacts' AS check_name,
  (SELECT COUNT(*)::BIGINT FROM violations) AS issue_count,
  COALESCE(
    (
      SELECT JSON_AGG(JSON_BUILD_OBJECT(
        'link_id', s.id,
        'link_org_id', s.org_id,
        'account_org_id', s.account_org_id,
        'contact_org_id', s.contact_org_id
      ))
      FROM (SELECT * FROM violations LIMIT 25) s
    ),
    '[]'::JSON
  ) AS sample;

-- 3) Backfill mismatch from legacy contacts.account_id -> account_contacts(primary)
WITH violations AS (
  SELECT c.id, c.account_id
  FROM contacts c
  LEFT JOIN account_contacts ac
    ON ac.contact_id = c.id
   AND ac.account_id = c.account_id
   AND ac.is_primary = TRUE
  WHERE c.account_id IS NOT NULL
    AND ac.id IS NULL
)
SELECT
  'legacy_primary_backfill_mismatch' AS check_name,
  (SELECT COUNT(*)::BIGINT FROM violations) AS issue_count,
  COALESCE(
    (
      SELECT JSON_AGG(JSON_BUILD_OBJECT('contact_id', s.id, 'legacy_account_id', s.account_id))
      FROM (SELECT * FROM violations LIMIT 25) s
    ),
    '[]'::JSON
  ) AS sample;
