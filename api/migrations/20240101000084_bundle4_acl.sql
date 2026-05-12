-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS crm_roles (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    description TEXT,
    system_key  TEXT,
    parent_id   UUID        REFERENCES crm_roles(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, name),
    UNIQUE (org_id, system_key)
);

CREATE TABLE IF NOT EXISTS crm_role_closure (
    org_id        UUID    NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    ancestor_id   UUID    NOT NULL REFERENCES crm_roles(id) ON DELETE CASCADE,
    descendant_id UUID    NOT NULL REFERENCES crm_roles(id) ON DELETE CASCADE,
    depth         INTEGER NOT NULL CHECK (depth >= 0),
    PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE TABLE IF NOT EXISTS crm_profiles (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    description TEXT,
    system_key  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, name),
    UNIQUE (org_id, system_key)
);

CREATE TABLE IF NOT EXISTS crm_profile_permissions (
    org_id     UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES crm_profiles(id) ON DELETE CASCADE,
    module     TEXT NOT NULL,
    action     TEXT NOT NULL CHECK (action IN ('read', 'create', 'update', 'delete', 'export', 'admin')),
    allowed    BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (profile_id, module, action)
);

CREATE TABLE IF NOT EXISTS crm_profile_field_permissions (
    org_id      UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    profile_id  UUID NOT NULL REFERENCES crm_profiles(id) ON DELETE CASCADE,
    module      TEXT NOT NULL,
    field_name  TEXT NOT NULL,
    can_write   BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (profile_id, module, field_name)
);

CREATE TABLE IF NOT EXISTS crm_groups (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, name)
);

