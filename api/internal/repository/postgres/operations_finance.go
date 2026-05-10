package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type OperationsFinanceRepo struct{ db *pgxpool.Pool }

func NewOperationsFinanceRepo(db *pgxpool.Pool) *OperationsFinanceRepo {
	return &OperationsFinanceRepo{db: db}
}

func scanErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func (r *OperationsFinanceRepo) CreateServiceContract(ctx context.Context, v *domain.ServiceContract) (*domain.ServiceContract, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.Currency == "" {
		cur, err := getOrgDefaultCurrency(ctx, r.db, v.OrgID)
		if err != nil {
			return nil, err
		}
		v.Currency = cur
	}
	if v.Status == "" {
		v.Status = domain.ServiceContractStatusDraft
	}
	row := r.db.QueryRow(ctx, `INSERT INTO service_contracts (id, org_id, account_id, contact_id, owner_id, created_by, updated_by, status, name, contract_number, billing_cycle, rate_cents, currency, start_date, end_date)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
RETURNING id, org_id, account_id, contact_id, owner_id, created_by, updated_by, status, name, contract_number, billing_cycle, rate_cents, currency, start_date, end_date, created_at, updated_at`,
		v.ID, v.OrgID, v.AccountID, v.ContactID, v.OwnerID, v.CreatedBy, v.UpdatedBy, v.Status, v.Name, v.ContractNumber, v.BillingCycle, v.RateCents, v.Currency, v.StartDate, v.EndDate)
	out := domain.ServiceContract{}
	err := row.Scan(&out.ID, &out.OrgID, &out.AccountID, &out.ContactID, &out.OwnerID, &out.CreatedBy, &out.UpdatedBy, &out.Status, &out.Name, &out.ContractNumber, &out.BillingCycle, &out.RateCents, &out.Currency, &out.StartDate, &out.EndDate, &out.CreatedAt, &out.UpdatedAt)
	return &out, scanErr(err)
}

