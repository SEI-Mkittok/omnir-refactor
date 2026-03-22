-- +goose Up
-- +goose StatementBegin
CREATE TABLE email_connections (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID        NOT NULL,
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider       TEXT        NOT NULL CHECK (provider IN ('gmail', 'outlook')),
    email_address  TEXT        NOT NULL,
    access_token   TEXT        NOT NULL,
    refresh_token  TEXT,
    token_expiry   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sync_cursor    TEXT,
    last_synced_at TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, user_id, provider)
);

CREATE INDEX idx_email_connections_org_id  ON email_connections(org_id);
CREATE INDEX idx_email_connections_user_id ON email_connections(user_id);

CREATE TABLE email_inbox_messages (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL,
    connection_id UUID       NOT NULL REFERENCES email_connections(id) ON DELETE CASCADE,
    message_id   TEXT        NOT NULL,
    thread_id    TEXT        NOT NULL,
    from_addr    TEXT        NOT NULL,
    to_addrs     JSONB       NOT NULL DEFAULT '[]',
    subject      TEXT        NOT NULL DEFAULT '',
    body_text    TEXT,
    body_html    TEXT,
    contact_id   UUID        REFERENCES contacts(id) ON DELETE SET NULL,
    direction    TEXT        NOT NULL CHECK (direction IN ('inbound', 'outbound')) DEFAULT 'inbound',
    sent_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (connection_id, message_id)
);

CREATE INDEX idx_email_inbox_messages_org_id       ON email_inbox_messages(org_id);
CREATE INDEX idx_email_inbox_messages_connection_id ON email_inbox_messages(connection_id);
CREATE INDEX idx_email_inbox_messages_thread_id    ON email_inbox_messages(thread_id);
CREATE INDEX idx_email_inbox_messages_contact_id   ON email_inbox_messages(contact_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS email_inbox_messages;
DROP TABLE IF EXISTS email_connections;
-- +goose StatementEnd
