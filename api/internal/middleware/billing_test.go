package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func makeReqWithOrg(orgID uuid.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	return req.WithContext(domain.WithOrgID(req.Context(), orgID))
}

func planRecord(plan domain.BillingPlan) *domain.OrgPlanRecord {
	return &domain.OrgPlanRecord{
		ID:     uuid.New(),
		OrgID:  uuid.New(),
		Plan:   plan,
		Status: domain.SubscriptionStatusActive,
	}
}

func TestRequirePlan_AllowsExactPlan(t *testing.T) {
	orgID := uuid.New()
	repo := &mocks.MockBillingRepository{}
	repo.On("GetOrCreatePlan", mock.Anything, orgID).Return(planRecord(domain.BillingPlanPro), nil)

	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	h := middleware.RequirePlan(repo, domain.BillingPlanPro)(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, makeReqWithOrg(orgID))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, called)
}

func TestRequirePlan_AllowsHigherPlan(t *testing.T) {
	orgID := uuid.New()
	repo := &mocks.MockBillingRepository{}
	repo.On("GetOrCreatePlan", mock.Anything, orgID).Return(planRecord(domain.BillingPlanEnterprise), nil)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	h := middleware.RequirePlan(repo, domain.BillingPlanPro)(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, makeReqWithOrg(orgID))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequirePlan_BlocksLowerPlan(t *testing.T) {
	orgID := uuid.New()
	repo := &mocks.MockBillingRepository{}
	repo.On("GetOrCreatePlan", mock.Anything, orgID).Return(planRecord(domain.BillingPlanFree), nil)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	h := middleware.RequirePlan(repo, domain.BillingPlanPro)(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, makeReqWithOrg(orgID))

	assert.Equal(t, http.StatusPaymentRequired, rec.Code)
	assert.Contains(t, rec.Body.String(), "feature requires plan: pro")
}

func TestRequirePlan_EnterpriseBlocksPro(t *testing.T) {
	orgID := uuid.New()
	repo := &mocks.MockBillingRepository{}
	repo.On("GetOrCreatePlan", mock.Anything, orgID).Return(planRecord(domain.BillingPlanPro), nil)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	h := middleware.RequirePlan(repo, domain.BillingPlanEnterprise)(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, makeReqWithOrg(orgID))

	assert.Equal(t, http.StatusPaymentRequired, rec.Code)
	assert.Contains(t, rec.Body.String(), "feature requires plan: enterprise")
}

func TestRequirePlan_NoOrgIDAllowsFree(t *testing.T) {
	repo := &mocks.MockBillingRepository{}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	h := middleware.RequirePlan(repo, domain.BillingPlanFree)(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	repo.AssertNotCalled(t, "GetOrCreatePlan")
}

func TestRequirePlan_RepoErrorReturns500(t *testing.T) {
	orgID := uuid.New()
	repo := &mocks.MockBillingRepository{}
	repo.On("GetOrCreatePlan", mock.Anything, orgID).Return(nil, errors.New("db error"))

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	h := middleware.RequirePlan(repo, domain.BillingPlanPro)(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, makeReqWithOrg(orgID))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
