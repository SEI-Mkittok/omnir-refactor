package handler_test

import (
	"bytes"
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

func TestDealHandler_GetByID_IncludesContacts(t *testing.T) {
	dealID := uuid.New()
	contactID := uuid.New()
	ownerID := uuid.New()
	pipelineID := uuid.New()

	email := "alice@example.com"
	contacts := []domain.Contact{
		{
			ID:        contactID,
			FirstName: "Alice",
			LastName:  "Smith",
			Email:     &email,
			OwnerID:   ownerID,
		},
	}

	mockRepo := new(mocks.MockDealRepository)
	mockRepo.On("GetByID", mock.Anything, dealID).Return(&domain.Deal{
		ID:         dealID,
		Title:      "Big Deal",
		OwnerID:    ownerID,
		PipelineID: pipelineID,
		Contacts:   contacts,
	}, nil)

	h := handler.NewDealHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/deals/"+dealID.String(), nil)
	req = withURLParam(req, "id", dealID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	contactsArr, ok := resp["contacts"].([]any)
	require.True(t, ok, "expected contacts array in response")
	assert.Len(t, contactsArr, 1)

	mockRepo.AssertExpectations(t)
}

func TestDealHandler_AddContact(t *testing.T) {
	dealID := uuid.New()
	contactID := uuid.New()
	ownerID := uuid.New()
	pipelineID := uuid.New()

	tests := []struct {
		name          string
		body          map[string]any
		setupMock     func(*mocks.MockDealRepository)
		setupContacts func(*mocks.MockContactRepository)
		wantStatus    int
	}{
		{
			name: "adds contact successfully",
			body: map[string]any{
				"contact_id": contactID.String(),
				"role":       "decision_maker",
			},
			setupMock: func(m *mocks.MockDealRepository) {
				m.On("AddContact", mock.Anything, dealID, contactID, "decision_maker").Return(nil)
				m.On("GetByID", mock.Anything, dealID).Return(&domain.Deal{
					ID:         dealID,
					Title:      "Test Deal",
					OwnerID:    ownerID,
					PipelineID: pipelineID,
					Contacts: []domain.Contact{
						{ID: contactID, FirstName: "Bob", LastName: "Jones", OwnerID: ownerID},
					},
				}, nil)
			},
			setupContacts: func(m *mocks.MockContactRepository) {
				m.On("GetByID", mock.Anything, contactID).Return(&domain.Contact{ID: contactID, OwnerID: ownerID}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "returns 404 when contact is not visible",
			body: map[string]any{
				"contact_id": contactID.String(),
				"role":       "decision_maker",
			},
			setupMock: func(_ *mocks.MockDealRepository) {},
			setupContacts: func(m *mocks.MockContactRepository) {
				m.On("GetByID", mock.Anything, contactID).Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:          "returns 422 for missing contact_id",
			body:          map[string]any{"role": "influencer"},
			setupMock:     func(_ *mocks.MockDealRepository) {},
			setupContacts: func(_ *mocks.MockContactRepository) {},
			wantStatus:    http.StatusUnprocessableEntity,
		},
		{
			name:          "returns 400 for invalid JSON",
			body:          nil, // will send raw invalid bytes
			setupMock:     func(_ *mocks.MockDealRepository) {},
			setupContacts: func(_ *mocks.MockContactRepository) {},
			wantStatus:    http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockDealRepository)
			mockContacts := new(mocks.MockContactRepository)
			tt.setupMock(mockRepo)
			tt.setupContacts(mockContacts)

			h := handler.NewDealHandler(mockRepo).WithContacts(mockContacts)

			var bodyReader *bytes.Reader
			if tt.name == "returns 400 for invalid JSON" {
				bodyReader = bytes.NewReader([]byte("{not-json}"))
			} else {
				bodyBytes, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyReader = bytes.NewReader(bodyBytes)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/deals/"+dealID.String()+"/contacts", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			req = withURLParam(req, "id", dealID.String())
			w := httptest.NewRecorder()

			h.AddContact(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
			mockContacts.AssertExpectations(t)
		})
	}
}

func TestDealHandler_AddContact_InvalidDealID(t *testing.T) {
	mockRepo := new(mocks.MockDealRepository)
	h := handler.NewDealHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/deals/not-a-uuid/contacts",
		bytes.NewReader([]byte(`{"contact_id":"`+uuid.New().String()+`"}`)))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.AddContact(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestDealHandler_Update_InvalidStageReturns422(t *testing.T) {
	mockRepo := new(mocks.MockDealRepository)
	h := handler.NewDealHandler(mockRepo)

	dealID := uuid.New()
	body := bytes.NewReader([]byte(`{"stage":"won"}`))
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/deals/"+dealID.String(), body)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", dealID.String())
	w := httptest.NewRecorder()

	h.Update(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "closed_won")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestDealHandler_Create_RejectsUnrelatedContactAccountPair(t *testing.T) {
	dealRepo := new(mocks.MockDealRepository)
	contactRepo := new(mocks.MockContactRepository)
	h := handler.NewDealHandler(dealRepo).WithContacts(contactRepo)

	contactID := uuid.New()
	accountID := uuid.New()
	ownerID := uuid.New()
	pipelineID := uuid.New()

	contactRepo.On("IsRelatedToAccount", mock.Anything, contactID, accountID).Return(false, nil)

	reqBody := map[string]any{
		"title":       "Bad Pair",
		"owner_id":    ownerID.String(),
		"pipeline_id": pipelineID.String(),
		"contact_id":  contactID.String(),
		"account_id":  accountID.String(),
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/deals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Create(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	dealRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	contactRepo.AssertExpectations(t)
}
