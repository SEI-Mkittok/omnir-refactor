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
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNoteHandler_Create(t *testing.T) {
	contactID := uuid.New()
	authorID := uuid.New()

	tests := []struct {
		name       string
		body       map[string]any
		setupMock  func(*mocks.MockNoteRepository)
		wantStatus int
	}{
		{
			name: "creates note successfully",
			body: map[string]any{
				"content":   "Met at SaaStr. Very interested.",
				"author_id": authorID.String(),
			},
			setupMock: func(m *mocks.MockNoteRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Note")).
					Return(&domain.Note{
						ID:         uuid.New(),
						Content:    "Met at SaaStr. Very interested.",
						EntityType: domain.NoteEntityContact,
						EntityID:   contactID,
						AuthorID:   authorID,
					}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "returns 422 for missing content",
			body: map[string]any{
				"author_id": authorID.String(),
			},
			setupMock:  func(m *mocks.MockNoteRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "returns 422 for missing author_id",
			body: map[string]any{
				"content": "A note with no author",
			},
			setupMock:  func(m *mocks.MockNoteRepository) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockNoteRepository)
			tt.setupMock(mockRepo)

			h := handler.NewNoteHandler(mockRepo, domain.NoteEntityContact, "id")

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", contactID.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.Create(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNoteHandler_List(t *testing.T) {
	contactID := uuid.New()
	authorID := uuid.New()

	mockRepo := new(mocks.MockNoteRepository)
	mockRepo.On("ListByEntity", mock.Anything, mock.MatchedBy(func(f domain.NoteFilter) bool {
		return f.EntityType == domain.NoteEntityContact && f.EntityID == contactID
	})).Return([]*domain.Note{
		{
			ID:         uuid.New(),
			Content:    "Test note",
			EntityType: domain.NoteEntityContact,
			EntityID:   contactID,
			AuthorID:   authorID,
		},
	}, 1, nil)

	h := handler.NewNoteHandler(mockRepo, domain.NoteEntityContact, "id")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", contactID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestNoteHandler_Delete(t *testing.T) {
	noteID := uuid.New()
	contactID := uuid.New()

	t.Run("deletes note successfully", func(t *testing.T) {
		mockRepo := new(mocks.MockNoteRepository)
		mockRepo.On("Delete", mock.Anything, noteID).Return(nil)

		h := handler.NewNoteHandler(mockRepo, domain.NoteEntityContact, "id")

		req := httptest.NewRequest(http.MethodDelete, "/"+noteID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", contactID.String())
		rctx.URLParams.Add("noteID", noteID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent note", func(t *testing.T) {
		mockRepo := new(mocks.MockNoteRepository)
		mockRepo.On("Delete", mock.Anything, noteID).Return(domain.ErrNotFound)

		h := handler.NewNoteHandler(mockRepo, domain.NoteEntityContact, "id")

		req := httptest.NewRequest(http.MethodDelete, "/"+noteID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", contactID.String())
		rctx.URLParams.Add("noteID", noteID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockRepo.AssertExpectations(t)
	})
}
