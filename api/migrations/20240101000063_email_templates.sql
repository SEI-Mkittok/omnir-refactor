CREATE TABLE email_templates (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_templates_org_id_created_at ON email_templates(org_id, created_at DESC);

ALTER TABLE sequence_steps ADD COLUMN template_id UUID REFERENCES email_templates(id) ON DELETE SET NULL;
