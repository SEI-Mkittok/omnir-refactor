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

func TestAccountHandler_Create(t *testing.T) {
	ownerID := uuid.New()

	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockAccountRepository)
		wantStatus int
	}{
		{
			name: "creates account successfully",
			body: map[string]any{
				"name":     "Acme Corp",
				"domain":   "acme.com",
				"industry": "Technology",
				"size":     "51-200",
				"owner_id": ownerID.String(),
			},
			setupMock: func(m *mocks.MockAccountRepository) {
				domainStr := "acme.com"
				industry := "Technology"
				size := domain.AccountSize51_200
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).
					Return(&domain.Account{
						ID:       uuid.New(),
						Name:     "Acme Corp",
						Domain:   &domainStr,
						Industry: &industry,
						Size:     &size,
						OwnerID:  ownerID,
					}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 422 for missing name",
			body:       map[string]any{"owner_id": ownerID.String()},
			setupMock:  func(m *mocks.MockAccountRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for missing owner_id",
			body:       map[string]any{"name": "Acme Corp"},
			setupMock:  func(m *mocks.MockAccountRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       nil,
			setupMock:  func(m *mocks.MockAccountRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockAccountRepository)
			tt.setupMock(mockRepo)

			h := handler.NewAccountHandler(mockRepo)

			var bodyReader *bytes.Reader
			if tt.body != nil {
				bodyBytes, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyReader = bytes.NewReader(bodyBytes)
			} else {
				bodyReader = bytes.NewReader([]byte("invalid-json{"))
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountHandler_GetByID(t *testing.T) {
	accountID := uuid.New()

	tests := []struct {
		name       string
		accountID  string
		setupMock  func(*mocks.MockAccountRepository)
		wantStatus int
	}{
		{
			name:      "returns account by id",
			accountID: accountID.String(),
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("GetByID", mock.Anything, accountID).
					Return(&domain.Account{ID: accountID, Name: "Acme Corp"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "returns 404 for unknown id",
			accountID: uuid.New().String(),
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid uuid",
			accountID:  "not-a-uuid",
			setupMock:  func(m *mocks.MockAccountRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockAccountRepository)
			tt.setupMock(mockRepo)

			h := handler.NewAccountHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+tt.accountID, nil)
			req = withURLParam(req, "id", tt.accountID)
			w := httptest.NewRecorder()

			h.GetByID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountHandler_Update(t *testing.T) {
	accountID := uuid.New()

	tests := []struct {
		name       string
		accountID  string
		body       map[string]any
		setupMock  func(*mocks.MockAccountRepository)
		wantStatus int
	}{
		{
			name:      "updates account successfully",
			accountID: accountID.String(),
			body:      map[string]any{"name": "Updated Corp", "industry": "Finance"},
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("GetByID", mock.Anything, accountID).
					Return(&domain.Account{ID: accountID, Name: "Acme Corp"}, nil)
				m.On("Update", mock.Anything, accountID, mock.AnythingOfType("domain.AccountPatch")).
					Return(&domain.Account{ID: accountID, Name: "Updated Corp"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "returns 404 for unknown id",
			accountID: uuid.New().String(),
			body:      map[string]any{"name": "Updated"},
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid uuid",
			accountID:  "not-a-uuid",
			body:       map[string]any{"name": "Updated"},
			setupMock:  func(m *mocks.MockAccountRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockAccountRepository)
			tt.setupMock(mockRepo)

			h := handler.NewAccountHandler(mockRepo)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/"+tt.accountID, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = withURLParam(req, "id", tt.accountID)
			w := httptest.NewRecorder()

			h.Update(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountHandler_Delete(t *testing.T) {
	accountID := uuid.New()

	tests := []struct {
		name       string
		accountID  string
		setupMock  func(*mocks.MockAccountRepository)
		wantStatus int
	}{
		{
			name:      "soft-deletes account",
			accountID: accountID.String(),
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("Delete", mock.Anything, accountID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:      "returns 404 for unknown id",
			accountID: uuid.New().String(),
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid uuid",
			accountID:  "not-a-uuid",
			setupMock:  func(m *mocks.MockAccountRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockAccountRepository)
			tt.setupMock(mockRepo)

			h := handler.NewAccountHandler(mockRepo)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/"+tt.accountID, nil)
			req = withURLParam(req, "id", tt.accountID)
			w := httptest.NewRecorder()

			h.Delete(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAccountHandler_List(t *testing.T) {
	ownerID := uuid.New()

	tests := []struct {
		name       string
		query      string
		setupMock  func(*mocks.MockAccountRepository)
		wantStatus int
		wantTotal  int
	}{
		{
			name:  "lists accounts with defaults",
			query: "",
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.AccountFilter")).
					Return([]*domain.Account{
						{ID: uuid.New(), Name: "Acme Corp", OwnerID: ownerID},
					}, 1, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name:  "filters by industry",
			query: "?industry=Technology",
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.AccountFilter) bool {
					return f.Industry != nil && *f.Industry == "Technology"
				})).Return([]*domain.Account{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
		{
			name:  "filters by size",
			query: "?size=51-200",
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.AccountFilter) bool {
					return f.Size != nil && *f.Size == domain.AccountSize51_200
				})).Return([]*domain.Account{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
		{
			name:  "filters by owner_id",
			query: "?owner_id=" + ownerID.String(),
			setupMock: func(m *mocks.MockAccountRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f domain.AccountFilter) bool {
					return f.OwnerID != nil && *f.OwnerID == ownerID
				})).Return([]*domain.Account{}, 0, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockAccountRepository)
			tt.setupMock(mockRepo)

			h := handler.NewAccountHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts"+tt.query, nil)
			w := httptest.NewRecorder()

			h.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var resp map[string]any
			require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
			assert.Equal(t, float64(tt.wantTotal), resp["total"])
			mockRepo.AssertExpectations(t)
		})
	}
}
