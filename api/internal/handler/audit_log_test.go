package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func auditRequest(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req = req.WithContext(middleware.WithClaims(req.Context(), &auth.Claims{
		UserID: uuid.New(),
		OrgID:  domain.DefaultOrgID,
		Role:   string(domain.UserRoleAdmin),
	}))
	return req
}

func TestAuditLogListParsesSearchQuery(t *testing.T) {
	repo := new(mocks.MockAuditLogRepository)
	repo.On("List", mock.Anything, mock.MatchedBy(func(filter domain.AuditLogFilter) bool {
		return filter.OrgID == domain.DefaultOrgID && filter.Q == "Ada" && filter.Page == 1 && filter.Limit == 50
	})).Return([]*domain.AuditLog{}, 0, nil)

	handler := NewAuditLogHandler(repo).Router()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, auditRequest(http.MethodGet, "/?q=Ada"))

	require.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestAuditLogGetByIDUsesOrgScope(t *testing.T) {
	entryID := uuid.New()
	repo := new(mocks.MockAuditLogRepository)
	repo.On("GetByID", mock.Anything, domain.DefaultOrgID, entryID).Return(&domain.AuditLog{
		ID:           entryID,
		OrgID:        domain.DefaultOrgID,
		ActorType:    "system",
		ActorDisplay: "System",
		Action:       domain.AuditActionUpdated,
		EntityType:   domain.AuditEntityAccount,
		CreatedAt:    time.Now(),
	}, nil)

	handler := NewAuditLogHandler(repo).Router()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, auditRequest(http.MethodGet, "/"+entryID.String()))

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"actor_display":"System"`)
	repo.AssertExpectations(t)
}

func TestAuditLogExportIncludesActorColumns(t *testing.T) {
	userID := uuid.New()
	repo := new(mocks.MockAuditLogRepository)
	repo.On("List", mock.Anything, mock.MatchedBy(func(filter domain.AuditLogFilter) bool {
		return filter.OrgID == domain.DefaultOrgID && filter.Q == "Ada" && filter.Limit == 10000
	})).Return([]*domain.AuditLog{
		{
			ID:           uuid.New(),
			OrgID:        domain.DefaultOrgID,
			UserID:       &userID,
			ActorType:    "user",
			ActorDisplay: "Ada Lovelace",
			Action:       domain.AuditActionUpdated,
			EntityType:   domain.AuditEntityAccount,
			CreatedAt:    time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC),
		},
	}, 1, nil)

	handler := NewAuditLogHandler(repo).Router()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, auditRequest(http.MethodGet, "/export?q=Ada"))

	require.Equal(t, http.StatusOK, w.Code)
	csv := w.Body.String()
	require.True(t, strings.Contains(csv, "actor_type,actor_display,actor_name,actor_email,user_id,agent_id"))
	require.Contains(t, csv, "Ada Lovelace")
	repo.AssertExpectations(t)
}
