-- Phase 3 gate: read-switch parity validation (feature-flag readiness)

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

-- 3) Legacy source vs new source contact->account projection mismatch
WITH legacy AS (
  SELECT c.id AS contact_id, c.account_id AS account_id
  FROM contacts c
),
new_model AS (
  SELECT ac.contact_id,
         (ARRAY_AGG(ac.account_id ORDER BY ac.is_primary DESC, ac.updated_at DESC NULLS LAST, ac.created_at DESC))[1] AS account_id
  FROM account_contacts ac
  GROUP BY ac.contact_id
),
violations AS (
  SELECT l.contact_id, l.account_id AS legacy_account_id, n.account_id AS new_model_account_id
  FROM legacy l
  LEFT JOIN new_model n ON n.contact_id = l.contact_id
  WHERE l.account_id IS DISTINCT FROM n.account_id
)
SELECT
  'read_projection_parity_mismatch' AS check_name,
  (SELECT COUNT(*)::BIGINT FROM violations) AS issue_count,
  COALESCE(
    (
      SELECT JSON_AGG(JSON_BUILD_OBJECT('contact_id', s.contact_id, 'legacy_account_id', s.legacy_account_id, 'new_model_account_id', s.new_model_account_id))
      FROM (SELECT * FROM violations LIMIT 25) s
    ),
    '[]'::JSON
  ) AS sample;
