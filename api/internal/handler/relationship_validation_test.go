package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

type quoteRepoStub struct{ createCalled bool }

func (s *quoteRepoStub) Create(_ context.Context, _ *domain.Quote) (*domain.Quote, error) {
	s.createCalled = true
	return &domain.Quote{}, nil
}
func (s *quoteRepoStub) GetByID(context.Context, uuid.UUID) (*domain.Quote, error) {
	return &domain.Quote{}, nil
}
func (s *quoteRepoStub) Update(context.Context, uuid.UUID, domain.QuotePatch) (*domain.Quote, error) {
	return &domain.Quote{}, nil
}
func (s *quoteRepoStub) Delete(context.Context, uuid.UUID) error { return nil }
func (s *quoteRepoStub) List(context.Context, domain.QuoteFilter) ([]*domain.Quote, int, error) {
	return nil, 0, nil
}
func (s *quoteRepoStub) ReplaceLineItems(context.Context, uuid.UUID, []domain.QuoteLineItemInput) ([]domain.QuoteLineItem, error) {
	return nil, nil
}
func (s *quoteRepoStub) MarkSent(context.Context, uuid.UUID) (*domain.Quote, error) {
	return &domain.Quote{}, nil
}
func (s *quoteRepoStub) MarkApproved(context.Context, uuid.UUID) (*domain.Quote, error) {
	return &domain.Quote{}, nil
}
func (s *quoteRepoStub) MarkRejected(context.Context, uuid.UUID) (*domain.Quote, error) {
	return &domain.Quote{}, nil
}

type emailRepoStub struct{ createCalled bool }

func (s *emailRepoStub) Create(_ context.Context, _ *domain.ContactEmail) (*domain.ContactEmail, error) {
	s.createCalled = true
	return &domain.ContactEmail{}, nil
}
func (s *emailRepoStub) List(context.Context, domain.EmailFilter) ([]*domain.ContactEmail, int, error) {
	return nil, 0, nil
}

func TestQuoteHandler_Create_RejectsUnrelatedContactDealPair(t *testing.T) {
	quoteRepo := &quoteRepoStub{}
	dealRepo := new(mocks.MockDealRepository)
	contactRepo := new(mocks.MockContactRepository)
	h := handler.NewQuoteHandler(quoteRepo).WithRelations(contactRepo, dealRepo)

	dealID := uuid.New()
	contactID := uuid.New()
	accountID := uuid.New()
	orgID := uuid.New()
	dealRepo.On("GetByID", mock.Anything, dealID).Return(&domain.Deal{ID: dealID, AccountID: &accountID}, nil)
	contactRepo.On("IsRelatedToAccount", mock.Anything, contactID, accountID).Return(false, nil)

	reqBody := map[string]any{"title": "Q-1", "deal_id": dealID.String(), "contact_id": contactID.String()}
	b, err := json.Marshal(reqBody)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/quotes", bytes.NewReader(b))
	req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
	w := httptest.NewRecorder()

	h.Create(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.False(t, quoteRepo.createCalled)
}

func TestEmailHandler_Send_RejectsUnrelatedContactDealPair(t *testing.T) {
	emailRepo := &emailRepoStub{}
	activityRepo := new(mocks.MockActivityRepository)
	dealRepo := new(mocks.MockDealRepository)
	contactRepo := new(mocks.MockContactRepository)
	h := handler.NewEmailHandler(emailRepo, activityRepo, contactRepo, dealRepo, email.NewMailer(nil, ""), "from@omnir.test")

	dealID := uuid.New()
	contactID := uuid.New()
	accountID := uuid.New()
	dealRepo.On("GetByID", mock.Anything, dealID).Return(&domain.Deal{ID: dealID, AccountID: &accountID}, nil)
	contactRepo.On("IsRelatedToAccount", mock.Anything, contactID, accountID).Return(false, nil)

	reqBody := map[string]any{
		"to":         "client@example.com",
		"subject":    "Hello",
		"body":       "Body",
		"contact_id": contactID.String(),
		"deal_id":    dealID.String(),
	}
	b, err := json.Marshal(reqBody)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/emails", bytes.NewReader(b))
	w := httptest.NewRecorder()

	h.Send(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.False(t, emailRepo.createCalled)
}
