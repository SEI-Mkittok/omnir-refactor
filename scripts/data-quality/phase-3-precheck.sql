-- Phase 3 gate: read-switch parity validation (feature-flag readiness)

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
)
SELECT
  'read_projection_parity_mismatch' AS check_name,
  COUNT(*)::BIGINT AS issue_count,
  COALESCE(
    JSON_AGG(JSON_BUILD_OBJECT('contact_id', l.contact_id, 'legacy_account_id', l.account_id, 'new_model_account_id', n.account_id))
      FILTER (WHERE l.contact_id IS NOT NULL),
    '[]'::JSON
  ) AS sample
FROM legacy l
LEFT JOIN new_model n ON n.contact_id = l.contact_id
WHERE l.account_id IS DISTINCT FROM n.account_id;
