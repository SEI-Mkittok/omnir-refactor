package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ServiceContractStatus string

type AssetStatus string

type ProjectStatus string

type ProjectTaskStatus string

type TimeEntryStatus string

type ExpenseStatus string

type OpsInvoiceStatus string

type PaymentStatus string

const (
	ServiceContractStatusDraft    ServiceContractStatus = "draft"
	ServiceContractStatusActive   ServiceContractStatus = "active"
	ServiceContractStatusExpired  ServiceContractStatus = "expired"
	ServiceContractStatusCanceled ServiceContractStatus = "canceled"

	AssetStatusPlanned  AssetStatus = "planned"
	AssetStatusActive   AssetStatus = "active"
	AssetStatusRetired  AssetStatus = "retired"
	AssetStatusDisposed AssetStatus = "disposed"

	ProjectStatusPlanned   ProjectStatus = "planned"
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusOnHold    ProjectStatus = "on_hold"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusCanceled  ProjectStatus = "canceled"

	ProjectTaskStatusTodo       ProjectTaskStatus = "todo"
	ProjectTaskStatusInProgress ProjectTaskStatus = "in_progress"
	ProjectTaskStatusBlocked    ProjectTaskStatus = "blocked"
	ProjectTaskStatusDone       ProjectTaskStatus = "done"

	TimeEntryStatusDraft    TimeEntryStatus = "draft"
	TimeEntryStatusApproved TimeEntryStatus = "approved"
	TimeEntryStatusBilled   TimeEntryStatus = "billed"

	ExpenseStatusDraft    ExpenseStatus = "draft"
	ExpenseStatusApproved ExpenseStatus = "approved"
	ExpenseStatusBilled   ExpenseStatus = "billed"

	OpsInvoiceStatusDraft     OpsInvoiceStatus = "draft"
	OpsInvoiceStatusIssued    OpsInvoiceStatus = "issued"
	OpsInvoiceStatusPartially OpsInvoiceStatus = "partially_paid"
	OpsInvoiceStatusPaid      OpsInvoiceStatus = "paid"
	OpsInvoiceStatusVoid      OpsInvoiceStatus = "void"

	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusPosted     PaymentStatus = "posted"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusReconciled PaymentStatus = "reconciled"
)

