package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

// withURLParam injects a Chi URL param into a request context.
func withURLParam(r *http.Request, key, value string) *http.Request { //nolint:unparam // key is always "id" in current tests, but keeping generic for future use
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

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
			setupMock:  func(m *mocks.MockContactRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for missing owner_id",
			body:       map[string]any{"first_name": "Ada", "last_name": "Lovelace"},
			setupMock:  func(m *mocks.MockContactRepository) {},
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
			req = withURLParam(req, "id", tt.contactID)
			w := httptest.NewRecorder()

			h.GetByID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
