-- Phase 2 gate: dual-write consistency validation

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
      SELECT JSON_AGG(JSON_BUILD_OBJECT('link_id', s.id, 'link_org_id', s.org_id, 'account_org_id', s.account_org_id, 'contact_org_id', s.contact_org_id))
      FROM (SELECT * FROM violations LIMIT 25) s
    ),
    '[]'::JSON
  ) AS sample;

-- 3) Legacy/new parity for primary link
WITH violations AS (
  SELECT c.id, c.account_id, p.account_id AS new_primary_account_id
  FROM contacts c
  LEFT JOIN LATERAL (
    SELECT ac.account_id
    FROM account_contacts ac
    WHERE ac.contact_id = c.id
      AND ac.is_primary = TRUE
    ORDER BY ac.updated_at DESC NULLS LAST, ac.created_at DESC
    LIMIT 1
  ) p ON TRUE
  WHERE c.account_id IS DISTINCT FROM p.account_id
)
SELECT
  'legacy_new_primary_parity_mismatch' AS check_name,
  (SELECT COUNT(*)::BIGINT FROM violations) AS issue_count,
  COALESCE(
    (
      SELECT JSON_AGG(JSON_BUILD_OBJECT('contact_id', s.id, 'legacy_account_id', s.account_id, 'new_primary_account_id', s.new_primary_account_id))
      FROM (SELECT * FROM violations LIMIT 25) s
    ),
    '[]'::JSON
  ) AS sample;
