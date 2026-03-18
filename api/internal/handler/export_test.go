package handler_test

import (
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func newExportHandler() (*handler.ExportHandler, *mocks.MockContactRepository, *mocks.MockAccountRepository, *mocks.MockDealRepository, *mocks.MockReportsRepository) {
	contacts := &mocks.MockContactRepository{}
	accounts := &mocks.MockAccountRepository{}
	deals := &mocks.MockDealRepository{}
	reports := &mocks.MockReportsRepository{}
	h := handler.NewExportHandler(contacts, accounts, deals, reports)
	return h, contacts, accounts, deals, reports
}

func TestExportHandler_Contacts(t *testing.T) {
	ownerID := uuid.New()
	contact := testutil.SeedContact(ownerID)

	h, contactRepo, accountRepo, dealRepo, reportsRepo := newExportHandler()
	_ = accountRepo
	_ = dealRepo
	_ = reportsRepo

	// First page returns one contact; second page (pagination sentinel) returns empty.
	contactRepo.On("List", mock.Anything, mock.MatchedBy(func(f domain.ContactFilter) bool { return f.Page == 1 })).
		Return([]*domain.Contact{contact}, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/export/contacts", nil)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/csv")
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "contacts.csv")

	rows := parseCSV(t, rec.Body.String())
	require.Len(t, rows, 2) // header + 1 data row
	assert.Equal(t, "id", rows[0][0])
	assert.Equal(t, contact.ID.String(), rows[1][0])
	assert.Equal(t, "Ada", rows[1][1])
	assert.Equal(t, "Lovelace", rows[1][2])

	contactRepo.AssertExpectations(t)
}

func TestExportHandler_Accounts(t *testing.T) {
	ownerID := uuid.New()
	account := testutil.SeedAccount(ownerID)

	h, contactRepo, accountRepo, dealRepo, reportsRepo := newExportHandler()
	_ = contactRepo
	_ = dealRepo
	_ = reportsRepo

	accountRepo.On("List", mock.Anything, mock.MatchedBy(func(f domain.AccountFilter) bool { return f.Page == 1 })).
		Return([]*domain.Account{account}, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/export/accounts", nil)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "accounts.csv")

	rows := parseCSV(t, rec.Body.String())
	require.Len(t, rows, 2)
	assert.Equal(t, "id", rows[0][0])
	assert.Equal(t, account.ID.String(), rows[1][0])
	assert.Equal(t, "Acme Corp", rows[1][1])

	accountRepo.AssertExpectations(t)
}

func TestExportHandler_Deals(t *testing.T) {
	ownerID := uuid.New()
	pipelineID := uuid.New()
	deal := testutil.SeedDeal(ownerID, pipelineID)

	h, contactRepo, accountRepo, dealRepo, reportsRepo := newExportHandler()
	_ = contactRepo
	_ = accountRepo
	_ = reportsRepo

	dealRepo.On("List", mock.Anything, mock.MatchedBy(func(f domain.DealFilter) bool { return f.Page == 1 })).
		Return([]*domain.Deal{deal}, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/export/deals", nil)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "deals.csv")

	rows := parseCSV(t, rec.Body.String())
	require.Len(t, rows, 2)
	assert.Equal(t, "id", rows[0][0])
	assert.Equal(t, deal.ID.String(), rows[1][0])
	assert.Equal(t, "New Enterprise Deal", rows[1][1])
	assert.Equal(t, "500000", rows[1][3]) // value_cents

	dealRepo.AssertExpectations(t)
}

func TestExportHandler_Reports(t *testing.T) {
	h, contactRepo, accountRepo, dealRepo, reportsRepo := newExportHandler()
	_ = contactRepo
	_ = accountRepo
	_ = dealRepo

	reportsRepo.On("DealsByStage", mock.Anything).
		Return([]domain.DealStageMetric{
			{Stage: domain.DealStageQualified, Count: 3, TotalValueCents: 150000},
		}, nil)
	reportsRepo.On("ContactsMonthly", mock.Anything).
		Return([]domain.ContactMonthlyMetric{
			{Month: "2026-03", Count: 10},
		}, nil)
	reportsRepo.On("ActivitiesByType", mock.Anything).
		Return([]domain.ActivityTypeMetric{
			{Type: domain.ActivityTypeCall, Count: 5},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/export/reports", nil)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "reports.csv")

	body := rec.Body.String()
	assert.Contains(t, body, "deals_by_stage")
	assert.Contains(t, body, "contacts_monthly")
	assert.Contains(t, body, "activities_by_type")

	reportsRepo.AssertExpectations(t)
}

func TestExportHandler_Contacts_FilterByStage(t *testing.T) {
	ownerID := uuid.New()
	contact := testutil.SeedContact(ownerID)

	h, contactRepo, _, _, _ := newExportHandler()

	stage := domain.ContactStageProspect
	contactRepo.On("List", mock.Anything, mock.MatchedBy(func(f domain.ContactFilter) bool {
		return f.Page == 1 && f.Stage != nil && *f.Stage == stage
	})).Return([]*domain.Contact{contact}, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/export/contacts?stage=prospect", nil)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	contactRepo.AssertExpectations(t)
}

func TestExportHandler_Contacts_RepoError(t *testing.T) {
	h, contactRepo, _, _, _ := newExportHandler()

	contactRepo.On("List", mock.Anything, mock.Anything).
		Return(nil, 0, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/export/contacts", nil)
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	// On repo error we still return 200 with just the header row (error breaks the loop).
	assert.Equal(t, http.StatusOK, rec.Code)
	rows := parseCSV(t, rec.Body.String())
	assert.Len(t, rows, 1) // header only

	contactRepo.AssertExpectations(t)
}

// parseCSV parses a CSV string into rows, ignoring blank lines.
func parseCSV(t *testing.T, body string) [][]string {
	t.Helper()
	r := csv.NewReader(strings.NewReader(body))
	rows, err := r.ReadAll()
	require.NoError(t, err)
	return rows
}
