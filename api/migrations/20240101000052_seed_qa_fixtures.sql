-- +goose Up
-- QA fixture data for staging: deals, contacts, tickets.
-- All inserts are idempotent via WHERE NOT EXISTS.
-- org_id: 00000000-0000-0000-0000-000000000002 (Default org)

-- -------------------------------------------------------
-- 1. QA admin user (owner for deals/contacts, assignee for tickets)
-- -------------------------------------------------------
INSERT INTO users (id, org_id, name, email, password_hash, role, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000011',
    '00000000-0000-0000-0000-000000000002',
    'QA Admin',
    'admin@omnir.test',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'admin',
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = '00000000-0000-0000-0000-000000000011');

-- -------------------------------------------------------
-- 2. QA accounts (companies)
-- -------------------------------------------------------
INSERT INTO accounts (id, org_id, name, domain, industry, size, owner_id, created_at, updated_at)
SELECT '00000000-0000-0000-0001-000000000001', '00000000-0000-0000-0000-000000000002',
       'Acme Corp', 'acme.com', 'Manufacturing', '201-500',
       '00000000-0000-0000-0000-000000000011', NOW() - INTERVAL '85 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE id = '00000000-0000-0000-0001-000000000001');

INSERT INTO accounts (id, org_id, name, domain, industry, size, owner_id, created_at, updated_at)
SELECT '00000000-0000-0000-0001-000000000002', '00000000-0000-0000-0000-000000000002',
       'TechVentures Inc', 'techventures.io', 'Technology', '51-200',
       '00000000-0000-0000-0000-000000000011', NOW() - INTERVAL '70 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE id = '00000000-0000-0000-0001-000000000002');

INSERT INTO accounts (id, org_id, name, domain, industry, size, owner_id, created_at, updated_at)
SELECT '00000000-0000-0000-0001-000000000003', '00000000-0000-0000-0000-000000000002',
       'GlobalRetail Ltd', 'globalretail.com', 'Retail', '501+',
       '00000000-0000-0000-0000-000000000011', NOW() - INTERVAL '55 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE id = '00000000-0000-0000-0001-000000000003');

-- -------------------------------------------------------
-- 3. QA contacts (10 contacts)
-- -------------------------------------------------------
INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000001', '00000000-0000-0000-0000-000000000002',
       'Alice', 'Johnson', 'alice.johnson@acme.com', '+1-555-0101',
       '00000000-0000-0000-0001-000000000001', '00000000-0000-0000-0000-000000000011',
       'website', 'customer', NOW() - INTERVAL '80 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000001');

INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0000-000000000002',
       'Bob', 'Martinez', 'bob.martinez@acme.com', '+1-555-0102',
       '00000000-0000-0000-0001-000000000001', '00000000-0000-0000-0000-000000000011',
       'referral', 'prospect', NOW() - INTERVAL '72 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000002');

INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0000-000000000002',
       'Carol', 'Lee', 'carol.lee@techventures.io', '+1-555-0103',
       '00000000-0000-0000-0001-000000000002', '00000000-0000-0000-0000-000000000011',
       'cold_call', 'lead', NOW() - INTERVAL '65 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000003');

INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000004', '00000000-0000-0000-0000-000000000002',
       'David', 'Kim', 'david.kim@techventures.io', '+1-555-0104',
       '00000000-0000-0000-0001-000000000002', '00000000-0000-0000-0000-000000000011',
       'linkedin', 'prospect', NOW() - INTERVAL '58 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000004');

INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000005', '00000000-0000-0000-0000-000000000002',
       'Eve', 'Williams', 'eve.williams@globalretail.com', '+1-555-0105',
       '00000000-0000-0000-0001-000000000003', '00000000-0000-0000-0000-000000000011',
       'conference', 'customer', NOW() - INTERVAL '50 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000005');

INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000006', '00000000-0000-0000-0000-000000000002',
       'Frank', 'Brown', 'frank.brown@globalretail.com', '+1-555-0106',
       '00000000-0000-0000-0001-000000000003', '00000000-0000-0000-0000-000000000011',
       'website', 'lead', NOW() - INTERVAL '43 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000006');

INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000007', '00000000-0000-0000-0000-000000000002',
       'Grace', 'Taylor', 'grace.taylor@acme.com', '+1-555-0107',
       '00000000-0000-0000-0001-000000000001', '00000000-0000-0000-0000-000000000011',
       'partner', 'prospect', NOW() - INTERVAL '36 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000007');

INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000008', '00000000-0000-0000-0000-000000000002',
       'Henry', 'Davis', 'henry.davis@techventures.io', '+1-555-0108',
       '00000000-0000-0000-0001-000000000002', '00000000-0000-0000-0000-000000000011',
       'referral', 'customer', NOW() - INTERVAL '29 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000008');

INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000009', '00000000-0000-0000-0000-000000000002',
       'Iris', 'Wilson', 'iris.wilson@globalretail.com', '+1-555-0109',
       '00000000-0000-0000-0001-000000000003', '00000000-0000-0000-0000-000000000011',
       'cold_call', 'lead', NOW() - INTERVAL '22 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000009');

INSERT INTO contacts (id, org_id, first_name, last_name, email, phone, account_id, owner_id, lead_source, stage, created_at, updated_at)
SELECT '00000000-0000-0000-0002-000000000010', '00000000-0000-0000-0000-000000000002',
       'Jack', 'Moore', 'jack.moore@acme.com', '+1-555-0110',
       '00000000-0000-0000-0001-000000000001', '00000000-0000-0000-0000-000000000011',
       'website', 'prospect', NOW() - INTERVAL '15 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM contacts WHERE id = '00000000-0000-0000-0002-000000000010');

-- -------------------------------------------------------
-- 4. QA deals (25 deals: 16 in 'lead', rest spread across stages)
-- created_at spread over 90 days to exercise chart/reporting views
-- pipeline_id: 00000000-0000-0000-0000-000000000001 (Default)
-- -------------------------------------------------------

-- Lead stage: 16 deals
INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0000-000000000002',
       'Website Redesign Project', 1200000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '30 days',
       '00000000-0000-0000-0002-000000000001', '00000000-0000-0000-0001-000000000001',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '88 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000001');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000002', '00000000-0000-0000-0000-000000000002',
       'ERP Integration', 4500000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '45 days',
       '00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0001-000000000001',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '84 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000002');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000003', '00000000-0000-0000-0000-000000000002',
       'Cloud Migration Phase 1', 8750000, 'USD', 'lead', 15,
       CURRENT_DATE + INTERVAL '60 days',
       '00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0001-000000000002',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '80 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000003');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000004', '00000000-0000-0000-0000-000000000002',
       'Data Analytics Platform', 3200000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '35 days',
       '00000000-0000-0000-0002-000000000004', '00000000-0000-0000-0001-000000000002',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '76 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000004');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000005', '00000000-0000-0000-0000-000000000002',
       'Retail POS Upgrade', 1900000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '50 days',
       '00000000-0000-0000-0002-000000000005', '00000000-0000-0000-0001-000000000003',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '72 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000005');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000006', '00000000-0000-0000-0000-000000000002',
       'Inventory Management System', 2600000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '40 days',
       '00000000-0000-0000-0002-000000000006', '00000000-0000-0000-0001-000000000003',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '68 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000006');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000007', '00000000-0000-0000-0000-000000000002',
       'Security Audit & Compliance', 950000, 'USD', 'lead', 15,
       CURRENT_DATE + INTERVAL '25 days',
       '00000000-0000-0000-0002-000000000007', '00000000-0000-0000-0001-000000000001',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '64 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000007');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000008', '00000000-0000-0000-0000-000000000002',
       'Mobile App Development', 5500000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '55 days',
       '00000000-0000-0000-0002-000000000008', '00000000-0000-0000-0001-000000000002',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '60 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000008');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000009', '00000000-0000-0000-0000-000000000002',
       'HR Software Rollout', 1750000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '45 days',
       '00000000-0000-0000-0002-000000000009', '00000000-0000-0000-0001-000000000003',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '56 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000009');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000010', '00000000-0000-0000-0000-000000000002',
       'Customer Portal Build', 3100000, 'USD', 'lead', 15,
       CURRENT_DATE + INTERVAL '30 days',
       '00000000-0000-0000-0002-000000000010', '00000000-0000-0000-0001-000000000001',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '52 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000010');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000011', '00000000-0000-0000-0000-000000000002',
       'Supply Chain Automation', 7200000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '60 days',
       '00000000-0000-0000-0002-000000000001', '00000000-0000-0000-0001-000000000001',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '48 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000011');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000012', '00000000-0000-0000-0000-000000000002',
       'DevOps Transformation', 2850000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '40 days',
       '00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0001-000000000002',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '44 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000012');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000013', '00000000-0000-0000-0000-000000000002',
       'Omnichannel Commerce Platform', 6400000, 'USD', 'lead', 15,
       CURRENT_DATE + INTERVAL '50 days',
       '00000000-0000-0000-0002-000000000005', '00000000-0000-0000-0001-000000000003',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '40 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000013');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000014', '00000000-0000-0000-0000-000000000002',
       'AI Chatbot Integration', 1400000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '35 days',
       '00000000-0000-0000-0002-000000000004', '00000000-0000-0000-0001-000000000002',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '36 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000014');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000015', '00000000-0000-0000-0000-000000000002',
       'B2B Marketplace Launch', 9800000, 'USD', 'lead', 10,
       CURRENT_DATE + INTERVAL '75 days',
       '00000000-0000-0000-0002-000000000006', '00000000-0000-0000-0001-000000000003',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '32 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000015');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000016', '00000000-0000-0000-0000-000000000002',
       'Workforce Management Suite', 2100000, 'USD', 'lead', 15,
       CURRENT_DATE + INTERVAL '45 days',
       '00000000-0000-0000-0002-000000000007', '00000000-0000-0000-0001-000000000001',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '28 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000016');

