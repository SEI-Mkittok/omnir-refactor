package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v76"
	portalsession "github.com/stripe/stripe-go/v76/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	stripecustomer "github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// BillingHandler serves Stripe billing endpoints.
type BillingHandler struct {
	billing   repository.BillingRepository
	stripeCfg config.StripeConfig
	appURL    string
}

// NewBillingHandler constructs a BillingHandler and sets the Stripe API key.
func NewBillingHandler(billing repository.BillingRepository, stripeCfg config.StripeConfig, appURL string) *BillingHandler {
	if stripeCfg.SecretKey != "" {
		stripe.Key = stripeCfg.SecretKey
	}
	return &BillingHandler{billing: billing, stripeCfg: stripeCfg, appURL: appURL}
}

// Router returns the authenticated billing sub-router (mounted under /api/v1/billing).
func (h *BillingHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/plans", h.GetPlans)
	r.Get("/usage", h.GetUsage)
	r.Get("/subscription", h.GetSubscription)
	r.Get("/invoices", h.ListInvoices)
	r.Post("/checkout", h.CreateCheckout)
	r.Post("/subscription/cancel", h.CancelSubscription)
	r.Post("/portal", h.CreatePortal)
	return r
}

// WebhookRouter returns the unauthenticated webhook router (mounted under /api/billing/webhook).
func (h *BillingHandler) WebhookRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.HandleWebhook)
	return r
}

// GetPlans returns the public plan catalog used by the billing UI.
// GET /api/v1/billing/plans
func (h *BillingHandler) GetPlans(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, []map[string]any{
		{
			"tier":             "free",
			"name":             "Free",
			"priceMonthly":     0,
			"priceAnnual":      0,
			"seats":            2,
			"ticketsPerMonth":  100,
			"apiCallsPerMonth": nil,
			"features": map[string]bool{
				"slaRules": false, "reports": false, "apiAccess": false, "customFields": false, "whiteLabel": false,
			},
		},
		{
			"tier":             "pro",
			"name":             "Pro",
			"priceMonthly":     49,
			"priceAnnual":      44,
			"seats":            25,
			"ticketsPerMonth":  nil,
			"apiCallsPerMonth": 10000,
			"features": map[string]bool{
				"slaRules": true, "reports": true, "apiAccess": true, "customFields": false, "whiteLabel": false,
			},
		},
		{
			"tier":             "enterprise",
			"name":             "Enterprise",
			"priceMonthly":     nil,
			"priceAnnual":      nil,
			"seats":            nil,
			"ticketsPerMonth":  nil,
			"apiCallsPerMonth": nil,
			"features": map[string]bool{
				"slaRules": true, "reports": true, "apiAccess": true, "customFields": true, "whiteLabel": true,
			},
		},
	})
}

// GetUsage returns current usage stats for the authenticated org.
// GET /api/v1/billing/usage
func (h *BillingHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}

	plan, err := h.billing.GetOrCreatePlan(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch billing info")
		return
	}

	userCount, contactCount, err := h.billing.GetUsageStats(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch usage stats")
		return
	}

	resp := buildUsageResponse(plan, userCount, contactCount)
	writeJSON(w, http.StatusOK, resp)
}

