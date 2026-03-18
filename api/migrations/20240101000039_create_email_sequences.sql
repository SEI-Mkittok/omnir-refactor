-- Email sequences: automated multi-step email drip campaigns
-- Tables: email_sequences, sequence_steps, sequence_enrollments, sequence_events

CREATE TYPE sequence_status AS ENUM ('draft', 'active', 'paused', 'archived');
CREATE TYPE sequence_step_kind AS ENUM ('email', 'wait');
CREATE TYPE enrollment_status AS ENUM ('active', 'completed', 'unsubscribed', 'bounced', 'paused');
CREATE TYPE sequence_event_kind AS ENUM ('sent', 'opened', 'clicked', 'completed', 'bounced', 'unsubscribed');

CREATE TABLE email_sequences (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    status      sequence_status NOT NULL DEFAULT 'draft',
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE sequence_steps (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_id UUID NOT NULL REFERENCES email_sequences(id) ON DELETE CASCADE,
    org_id      UUID NOT NULL,
    position    INT NOT NULL,        -- 0-based ordering
    kind        sequence_step_kind NOT NULL,
    -- email step fields
    subject     TEXT,
    body        TEXT,
    -- wait step fields
    wait_duration_hours INT,        -- how long to wait before next step
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sequence_enrollments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_id  UUID NOT NULL REFERENCES email_sequences(id) ON DELETE CASCADE,
    contact_id   UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    org_id       UUID NOT NULL,
    status       enrollment_status NOT NULL DEFAULT 'active',
    current_step INT NOT NULL DEFAULT 0,  -- index of the next step to execute
    next_step_at TIMESTAMPTZ,             -- when the worker should process the next step
    enrolled_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    UNIQUE (sequence_id, contact_id)
);

CREATE TABLE sequence_events (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_id    UUID NOT NULL REFERENCES email_sequences(id) ON DELETE CASCADE,
    step_id        UUID REFERENCES sequence_steps(id) ON DELETE SET NULL,
    enrollment_id  UUID NOT NULL REFERENCES sequence_enrollments(id) ON DELETE CASCADE,
    contact_id     UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    org_id         UUID NOT NULL,
    kind           sequence_event_kind NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_email_sequences_org_id ON email_sequences(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_sequence_steps_sequence_id ON sequence_steps(sequence_id);
CREATE INDEX idx_sequence_enrollments_sequence_id ON sequence_enrollments(sequence_id);
CREATE INDEX idx_sequence_enrollments_contact_id ON sequence_enrollments(contact_id);
CREATE INDEX idx_sequence_events_sequence_id ON sequence_events(sequence_id);
CREATE INDEX idx_sequence_events_enrollment_id ON sequence_events(enrollment_id);

-- RLS
ALTER TABLE email_sequences ENABLE ROW LEVEL SECURITY;
ALTER TABLE sequence_steps ENABLE ROW LEVEL SECURITY;
ALTER TABLE sequence_enrollments ENABLE ROW LEVEL SECURITY;
ALTER TABLE sequence_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_isolation ON email_sequences
    USING (org_id = current_setting('app.org_id', true)::uuid);

CREATE POLICY org_isolation ON sequence_steps
    USING (org_id = current_setting('app.org_id', true)::uuid);

CREATE POLICY org_isolation ON sequence_enrollments
    USING (org_id = current_setting('app.org_id', true)::uuid);

CREATE POLICY org_isolation ON sequence_events
    USING (org_id = current_setting('app.org_id', true)::uuid);