type OwnershipFields struct {
	OwnerID   uuid.UUID  `json:"owner_id"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

type ServiceContract struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	AccountID uuid.UUID  `json:"account_id"`
	ContactID *uuid.UUID `json:"contact_id,omitempty"`
	OwnershipFields
	Status         ServiceContractStatus `json:"status"`
	Name           string                `json:"name"`
	ContractNumber string                `json:"contract_number"`
	BillingCycle   string                `json:"billing_cycle"`
	RateCents      int64                 `json:"rate_cents"`
	Currency       string                `json:"currency"`
	StartDate      time.Time             `json:"start_date"`
	EndDate        *time.Time            `json:"end_date,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

type Asset struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	AccountID uuid.UUID  `json:"account_id"`
	ContactID *uuid.UUID `json:"contact_id,omitempty"`
	OwnershipFields
	Status       AssetStatus `json:"status"`
	Name         string      `json:"name"`
	AssetType    string      `json:"asset_type"`
	SerialNumber string      `json:"serial_number,omitempty"`
	Description  string      `json:"description,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type Project struct {
	ID         uuid.UUID  `json:"id"`
	OrgID      uuid.UUID  `json:"org_id"`
	AccountID  uuid.UUID  `json:"account_id"`
	ContactID  *uuid.UUID `json:"contact_id,omitempty"`
	ContractID *uuid.UUID `json:"contract_id,omitempty"`
	OwnershipFields
	Status      ProjectStatus `json:"status"`
	Name        string        `json:"name"`
	Code        string        `json:"code"`
	BudgetCents int64         `json:"budget_cents"`
	Currency    string        `json:"currency"`
	StartDate   *time.Time    `json:"start_date,omitempty"`
	DueDate     *time.Time    `json:"due_date,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type ProjectTask struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	AccountID uuid.UUID  `json:"account_id"`
	ContactID *uuid.UUID `json:"contact_id,omitempty"`
	ProjectID uuid.UUID  `json:"project_id"`
	OwnershipFields
	Status      ProjectTaskStatus `json:"status"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	DueDate     *time.Time        `json:"due_date,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type TimeEntry struct {
	ID            uuid.UUID  `json:"id"`
	OrgID         uuid.UUID  `json:"org_id"`
	AccountID     uuid.UUID  `json:"account_id"`
	ContactID     *uuid.UUID `json:"contact_id,omitempty"`
	ProjectID     *uuid.UUID `json:"project_id,omitempty"`
	ProjectTaskID *uuid.UUID `json:"project_task_id,omitempty"`
	ContractID    *uuid.UUID `json:"contract_id,omitempty"`
	OwnershipFields
	Status          TimeEntryStatus `json:"status"`
	EntryDate       time.Time       `json:"entry_date"`
	Minutes         int             `json:"minutes"`
	RateCents       int64           `json:"rate_cents"`
	Currency        string          `json:"currency"`
	Description     string          `json:"description,omitempty"`
	BilledInvoiceID *uuid.UUID      `json:"billed_invoice_id,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type Expense struct {
	ID         uuid.UUID  `json:"id"`
	OrgID      uuid.UUID  `json:"org_id"`
	AccountID  uuid.UUID  `json:"account_id"`
	ContactID  *uuid.UUID `json:"contact_id,omitempty"`
	ProjectID  *uuid.UUID `json:"project_id,omitempty"`
	ContractID *uuid.UUID `json:"contract_id,omitempty"`
	OwnershipFields
	Status          ExpenseStatus `json:"status"`
	ExpenseDate     time.Time     `json:"expense_date"`
	Category        string        `json:"category"`
	AmountCents     int64         `json:"amount_cents"`
	Currency        string        `json:"currency"`
	Vendor          string        `json:"vendor,omitempty"`
	Notes           string        `json:"notes,omitempty"`
	BilledInvoiceID *uuid.UUID    `json:"billed_invoice_id,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type OpsInvoice struct {
	ID         uuid.UUID  `json:"id"`
	OrgID      uuid.UUID  `json:"org_id"`
	AccountID  uuid.UUID  `json:"account_id"`
	ContactID  *uuid.UUID `json:"contact_id,omitempty"`
	ContractID *uuid.UUID `json:"contract_id,omitempty"`
	OwnershipFields
	Status           OpsInvoiceStatus `json:"status"`
	InvoiceNumber    string           `json:"invoice_number"`
	IssueDate        time.Time        `json:"issue_date"`
	DueDate          *time.Time       `json:"due_date,omitempty"`
	Currency         string           `json:"currency"`
	SubtotalCents    int64            `json:"subtotal_cents"`
	TaxCents         int64            `json:"tax_cents"`
	TotalCents       int64            `json:"total_cents"`
	PaidCents        int64            `json:"paid_cents"`
	OutstandingCents int64            `json:"outstanding_cents"`
	BillingContext   string           `json:"billing_context"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type InvoiceLineItem struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	AccountID uuid.UUID  `json:"account_id"`
	ContactID *uuid.UUID `json:"contact_id,omitempty"`
	InvoiceID uuid.UUID  `json:"invoice_id"`
	OwnershipFields
	SourceType     string     `json:"source_type"`
	SourceID       *uuid.UUID `json:"source_id,omitempty"`
	Description    string     `json:"description"`
	Quantity       float64    `json:"quantity"`
	UnitPriceCents int64      `json:"unit_price_cents"`
	AmountCents    int64      `json:"amount_cents"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Payment struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	AccountID uuid.UUID  `json:"account_id"`
	ContactID *uuid.UUID `json:"contact_id,omitempty"`
	InvoiceID uuid.UUID  `json:"invoice_id"`
	OwnershipFields
	Status      PaymentStatus `json:"status"`
	PaymentDate time.Time     `json:"payment_date"`
	Method      string        `json:"method"`
	Reference   string        `json:"reference,omitempty"`
	AmountCents int64         `json:"amount_cents"`
	Currency    string        `json:"currency"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type InvoiceAssembly struct {
	Invoice   *OpsInvoice        `json:"invoice"`
	LineItems []*InvoiceLineItem `json:"line_items"`
}

func (a *InvoiceAssembly) Validate() error {
	if a == nil || a.Invoice == nil {
		return fmt.Errorf("%w: invoice is required", ErrValidation)
	}
	if a.Invoice.AccountID == uuid.Nil {
		return fmt.Errorf("%w: account_id is required", ErrValidation)
	}
	return nil
}