// buildUsageResponse assembles the BillingUsageStats response from a plan and live counts.
func buildUsageResponse(plan *domain.OrgPlanRecord, userCount, contactCount int) domain.BillingUsageStats {
	var resp domain.BillingUsageStats
	resp.Plan = plan.Plan

	// Hardcoded limits by plan tier.
	switch plan.Plan {
	case domain.BillingPlanPro:
		limit10 := 10
		limit1000 := 1000
		resp.Usage.Users.Limit = &limit10
		resp.Usage.StorageMB.Limit = &limit1000
		// contacts unlimited (nil)
	case domain.BillingPlanEnterprise:
		// all unlimited
	default: // free
		limit5 := 5
		limit500 := 500
		limit100 := 100
		resp.Usage.Users.Limit = &limit5
		resp.Usage.Contacts.Limit = &limit500
		resp.Usage.StorageMB.Limit = &limit100
	}

	resp.Usage.Users.Used = userCount
	resp.Usage.Contacts.Used = contactCount
	resp.Usage.StorageMB.Used = 0 // storage tracking not yet implemented

	// Derive billing cycle dates.
	now := time.Now()
	if plan.CurrentPeriodEnd != nil {
		end := *plan.CurrentPeriodEnd
		// start = first day of the same month as end
		start := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)
		resp.BillingCycle.CurrentPeriodStart = start.Format("2006-01-02")
		resp.BillingCycle.CurrentPeriodEnd = end.Format("2006-01-02")
	} else {
		// Fall back to the current calendar month.
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)
		resp.BillingCycle.CurrentPeriodStart = start.Format("2006-01-02")
		resp.BillingCycle.CurrentPeriodEnd = end.Format("2006-01-02")
	}

	return resp
}

// GetSubscription returns the current plan for the authenticated org.
// GET /api/v1/billing/subscription
func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}

	plan, err := h.billing.GetOrCreatePlan(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch subscription")
		return
	}

	writeJSON(w, http.StatusOK, plan)
}

// ListInvoices returns invoice history for the authenticated org.
// GET /api/v1/billing/invoices
func (h *BillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}

	page, limit := 1, 25
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	offset := (page - 1) * limit

	invs, total, err := h.billing.ListInvoices(r.Context(), orgID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list invoices")
		return
	}

	writeJSON(w, http.StatusOK, paginated(invs, total, page, limit))
}

type checkoutRequest struct {
	Plan      string `json:"plan"`
	Tier      string `json:"tier"`
	ReturnURL string `json:"return_url"`
}

// CreateCheckout creates a Stripe Checkout session for plan upgrade.
// POST /api/v1/billing/checkout
func (h *BillingHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	if h.stripeCfg.SecretKey == "" {
		writeError(w, http.StatusServiceUnavailable, "billing not configured")
		return
	}

	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}

	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	planName := req.Plan
	if planName == "" {
		planName = req.Tier
	}
	priceID := h.priceIDForPlan(planName)
	if priceID == "" {
		writeError(w, http.StatusBadRequest, "invalid plan: must be 'pro' or 'enterprise'")
		return
	}

	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = h.appURL + "/settings/billing"
	}

	plan, err := h.billing.GetOrCreatePlan(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch billing info")
		return
	}

	customerID, err := h.ensureStripeCustomer(r.Context(), plan, orgID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create Stripe customer")
		return
	}

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(customerID),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(returnURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(returnURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(priceID), Quantity: stripe.Int64(1)},
		},
	}

	s, err := checkoutsession.New(params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create checkout session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": s.URL})
}

// CancelSubscription marks the org subscription as canceled locally.
// POST /api/v1/billing/subscription/cancel
func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}

	status := domain.SubscriptionStatusCanceled
	plan, err := h.billing.UpsertPlan(r.Context(), orgID, domain.OrgPlanPatch{Status: &status})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to cancel subscription")
		return
	}

	writeJSON(w, http.StatusOK, plan)
}

type portalRequest struct {
	ReturnURL string `json:"return_url"`
}

