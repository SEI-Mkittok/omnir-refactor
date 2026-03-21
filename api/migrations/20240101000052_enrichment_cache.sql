-- +goose Up

-- enrichment_cache stores domain-level company data for contact enrichment.
CREATE TABLE enrichment_cache (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    domain       TEXT        NOT NULL,
    data         JSONB       NOT NULL DEFAULT '{}',
    fetched_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One cache entry per (org, domain) pair so enrichment is org-scoped.
CREATE UNIQUE INDEX idx_enrichment_cache_org_domain ON enrichment_cache (org_id, domain);
CREATE INDEX idx_enrichment_cache_domain ON enrichment_cache (domain);

ALTER TABLE enrichment_cache ENABLE ROW LEVEL SECURITY;
CREATE POLICY enrichment_cache_org_isolation ON enrichment_cache
    USING (org_id = current_setting('app.current_org_id', true)::uuid);

-- +goose Down

DROP TABLE IF EXISTS enrichment_cache;
