package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// BillingRepo is a PostgreSQL-backed BillingRepository.
type BillingRepo struct {
	db *pgxpool.Pool
}

// NewBillingRepo constructs a BillingRepo.
func NewBillingRepo(db *pgxpool.Pool) *BillingRepo {
	return &BillingRepo{db: db}
}

const orgPlanCols = `id, org_id, plan, stripe_customer_id, stripe_subscription_id, status, current_period_end, created_at, updated_at`

func scanOrgPlan(row pgx.Row) (*domain.OrgPlanRecord, error) {
	var p domain.OrgPlanRecord
	err := row.Scan(
		&p.ID, &p.OrgID, &p.Plan, &p.StripeCustomerID, &p.StripeSubscriptionID,
		&p.Status, &p.CurrentPeriodEnd, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// GetOrCreatePlan returns the billing plan for an org, creating a free-tier record if absent.
func (r *BillingRepo) GetOrCreatePlan(ctx context.Context, orgID uuid.UUID) (*domain.OrgPlanRecord, error) {
	const q = `
		INSERT INTO org_plans (org_id, plan, status)
		VALUES ($1, 'free', 'active')
		ON CONFLICT (org_id) DO UPDATE SET updated_at = NOW()
		RETURNING ` + orgPlanCols
	return scanOrgPlan(r.db.QueryRow(ctx, q, orgID))
}

// UpsertPlan applies a patch to the org's billing plan row.
func (r *BillingRepo) UpsertPlan(ctx context.Context, orgID uuid.UUID, patch domain.OrgPlanPatch) (*domain.OrgPlanRecord, error) {
	// Ensure row exists first.
	if _, err := r.GetOrCreatePlan(ctx, orgID); err != nil {
		return nil, err
	}

	const q = `
		UPDATE org_plans SET
			plan                  = COALESCE($2, plan),
			stripe_customer_id    = COALESCE($3, stripe_customer_id),
			stripe_subscription_id = COALESCE($4, stripe_subscription_id),
			status                = COALESCE($5, status),
			current_period_end    = COALESCE($6, current_period_end),
			updated_at            = NOW()
		WHERE org_id = $1
		RETURNING ` + orgPlanCols

	var planStr *string
	if patch.Plan != nil {
		s := string(*patch.Plan)
		planStr = &s
	}
	var statusStr *string
	if patch.Status != nil {
		s := string(*patch.Status)
		statusStr = &s
	}

	return scanOrgPlan(r.db.QueryRow(ctx, q,
		orgID,
		planStr,
		patch.StripeCustomerID,
		patch.StripeSubscriptionID,
		statusStr,
		patch.CurrentPeriodEnd,
	))
}

// GetPlanByStripeSubscriptionID looks up a plan by Stripe subscription ID.
func (r *BillingRepo) GetPlanByStripeSubscriptionID(ctx context.Context, subID string) (*domain.OrgPlanRecord, error) {
	const q = `SELECT ` + orgPlanCols + ` FROM org_plans WHERE stripe_subscription_id = $1`
	return scanOrgPlan(r.db.QueryRow(ctx, q, subID))
}

// GetPlanByStripeCustomerID looks up a plan by Stripe customer ID.
func (r *BillingRepo) GetPlanByStripeCustomerID(ctx context.Context, customerID string) (*domain.OrgPlanRecord, error) {
	const q = `SELECT ` + orgPlanCols + ` FROM org_plans WHERE stripe_customer_id = $1`
	return scanOrgPlan(r.db.QueryRow(ctx, q, customerID))
}

const invoiceCols = `id, org_id, stripe_invoice_id, amount_cents, currency, status, pdf_url, created_at`

func scanInvoice(row pgx.Row) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := row.Scan(
		&inv.ID, &inv.OrgID, &inv.StripeInvoiceID, &inv.AmountCents,
		&inv.Currency, &inv.Status, &inv.PDFURL, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

// UpsertInvoice inserts or updates an invoice by stripe_invoice_id.
func (r *BillingRepo) UpsertInvoice(ctx context.Context, inv *domain.Invoice) (*domain.Invoice, error) {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	if inv.CreatedAt.IsZero() {
		inv.CreatedAt = time.Now()
	}
	const q = `
		INSERT INTO invoices (id, org_id, stripe_invoice_id, amount_cents, currency, status, pdf_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (stripe_invoice_id) DO UPDATE SET
			amount_cents = EXCLUDED.amount_cents,
			currency     = EXCLUDED.currency,
			status       = EXCLUDED.status,
			pdf_url      = EXCLUDED.pdf_url
		RETURNING ` + invoiceCols
	return scanInvoice(r.db.QueryRow(ctx, q,
		inv.ID, inv.OrgID, inv.StripeInvoiceID, inv.AmountCents,
		inv.Currency, inv.Status, inv.PDFURL, inv.CreatedAt,
	))
}

// GetUsageStats returns current user and contact counts for the org.
func (r *BillingRepo) GetUsageStats(ctx context.Context, orgID uuid.UUID) (userCount, contactCount int, err error) {
	const q = `
		SELECT
			(SELECT COUNT(*) FROM users    WHERE org_id = $1 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM contacts WHERE org_id = $1)`
	err = r.db.QueryRow(ctx, q, orgID).Scan(&userCount, &contactCount)
	return
}

// ListInvoices returns paginated invoices for an org, most recent first.
func (r *BillingRepo) ListInvoices(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.Invoice, int, error) {
	const countQ = `SELECT COUNT(*) FROM invoices WHERE org_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQ, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	const q = `SELECT ` + invoiceCols + ` FROM invoices WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invs []*domain.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, 0, err
		}
		invs = append(invs, inv)
	}
	return invs, total, rows.Err()
}
