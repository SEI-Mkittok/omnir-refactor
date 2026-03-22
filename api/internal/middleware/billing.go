package middleware

import (
	"context"
	"net/http"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

type billingPlanKey struct{}

// planOrder maps billing plans to a comparable integer tier.
var planOrder = map[domain.BillingPlan]int{
	domain.BillingPlanFree:       0,
	domain.BillingPlanPro:        1,
	domain.BillingPlanEnterprise: 2,
}

// RequirePlan returns a chi middleware that enforces a minimum billing plan.
// Requests from orgs below the required plan receive 402 Payment Required.
// The plan is cached in the request context to avoid N+1 DB lookups when
// multiple plan-gated routes share a request lifecycle.
func RequirePlan(billingRepo repository.BillingRepository, required domain.BillingPlan) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			plan, err := resolveOrgPlan(ctx, billingRepo)
			if err != nil {
				http.Error(w, `{"error":"billing unavailable"}`, http.StatusInternalServerError)
				return
			}

			if planOrder[plan] < planOrder[required] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusPaymentRequired)
				_, _ = w.Write([]byte(`{"error":"feature requires plan: ` + string(required) + `"}`))
				return
			}

			// Cache the resolved plan in context for downstream use.
			if _, cached := ctx.Value(billingPlanKey{}).(domain.BillingPlan); !cached {
				r = r.WithContext(context.WithValue(ctx, billingPlanKey{}, plan))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// resolveOrgPlan fetches the org's billing plan, using the context cache when available.
func resolveOrgPlan(ctx context.Context, billingRepo repository.BillingRepository) (domain.BillingPlan, error) {
	if cached, ok := ctx.Value(billingPlanKey{}).(domain.BillingPlan); ok {
		return cached, nil
	}

	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return domain.BillingPlanFree, nil
	}

	record, err := billingRepo.GetOrCreatePlan(ctx, orgID)
	if err != nil {
		return "", err
	}
	return record.Plan, nil
}
