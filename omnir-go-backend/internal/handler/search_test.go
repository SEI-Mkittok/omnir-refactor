package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func newSearchHandler() (*handler.SearchHandler, *mocks.MockContactRepository, *mocks.MockAccountRepository, *mocks.MockDealRepository) {
	contacts := new(mocks.MockContactRepository)
	accounts := new(mocks.MockAccountRepository)
	deals := new(mocks.MockDealRepository)
	return handler.NewSearchHandler(contacts, accounts, deals), contacts, accounts, deals
}

func TestSearchHandler_Search_Success(t *testing.T) {
	h, contacts, accounts, deals := newSearchHandler()

	ownerID := uuid.New()
	orgID := domain.DefaultOrgID
	now := time.Now()

	contacts.On("List", mock.Anything, domain.ContactFilter{Q: "acme", Page: 1, Limit: 5}).
		Return([]*domain.Contact{
			{ID: uuid.New(), OrgID: orgID, FirstName: "Acme", LastName: "Contact", OwnerID: ownerID, Stage: domain.ContactStageLead, Tags: []string{}, CreatedAt: now, UpdatedAt: now},
		}, 1, nil)
	accounts.On("List", mock.Anything, domain.AccountFilter{Q: "acme", Page: 1, Limit: 5}).
		Return([]*domain.Account{
			{ID: uuid.New(), OrgID: orgID, Name: "Acme Corp", OwnerID: ownerID, Tags: []string{}, CreatedAt: now, UpdatedAt: now},
		}, 1, nil)
	deals.On("List", mock.Anything, domain.DealFilter{Q: "acme", Page: 1, Limit: 5}).
		Return([]*domain.Deal{
			{ID: uuid.New(), OrgID: orgID, Title: "Acme Deal", Stage: domain.DealStageLead, OwnerID: ownerID, PipelineID: uuid.New(), CreatedAt: now, UpdatedAt: now},
		}, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/?q=acme", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result["contacts"], 1)
	assert.Len(t, result["accounts"], 1)
	assert.Len(t, result["deals"], 1)
	assert.Equal(t, float64(3), result["total"])

	contacts.AssertExpectations(t)
	accounts.AssertExpectations(t)
	deals.AssertExpectations(t)
}

func TestSearchHandler_Search_EmptyQuery(t *testing.T) {
	h, _, _, _ := newSearchHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchHandler_Search_RequiresAuth(t *testing.T) {
	h, _, _, _ := newSearchHandler()

	jwtSvc := auth.NewJWTService("test-secret")
	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc))
	r.Mount("/", h.Router())

	// No Authorization header — expect 401.
	req := httptest.NewRequest(http.MethodGet, "/?q=test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSearchHandler_Search_OrgIDScoping(t *testing.T) {
	h, contacts, accounts, deals := newSearchHandler()

	orgID := uuid.New()
	otherOrgID := uuid.New()

	// Mocks verify they are called with the correct org_id in context.
	contacts.On("List", mock.MatchedBy(func(ctx context.Context) bool {
		id, ok := domain.OrgIDFromContext(ctx)
		return ok && id == orgID
	}), mock.Anything).Return([]*domain.Contact{}, 0, nil)
	accounts.On("List", mock.MatchedBy(func(ctx context.Context) bool {
		id, ok := domain.OrgIDFromContext(ctx)
		return ok && id == orgID
	}), mock.Anything).Return([]*domain.Account{}, 0, nil)
	deals.On("List", mock.MatchedBy(func(ctx context.Context) bool {
		id, ok := domain.OrgIDFromContext(ctx)
		return ok && id == orgID
	}), mock.Anything).Return([]*domain.Deal{}, 0, nil)

	// Inject org_id and claims for orgID (not otherOrgID) into context.
	ctx := domain.WithOrgID(context.Background(), orgID)
	ctx = middleware.WithClaims(ctx, &auth.Claims{UserID: uuid.New(), OrgID: orgID, Role: "user"})
	req := httptest.NewRequest(http.MethodGet, "/?q=test", nil).WithContext(ctx)
	_ = otherOrgID // verified implicitly — mocks only match orgID
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	contacts.AssertExpectations(t)
	accounts.AssertExpectations(t)
	deals.AssertExpectations(t)
}

func TestSearchHandler_Search_PartialMatch(t *testing.T) {
	h, contacts, accounts, deals := newSearchHandler()

	ownerID := uuid.New()
	now := time.Now()

	contacts.On("List", mock.Anything, domain.ContactFilter{Q: "acm", Page: 1, Limit: 5}).
		Return([]*domain.Contact{
			{ID: uuid.New(), FirstName: "Acme", LastName: "Person", OwnerID: ownerID, Stage: domain.ContactStageLead, Tags: []string{}, CreatedAt: now, UpdatedAt: now},
		}, 1, nil)
	accounts.On("List", mock.Anything, domain.AccountFilter{Q: "acm", Page: 1, Limit: 5}).
		Return([]*domain.Account{}, 0, nil)
	deals.On("List", mock.Anything, domain.DealFilter{Q: "acm", Page: 1, Limit: 5}).
		Return([]*domain.Deal{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/?q=acm", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Len(t, result["contacts"], 1)
	assert.Equal(t, float64(1), result["total"])

	contacts.AssertExpectations(t)
}

func TestSearchHandler_Search_EmptyResults(t *testing.T) {
	h, contacts, accounts, deals := newSearchHandler()

	contacts.On("List", mock.Anything, domain.ContactFilter{Q: "nomatch", Page: 1, Limit: 5}).
		Return(nil, 0, nil)
	accounts.On("List", mock.Anything, domain.AccountFilter{Q: "nomatch", Page: 1, Limit: 5}).
		Return(nil, 0, nil)
	deals.On("List", mock.Anything, domain.DealFilter{Q: "nomatch", Page: 1, Limit: 5}).
		Return(nil, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/?q=nomatch", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify nil slices are serialised as empty arrays, not null.
	body := w.Body.String()
	assert.Contains(t, body, `"contacts":[]`)
	assert.Contains(t, body, `"accounts":[]`)
	assert.Contains(t, body, `"deals":[]`)
	assert.Contains(t, body, `"total":0`)
}

func TestSearchHandler_Search_LimitParam(t *testing.T) {
	h, contacts, accounts, deals := newSearchHandler()

	// With limit=10 the handler should pass Limit:10 to each repo.
	contacts.On("List", mock.Anything, domain.ContactFilter{Q: "test", Page: 1, Limit: 10}).
		Return([]*domain.Contact{}, 0, nil)
	accounts.On("List", mock.Anything, domain.AccountFilter{Q: "test", Page: 1, Limit: 10}).
		Return([]*domain.Account{}, 0, nil)
	deals.On("List", mock.Anything, domain.DealFilter{Q: "test", Page: 1, Limit: 10}).
		Return([]*domain.Deal{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/?q=test&limit=10", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	contacts.AssertExpectations(t)
	accounts.AssertExpectations(t)
	deals.AssertExpectations(t)
}

func TestSearchHandler_Search_RepoError(t *testing.T) {
	h, contacts, _, _ := newSearchHandler()

	contacts.On("List", mock.Anything, mock.Anything).
		Return(nil, 0, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/?q=test", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
