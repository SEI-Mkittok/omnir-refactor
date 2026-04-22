-- Phase 2 gate: dual-write consistency validation

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
  COALESCE(
    JSON_AGG(JSON_BUILD_OBJECT('link_id', ac.id, 'link_org_id', ac.org_id, 'account_org_id', a.org_id, 'contact_org_id', c.org_id))
      FILTER (WHERE ac.id IS NOT NULL),
    '[]'::JSON
  ) AS sample
FROM account_contacts ac
JOIN accounts a ON a.id = ac.account_id
JOIN contacts c ON c.id = ac.contact_id
WHERE ac.org_id <> a.org_id OR ac.org_id <> c.org_id;

-- 3) Legacy/new parity for primary link
SELECT
  'legacy_new_primary_parity_mismatch' AS check_name,
  COUNT(*)::BIGINT AS issue_count,
  COALESCE(
    JSON_AGG(JSON_BUILD_OBJECT('contact_id', c.id, 'legacy_account_id', c.account_id, 'new_primary_account_id', p.account_id))
      FILTER (WHERE c.id IS NOT NULL),
    '[]'::JSON
  ) AS sample
FROM contacts c
LEFT JOIN LATERAL (
  SELECT ac.account_id
  FROM account_contacts ac
  WHERE ac.contact_id = c.id
    AND ac.is_primary = TRUE
  ORDER BY ac.updated_at DESC NULLS LAST, ac.created_at DESC
  LIMIT 1
) p ON TRUE
WHERE c.account_id IS DISTINCT FROM p.account_id;
