-- +goose Up

-- outbound_webhooks stores registered webhook endpoints per org.
CREATE TABLE outbound_webhooks (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID         NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    url        TEXT         NOT NULL,
    events     TEXT[]       NOT NULL DEFAULT '{}',
    secret     TEXT         NOT NULL,
    active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbound_webhooks_org_id ON outbound_webhooks(org_id);
CREATE UNIQUE INDEX idx_outbound_webhooks_org_url ON outbound_webhooks(org_id, url);

-- webhook_deliveries tracks each outbound delivery attempt.
CREATE TABLE webhook_deliveries (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id    UUID           NOT NULL REFERENCES outbound_webhooks(id) ON DELETE CASCADE,
    event         TEXT           NOT NULL,
    payload       JSONB          NOT NULL,
    status        TEXT           NOT NULL DEFAULT 'pending',
    attempts      INT            NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    delivered_at  TIMESTAMPTZ,
    last_error    TEXT,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_status      ON webhook_deliveries(status) WHERE status IN ('pending', 'failed');
CREATE INDEX idx_webhook_deliveries_next_retry  ON webhook_deliveries(next_retry_at) WHERE status = 'pending';

-- Enable RLS so org-scoping via set_config works automatically.
ALTER TABLE outbound_webhooks  ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_deliveries ENABLE ROW LEVEL SECURITY;

CREATE POLICY outbound_webhooks_org_isolation ON outbound_webhooks
    USING (org_id = current_setting('app.current_org_id', TRUE)::UUID OR
           current_setting('app.current_org_id', TRUE) IS NULL OR
           current_setting('app.current_org_id', TRUE) = '');

CREATE POLICY webhook_deliveries_org_isolation ON webhook_deliveries
    USING (webhook_id IN (
        SELECT id FROM outbound_webhooks
        WHERE org_id = current_setting('app.current_org_id', TRUE)::UUID OR
              current_setting('app.current_org_id', TRUE) IS NULL OR
              current_setting('app.current_org_id', TRUE) = ''
    ));

-- +goose Down
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS outbound_webhooks;
