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
	"omnir/internal/domain"
	"omnir/internal/handler"
	"omnir/internal/testutil/mocks"
)

func TestContactHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockContactRepository)
		wantStatus int
	}{
		{
			name: "creates contact successfully",
			body: map[string]any{
				"firstName": "Ada",
				"lastName":  "Lovelace",
				"email":     "ada@example.com",
			},
			setupMock: func(m *mocks.MockContactRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Contact")).
					Return(&domain.Contact{
						ID:        uuid.New(),
						FirstName: "Ada",
						LastName:  "Lovelace",
						Email:     "ada@example.com",
					}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 400 for missing firstName",
			body:       map[string]any{"email": "ada@example.com"},
			setupMock:  func(m *mocks.MockContactRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid email",
			body:       map[string]any{"firstName": "Ada", "email": "not-an-email"},
			setupMock:  func(m *mocks.MockContactRepository) {},
			wantStatus: http.StatusBadRequest,
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
			setupMock:  func(m *mocks.MockContactRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockContactRepository)
			tt.setupMock(mockRepo)

			h := handler.NewContactHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts/"+tt.contactID, nil)
			w := httptest.NewRecorder()

			// Inject URL param (Chi router context would normally do this)
			h.GetByID(w, req, tt.contactID)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