// CreatePortal creates a Stripe customer portal session.
// POST /api/v1/billing/portal
func (h *BillingHandler) CreatePortal(w http.ResponseWriter, r *http.Request) {
	if h.stripeCfg.SecretKey == "" {
		writeError(w, http.StatusServiceUnavailable, "billing not configured")
		return
	}

	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}

	plan, err := h.billing.GetOrCreatePlan(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch billing info")
		return
	}

	if plan.StripeCustomerID == nil {
		writeError(w, http.StatusBadRequest, "no Stripe customer found — upgrade first")
		return
	}

	var req portalRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = h.appURL + "/settings/billing"
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(*plan.StripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	s, err := portalsession.New(params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create portal session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": s.URL})
}

// HandleWebhook processes Stripe webhook events.
// POST /api/billing/webhook
func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if h.stripeCfg.WebhookSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "billing not configured")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), h.stripeCfg.WebhookSecret)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook signature")
		return
	}

	switch event.Type {
	case "customer.subscription.updated", "customer.subscription.deleted":
		h.handleSubscriptionEvent(w, r, event)
	case "invoice.paid", "invoice.payment_failed":
		h.handleInvoiceEvent(w, r, event)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (h *BillingHandler) priceIDForPlan(plan string) string {
	switch plan {
	case "pro":
		return h.stripeCfg.ProPriceID
	case "enterprise":
		return h.stripeCfg.EnterprisePriceID
	default:
		return ""
	}
}

// ensureStripeCustomer returns the existing Stripe customer ID or creates a new one,
// persisting it to the billing repo.
func (h *BillingHandler) ensureStripeCustomer(ctx context.Context, plan *domain.OrgPlanRecord, orgID string) (string, error) {
	if plan.StripeCustomerID != nil && *plan.StripeCustomerID != "" {
		return *plan.StripeCustomerID, nil
	}

	params := &stripe.CustomerParams{
		Metadata: map[string]string{
			"org_id": orgID,
		},
	}
	c, err := stripecustomer.New(params)
	if err != nil {
		return "", err
	}

	// Persist the customer ID.
	if _, err := h.billing.UpsertPlan(ctx, plan.OrgID, domain.OrgPlanPatch{
		StripeCustomerID: stripe.String(c.ID),
	}); err != nil {
		return "", err
	}

	return c.ID, nil
}

func (h *BillingHandler) handleSubscriptionEvent(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse subscription event")
		return
	}

	// Look up org by subscription ID first, fall back to customer ID.
	plan, err := h.billing.GetPlanByStripeSubscriptionID(r.Context(), sub.ID)
	if err != nil {
		plan, err = h.billing.GetPlanByStripeCustomerID(r.Context(), sub.Customer.ID)
		if err != nil {
			// Unknown org — acknowledge and skip.
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	billingPlan := domain.BillingPlanFree
	if len(sub.Items.Data) > 0 {
		priceID := sub.Items.Data[0].Price.ID
		switch priceID {
		case h.stripeCfg.ProPriceID:
			billingPlan = domain.BillingPlanPro
		case h.stripeCfg.EnterprisePriceID:
			billingPlan = domain.BillingPlanEnterprise
		}
	}

	status := domain.SubscriptionStatus(sub.Status)
	subID := sub.ID
	customerID := sub.Customer.ID
	t := time.Unix(sub.CurrentPeriodEnd, 0)

	patch := domain.OrgPlanPatch{
		Plan:                 &billingPlan,
		StripeSubscriptionID: &subID,
		StripeCustomerID:     &customerID,
		Status:               &status,
		CurrentPeriodEnd:     &t,
	}

	if _, err := h.billing.UpsertPlan(r.Context(), plan.OrgID, patch); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update plan")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) handleInvoiceEvent(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse invoice event")
		return
	}

	plan, err := h.billing.GetPlanByStripeCustomerID(r.Context(), inv.Customer.ID)
	if err != nil {
		// Unknown org — acknowledge and skip.
		w.WriteHeader(http.StatusOK)
		return
	}

	var pdfURL *string
	if inv.InvoicePDF != "" {
		u := inv.InvoicePDF
		pdfURL = &u
	}

	record := &domain.Invoice{
		OrgID:           plan.OrgID,
		StripeInvoiceID: inv.ID,
		AmountCents:     inv.AmountPaid,
		Currency:        string(inv.Currency),
		Status:          domain.InvoiceStatus(inv.Status),
		PDFURL:          pdfURL,
	}

	if _, err := h.billing.UpsertInvoice(r.Context(), record); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upsert invoice")
		return
	}

	w.WriteHeader(http.StatusOK)
}
