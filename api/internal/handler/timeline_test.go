package handler_test

import (
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

func TestTimelineHandler_List_Success(t *testing.T) {
	repo := new(mocks.MockTimelineRepository)
	h := handler.NewTimelineHandler(repo)

	contactID := uuid.New()
	eventID := uuid.New()
	now := time.Now().UTC()

	repo.On("List", mock.Anything, mock.MatchedBy(func(f domain.TimelineFilter) bool {
		return f.ContactID != nil && *f.ContactID == contactID && f.Page == 1 && f.Limit == 50
	})).Return([]*domain.TimelineEvent{{
		EventType:  "ticket.created",
		EventID:    eventID,
		OccurredAt: now,
		Preview:    "New ticket",
	}}, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/?contact_id="+contactID.String(), nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	meta, ok := body["meta"].(map[string]any)
	require.True(t, ok)
	assert.EqualValues(t, 1, meta["total"])
	repo.AssertExpectations(t)
}

func TestTimelineHandler_List_InvalidAccountID(t *testing.T) {
	h := handler.NewTimelineHandler(new(mocks.MockTimelineRepository))

	req := httptest.NewRequest(http.MethodGet, "/?account_id=bad-uuid", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTimelineHandler_List_InvalidStartAt(t *testing.T) {
	h := handler.NewTimelineHandler(new(mocks.MockTimelineRepository))

	req := httptest.NewRequest(http.MethodGet, "/?start_at=tomorrow", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTimelineHandler_List_NormalizesPaginationMetadata(t *testing.T) {
	repo := new(mocks.MockTimelineRepository)
	h := handler.NewTimelineHandler(repo)

	repo.On("List", mock.Anything, mock.MatchedBy(func(f domain.TimelineFilter) bool {
		return f.Page == 1 && f.Limit == 50
	})).Return([]*domain.TimelineEvent{}, 3, nil)

	req := httptest.NewRequest(http.MethodGet, "/?page=-2&limit=-10", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	meta, ok := body["meta"].(map[string]any)
	require.True(t, ok)
	assert.EqualValues(t, 1, meta["page"])
	assert.EqualValues(t, 50, meta["per_page"])
	assert.EqualValues(t, 3, meta["total"])
	repo.AssertExpectations(t)
}
