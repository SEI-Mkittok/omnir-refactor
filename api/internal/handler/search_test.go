package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func newSearchHandler() (*handler.SearchHandler, *mocks.MockSearchRepository) {
	repo := new(mocks.MockSearchRepository)
	return handler.NewSearchHandler(repo), repo
}

func emptySearchResult() *domain.SearchGroupedResult {
	return &domain.SearchGroupedResult{
		Contacts: []domain.SearchContact{},
		Accounts: []domain.SearchAccount{},
		Deals:    []domain.SearchDeal{},
		Tickets:  []domain.SearchTicket{},
	}
}

func TestSearchHandler_Search_Success(t *testing.T) {
	h, repo := newSearchHandler()

	result := &domain.SearchGroupedResult{
		Contacts: []domain.SearchContact{
			{ID: "1", FirstName: "Acme", LastName: "Contact", Email: "acme@example.com", Stage: "lead"},
		},
		Accounts: []domain.SearchAccount{
			{ID: "2", Name: "Acme Corp"},
		},
		Deals:   []domain.SearchDeal{},
		Tickets: []domain.SearchTicket{},
	}
	repo.On("Search", mock.Anything, "acme", 20).Return(result, nil)

	req := httptest.NewRequest(http.MethodGet, "/?q=acme", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Len(t, body["contacts"], 1)
	assert.Len(t, body["accounts"], 1)

	repo.AssertExpectations(t)
}

func TestSearchHandler_Search_EmptyQuery(t *testing.T) {
	h, _ := newSearchHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchHandler_Search_RequiresAuth(t *testing.T) {
	h, _ := newSearchHandler()

	jwtSvc := auth.NewJWTService("test-secret")
	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc, nil, nil))
	r.Mount("/", h.Router())

	// No Authorization header — expect 401.
	req := httptest.NewRequest(http.MethodGet, "/?q=test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSearchHandler_Search_OrgIDScoping(t *testing.T) {
	h, repo := newSearchHandler()

	orgID := domain.DefaultOrgID

	repo.On("Search", mock.MatchedBy(func(ctx context.Context) bool {
		id, ok := domain.OrgIDFromContext(ctx)
		return ok && id == orgID
	}), "test", 20).Return(emptySearchResult(), nil)

	ctx := domain.WithOrgID(context.Background(), orgID)
	ctx = middleware.WithClaims(ctx, &auth.Claims{OrgID: orgID, Role: "user"})
	req := httptest.NewRequest(http.MethodGet, "/?q=test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestSearchHandler_Search_EmptyResults(t *testing.T) {
	h, repo := newSearchHandler()

	repo.On("Search", mock.Anything, "nomatch", 20).Return(emptySearchResult(), nil)

	req := httptest.NewRequest(http.MethodGet, "/?q=nomatch", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, `"contacts":[]`)
	assert.Contains(t, body, `"accounts":[]`)
	assert.Contains(t, body, `"deals":[]`)
	assert.Contains(t, body, `"tickets":[]`)
}

func TestSearchHandler_Search_LimitParam(t *testing.T) {
	h, repo := newSearchHandler()

	repo.On("Search", mock.Anything, "test", 10).Return(emptySearchResult(), nil)

	req := httptest.NewRequest(http.MethodGet, "/?q=test&limit=10", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestSearchHandler_Search_RepoError(t *testing.T) {
	h, repo := newSearchHandler()

	repo.On("Search", mock.Anything, "test", 20).Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/?q=test", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