func (r *OperationsFinanceRepo) CreateAsset(ctx context.Context, v *domain.Asset) (*domain.Asset, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.Status == "" {
		v.Status = domain.AssetStatusPlanned
	}
	row := r.db.QueryRow(ctx, `INSERT INTO assets (id, org_id, account_id, contact_id, owner_id, created_by, updated_by, status, name, asset_type, serial_number, description)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING id, org_id, account_id, contact_id, owner_id, created_by, updated_by, status, name, asset_type, serial_number, description, created_at, updated_at`,
		v.ID, v.OrgID, v.AccountID, v.ContactID, v.OwnerID, v.CreatedBy, v.UpdatedBy, v.Status, v.Name, v.AssetType, nilIfEmpty(v.SerialNumber), nilIfEmpty(v.Description))
	out := domain.Asset{}
	err := row.Scan(&out.ID, &out.OrgID, &out.AccountID, &out.ContactID, &out.OwnerID, &out.CreatedBy, &out.UpdatedBy, &out.Status, &out.Name, &out.AssetType, &out.SerialNumber, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	return &out, scanErr(err)
}

func (r *OperationsFinanceRepo) CreateProject(ctx context.Context, v *domain.Project) (*domain.Project, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.Currency == "" {
		cur, err := getOrgDefaultCurrency(ctx, r.db, v.OrgID)
		if err != nil {
			return nil, err
		}
		v.Currency = cur
	}
	if v.Status == "" {
		v.Status = domain.ProjectStatusPlanned
	}
	row := r.db.QueryRow(ctx, `INSERT INTO projects (id, org_id, account_id, contact_id, contract_id, owner_id, created_by, updated_by, status, name, code, budget_cents, currency, start_date, due_date, completed_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
RETURNING id, org_id, account_id, contact_id, contract_id, owner_id, created_by, updated_by, status, name, code, budget_cents, currency, start_date, due_date, completed_at, created_at, updated_at`,
		v.ID, v.OrgID, v.AccountID, v.ContactID, v.ContractID, v.OwnerID, v.CreatedBy, v.UpdatedBy, v.Status, v.Name, v.Code, v.BudgetCents, v.Currency, v.StartDate, v.DueDate, v.CompletedAt)
	out := domain.Project{}
	err := row.Scan(&out.ID, &out.OrgID, &out.AccountID, &out.ContactID, &out.ContractID, &out.OwnerID, &out.CreatedBy, &out.UpdatedBy, &out.Status, &out.Name, &out.Code, &out.BudgetCents, &out.Currency, &out.StartDate, &out.DueDate, &out.CompletedAt, &out.CreatedAt, &out.UpdatedAt)
	return &out, scanErr(err)
}

func (r *OperationsFinanceRepo) CreateProjectTask(ctx context.Context, v *domain.ProjectTask) (*domain.ProjectTask, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.Status == "" {
		v.Status = domain.ProjectTaskStatusTodo
	}
	row := r.db.QueryRow(ctx, `INSERT INTO project_tasks (id, org_id, account_id, contact_id, project_id, owner_id, created_by, updated_by, status, title, description, due_date, completed_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
RETURNING id, org_id, account_id, contact_id, project_id, owner_id, created_by, updated_by, status, title, description, due_date, completed_at, created_at, updated_at`,
		v.ID, v.OrgID, v.AccountID, v.ContactID, v.ProjectID, v.OwnerID, v.CreatedBy, v.UpdatedBy, v.Status, v.Title, nilIfEmpty(v.Description), v.DueDate, v.CompletedAt)
	out := domain.ProjectTask{}
	err := row.Scan(&out.ID, &out.OrgID, &out.AccountID, &out.ContactID, &out.ProjectID, &out.OwnerID, &out.CreatedBy, &out.UpdatedBy, &out.Status, &out.Title, &out.Description, &out.DueDate, &out.CompletedAt, &out.CreatedAt, &out.UpdatedAt)
	return &out, scanErr(err)
}

func (r *OperationsFinanceRepo) CreateTimeEntry(ctx context.Context, v *domain.TimeEntry) (*domain.TimeEntry, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.Currency == "" {
		cur, err := getOrgDefaultCurrency(ctx, r.db, v.OrgID)
		if err != nil {
			return nil, err
		}
		v.Currency = cur
	}
	if v.Status == "" {
		v.Status = domain.TimeEntryStatusDraft
	}
	row := r.db.QueryRow(ctx, `INSERT INTO time_entries (id, org_id, account_id, contact_id, project_id, project_task_id, contract_id, owner_id, created_by, updated_by, status, entry_date, minutes, rate_cents, currency, description, billed_invoice_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
RETURNING id, org_id, account_id, contact_id, project_id, project_task_id, contract_id, owner_id, created_by, updated_by, status, entry_date, minutes, rate_cents, currency, description, billed_invoice_id, created_at, updated_at`,
		v.ID, v.OrgID, v.AccountID, v.ContactID, v.ProjectID, v.ProjectTaskID, v.ContractID, v.OwnerID, v.CreatedBy, v.UpdatedBy, v.Status, v.EntryDate, v.Minutes, v.RateCents, v.Currency, nilIfEmpty(v.Description), v.BilledInvoiceID)
	out := domain.TimeEntry{}
	err := row.Scan(&out.ID, &out.OrgID, &out.AccountID, &out.ContactID, &out.ProjectID, &out.ProjectTaskID, &out.ContractID, &out.OwnerID, &out.CreatedBy, &out.UpdatedBy, &out.Status, &out.EntryDate, &out.Minutes, &out.RateCents, &out.Currency, &out.Description, &out.BilledInvoiceID, &out.CreatedAt, &out.UpdatedAt)
	return &out, scanErr(err)
}

func (r *OperationsFinanceRepo) CreateExpense(ctx context.Context, v *domain.Expense) (*domain.Expense, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.Currency == "" {
		cur, err := getOrgDefaultCurrency(ctx, r.db, v.OrgID)
		if err != nil {
			return nil, err
		}
		v.Currency = cur
	}
	if v.Status == "" {
		v.Status = domain.ExpenseStatusDraft
	}
	row := r.db.QueryRow(ctx, `INSERT INTO expenses (id, org_id, account_id, contact_id, project_id, contract_id, owner_id, created_by, updated_by, status, expense_date, category, amount_cents, currency, vendor, notes, billed_invoice_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
RETURNING id, org_id, account_id, contact_id, project_id, contract_id, owner_id, created_by, updated_by, status, expense_date, category, amount_cents, currency, vendor, notes, billed_invoice_id, created_at, updated_at`,
		v.ID, v.OrgID, v.AccountID, v.ContactID, v.ProjectID, v.ContractID, v.OwnerID, v.CreatedBy, v.UpdatedBy, v.Status, v.ExpenseDate, v.Category, v.AmountCents, v.Currency, nilIfEmpty(v.Vendor), nilIfEmpty(v.Notes), v.BilledInvoiceID)
	out := domain.Expense{}
	err := row.Scan(&out.ID, &out.OrgID, &out.AccountID, &out.ContactID, &out.ProjectID, &out.ContractID, &out.OwnerID, &out.CreatedBy, &out.UpdatedBy, &out.Status, &out.ExpenseDate, &out.Category, &out.AmountCents, &out.Currency, &out.Vendor, &out.Notes, &out.BilledInvoiceID, &out.CreatedAt, &out.UpdatedAt)
	return &out, scanErr(err)
}

func (r *OperationsFinanceRepo) CreateInvoice(ctx context.Context, v *domain.OpsInvoice) (*domain.OpsInvoice, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.Currency == "" {
		cur, err := getOrgDefaultCurrency(ctx, r.db, v.OrgID)
		if err != nil {
			return nil, err
		}
		v.Currency = cur
	}
	if v.Status == "" {
		v.Status = domain.OpsInvoiceStatusDraft
	}
	if strings.TrimSpace(v.InvoiceNumber) == "" {
		nextNum, err := getNextDocNumber(ctx, r.db, v.OrgID, domain.DocTypeInvoice)
		if err != nil {
			return nil, err
		}
		prefix, err := getDocPrefix(ctx, r.db, v.OrgID, domain.DocTypeInvoice)
		if err != nil {
			return nil, err
		}
		v.InvoiceNumber = domain.FormattedDocNumber(prefix, nextNum)
	}
	row := r.db.QueryRow(ctx, `INSERT INTO ops_invoices (id, org_id, account_id, contact_id, contract_id, owner_id, created_by, updated_by, status, invoice_number, issue_date, due_date, currency, subtotal_cents, tax_cents, total_cents, paid_cents, outstanding_cents, billing_context)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
RETURNING id, org_id, account_id, contact_id, contract_id, owner_id, created_by, updated_by, status, invoice_number, issue_date, due_date, currency, subtotal_cents, tax_cents, total_cents, paid_cents, outstanding_cents, billing_context, created_at, updated_at`,
		v.ID, v.OrgID, v.AccountID, v.ContactID, v.ContractID, v.OwnerID, v.CreatedBy, v.UpdatedBy, v.Status, v.InvoiceNumber, v.IssueDate, v.DueDate, v.Currency, v.SubtotalCents, v.TaxCents, v.TotalCents, v.PaidCents, v.OutstandingCents, v.BillingContext)
	out := domain.OpsInvoice{}
	err := row.Scan(&out.ID, &out.OrgID, &out.AccountID, &out.ContactID, &out.ContractID, &out.OwnerID, &out.CreatedBy, &out.UpdatedBy, &out.Status, &out.InvoiceNumber, &out.IssueDate, &out.DueDate, &out.Currency, &out.SubtotalCents, &out.TaxCents, &out.TotalCents, &out.PaidCents, &out.OutstandingCents, &out.BillingContext, &out.CreatedAt, &out.UpdatedAt)
	return &out, scanErr(err)
}

func (r *OperationsFinanceRepo) CreateInvoiceLineItem(ctx context.Context, v *domain.InvoiceLineItem) (*domain.InvoiceLineItem, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	row := r.db.QueryRow(ctx, `INSERT INTO invoice_line_items (id, org_id, account_id, contact_id, invoice_id, owner_id, created_by, updated_by, source_type, source_id, description, quantity, unit_price_cents, amount_cents)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
RETURNING id, org_id, account_id, contact_id, invoice_id, owner_id, created_by, updated_by, source_type, source_id, description, quantity, unit_price_cents, amount_cents, created_at, updated_at`,
		v.ID, v.OrgID, v.AccountID, v.ContactID, v.InvoiceID, v.OwnerID, v.CreatedBy, v.UpdatedBy, v.SourceType, v.SourceID, v.Description, v.Quantity, v.UnitPriceCents, v.AmountCents)
	out := domain.InvoiceLineItem{}
	err := row.Scan(&out.ID, &out.OrgID, &out.AccountID, &out.ContactID, &out.InvoiceID, &out.OwnerID, &out.CreatedBy, &out.UpdatedBy, &out.SourceType, &out.SourceID, &out.Description, &out.Quantity, &out.UnitPriceCents, &out.AmountCents, &out.CreatedAt, &out.UpdatedAt)
	return &out, scanErr(err)
}

func (r *OperationsFinanceRepo) CreatePayment(ctx context.Context, v *domain.Payment) (*domain.Payment, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.Currency == "" {
		cur, err := getOrgDefaultCurrency(ctx, r.db, v.OrgID)
		if err != nil {
			return nil, err
		}
		v.Currency = cur
	}
	if v.Status == "" {
		v.Status = domain.PaymentStatusPending
	}
	row := r.db.QueryRow(ctx, `INSERT INTO payments (id, org_id, account_id, contact_id, invoice_id, owner_id, created_by, updated_by, status, payment_date, method, reference, amount_cents, currency)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
RETURNING id, org_id, account_id, contact_id, invoice_id, owner_id, created_by, updated_by, status, payment_date, method, reference, amount_cents, currency, created_at, updated_at`,
		v.ID, v.OrgID, v.AccountID, v.ContactID, v.InvoiceID, v.OwnerID, v.CreatedBy, v.UpdatedBy, v.Status, v.PaymentDate, v.Method, nilIfEmpty(v.Reference), v.AmountCents, v.Currency)
	out := domain.Payment{}
	err := row.Scan(&out.ID, &out.OrgID, &out.AccountID, &out.ContactID, &out.InvoiceID, &out.OwnerID, &out.CreatedBy, &out.UpdatedBy, &out.Status, &out.PaymentDate, &out.Method, &out.Reference, &out.AmountCents, &out.Currency, &out.CreatedAt, &out.UpdatedAt)
	return &out, scanErr(err)
}

func (r *OperationsFinanceRepo) ListServiceContracts(ctx context.Context, orgID uuid.UUID) ([]*domain.ServiceContract, error) {
	rows, err := r.db.Query(ctx, `SELECT id, org_id, account_id, contact_id, owner_id, created_by, updated_by, status, name, contract_number, billing_cycle, rate_cents, currency, start_date, end_date, created_at, updated_at FROM service_contracts WHERE org_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ServiceContract
	for rows.Next() {
		v := domain.ServiceContract{}
		if err := rows.Scan(&v.ID, &v.OrgID, &v.AccountID, &v.ContactID, &v.OwnerID, &v.CreatedBy, &v.UpdatedBy, &v.Status, &v.Name, &v.ContractNumber, &v.BillingCycle, &v.RateCents, &v.Currency, &v.StartDate, &v.EndDate, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}
func (r *OperationsFinanceRepo) ListAssets(ctx context.Context, orgID uuid.UUID) ([]*domain.Asset, error) {
	rows, err := r.db.Query(ctx, `SELECT id, org_id, account_id, contact_id, owner_id, created_by, updated_by, status, name, asset_type, serial_number, description, created_at, updated_at FROM assets WHERE org_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Asset
	for rows.Next() {
		v := domain.Asset{}
		if err := rows.Scan(&v.ID, &v.OrgID, &v.AccountID, &v.ContactID, &v.OwnerID, &v.CreatedBy, &v.UpdatedBy, &v.Status, &v.Name, &v.AssetType, &v.SerialNumber, &v.Description, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}
func (r *OperationsFinanceRepo) ListProjects(ctx context.Context, orgID uuid.UUID) ([]*domain.Project, error) {
	rows, err := r.db.Query(ctx, `SELECT id, org_id, account_id, contact_id, contract_id, owner_id, created_by, updated_by, status, name, code, budget_cents, currency, start_date, due_date, completed_at, created_at, updated_at FROM projects WHERE org_id=$1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Project
	for rows.Next() {
		v := domain.Project{}
		if err := rows.Scan(&v.ID, &v.OrgID, &v.AccountID, &v.ContactID, &v.ContractID, &v.OwnerID, &v.CreatedBy, &v.UpdatedBy, &v.Status, &v.Name, &v.Code, &v.BudgetCents, &v.Currency, &v.StartDate, &v.DueDate, &v.CompletedAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}
func (r *OperationsFinanceRepo) ListProjectTasks(ctx context.Context, orgID uuid.UUID, projectID uuid.UUID) ([]*domain.ProjectTask, error) {
	rows, err := r.db.Query(ctx, `SELECT id, org_id, account_id, contact_id, project_id, owner_id, created_by, updated_by, status, title, description, due_date, completed_at, created_at, updated_at FROM project_tasks WHERE org_id=$1 AND project_id=$2 ORDER BY created_at DESC`, orgID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ProjectTask
	for rows.Next() {
		v := domain.ProjectTask{}
		if err := rows.Scan(&v.ID, &v.OrgID, &v.AccountID, &v.ContactID, &v.ProjectID, &v.OwnerID, &v.CreatedBy, &v.UpdatedBy, &v.Status, &v.Title, &v.Description, &v.DueDate, &v.CompletedAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}
func (r *OperationsFinanceRepo) ListTimeEntries(ctx context.Context, orgID uuid.UUID, accountID uuid.UUID) ([]*domain.TimeEntry, error) {
	rows, err := r.db.Query(ctx, `SELECT id, org_id, account_id, contact_id, project_id, project_task_id, contract_id, owner_id, created_by, updated_by, status, entry_date, minutes, rate_cents, currency, description, billed_invoice_id, created_at, updated_at FROM time_entries WHERE org_id=$1 AND account_id=$2 ORDER BY entry_date DESC`, orgID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.TimeEntry
	for rows.Next() {
		v := domain.TimeEntry{}
		if err := rows.Scan(&v.ID, &v.OrgID, &v.AccountID, &v.ContactID, &v.ProjectID, &v.ProjectTaskID, &v.ContractID, &v.OwnerID, &v.CreatedBy, &v.UpdatedBy, &v.Status, &v.EntryDate, &v.Minutes, &v.RateCents, &v.Currency, &v.Description, &v.BilledInvoiceID, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}
func (r *OperationsFinanceRepo) ListExpenses(ctx context.Context, orgID uuid.UUID, accountID uuid.UUID) ([]*domain.Expense, error) {
	rows, err := r.db.Query(ctx, `SELECT id, org_id, account_id, contact_id, project_id, contract_id, owner_id, created_by, updated_by, status, expense_date, category, amount_cents, currency, vendor, notes, billed_invoice_id, created_at, updated_at FROM expenses WHERE org_id=$1 AND account_id=$2 ORDER BY expense_date DESC`, orgID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Expense
	for rows.Next() {
		v := domain.Expense{}
		if err := rows.Scan(&v.ID, &v.OrgID, &v.AccountID, &v.ContactID, &v.ProjectID, &v.ContractID, &v.OwnerID, &v.CreatedBy, &v.UpdatedBy, &v.Status, &v.ExpenseDate, &v.Category, &v.AmountCents, &v.Currency, &v.Vendor, &v.Notes, &v.BilledInvoiceID, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}

func (r *OperationsFinanceRepo) GetInvoiceAssembly(ctx context.Context, orgID uuid.UUID, invoiceID uuid.UUID) (*domain.InvoiceAssembly, error) {
	inv := domain.OpsInvoice{}
	err := r.db.QueryRow(ctx, `SELECT id, org_id, account_id, contact_id, contract_id, owner_id, created_by, updated_by, status, invoice_number, issue_date, due_date, currency, subtotal_cents, tax_cents, total_cents, paid_cents, outstanding_cents, billing_context, created_at, updated_at FROM ops_invoices WHERE org_id=$1 AND id=$2`, orgID, invoiceID).
		Scan(&inv.ID, &inv.OrgID, &inv.AccountID, &inv.ContactID, &inv.ContractID, &inv.OwnerID, &inv.CreatedBy, &inv.UpdatedBy, &inv.Status, &inv.InvoiceNumber, &inv.IssueDate, &inv.DueDate, &inv.Currency, &inv.SubtotalCents, &inv.TaxCents, &inv.TotalCents, &inv.PaidCents, &inv.OutstandingCents, &inv.BillingContext, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return nil, scanErr(err)
	}
	rows, err := r.db.Query(ctx, `SELECT id, org_id, account_id, contact_id, invoice_id, owner_id, created_by, updated_by, source_type, source_id, description, quantity, unit_price_cents, amount_cents, created_at, updated_at FROM invoice_line_items WHERE org_id=$1 AND invoice_id=$2 ORDER BY created_at ASC`, orgID, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*domain.InvoiceLineItem{}
	for rows.Next() {
		v := domain.InvoiceLineItem{}
		if err := rows.Scan(&v.ID, &v.OrgID, &v.AccountID, &v.ContactID, &v.InvoiceID, &v.OwnerID, &v.CreatedBy, &v.UpdatedBy, &v.SourceType, &v.SourceID, &v.Description, &v.Quantity, &v.UnitPriceCents, &v.AmountCents, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, &v)
	}
	return &domain.InvoiceAssembly{Invoice: &inv, LineItems: items}, rows.Err()
}

func (r *OperationsFinanceRepo) BuildInvoiceAssembly(ctx context.Context, orgID uuid.UUID, accountID uuid.UUID, contactID *uuid.UUID, contractID *uuid.UUID, ownerID uuid.UUID, createdBy uuid.UUID) (*domain.InvoiceAssembly, error) {
	timeRows, err := r.db.Query(ctx, `SELECT id, minutes, rate_cents, COALESCE(description, 'Time entry') FROM time_entries WHERE org_id=$1 AND account_id=$2 AND billed_invoice_id IS NULL AND status IN ('approved','draft')`, orgID, accountID)
	if err != nil {
		return nil, err
	}
	defer timeRows.Close()
	items := []*domain.InvoiceLineItem{}
	var subtotal int64
	for timeRows.Next() {
		var srcID uuid.UUID
		var minutes int
		var rate int64
		var desc string
		if err := timeRows.Scan(&srcID, &minutes, &rate, &desc); err != nil {
			return nil, err
		}
		amount := int64(minutes) * rate / 60
		subtotal += amount
		items = append(items, &domain.InvoiceLineItem{ID: uuid.New(), OrgID: orgID, AccountID: accountID, ContactID: contactID, OwnershipFields: domain.OwnershipFields{OwnerID: ownerID, CreatedBy: createdBy}, SourceType: "time_entry", SourceID: &srcID, Description: desc, Quantity: float64(minutes) / 60, UnitPriceCents: rate, AmountCents: amount})
	}
	expRows, err := r.db.Query(ctx, `SELECT id, amount_cents, COALESCE(notes, category) FROM expenses WHERE org_id=$1 AND account_id=$2 AND billed_invoice_id IS NULL AND status IN ('approved','draft')`, orgID, accountID)
	if err != nil {
		return nil, err
	}
	defer expRows.Close()
	for expRows.Next() {
		var srcID uuid.UUID
		var amount int64
		var desc string
		if err := expRows.Scan(&srcID, &amount, &desc); err != nil {
			return nil, err
		}
		subtotal += amount
		items = append(items, &domain.InvoiceLineItem{ID: uuid.New(), OrgID: orgID, AccountID: accountID, ContactID: contactID, OwnershipFields: domain.OwnershipFields{OwnerID: ownerID, CreatedBy: createdBy}, SourceType: "expense", SourceID: &srcID, Description: desc, Quantity: 1, UnitPriceCents: amount, AmountCents: amount})
	}
	if contractID != nil {
		var rate int64
		var name string
		err = r.db.QueryRow(ctx, `SELECT rate_cents, name FROM service_contracts WHERE org_id=$1 AND id=$2`, orgID, *contractID).Scan(&rate, &name)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			subtotal += rate
			items = append(items, &domain.InvoiceLineItem{ID: uuid.New(), OrgID: orgID, AccountID: accountID, ContactID: contactID, OwnershipFields: domain.OwnershipFields{OwnerID: ownerID, CreatedBy: createdBy}, SourceType: "service_contract", SourceID: contractID, Description: fmt.Sprintf("Contract fee: %s", name), Quantity: 1, UnitPriceCents: rate, AmountCents: rate})
		}
	}
	nextNum, err := getNextDocNumber(ctx, r.db, orgID, domain.DocTypeInvoice)
	if err != nil {
		return nil, err
	}
	prefix, err := getDocPrefix(ctx, r.db, orgID, domain.DocTypeInvoice)
	if err != nil {
		return nil, err
	}
	defaultCurrency, err := getOrgDefaultCurrency(ctx, r.db, orgID)
	if err != nil {
		return nil, err
	}
	inv := &domain.OpsInvoice{
		ID:              uuid.New(),
		OrgID:           orgID,
		AccountID:       accountID,
		ContactID:       contactID,
		ContractID:      contractID,
		OwnershipFields: domain.OwnershipFields{OwnerID: ownerID, CreatedBy: createdBy},
		Status:          domain.OpsInvoiceStatusDraft,
		InvoiceNumber:   domain.FormattedDocNumber(prefix, nextNum),
		IssueDate:       time.Now().UTC(),
		Currency:        defaultCurrency,
		SubtotalCents:   subtotal,
		TaxCents:        0,
		TotalCents:      subtotal,
		OutstandingCents: subtotal,
		BillingContext:  "direct_account_billing",
	}
	asm := &domain.InvoiceAssembly{Invoice: inv, LineItems: items}
	return asm, asm.Validate()
}
