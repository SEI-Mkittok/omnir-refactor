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

// makeLeadContact returns a minimal contact in stage='lead'.
func makeLeadContact(id, ownerID uuid.UUID) *domain.Contact {
	return &domain.Contact{
		ID:        id,
		OrgID:     uuid.New(),
		FirstName: "Jane",
		LastName:  "Doe",
		OwnerID:   ownerID,
		Stage:     domain.ContactStageLead,
		LeadScore: 0,
		Tags:      []string{},
	}
}

// --- LeadSources ---

func TestContactHandler_LeadSources(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mocks.MockContactRepository)
		wantStatus int
		wantBody   []string
	}{
		{
			name: "returns sources list",
			setupMock: func(m *mocks.MockContactRepository) {
				m.On("ListLeadSources", mock.Anything).
					Return([]string{"referral", "web_form"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   []string{"referral", "web_form"},
		},
		{
			name: "returns empty list when no sources",
			setupMock: func(m *mocks.MockContactRepository) {
				m.On("ListLeadSources", mock.Anything).
					Return([]string{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockContactRepository)
			tt.setupMock(mockRepo)

			h := handler.NewContactHandler(mockRepo)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts/lead-sources", nil)
			w := httptest.NewRecorder()

			h.LeadSources(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			var resp map[string][]string
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, tt.wantBody, resp["sources"])
			mockRepo.AssertExpectations(t)
		})
	}
}

// --- UpdateLeadScore ---

func TestContactHandler_UpdateLeadScore(t *testing.T) {
	contactID := uuid.New()
	ownerID := uuid.New()

	tests := []struct {
		name       string
		contactID  string
		body       map[string]any
		setupMock  func(*mocks.MockContactRepository)
		wantStatus int
	}{
		{
			name:      "sets absolute score",
			contactID: contactID.String(),
			body:      map[string]any{"score": 75},
			setupMock: func(m *mocks.MockContactRepository) {
				score := 75
				m.On("UpdateLeadScore", mock.Anything, contactID, domain.LeadScorePatch{Score: &score}).
					Return(makeLeadContact(contactID, ownerID), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "applies delta",
			contactID: contactID.String(),
			body:      map[string]any{"delta": 10},
			setupMock: func(m *mocks.MockContactRepository) {
				delta := 10
				m.On("UpdateLeadScore", mock.Anything, contactID, domain.LeadScorePatch{Delta: &delta}).
					Return(makeLeadContact(contactID, ownerID), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid id",
			contactID:  "not-a-uuid",
			body:       map[string]any{"score": 50},
			setupMock:  func(_ *mocks.MockContactRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for missing score and delta",
			contactID:  contactID.String(),
			body:       map[string]any{},
			setupMock:  func(_ *mocks.MockContactRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 404 when contact not found",
			contactID: contactID.String(),
			body:      map[string]any{"score": 50},
			setupMock: func(m *mocks.MockContactRepository) {
				score := 50
				m.On("UpdateLeadScore", mock.Anything, contactID, domain.LeadScorePatch{Score: &score}).
					Return(nil, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockContactRepository)
			tt.setupMock(mockRepo)

			h := handler.NewContactHandler(mockRepo)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/contacts/"+tt.contactID+"/score", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = withURLParam(req, "id", tt.contactID)
			w := httptest.NewRecorder()

			h.UpdateLeadScore(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// --- ConvertLead ---

func TestContactHandler_ConvertLead(t *testing.T) {
	contactID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	convertedContact := makeLeadContact(contactID, ownerID)
	convertedAt := time.Now()
	convertedContact.Stage = domain.ContactStageProspect
	convertedContact.ConvertedAt = &convertedAt
	convertedContact.ConvertedBy = &userID

	tests := []struct {
		name       string
		contactID  string
		body       map[string]any
		claims     *auth.Claims
		setupMocks func(*mocks.MockContactRepository, *mocks.MockDealRepository)
		withDeals  bool
		wantStatus int
	}{
		{
			name:      "converts lead without deal creation",
			contactID: contactID.String(),
			body:      map[string]any{"create_deal": false},
			claims:    &auth.Claims{UserID: userID},
			setupMocks: func(mc *mocks.MockContactRepository, _ *mocks.MockDealRepository) {
				mc.On("ConvertLead", mock.Anything, contactID, userID, (*uuid.UUID)(nil)).
					Return(convertedContact, nil)
			},
			withDeals:  false,
			wantStatus: http.StatusOK,
		},
		{
			name:      "converts lead with deal creation",
			contactID: contactID.String(),
			body:      map[string]any{"create_deal": true, "deal_title": "Jane Doe Deal"},
			claims:    &auth.Claims{UserID: userID},
			setupMocks: func(mc *mocks.MockContactRepository, md *mocks.MockDealRepository) {
				mc.On("GetByID", mock.Anything, contactID).
					Return(makeLeadContact(contactID, ownerID), nil)
				dealID := uuid.New()
				md.On("Create", mock.Anything, mock.AnythingOfType("*domain.Deal")).
					Return(&domain.Deal{ID: dealID, Title: "Jane Doe Deal"}, nil)
				mc.On("ConvertLead", mock.Anything, contactID, userID, mock.AnythingOfType("*uuid.UUID")).
					Return(convertedContact, nil)
			},
			withDeals:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid id",
			contactID:  "bad-uuid",
			body:       map[string]any{},
			claims:     &auth.Claims{UserID: userID},
			setupMocks: func(_ *mocks.MockContactRepository, _ *mocks.MockDealRepository) {},
			withDeals:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "double-convert is idempotent (already converted)",
			contactID: contactID.String(),
			body:      map[string]any{"create_deal": false},
			claims:    &auth.Claims{UserID: userID},
			setupMocks: func(mc *mocks.MockContactRepository, _ *mocks.MockDealRepository) {
				// ConvertLead uses COALESCE so calling it again returns unchanged contact.
				mc.On("ConvertLead", mock.Anything, contactID, userID, (*uuid.UUID)(nil)).
					Return(convertedContact, nil)
			},
			withDeals:  false,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockContactRepository)
			mockDeals := new(mocks.MockDealRepository)
			tt.setupMocks(mockRepo, mockDeals)

			h := handler.NewContactHandler(mockRepo)
			if tt.withDeals {
				h = h.WithDeals(mockDeals)
			}

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts/"+tt.contactID+"/convert", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = withURLParam(req, "id", tt.contactID)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			w := httptest.NewRecorder()

			h.ConvertLead(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
			mockDeals.AssertExpectations(t)
		})
	}
}
