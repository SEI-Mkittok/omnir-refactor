-- +goose Up
-- +goose StatementBegin
CREATE TABLE email_templates (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_email_templates_org_id_created_at ON email_templates(org_id, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE sequence_steps ADD COLUMN template_id UUID REFERENCES email_templates(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sequence_steps DROP COLUMN IF EXISTS template_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_email_templates_org_id_created_at;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS email_templates;
-- +goose StatementEnd