-- Qualified stage: 3 deals
INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000017', '00000000-0000-0000-0000-000000000002',
       'SaaS Platform Expansion', 5100000, 'USD', 'qualified', 30,
       CURRENT_DATE + INTERVAL '25 days',
       '00000000-0000-0000-0002-000000000008', '00000000-0000-0000-0001-000000000002',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '55 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000017');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000018', '00000000-0000-0000-0000-000000000002',
       'Enterprise License Renewal', 3600000, 'USD', 'qualified', 35,
       CURRENT_DATE + INTERVAL '20 days',
       '00000000-0000-0000-0002-000000000009', '00000000-0000-0000-0001-000000000003',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '48 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000018');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000019', '00000000-0000-0000-0000-000000000002',
       'Digital Transformation Roadmap', 2400000, 'USD', 'qualified', 30,
       CURRENT_DATE + INTERVAL '30 days',
       '00000000-0000-0000-0002-000000000010', '00000000-0000-0000-0001-000000000001',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '42 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000019');

-- Proposal stage: 3 deals
INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000020', '00000000-0000-0000-0000-000000000002',
       'Managed Services Contract', 4800000, 'USD', 'proposal', 60,
       CURRENT_DATE + INTERVAL '15 days',
       '00000000-0000-0000-0002-000000000001', '00000000-0000-0000-0001-000000000001',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '65 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000020');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000021', '00000000-0000-0000-0000-000000000002',
       'Cybersecurity Package', 1650000, 'USD', 'proposal', 55,
       CURRENT_DATE + INTERVAL '10 days',
       '00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0001-000000000001',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '58 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000021');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000022', '00000000-0000-0000-0000-000000000002',
       'Custom Reporting Dashboard', 980000, 'USD', 'proposal', 60,
       CURRENT_DATE + INTERVAL '12 days',
       '00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0001-000000000002',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '50 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000022');

-- Negotiation stage: 2 deals
INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000023', '00000000-0000-0000-0000-000000000002',
       'Annual Support Agreement', 2200000, 'USD', 'negotiation', 75,
       CURRENT_DATE + INTERVAL '7 days',
       '00000000-0000-0000-0002-000000000004', '00000000-0000-0000-0001-000000000002',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '75 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000023');

INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000024', '00000000-0000-0000-0000-000000000002',
       'International Expansion Suite', 11500000, 'USD', 'negotiation', 80,
       CURRENT_DATE + INTERVAL '5 days',
       '00000000-0000-0000-0002-000000000005', '00000000-0000-0000-0001-000000000003',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '82 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000024');

-- Closed Won: 1 deal
INSERT INTO deals (id, org_id, title, value_cents, currency, stage, probability, expected_close_date, contact_id, account_id, owner_id, pipeline_id, created_at, updated_at)
SELECT '00000000-0000-0000-0003-000000000025', '00000000-0000-0000-0000-000000000002',
       'Legacy System Migration', 6700000, 'USD', 'closed_won', 100,
       CURRENT_DATE - INTERVAL '5 days',
       '00000000-0000-0000-0002-000000000006', '00000000-0000-0000-0001-000000000003',
       '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001',
       NOW() - INTERVAL '90 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM deals WHERE id = '00000000-0000-0000-0003-000000000025');

