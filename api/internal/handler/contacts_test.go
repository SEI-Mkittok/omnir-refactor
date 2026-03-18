package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func TestContactHandler_Create(t *testing.T) {
	ownerID := uuid.New()
	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockContactRepository)
		wantStatus int
	}{
		{
			name: "creates contact successfully",
			body: map[string]any{
				"first_name": "Ada",
				"last_name":  "Lovelace",
				"email":      "ada@example.com",
				"owner_id":   ownerID.String(),
				"stage":      "lead",
			},
			setupMock: func(m *mocks.MockContactRepository) {
				email := "ada@example.com"
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Contact")).
					Return(&domain.Contact{
						ID:        uuid.New(),
						FirstName: "Ada",
						LastName:  "Lovelace",
						Email:     &email,
						OwnerID:   ownerID,
					}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 422 for missing first_name",
			body:       map[string]any{"last_name": "Lovelace", "owner_id": ownerID.String()},
			setupMock:  func(_ *mocks.MockContactRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for missing owner_id",
			body:       map[string]any{"first_name": "Ada", "last_name": "Lovelace"},
			setupMock:  func(_ *mocks.MockContactRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockContactRepository)
			tt.setupMock(mockRepo)

			h := handler.NewContactHandler(mockRepo)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestContactHandler_GetByID(t *testing.T) {
	contactID := uuid.New()

	tests := []struct {
		name       string
		contactID  string
		setupMock  func(*mocks.MockContactRepository)
		wantStatus int
	}{
		{
			name:      "returns contact by id",
			contactID: contactID.String(),
			setupMock: func(m *mocks.MockContactRepository) {
				m.On("GetByID", mock.Anything, contactID).
					Return(&domain.Contact{ID: contactID, FirstName: "Ada"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "returns 404 for unknown id",
			contactID: uuid.New().String(),
			setupMock: func(m *mocks.MockContactRepository) {
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid uuid",
			contactID:  "not-a-uuid",
			setupMock:  func(_ *mocks.MockContactRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockContactRepository)
			tt.setupMock(mockRepo)

			h := handler.NewContactHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts/"+tt.contactID, nil)
			req = withURLParam(req, "id", tt.contactID)
			w := httptest.NewRecorder()

			h.GetByID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestContactHandler_Update_CustomFields(t *testing.T) {
	contactID := uuid.New()
	ownerID := uuid.New()
	textDef := &domain.CustomFieldDefinition{
		ID: uuid.New(), Name: "notes", FieldType: domain.CustomFieldTypeText,
		EntityType: domain.CustomFieldEntityContact, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	tests := []struct {
		name       string
		body       map[string]any
		setupMocks func(*mocks.MockContactRepository, *mocks.MockCustomFieldDefinitionRepository)
		wantStatus int
	}{
		{
			name: "set valid custom field",
			body: map[string]any{"custom_fields": map[string]any{"notes": "hello"}},
			setupMocks: func(cr *mocks.MockContactRepository, cfr *mocks.MockCustomFieldDefinitionRepository) {
				cr.On("GetByID", mock.Anything, contactID).
					Return(&domain.Contact{ID: contactID, OwnerID: ownerID}, nil)
				cfr.On("List", mock.Anything, mock.Anything).Return([]*domain.CustomFieldDefinition{textDef}, nil)
				cr.On("Update", mock.Anything, contactID, mock.Anything).
					Return(&domain.Contact{ID: contactID, OwnerID: ownerID}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "reject unknown custom field",
			body: map[string]any{"custom_fields": map[string]any{"unknown_field": "value"}},
			setupMocks: func(_ *mocks.MockContactRepository, cfr *mocks.MockCustomFieldDefinitionRepository) {
				cfr.On("List", mock.Anything, mock.Anything).Return([]*domain.CustomFieldDefinition{textDef}, nil)
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "reject invalid type",
			body: map[string]any{"custom_fields": map[string]any{"notes": 42}},
			setupMocks: func(_ *mocks.MockContactRepository, cfr *mocks.MockCustomFieldDefinitionRepository) {
				cfr.On("List", mock.Anything, mock.Anything).Return([]*domain.CustomFieldDefinition{textDef}, nil)
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "clear custom field (null value)",
			body: map[string]any{"custom_fields": map[string]any{"notes": nil}},
			setupMocks: func(cr *mocks.MockContactRepository, cfr *mocks.MockCustomFieldDefinitionRepository) {
				cr.On("GetByID", mock.Anything, contactID).
					Return(&domain.Contact{ID: contactID, OwnerID: ownerID}, nil)
				cfr.On("List", mock.Anything, mock.Anything).Return([]*domain.CustomFieldDefinition{textDef}, nil)
				cr.On("Update", mock.Anything, contactID, mock.Anything).
					Return(&domain.Contact{ID: contactID, OwnerID: ownerID}, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockContactRepository)
			mockCF := new(mocks.MockCustomFieldDefinitionRepository)
			tt.setupMocks(mockRepo, mockCF)

			h := handler.NewContactHandler(mockRepo).WithCustomFields(mockCF)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/contacts/"+contactID.String(), bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = withURLParam(req, "id", contactID.String())
			w := httptest.NewRecorder()

			h.Update(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
			mockCF.AssertExpectations(t)
		})
	}
}

func TestContactHandler_GetByID_ExpandsCustomFields(t *testing.T) {
	contactID := uuid.New()
	ownerID := uuid.New()
	textDef := &domain.CustomFieldDefinition{
		ID: uuid.New(), Name: "notes", FieldType: domain.CustomFieldTypeText,
		EntityType: domain.CustomFieldEntityContact, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	mockRepo := new(mocks.MockContactRepository)
	mockCF := new(mocks.MockCustomFieldDefinitionRepository)
	mockRepo.On("GetByID", mock.Anything, contactID).
		Return(&domain.Contact{ID: contactID, OwnerID: ownerID}, nil)
	mockCF.On("List", mock.Anything, mock.Anything).
		Return([]*domain.CustomFieldDefinition{textDef}, nil)

	h := handler.NewContactHandler(mockRepo).WithCustomFields(mockCF)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts/"+contactID.String(), nil)
	req = withURLParam(req, "id", contactID.String())
	w := httptest.NewRecorder()
	h.GetByID(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	cf, ok := body["custom_fields"].(map[string]any)
	require.True(t, ok, "custom_fields should be an object")
	_, hasNotes := cf["notes"]
	assert.True(t, hasNotes, "notes field should be present (even if null)")
}
