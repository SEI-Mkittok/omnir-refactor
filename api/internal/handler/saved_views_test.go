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

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func newSavedView(ownerID uuid.UUID) *domain.SavedView {
	return &domain.SavedView{
		ID:         uuid.New(),
		OrgID:      uuid.New(),
		CreatedBy:  ownerID,
		EntityType: domain.SavedViewEntityContacts,
		Name:       "My View",
		Filters:    json.RawMessage(`[]`),
		IsShared:   false,
		IsPinned:   false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// --- List ---

func TestSavedViewHandler_List(t *testing.T) {
	ownerID := uuid.New()
	views := []*domain.SavedView{newSavedView(ownerID)}

	tests := []struct {
		name       string
		entityType string
		setupMock  func(*mocks.MockSavedViewRepository)
		wantStatus int
	}{
		{
			name:       "returns views for valid entity type",
			entityType: "entityType=contacts",
			setupMock: func(m *mocks.MockSavedViewRepository) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.SavedViewFilter")).
					Return(views, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "accepts snake case entity_type",
			entityType: "entity_type=quotes",
			setupMock: func(m *mocks.MockSavedViewRepository) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.SavedViewFilter")).
					Return(views, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid entity type",
			entityType: "entity_type=invalid",
			setupMock:  func(_ *mocks.MockSavedViewRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for missing entity type",
			entityType: "",
			setupMock:  func(_ *mocks.MockSavedViewRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSavedViewRepository)
			tt.setupMock(mockRepo)

			h := handler.NewSavedViewHandler(mockRepo)

			url := "/api/v1/views"
			if tt.entityType != "" {
				url += "?" + tt.entityType
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = withClaims(req, userClaims(ownerID))
			w := httptest.NewRecorder()

			h.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// --- Create ---

func TestSavedViewHandler_Create(t *testing.T) {
	ownerID := uuid.New()

	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockSavedViewRepository)
		wantStatus int
	}{
		{
			name: "creates view successfully",
			body: map[string]any{
				"entity_type": "contacts",
				"name":        "High-value leads",
				"filters":     map[string]any{"account_id": uuid.New().String(), "unknown_filter": "kept"},
			},
			setupMock: func(m *mocks.MockSavedViewRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.SavedView")).
					Return(newSavedView(ownerID), nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "normalizes legacy array filters",
			body: map[string]any{
				"entity_type": "tickets",
				"name":        "Ticket queue",
				"filters":     []any{},
			},
			setupMock: func(m *mocks.MockSavedViewRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(v *domain.SavedView) bool {
					return string(v.Filters) == "{}"
				})).Return(newSavedView(ownerID), nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 422 for missing name",
			body:       map[string]any{"entity_type": "contacts"},
			setupMock:  func(_ *mocks.MockSavedViewRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for invalid entity_type",
			body:       map[string]any{"entity_type": "widgets", "name": "My View"},
			setupMock:  func(_ *mocks.MockSavedViewRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       nil,
			setupMock:  func(_ *mocks.MockSavedViewRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSavedViewRepository)
			tt.setupMock(mockRepo)

			h := handler.NewSavedViewHandler(mockRepo)

			var bodyBytes []byte
			var err error
			if tt.body != nil {
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			} else {
				bodyBytes = []byte("not-json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/views", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = withClaims(req, userClaims(ownerID))
			w := httptest.NewRecorder()

			h.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// --- Update ---

func TestSavedViewHandler_Update(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()

	tests := []struct {
		name       string
		viewID     string
		claims     *auth.Claims
		body       map[string]any
		setupMock  func(*mocks.MockSavedViewRepository)
		wantStatus int
	}{
		{
			name:   "owner can update their view",
			viewID: uuid.New().String(),
			claims: userClaims(ownerID),
			body:   map[string]any{"name": "Updated"},
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("domain.SavedViewPatch")).
					Return(v, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "admin can update any view",
			viewID: uuid.New().String(),
			claims: adminClaims(otherID),
			body:   map[string]any{"name": "Admin update"},
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("domain.SavedViewPatch")).
					Return(v, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "non-owner agent cannot update another user's view",
			viewID: uuid.New().String(),
			claims: userClaims(otherID),
			body:   map[string]any{"name": "Sneaky update"},
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "returns 400 for invalid UUID",
			viewID:     "not-a-uuid",
			claims:     userClaims(ownerID),
			body:       map[string]any{"name": "Updated"},
			setupMock:  func(_ *mocks.MockSavedViewRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSavedViewRepository)
			tt.setupMock(mockRepo)

			h := handler.NewSavedViewHandler(mockRepo)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/views/"+tt.viewID, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = withClaims(req, tt.claims)
			req = withURLParam(req, "id", tt.viewID)
			w := httptest.NewRecorder()

			h.Update(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// --- Delete ---

func TestSavedViewHandler_Delete(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()

	tests := []struct {
		name       string
		viewID     string
		claims     *auth.Claims
		setupMock  func(*mocks.MockSavedViewRepository)
		wantStatus int
	}{
		{
			name:   "owner can delete their view",
			viewID: uuid.New().String(),
			claims: userClaims(ownerID),
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
				m.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "admin can delete any view",
			viewID: uuid.New().String(),
			claims: adminClaims(otherID),
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
				m.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "non-owner agent cannot delete another user's view",
			viewID: uuid.New().String(),
			claims: userClaims(otherID),
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "returns 404 when view not found",
			viewID: uuid.New().String(),
			claims: userClaims(ownerID),
			setupMock: func(m *mocks.MockSavedViewRepository) {
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSavedViewRepository)
			tt.setupMock(mockRepo)

			h := handler.NewSavedViewHandler(mockRepo)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/views/"+tt.viewID, nil)
			req = withClaims(req, tt.claims)
			req = withURLParam(req, "id", tt.viewID)
			w := httptest.NewRecorder()

			h.Delete(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// --- Pin ---

func TestSavedViewHandler_Pin(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()

	tests := []struct {
		name       string
		viewID     string
		claims     *auth.Claims
		body       map[string]any
		setupMock  func(*mocks.MockSavedViewRepository)
		wantStatus int
	}{
		{
			name:   "owner can toggle pin on their view",
			viewID: uuid.New().String(),
			claims: userClaims(ownerID),
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				pinned := newSavedView(ownerID)
				pinned.IsPinned = true
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
				m.On("Pin", mock.Anything, mock.AnythingOfType("uuid.UUID"), true).Return(pinned, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "owner can set explicit pin order without toggling",
			viewID: uuid.New().String(),
			claims: userClaims(ownerID),
			body:   map[string]any{"pin_order": 2},
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				v.IsPinned = true
				order := 2
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.MatchedBy(func(p domain.SavedViewPatch) bool {
					return p.IsPinned != nil && *p.IsPinned && p.PinnedOrder != nil && *p.PinnedOrder == order
				})).Return(v, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "non-owner cannot pin another user's view",
			viewID: uuid.New().String(),
			claims: userClaims(otherID),
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "admin can pin any view",
			viewID: uuid.New().String(),
			claims: adminClaims(otherID),
			setupMock: func(m *mocks.MockSavedViewRepository) {
				v := newSavedView(ownerID)
				pinned := newSavedView(ownerID)
				pinned.IsPinned = true
				m.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(v, nil)
				m.On("Pin", mock.Anything, mock.AnythingOfType("uuid.UUID"), true).Return(pinned, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSavedViewRepository)
			tt.setupMock(mockRepo)

			h := handler.NewSavedViewHandler(mockRepo)

			var bodyBytes []byte
			if tt.body != nil {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/views/"+tt.viewID+"/pin", bytes.NewReader(bodyBytes))
			req = withClaims(req, tt.claims)
			req = withURLParam(req, "id", tt.viewID)
			w := httptest.NewRecorder()

			h.Pin(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
