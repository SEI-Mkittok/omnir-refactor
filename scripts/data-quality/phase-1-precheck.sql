-- Phase 1 gate: schema+backfill validation

-- 1) Orphaned links (missing account/contact)
SELECT
  'orphaned_account_contacts' AS check_name,
  COUNT(*)::BIGINT AS issue_count,
  COALESCE(
    JSON_AGG(JSON_BUILD_OBJECT('link_id', ac.id, 'org_id', ac.org_id, 'account_id', ac.account_id, 'contact_id', ac.contact_id))
      FILTER (WHERE ac.id IS NOT NULL),
    '[]'::JSON
  ) AS sample
FROM account_contacts ac
LEFT JOIN accounts a ON a.id = ac.account_id
LEFT JOIN contacts c ON c.id = ac.contact_id
WHERE a.id IS NULL OR c.id IS NULL;

-- 2) Cross-org leaks
SELECT
  'cross_org_account_contacts' AS check_name,
  COUNT(*)::BIGINT AS issue_count,
  COALESCE(
    JSON_AGG(JSON_BUILD_OBJECT(
      'link_id', ac.id,
      'link_org_id', ac.org_id,
      'account_org_id', a.org_id,
      'contact_org_id', c.org_id
    )) FILTER (WHERE ac.id IS NOT NULL),
    '[]'::JSON
  ) AS sample
FROM account_contacts ac
JOIN accounts a ON a.id = ac.account_id
JOIN contacts c ON c.id = ac.contact_id
WHERE ac.org_id <> a.org_id OR ac.org_id <> c.org_id;

-- 3) Backfill mismatch from legacy contacts.account_id -> account_contacts(primary)
SELECT
  'legacy_primary_backfill_mismatch' AS check_name,
  COUNT(*)::BIGINT AS issue_count,
  COALESCE(
    JSON_AGG(JSON_BUILD_OBJECT('contact_id', c.id, 'legacy_account_id', c.account_id))
      FILTER (WHERE c.id IS NOT NULL),
    '[]'::JSON
  ) AS sample
FROM contacts c
LEFT JOIN account_contacts ac
  ON ac.contact_id = c.id
 AND ac.account_id = c.account_id
 AND ac.is_primary = TRUE
WHERE c.account_id IS NOT NULL
  AND ac.id IS NULL;
