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

func sampleSLAPolicy() *domain.SLAPolicy {
	return &domain.SLAPolicy{
		ID:                  uuid.New(),
		OrgID:               uuid.New(),
		Name:                "Standard",
		ResponseTimeHours:   8,
		ResolutionTimeHours: 24,
		PriorityFilter:      []domain.TicketPriority{domain.TicketPriorityHigh},
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
}

func TestSLAPolicyHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockSLAPolicyRepository)
		wantStatus int
	}{
		{
			name: "creates policy successfully",
			body: map[string]any{
				"name":                  "Standard",
				"response_time_hours":   8,
				"resolution_time_hours": 24,
				"priority_filter":       []string{"high"},
			},
			setupMock: func(m *mocks.MockSLAPolicyRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.SLAPolicy")).
					Return(sampleSLAPolicy(), nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 422 for missing name",
			body:       map[string]any{"response_time_hours": 8, "resolution_time_hours": 24},
			setupMock:  func(m *mocks.MockSLAPolicyRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for zero response_time_hours",
			body:       map[string]any{"name": "Bad", "response_time_hours": 0, "resolution_time_hours": 24},
			setupMock:  func(m *mocks.MockSLAPolicyRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "returns 422 for zero resolution_time_hours",
			body:       map[string]any{"name": "Bad", "response_time_hours": 4, "resolution_time_hours": 0},
			setupMock:  func(m *mocks.MockSLAPolicyRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mocks.MockSLAPolicyRepository)
			tt.setupMock(m)

			h := handler.NewSLAPolicyHandler(m)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sla-policies", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			m.AssertExpectations(t)
		})
	}
}

func TestSLAPolicyHandler_List(t *testing.T) {
	m := new(mocks.MockSLAPolicyRepository)
	policies := []*domain.SLAPolicy{sampleSLAPolicy()}
	m.On("List", mock.Anything, uuid.Nil).Return(policies, nil)

	h := handler.NewSLAPolicyHandler(m)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sla-policies", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestSLAPolicyHandler_GetByID(t *testing.T) {
	t.Run("returns policy by id", func(t *testing.T) {
		p := sampleSLAPolicy()
		m := new(mocks.MockSLAPolicyRepository)
		m.On("GetByID", mock.Anything, p.ID).Return(p, nil)

		h := handler.NewSLAPolicyHandler(m)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sla-policies/"+p.ID.String(), nil)
		req = withURLParam(req, "id", p.ID.String())
		w := httptest.NewRecorder()

		h.GetByID(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		m.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid id", func(t *testing.T) {
		m := new(mocks.MockSLAPolicyRepository)
		h := handler.NewSLAPolicyHandler(m)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/sla-policies/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		w := httptest.NewRecorder()

		h.GetByID(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		id := uuid.New()
		m := new(mocks.MockSLAPolicyRepository)
		m.On("GetByID", mock.Anything, id).Return(nil, domain.ErrNotFound)

		h := handler.NewSLAPolicyHandler(m)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sla-policies/"+id.String(), nil)
		req = withURLParam(req, "id", id.String())
		w := httptest.NewRecorder()

		h.GetByID(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		m.AssertExpectations(t)
	})
}

func TestSLAPolicyHandler_Update(t *testing.T) {
	t.Run("updates policy successfully", func(t *testing.T) {
		p := sampleSLAPolicy()
		m := new(mocks.MockSLAPolicyRepository)
		m.On("Update", mock.Anything, p.ID, mock.AnythingOfType("domain.SLAPolicyPatch")).
			Return(p, nil)

		h := handler.NewSLAPolicyHandler(m)
		bodyBytes, _ := json.Marshal(map[string]any{"name": "Updated"})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/sla-policies/"+p.ID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = withURLParam(req, "id", p.ID.String())
		w := httptest.NewRecorder()

		h.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		m.AssertExpectations(t)
	})

	t.Run("returns 422 for invalid response_time_hours", func(t *testing.T) {
		id := uuid.New()
		m := new(mocks.MockSLAPolicyRepository)
		h := handler.NewSLAPolicyHandler(m)

		bodyBytes, _ := json.Marshal(map[string]any{"response_time_hours": -1})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/sla-policies/"+id.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = withURLParam(req, "id", id.String())
		w := httptest.NewRecorder()

		h.Update(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestSLAPolicyHandler_Delete(t *testing.T) {
	t.Run("deletes policy successfully", func(t *testing.T) {
		id := uuid.New()
		m := new(mocks.MockSLAPolicyRepository)
		m.On("Delete", mock.Anything, id).Return(nil)

		h := handler.NewSLAPolicyHandler(m)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/sla-policies/"+id.String(), nil)
		req = withURLParam(req, "id", id.String())
		w := httptest.NewRecorder()

		h.Delete(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		m.AssertExpectations(t)
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		id := uuid.New()
		m := new(mocks.MockSLAPolicyRepository)
		m.On("Delete", mock.Anything, id).Return(domain.ErrNotFound)

		h := handler.NewSLAPolicyHandler(m)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/sla-policies/"+id.String(), nil)
		req = withURLParam(req, "id", id.String())
		w := httptest.NewRecorder()

		h.Delete(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		m.AssertExpectations(t)
	})
}
