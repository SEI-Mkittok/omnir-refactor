-- +goose Up
-- Phase 9 Area 3: deal client portal links
CREATE TABLE IF NOT EXISTS portal_links (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    deal_id             UUID        NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    created_by_user_id  UUID        REFERENCES users(id) ON DELETE SET NULL,
    token               TEXT        NOT NULL UNIQUE,
    label               TEXT,
    expires_at          TIMESTAMPTZ,
    revoked_at          TIMESTAMPTZ,
    view_count          INTEGER     NOT NULL DEFAULT 0,
    last_viewed_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_portal_links_token ON portal_links (token);
CREATE INDEX IF NOT EXISTS idx_portal_links_deal  ON portal_links (org_id, deal_id);

-- +goose Down
DROP INDEX IF EXISTS idx_portal_links_deal;
DROP INDEX IF EXISTS idx_portal_links_token;
DROP TABLE IF EXISTS portal_links;