CREATE TABLE IF NOT EXISTS crm_group_members (
    org_id     UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    group_id   UUID NOT NULL REFERENCES crm_groups(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS crm_sharing_defaults (
    org_id     UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    module     TEXT NOT NULL,
    mode       TEXT NOT NULL CHECK (mode IN ('private', 'public_read', 'public_rw')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_id, module)
);

CREATE TABLE IF NOT EXISTS crm_sharing_grants (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    module       TEXT NOT NULL,
    grantee_type TEXT NOT NULL CHECK (grantee_type IN ('role', 'group')),
    grantee_id   UUID NOT NULL,
    access_level TEXT NOT NULL CHECK (access_level IN ('read', 'write')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, module, grantee_type, grantee_id, access_level)
);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES crm_roles(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS profile_id UUID REFERENCES crm_profiles(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_crm_roles_org_parent ON crm_roles(org_id, parent_id);
CREATE INDEX IF NOT EXISTS idx_crm_role_closure_descendant ON crm_role_closure(descendant_id, ancestor_id);
CREATE INDEX IF NOT EXISTS idx_crm_profiles_org ON crm_profiles(org_id);
CREATE INDEX IF NOT EXISTS idx_crm_group_members_user ON crm_group_members(user_id, group_id);
CREATE INDEX IF NOT EXISTS idx_crm_sharing_grants_grantee ON crm_sharing_grants(org_id, module, grantee_type, grantee_id);
CREATE INDEX IF NOT EXISTS idx_users_acl_role_profile ON users(org_id, role_id, profile_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
WITH role_seed(system_key, name, description) AS (
    VALUES
        ('super_admin', 'Super Admin', 'Cross-organization operator role.'),
        ('admin', 'Administrator', 'Workspace administration role.'),
        ('agent', 'Agent', 'Standard internal CRM user role.'),
        ('client', 'Client', 'Customer portal role.')
)
INSERT INTO crm_roles (org_id, system_key, name, description)
SELECT o.id, s.system_key, s.name, s.description
FROM orgs o
CROSS JOIN role_seed s
ON CONFLICT (org_id, system_key) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();

UPDATE crm_roles child
SET parent_id = parent.id,
    updated_at = NOW()
FROM crm_roles parent
WHERE child.org_id = parent.org_id
  AND (
      (child.system_key = 'admin' AND parent.system_key = 'super_admin')
      OR (child.system_key = 'agent' AND parent.system_key = 'admin')
      OR (child.system_key = 'client' AND parent.system_key = 'agent')
  );
-- +goose StatementEnd

-- +goose StatementBegin
WITH profile_seed(system_key, name, description) AS (
    VALUES
        ('administrator', 'Administrator', 'Full organization administration and data access.'),
        ('sales', 'Sales', 'CRM sales operations access.'),
        ('operations', 'Operations', 'Support and operations access.'),
        ('executive', 'Executive', 'Read-only reporting and CRM oversight.'),
        ('partner', 'Partner', 'Limited partner-facing CRM visibility.'),
        ('guest', 'Guest', 'Minimal read-only access.')
)
INSERT INTO crm_profiles (org_id, system_key, name, description)
SELECT o.id, s.system_key, s.name, s.description
FROM orgs o
CROSS JOIN profile_seed s
ON CONFLICT (org_id, system_key) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();
-- +goose StatementEnd

-- +goose StatementBegin
WITH RECURSIVE walk AS (
    SELECT org_id, id AS ancestor_id, id AS descendant_id, 0 AS depth
    FROM crm_roles
    UNION ALL
    SELECT p.org_id, w.ancestor_id, c.id AS descendant_id, w.depth + 1
    FROM walk w
    JOIN crm_roles p ON p.id = w.descendant_id
    JOIN crm_roles c ON c.parent_id = p.id
)
INSERT INTO crm_role_closure (org_id, ancestor_id, descendant_id, depth)
SELECT org_id, ancestor_id, descendant_id, depth
FROM walk
ON CONFLICT (ancestor_id, descendant_id) DO UPDATE
SET depth = EXCLUDED.depth;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE users u
SET role_id = r.id,
    profile_id = p.id
FROM crm_roles r
JOIN crm_profiles p ON p.org_id = r.org_id
WHERE u.org_id = r.org_id
  AND r.system_key = u.role
  AND p.system_key = CASE
      WHEN u.role IN ('super_admin', 'admin') THEN 'administrator'
      WHEN u.role = 'agent' THEN 'sales'
      ELSE 'guest'
  END
  AND (u.role_id IS NULL OR u.profile_id IS NULL);
-- +goose StatementEnd

-- +goose StatementBegin
WITH modules(module) AS (
    VALUES
        ('accounts'), ('activities'), ('api_keys'), ('audit_log'), ('automations'),
        ('billing'), ('calendar'), ('contacts'), ('custom_fields'), ('dashboards'),
        ('deals'), ('email_templates'), ('emails'), ('export'), ('integrations'),
        ('kb'), ('leads'), ('notifications'), ('onboarding'), ('ops_finance'),
        ('products'), ('quotes'), ('reports'), ('search'), ('sequences'),
        ('settings'), ('sla'), ('tickets'), ('timeline'), ('users'), ('views'), ('webhooks')
),
actions(action) AS (
    VALUES ('read'), ('create'), ('update'), ('delete'), ('export'), ('admin')
),
profiles AS (
    SELECT p.*, m.module, a.action
    FROM crm_profiles p
    CROSS JOIN modules m
    CROSS JOIN actions a
)
INSERT INTO crm_profile_permissions (org_id, profile_id, module, action, allowed)
SELECT org_id, id, module, action,
    CASE
        WHEN system_key = 'administrator' THEN TRUE
        WHEN system_key = 'sales' THEN
            action = 'read'
            OR (module IN ('accounts', 'activities', 'automations', 'calendar', 'contacts', 'dashboards', 'deals', 'email_templates', 'emails', 'export', 'kb', 'leads', 'notifications', 'products', 'quotes', 'reports', 'search', 'sequences', 'tickets', 'timeline', 'views') AND action IN ('create', 'update', 'delete', 'export'))
        WHEN system_key = 'operations' THEN
            action = 'read'
            OR (module IN ('activities', 'calendar', 'dashboards', 'email_templates', 'emails', 'export', 'kb', 'notifications', 'reports', 'search', 'sla', 'tickets', 'timeline', 'views') AND action IN ('create', 'update', 'delete', 'export'))
        WHEN system_key = 'executive' THEN action IN ('read', 'export')
        WHEN system_key IN ('partner', 'guest') THEN module IN ('accounts', 'contacts', 'deals', 'kb', 'notifications', 'tickets') AND action = 'read'
        ELSE FALSE
    END
FROM profiles
ON CONFLICT (profile_id, module, action) DO UPDATE
SET allowed = EXCLUDED.allowed;
-- +goose StatementEnd

-- +goose StatementBegin
WITH sharing_modules(module) AS (
    VALUES ('accounts'), ('contacts'), ('deals'), ('leads'), ('tickets')
)
INSERT INTO crm_sharing_defaults (org_id, module, mode)
SELECT o.id, m.module, 'private'
FROM orgs o
CROSS JOIN sharing_modules m
ON CONFLICT (org_id, module) DO NOTHING;

WITH sharing_modules(module) AS (
    VALUES ('accounts'), ('contacts'), ('deals'), ('leads'), ('tickets')
)
INSERT INTO crm_sharing_grants (org_id, module, grantee_type, grantee_id, access_level)
SELECT r.org_id, m.module, 'role', r.id, 'write'
FROM crm_roles r
CROSS JOIN sharing_modules m
WHERE r.system_key IN ('super_admin', 'admin')
ON CONFLICT (org_id, module, grantee_type, grantee_id, access_level) DO NOTHING;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE crm_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE crm_role_closure ENABLE ROW LEVEL SECURITY;
ALTER TABLE crm_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE crm_profile_permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE crm_profile_field_permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE crm_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE crm_group_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE crm_sharing_defaults ENABLE ROW LEVEL SECURITY;
ALTER TABLE crm_sharing_grants ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS org_isolation ON crm_roles;
DROP POLICY IF EXISTS org_isolation ON crm_role_closure;
DROP POLICY IF EXISTS org_isolation ON crm_profiles;
DROP POLICY IF EXISTS org_isolation ON crm_profile_permissions;
DROP POLICY IF EXISTS org_isolation ON crm_profile_field_permissions;
DROP POLICY IF EXISTS org_isolation ON crm_groups;
DROP POLICY IF EXISTS org_isolation ON crm_group_members;
DROP POLICY IF EXISTS org_isolation ON crm_sharing_defaults;
DROP POLICY IF EXISTS org_isolation ON crm_sharing_grants;

CREATE POLICY org_isolation ON crm_roles USING (org_id = current_org_id());
CREATE POLICY org_isolation ON crm_role_closure USING (org_id = current_org_id());
CREATE POLICY org_isolation ON crm_profiles USING (org_id = current_org_id());
CREATE POLICY org_isolation ON crm_profile_permissions USING (org_id = current_org_id());
CREATE POLICY org_isolation ON crm_profile_field_permissions USING (org_id = current_org_id());
CREATE POLICY org_isolation ON crm_groups USING (org_id = current_org_id());
CREATE POLICY org_isolation ON crm_group_members USING (org_id = current_org_id());
CREATE POLICY org_isolation ON crm_sharing_defaults USING (org_id = current_org_id());
CREATE POLICY org_isolation ON crm_sharing_grants USING (org_id = current_org_id());
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION seed_bundle4_acl_for_org(p_org_id UUID)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    previous_org_id TEXT := current_setting('app.current_org_id', true);
BEGIN
    PERFORM set_config('app.current_org_id', p_org_id::TEXT, true);

    WITH role_seed(system_key, name, description) AS (
        VALUES
            ('super_admin', 'Super Admin', 'Cross-organization operator role.'),
            ('admin', 'Administrator', 'Workspace administration role.'),
            ('agent', 'Agent', 'Standard internal CRM user role.'),
            ('client', 'Client', 'Customer portal role.')
    )
    INSERT INTO crm_roles (org_id, system_key, name, description)
    SELECT p_org_id, s.system_key, s.name, s.description
    FROM role_seed s
    ON CONFLICT (org_id, system_key) DO UPDATE
    SET name = EXCLUDED.name,
        description = EXCLUDED.description,
        updated_at = NOW();

    UPDATE crm_roles child
    SET parent_id = parent.id,
        updated_at = NOW()
    FROM crm_roles parent
    WHERE child.org_id = p_org_id
      AND child.org_id = parent.org_id
      AND (
          (child.system_key = 'admin' AND parent.system_key = 'super_admin')
          OR (child.system_key = 'agent' AND parent.system_key = 'admin')
          OR (child.system_key = 'client' AND parent.system_key = 'agent')
      );

    WITH profile_seed(system_key, name, description) AS (
        VALUES
            ('administrator', 'Administrator', 'Full organization administration and data access.'),
            ('sales', 'Sales', 'CRM sales operations access.'),
            ('operations', 'Operations', 'Support and operations access.'),
            ('executive', 'Executive', 'Read-only reporting and CRM oversight.'),
            ('partner', 'Partner', 'Limited partner-facing CRM visibility.'),
            ('guest', 'Guest', 'Minimal read-only access.')
    )
    INSERT INTO crm_profiles (org_id, system_key, name, description)
    SELECT p_org_id, s.system_key, s.name, s.description
    FROM profile_seed s
    ON CONFLICT (org_id, system_key) DO UPDATE
    SET name = EXCLUDED.name,
        description = EXCLUDED.description,
        updated_at = NOW();

    DELETE FROM crm_role_closure WHERE org_id = p_org_id;
    WITH RECURSIVE walk AS (
        SELECT org_id, id AS ancestor_id, id AS descendant_id, 0 AS depth
        FROM crm_roles
        WHERE org_id = p_org_id
        UNION ALL
        SELECT p.org_id, w.ancestor_id, c.id AS descendant_id, w.depth + 1
        FROM walk w
        JOIN crm_roles p ON p.id = w.descendant_id
        JOIN crm_roles c ON c.parent_id = p.id
        WHERE c.org_id = p_org_id
    )
    INSERT INTO crm_role_closure (org_id, ancestor_id, descendant_id, depth)
    SELECT org_id, ancestor_id, descendant_id, depth
    FROM walk;

    WITH modules(module) AS (
        VALUES
            ('accounts'), ('activities'), ('api_keys'), ('audit_log'), ('automations'),
            ('billing'), ('calendar'), ('contacts'), ('custom_fields'), ('dashboards'),
            ('deals'), ('email_templates'), ('emails'), ('export'), ('integrations'),
            ('kb'), ('leads'), ('notifications'), ('onboarding'), ('ops_finance'),
            ('products'), ('quotes'), ('reports'), ('search'), ('sequences'),
            ('settings'), ('sla'), ('tickets'), ('timeline'), ('users'), ('views'), ('webhooks')
    ),
    actions(action) AS (
        VALUES ('read'), ('create'), ('update'), ('delete'), ('export'), ('admin')
    ),
    profiles AS (
        SELECT p.*, m.module, a.action
        FROM crm_profiles p
        CROSS JOIN modules m
        CROSS JOIN actions a
        WHERE p.org_id = p_org_id
    )
    INSERT INTO crm_profile_permissions (org_id, profile_id, module, action, allowed)
    SELECT org_id, id, module, action,
        CASE
            WHEN system_key = 'administrator' THEN TRUE
            WHEN system_key = 'sales' THEN
                action = 'read'
                OR (module IN ('accounts', 'activities', 'automations', 'calendar', 'contacts', 'dashboards', 'deals', 'email_templates', 'emails', 'export', 'kb', 'leads', 'notifications', 'products', 'quotes', 'reports', 'search', 'sequences', 'tickets', 'timeline', 'views') AND action IN ('create', 'update', 'delete', 'export'))
            WHEN system_key = 'operations' THEN
                action = 'read'
                OR (module IN ('activities', 'calendar', 'dashboards', 'email_templates', 'emails', 'export', 'kb', 'notifications', 'reports', 'search', 'sla', 'tickets', 'timeline', 'views') AND action IN ('create', 'update', 'delete', 'export'))
            WHEN system_key = 'executive' THEN action IN ('read', 'export')
            WHEN system_key IN ('partner', 'guest') THEN module IN ('accounts', 'contacts', 'deals', 'kb', 'notifications', 'tickets') AND action = 'read'
            ELSE FALSE
        END
    FROM profiles
    ON CONFLICT (profile_id, module, action) DO UPDATE
    SET allowed = EXCLUDED.allowed;

    WITH sharing_modules(module) AS (
        VALUES ('accounts'), ('contacts'), ('deals'), ('leads'), ('tickets')
    )
    INSERT INTO crm_sharing_defaults (org_id, module, mode)
    SELECT p_org_id, module, 'private'
    FROM sharing_modules
    ON CONFLICT (org_id, module) DO NOTHING;

    WITH sharing_modules(module) AS (
        VALUES ('accounts'), ('contacts'), ('deals'), ('leads'), ('tickets')
    )
    INSERT INTO crm_sharing_grants (org_id, module, grantee_type, grantee_id, access_level)
    SELECT r.org_id, m.module, 'role', r.id, 'write'
    FROM crm_roles r
    CROSS JOIN sharing_modules m
    WHERE r.org_id = p_org_id
      AND r.system_key IN ('super_admin', 'admin')
    ON CONFLICT (org_id, module, grantee_type, grantee_id, access_level) DO NOTHING;

    PERFORM set_config('app.current_org_id', COALESCE(previous_org_id, ''), true);
END;
$$;

DROP TRIGGER IF EXISTS trg_seed_bundle4_acl_on_org ON orgs;
CREATE OR REPLACE FUNCTION trigger_seed_bundle4_acl_for_org()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
    PERFORM seed_bundle4_acl_for_org(NEW.id);
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_seed_bundle4_acl_on_org
AFTER INSERT ON orgs
FOR EACH ROW
EXECUTE FUNCTION trigger_seed_bundle4_acl_for_org();

SELECT seed_bundle4_acl_for_org(id) FROM orgs;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_seed_bundle4_acl_on_org ON orgs;
DROP FUNCTION IF EXISTS trigger_seed_bundle4_acl_for_org();
DROP FUNCTION IF EXISTS seed_bundle4_acl_for_org(UUID);

ALTER TABLE users
    DROP COLUMN IF EXISTS profile_id,
    DROP COLUMN IF EXISTS role_id;

DROP TABLE IF EXISTS crm_sharing_grants;
DROP TABLE IF EXISTS crm_sharing_defaults;
DROP TABLE IF EXISTS crm_group_members;
DROP TABLE IF EXISTS crm_groups;
DROP TABLE IF EXISTS crm_profile_field_permissions;
DROP TABLE IF EXISTS crm_profile_permissions;
DROP TABLE IF EXISTS crm_profiles;
DROP TABLE IF EXISTS crm_role_closure;
DROP TABLE IF EXISTS crm_roles;
-- +goose StatementEnd
