package handler_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

type fakeImportLeadRepo struct {
	created bool
}

func (f *fakeImportLeadRepo) Create(context.Context, *domain.Lead) (*domain.Lead, error) {
	f.created = true
	return &domain.Lead{ID: uuid.New()}, nil
}

func (f *fakeImportLeadRepo) GetByID(context.Context, uuid.UUID) (*domain.Lead, error) {
	return nil, domain.ErrNotFound
}

func (f *fakeImportLeadRepo) Update(context.Context, uuid.UUID, domain.LeadPatch) (*domain.Lead, error) {
	return nil, domain.ErrNotFound
}

func (f *fakeImportLeadRepo) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (f *fakeImportLeadRepo) List(context.Context, domain.LeadFilter) ([]*domain.Lead, int, error) {
	return nil, 0, nil
}

func (f *fakeImportLeadRepo) ListSources(context.Context) ([]string, error) {
	return nil, nil
}

func newImportRequest(t *testing.T, path, csvBody string, module domain.ACLModule, deniedField string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "import.csv")
	require.NoError(t, err)
	_, err = file.Write([]byte(csvBody))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = withClaims(req, &auth.Claims{UserID: uuid.New(), Role: string(domain.UserRoleAgent)})
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: uuid.New(),
		FieldWrite: map[domain.ACLModule]map[string]bool{
			module: {deniedField: false},
		},
	}))
	return req
}

func TestImportContactsRejectsDeniedField(t *testing.T) {
	contacts := new(mocks.MockContactRepository)
	contacts.On("GetByEmail", mock.Anything, "alice@example.com").Return(nil, domain.ErrNotFound)
	h := handler.NewImportHandler(contacts, nil, nil)

	req := newImportRequest(t, "/api/v1/import/contacts", "first_name,last_name,email\nAlice,Smith,alice@example.com\n", domain.ACLModuleContacts, "email")
	w := httptest.NewRecorder()

	h.ImportContacts(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), `field \"email\" is not writable`)
	contacts.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	contacts.AssertExpectations(t)
}

func TestImportContactsPlatformAdminBypassesDeniedField(t *testing.T) {
	contacts := new(mocks.MockContactRepository)
	contacts.On("GetByEmail", mock.Anything, "alice@example.com").Return(nil, domain.ErrNotFound)
	contacts.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Contact) bool {
		return c.FirstName == "Alice" && c.LastName == "Smith" && c.Email != nil && *c.Email == "alice@example.com"
	})).Return(&domain.Contact{ID: uuid.New()}, nil)
	h := handler.NewImportHandler(contacts, nil, nil)

	req := newImportRequest(t, "/api/v1/import/contacts", "first_name,last_name,email\nAlice,Smith,alice@example.com\n", domain.ACLModuleContacts, "email")
	req = withClaims(req, &auth.Claims{UserID: uuid.New(), Role: string(domain.UserRoleAdmin)})
	w := httptest.NewRecorder()

	h.ImportContacts(w, req)

	assert.Equal(t, http.StatusOK, w.Code, strings.TrimSpace(w.Body.String()))
	contacts.AssertExpectations(t)
}

func TestImportContactsExistingRowsDoNotValidateLookupOnlyEmail(t *testing.T) {
	contactID := uuid.New()
	contacts := new(mocks.MockContactRepository)
	contacts.On("GetByEmail", mock.Anything, "alice@example.com").
		Return(&domain.Contact{ID: contactID, Email: strPtrForTest("alice@example.com")}, nil)
	contacts.On("Update", mock.Anything, contactID, mock.MatchedBy(func(p domain.ContactPatch) bool {
		return p.FirstName != nil && *p.FirstName == "Alicia" &&
			p.LastName != nil && *p.LastName == "Smith" &&
			p.Phone != nil && *p.Phone == "555-0100"
	})).Return(&domain.Contact{ID: contactID}, nil)
	h := handler.NewImportHandler(contacts, nil, nil)

	req := newImportRequest(t, "/api/v1/import/contacts", "first_name,last_name,email,phone\nAlicia,Smith,alice@example.com,555-0100\n", domain.ACLModuleContacts, "email")
	w := httptest.NewRecorder()

	h.ImportContacts(w, req)

	assert.Equal(t, http.StatusOK, w.Code, strings.TrimSpace(w.Body.String()))
	contacts.AssertExpectations(t)
	contacts.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestImportAccountsRejectsDeniedFieldBeforeUpsert(t *testing.T) {
	accounts := new(mocks.MockAccountRepository)
	h := handler.NewImportHandler(nil, accounts, nil)

	req := newImportRequest(t, "/api/v1/import/accounts", "name,industry\nAcme,SaaS\n", domain.ACLModuleAccounts, "industry")
	w := httptest.NewRecorder()

	h.ImportAccounts(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), `field \"industry\" is not writable`)
	accounts.AssertNotCalled(t, "GetByName", mock.Anything, mock.Anything)
	accounts.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestImportLeadsRejectsDeniedField(t *testing.T) {
	leads := &fakeImportLeadRepo{}
	h := handler.NewImportHandler(nil, nil, leads)

	req := newImportRequest(t, "/api/v1/import/leads", "first_name,last_name,company\nAlice,Smith,Acme\n", domain.ACLModuleLeads, "company")
	w := httptest.NewRecorder()

	h.ImportLeads(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), `field \"company\" is not writable`)
	assert.False(t, leads.created)
}

func TestImportContactsAllowsEmptyDeniedColumn(t *testing.T) {
	contacts := new(mocks.MockContactRepository)
	contacts.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Contact) bool {
		return c.FirstName == "Alice" && c.LastName == "Smith" && c.Email == nil
	})).Return(&domain.Contact{ID: uuid.New()}, nil)
	h := handler.NewImportHandler(contacts, nil, nil)

	req := newImportRequest(t, "/api/v1/import/contacts", "first_name,last_name,email\nAlice,Smith,\n", domain.ACLModuleContacts, "email")
	w := httptest.NewRecorder()

	h.ImportContacts(w, req)

	assert.Equal(t, http.StatusOK, w.Code, strings.TrimSpace(w.Body.String()))
	contacts.AssertExpectations(t)
}

func strPtrForTest(s string) *string {
	return &s
}
