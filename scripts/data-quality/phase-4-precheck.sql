-- Phase 4 gate: post-cutover data verification prior to deprecated-path removal

-- 1) Orphaned links
WITH violations AS (
  SELECT ac.id
  FROM account_contacts ac
  LEFT JOIN accounts a ON a.id = ac.account_id
  LEFT JOIN contacts c ON c.id = ac.contact_id
  WHERE a.id IS NULL OR c.id IS NULL
)
SELECT
  'orphaned_account_contacts' AS check_name,
  (SELECT COUNT(*)::BIGINT FROM violations) AS issue_count,
  COALESCE((SELECT JSON_AGG(s.id) FROM (SELECT * FROM violations LIMIT 25) s), '[]'::JSON) AS sample;

-- 2) Cross-org leaks
WITH violations AS (
  SELECT ac.id
  FROM account_contacts ac
  JOIN accounts a ON a.id = ac.account_id
  JOIN contacts c ON c.id = ac.contact_id
  WHERE ac.org_id <> a.org_id OR ac.org_id <> c.org_id
)
SELECT
  'cross_org_account_contacts' AS check_name,
  (SELECT COUNT(*)::BIGINT FROM violations) AS issue_count,
  COALESCE((SELECT JSON_AGG(s.id) FROM (SELECT * FROM violations LIMIT 25) s), '[]'::JSON) AS sample;

-- 3) Contact/account mismatch after cutover:
-- Any contact with >1 primary link is an integrity bug.
WITH violations AS (
  SELECT ac.contact_id, COUNT(*) AS primary_count
  FROM account_contacts ac
  WHERE ac.is_primary = TRUE
  GROUP BY ac.contact_id
  HAVING COUNT(*) > 1
)
SELECT
  'multiple_primary_links_per_contact' AS check_name,
  (SELECT COUNT(*)::BIGINT FROM violations) AS issue_count,
  COALESCE(
    (
      SELECT JSON_AGG(JSON_BUILD_OBJECT('contact_id', s.contact_id, 'primary_count', s.primary_count))
      FROM (SELECT * FROM violations LIMIT 25) s
    ),
    '[]'::JSON
  ) AS sample;
