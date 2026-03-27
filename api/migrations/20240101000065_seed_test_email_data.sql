-- +goose Up
-- Seed test email connection and inbox messages for QA/staging environments.
-- Creates a mock Gmail connection for admin@omnir.test and 3 threads:
--   - Thread 1: 2 unread messages (support inquiry)
--   - Thread 2: 1 unread message (billing question)
--   - Thread 3: 1 read message (already-read baseline)
-- Idempotent: uses ON CONFLICT DO NOTHING throughout.

-- +goose StatementBegin
DO $$
BEGIN
    -- Email connection (mock Gmail)
    INSERT INTO email_connections (id, org_id, user_id, provider, email_address, access_token, refresh_token, token_expiry, created_at, updated_at)
    VALUES (
        '00000000-0000-0000-0000-000000000100',
        '00000000-0000-0000-0000-000000000002',
        'a45fc555-75fa-4358-b265-577a9f976c84', -- admin@omnir.test
        'gmail', 'admin@omnir.test',
        'mock-access-token-qa', 'mock-refresh-token-qa',
        NOW() + INTERVAL '1 year',
        NOW(), NOW()
    )
    ON CONFLICT (org_id, user_id, provider) DO NOTHING;

    -- Thread 1: 2 unread messages (support inquiry)
    INSERT INTO email_inbox_messages (id, org_id, connection_id, message_id, thread_id, from_addr, to_addrs, subject, body_text, direction, sent_at, read_at)
    VALUES
    (
        '00000000-0000-0000-0001-000000000001',
        '00000000-0000-0000-0000-000000000002',
        '00000000-0000-0000-0000-000000000100',
        'msg-qa-001', 'thread-qa-001',
        'alice.customer@example.com',
        '["admin@omnir.test"]',
        'Support request: cannot export contacts',
        'Hi, I am trying to export my contacts to CSV but the button does not respond. Can you help?',
        'inbound',
        NOW() - INTERVAL '2 hours',
        NULL
    ),
    (
        '00000000-0000-0000-0001-000000000002',
        '00000000-0000-0000-0000-000000000002',
        '00000000-0000-0000-0000-000000000100',
        'msg-qa-002', 'thread-qa-001',
        'alice.customer@example.com',
        '["admin@omnir.test"]',
        'Re: Support request: cannot export contacts',
        'Following up — still having the issue. Is this a known bug?',
        'inbound',
        NOW() - INTERVAL '1 hour',
        NULL
    ),
    -- Thread 2: 1 unread message (pricing inquiry)
    (
        '00000000-0000-0000-0002-000000000001',
        '00000000-0000-0000-0000-000000000002',
        '00000000-0000-0000-0000-000000000100',
        'msg-qa-003', 'thread-qa-002',
        'bob.prospect@example.com',
        '["admin@omnir.test"]',
        'Pricing question for enterprise plan',
        'Hello, I would like to know more about the enterprise pricing. We have a team of 50.',
        'inbound',
        NOW() - INTERVAL '30 minutes',
        NULL
    ),
    -- Thread 3: already-read baseline
    (
        '00000000-0000-0000-0003-000000000001',
        '00000000-0000-0000-0000-000000000002',
        '00000000-0000-0000-0000-000000000100',
        'msg-qa-004', 'thread-qa-003',
        'carol.existing@example.com',
        '["admin@omnir.test"]',
        'Thank you for the onboarding call',
        'Thanks for the onboarding session yesterday. Very helpful!',
        'inbound',
        NOW() - INTERVAL '1 day',
        NOW() - INTERVAL '23 hours'
    )
    ON CONFLICT (connection_id, message_id) DO NOTHING;
END $$;
-- +goose StatementEnd

-- +goose Down
DELETE FROM email_connections WHERE id = '00000000-0000-0000-0000-000000000100';
