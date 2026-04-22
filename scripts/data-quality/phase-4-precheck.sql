-- Phase 4 gate: post-cutover data verification prior to deprecated-path removal

-- 1) Orphaned links
SELECT
  'orphaned_account_contacts' AS check_name,
  COUNT(*)::BIGINT AS issue_count,
  COALESCE(JSON_AGG(ac.id) FILTER (WHERE ac.id IS NOT NULL), '[]'::JSON) AS sample
FROM account_contacts ac
LEFT JOIN accounts a ON a.id = ac.account_id
LEFT JOIN contacts c ON c.id = ac.contact_id
WHERE a.id IS NULL OR c.id IS NULL;

-- 2) Cross-org leaks
SELECT
  'cross_org_account_contacts' AS check_name,
  COUNT(*)::BIGINT AS issue_count,
  COALESCE(JSON_AGG(ac.id) FILTER (WHERE ac.id IS NOT NULL), '[]'::JSON) AS sample
FROM account_contacts ac
JOIN accounts a ON a.id = ac.account_id
JOIN contacts c ON c.id = ac.contact_id
WHERE ac.org_id <> a.org_id OR ac.org_id <> c.org_id;

-- 3) Contact/account mismatch after cutover:
-- Any contact with >1 primary link is an integrity bug.
SELECT
  'multiple_primary_links_per_contact' AS check_name,
  COUNT(*)::BIGINT AS issue_count,
  COALESCE(
    JSON_AGG(JSON_BUILD_OBJECT('contact_id', contact_id, 'primary_count', primary_count))
      FILTER (WHERE contact_id IS NOT NULL),
    '[]'::JSON
  ) AS sample
FROM (
  SELECT ac.contact_id, COUNT(*) AS primary_count
  FROM account_contacts ac
  WHERE ac.is_primary = TRUE
  GROUP BY ac.contact_id
  HAVING COUNT(*) > 1
) x;
