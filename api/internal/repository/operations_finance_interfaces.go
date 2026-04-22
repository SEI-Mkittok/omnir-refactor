package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// OperationsFinanceRepository manages service delivery + invoice lifecycle entities.
type OperationsFinanceRepository interface {
	CreateServiceContract(ctx context.Context, v *domain.ServiceContract) (*domain.ServiceContract, error)
	CreateAsset(ctx context.Context, v *domain.Asset) (*domain.Asset, error)
	CreateProject(ctx context.Context, v *domain.Project) (*domain.Project, error)
	CreateProjectTask(ctx context.Context, v *domain.ProjectTask) (*domain.ProjectTask, error)
	CreateTimeEntry(ctx context.Context, v *domain.TimeEntry) (*domain.TimeEntry, error)
	CreateExpense(ctx context.Context, v *domain.Expense) (*domain.Expense, error)
	CreateInvoice(ctx context.Context, v *domain.OpsInvoice) (*domain.OpsInvoice, error)
	CreateInvoiceLineItem(ctx context.Context, v *domain.InvoiceLineItem) (*domain.InvoiceLineItem, error)
	CreatePayment(ctx context.Context, v *domain.Payment) (*domain.Payment, error)

	ListServiceContracts(ctx context.Context, orgID uuid.UUID) ([]*domain.ServiceContract, error)
	ListAssets(ctx context.Context, orgID uuid.UUID) ([]*domain.Asset, error)
	ListProjects(ctx context.Context, orgID uuid.UUID) ([]*domain.Project, error)
	ListProjectTasks(ctx context.Context, orgID uuid.UUID, projectID uuid.UUID) ([]*domain.ProjectTask, error)
	ListTimeEntries(ctx context.Context, orgID uuid.UUID, accountID uuid.UUID) ([]*domain.TimeEntry, error)
	ListExpenses(ctx context.Context, orgID uuid.UUID, accountID uuid.UUID) ([]*domain.Expense, error)
	GetInvoiceAssembly(ctx context.Context, orgID uuid.UUID, invoiceID uuid.UUID) (*domain.InvoiceAssembly, error)

	BuildInvoiceAssembly(ctx context.Context, orgID uuid.UUID, accountID uuid.UUID, contactID *uuid.UUID, contractID *uuid.UUID, ownerID uuid.UUID, createdBy uuid.UUID) (*domain.InvoiceAssembly, error)
}
