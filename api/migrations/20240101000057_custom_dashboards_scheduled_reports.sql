-- +goose Up
CREATE TABLE custom_dashboards (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID        NOT NULL,
    name       TEXT        NOT NULL,
    widgets    JSONB       NOT NULL DEFAULT '[]',
    created_by UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_custom_dashboards_org_id ON custom_dashboards (org_id);

CREATE TABLE scheduled_reports (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL,
    dashboard_id UUID        NOT NULL REFERENCES custom_dashboards(id) ON DELETE CASCADE,
    schedule     TEXT        NOT NULL,
    recipients   JSONB       NOT NULL DEFAULT '[]',
    last_sent_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scheduled_reports_org_id       ON scheduled_reports (org_id);
CREATE INDEX idx_scheduled_reports_dashboard_id ON scheduled_reports (dashboard_id);

-- +goose Down
DROP TABLE IF EXISTS scheduled_reports;
DROP TABLE IF EXISTS custom_dashboards;
