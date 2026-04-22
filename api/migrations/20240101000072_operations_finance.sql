-- +goose Up
CREATE TABLE IF NOT EXISTS service_contracts (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('draft','active','expired','canceled')),
    name TEXT NOT NULL,
    contract_number TEXT NOT NULL,
    billing_cycle TEXT NOT NULL,
    rate_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('planned','active','retired','disposed')),
    name TEXT NOT NULL,
    asset_type TEXT NOT NULL,
    serial_number TEXT,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    contract_id UUID REFERENCES service_contracts(id) ON DELETE SET NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('planned','active','on_hold','completed','canceled')),
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    budget_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    start_date TIMESTAMPTZ,
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS project_tasks (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('todo','in_progress','blocked','done')),
    title TEXT NOT NULL,
    description TEXT,
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ops_invoices (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    contract_id UUID REFERENCES service_contracts(id) ON DELETE SET NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('draft','issued','partially_paid','paid','void')),
    invoice_number TEXT NOT NULL,
    issue_date TIMESTAMPTZ NOT NULL,
    due_date TIMESTAMPTZ,
    currency TEXT NOT NULL DEFAULT 'USD',
    subtotal_cents BIGINT NOT NULL DEFAULT 0,
    tax_cents BIGINT NOT NULL DEFAULT 0,
    total_cents BIGINT NOT NULL DEFAULT 0,
    paid_cents BIGINT NOT NULL DEFAULT 0,
    outstanding_cents BIGINT NOT NULL DEFAULT 0,
    billing_context TEXT NOT NULL DEFAULT 'direct_account_billing',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS time_entries (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    project_task_id UUID REFERENCES project_tasks(id) ON DELETE SET NULL,
    contract_id UUID REFERENCES service_contracts(id) ON DELETE SET NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('draft','approved','billed')),
    entry_date TIMESTAMPTZ NOT NULL,
    minutes INTEGER NOT NULL,
    rate_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    description TEXT,
    billed_invoice_id UUID REFERENCES ops_invoices(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    contract_id UUID REFERENCES service_contracts(id) ON DELETE SET NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('draft','approved','billed')),
    expense_date TIMESTAMPTZ NOT NULL,
    category TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    vendor TEXT,
    notes TEXT,
    billed_invoice_id UUID REFERENCES ops_invoices(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invoice_line_items (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    invoice_id UUID NOT NULL REFERENCES ops_invoices(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    source_type TEXT NOT NULL CHECK (source_type IN ('service_contract','time_entry','expense','manual')),
    source_id UUID,
    description TEXT NOT NULL,
    quantity NUMERIC(12,2) NOT NULL DEFAULT 1,
    unit_price_cents BIGINT NOT NULL,
    amount_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    invoice_id UUID NOT NULL REFERENCES ops_invoices(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('pending','posted','failed','reconciled')),
    payment_date TIMESTAMPTZ NOT NULL,
    method TEXT NOT NULL,
    reference TEXT,
    amount_cents BIGINT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_contracts_org_account ON service_contracts(org_id, account_id);
CREATE INDEX IF NOT EXISTS idx_assets_org_account ON assets(org_id, account_id);
CREATE INDEX IF NOT EXISTS idx_projects_org_account ON projects(org_id, account_id);
CREATE INDEX IF NOT EXISTS idx_project_tasks_org_project ON project_tasks(org_id, project_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_org_account ON time_entries(org_id, account_id);
CREATE INDEX IF NOT EXISTS idx_expenses_org_account ON expenses(org_id, account_id);
CREATE INDEX IF NOT EXISTS idx_ops_invoices_org_account ON ops_invoices(org_id, account_id);
CREATE INDEX IF NOT EXISTS idx_invoice_line_items_org_invoice ON invoice_line_items(org_id, invoice_id);
CREATE INDEX IF NOT EXISTS idx_payments_org_invoice ON payments(org_id, invoice_id);

-- +goose Down
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoice_line_items;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS time_entries;
DROP TABLE IF EXISTS ops_invoices;
DROP TABLE IF EXISTS project_tasks;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS service_contracts;