-- -------------------------------------------------------
-- 5. QA tickets (10 tickets across statuses and priorities)
-- -------------------------------------------------------
INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000001', '00000000-0000-0000-0000-000000000002',
       'Login page not loading on mobile',
       'Users report blank screen when accessing login from iOS Safari.',
       'open', 'high', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000001', '00000000-0000-0000-0001-000000000001',
       'email', NOW() - INTERVAL '25 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000001');

INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000002', '00000000-0000-0000-0000-000000000002',
       'CSV export missing columns',
       'Contact export CSV is missing phone and lead_source columns.',
       'pending', 'medium', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0001-000000000001',
       'portal', NOW() - INTERVAL '20 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000002');

INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000003', '00000000-0000-0000-0000-000000000002',
       'Pipeline kanban drag-and-drop broken in Firefox',
       'Cards cannot be dragged between columns in Firefox 124.',
       'open', 'critical', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0001-000000000002',
       'chat', NOW() - INTERVAL '18 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000003');

INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000004', '00000000-0000-0000-0000-000000000002',
       'Email notifications delayed by 30+ minutes',
       'Deal stage change notifications arrive 30+ min late.',
       'in_progress', 'high', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000004', '00000000-0000-0000-0001-000000000002',
       'email', NOW() - INTERVAL '15 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000004');

INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000005', '00000000-0000-0000-0000-000000000002',
       'Report chart shows incorrect totals',
       'Pipeline report value bars do not match column sum totals.',
       'pending', 'high', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000005', '00000000-0000-0000-0001-000000000003',
       'portal', NOW() - INTERVAL '12 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000005');

INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000006', '00000000-0000-0000-0000-000000000002',
       'Account merge duplicates contacts',
       'Merging two accounts creates duplicate contact entries.',
       'open', 'medium', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000006', '00000000-0000-0000-0001-000000000003',
       'email', NOW() - INTERVAL '10 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000006');

INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000007', '00000000-0000-0000-0000-000000000002',
       'Custom field values not saved',
       'Custom field changes are discarded on page refresh.',
       'resolved', 'medium', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000007', '00000000-0000-0000-0001-000000000001',
       'chat', NOW() - INTERVAL '30 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000007');

INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000008', '00000000-0000-0000-0000-000000000002',
       'Slow search on large contact lists',
       'Global search takes 8+ seconds with > 500 contacts.',
       'resolved', 'low', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000008', '00000000-0000-0000-0001-000000000002',
       'portal', NOW() - INTERVAL '35 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000008');

INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000009', '00000000-0000-0000-0000-000000000002',
       'Webhook not firing on deal close',
       'Outbound webhook for deal.closed event not triggering.',
       'open', 'high', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000009', '00000000-0000-0000-0001-000000000003',
       'email', NOW() - INTERVAL '8 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000009');

INSERT INTO tickets (id, org_id, subject, description, status, priority, assignee_id, contact_id, account_id, source, created_at, updated_at)
SELECT '00000000-0000-0000-0004-000000000010', '00000000-0000-0000-0000-000000000002',
       'Lead score not updating after activity',
       'Lead score remains 0 after logging a call activity.',
       'closed', 'low', '00000000-0000-0000-0000-000000000011',
       '00000000-0000-0000-0002-000000000010', '00000000-0000-0000-0001-000000000001',
       'portal', NOW() - INTERVAL '40 days', NOW()
WHERE NOT EXISTS (SELECT 1 FROM tickets WHERE id = '00000000-0000-0000-0004-000000000010');

-- +goose Down
DELETE FROM tickets  WHERE id LIKE '00000000-0000-0000-0004-%';
DELETE FROM deals    WHERE id LIKE '00000000-0000-0000-0003-%';
DELETE FROM contacts WHERE id LIKE '00000000-0000-0000-0002-%';
DELETE FROM accounts WHERE id LIKE '00000000-0000-0000-0001-%';
DELETE FROM users    WHERE id = '00000000-0000-0000-0000-000000000011';
