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
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

type quoteCFRepoStub struct {
	created *domain.Quote
	updated domain.QuotePatch
}

func (s *quoteCFRepoStub) Create(_ context.Context, q *domain.Quote) (*domain.Quote, error) {
	s.created = q
	out := *q
	out.ID = uuid.New()
	return &out, nil
}

func (s *quoteCFRepoStub) GetByID(context.Context, uuid.UUID) (*domain.Quote, error) {
	return &domain.Quote{}, nil
}

func (s *quoteCFRepoStub) Update(_ context.Context, _ uuid.UUID, patch domain.QuotePatch) (*domain.Quote, error) {
	s.updated = patch
	return &domain.Quote{ID: uuid.New(), Title: "Updated", CustomFields: patch.CustomFields}, nil
}

func (s *quoteCFRepoStub) Delete(context.Context, uuid.UUID) error { return nil }

func (s *quoteCFRepoStub) List(context.Context, domain.QuoteFilter) ([]*domain.Quote, int, error) {
	return nil, 0, nil
}

func (s *quoteCFRepoStub) ReplaceLineItems(context.Context, uuid.UUID, []domain.QuoteLineItemInput) ([]domain.QuoteLineItem, error) {
	return nil, nil
}

func (s *quoteCFRepoStub) MarkSent(context.Context, uuid.UUID) (*domain.Quote, error) {
	return &domain.Quote{}, nil
}

func (s *quoteCFRepoStub) MarkApproved(context.Context, uuid.UUID) (*domain.Quote, error) {
	return &domain.Quote{}, nil
}

func (s *quoteCFRepoStub) MarkRejected(context.Context, uuid.UUID) (*domain.Quote, error) {
	return &domain.Quote{}, nil
}

func TestQuoteHandlerCreateValidatesAndExpandsCustomFields(t *testing.T) {
	orgID := uuid.New()
	repo := &quoteCFRepoStub{}
	cfRepo := new(mocks.MockCustomFieldDefinitionRepository)
	cfRepo.On("List", mock.Anything, mock.MatchedBy(func(filter domain.CustomFieldDefinitionFilter) bool {
		return filter.EntityType != nil && *filter.EntityType == domain.CustomFieldEntityQuote
	})).Return([]*domain.CustomFieldDefinition{
		{Name: "contract_type", Label: "Contract Type", FieldType: domain.CustomFieldTypeText, EntityType: domain.CustomFieldEntityQuote},
		{Name: "approval_code", Label: "Approval Code", FieldType: domain.CustomFieldTypeText, EntityType: domain.CustomFieldEntityQuote},
	}, nil).Twice()

	h := handler.NewQuoteHandler(repo).WithCustomFields(cfRepo)
	body := []byte(`{"title":"Enterprise quote","custom_fields":{"contract_type":"MSA"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/quotes", bytes.NewReader(body))
	req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.NotNil(t, repo.created)
	assert.JSONEq(t, `{"contract_type":"MSA"}`, string(repo.created.CustomFields))

	var response map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	customFields := response["custom_fields"].(map[string]any)
	assert.Equal(t, "MSA", customFields["contract_type"])
	assert.Nil(t, customFields["approval_code"])
	cfRepo.AssertExpectations(t)
}

func TestQuoteHandlerUpdateRejectsUnknownCustomField(t *testing.T) {
	quoteID := uuid.New()
	repo := &quoteCFRepoStub{}
	cfRepo := new(mocks.MockCustomFieldDefinitionRepository)
	cfRepo.On("List", mock.Anything, mock.Anything).Return([]*domain.CustomFieldDefinition{
		{Name: "contract_type", Label: "Contract Type", FieldType: domain.CustomFieldTypeText, EntityType: domain.CustomFieldEntityQuote},
	}, nil).Once()

	h := handler.NewQuoteHandler(repo).WithCustomFields(cfRepo)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/quotes/"+quoteID.String(), bytes.NewReader([]byte(`{"custom_fields":{"unknown":"x"}}`)))
	req = withURLParam(req, "id", quoteID.String())
	req = req.WithContext(domain.WithOrgID(req.Context(), uuid.New()))
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Empty(t, repo.updated.CustomFields)
	cfRepo.AssertExpectations(t)
}
