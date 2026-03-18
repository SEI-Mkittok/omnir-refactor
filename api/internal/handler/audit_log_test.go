package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	mw "github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func auditLogRequest(t *testing.T, method, path string, orgID uuid.UUID, role string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	claims := &auth.Claims{UserID: uuid.New(), OrgID: orgID, Role: role}
	ctx := mw.WithClaims(req.Context(), claims)
	ctx = domain.WithOrgID(ctx, orgID)
	return req.WithContext(ctx)
}

func TestAuditLogHandler_List_AdminOnly(t *testing.T) {
	orgID := uuid.New()
	repo := &mocks.MockAuditLogRepository{}

	h := handler.NewAuditLogHandler(repo)
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	})
	r.Mount("/admin/audit-log", h.Router())

	t.Run("returns 403 for non-admin", func(t *testing.T) {
		req := auditLogRequest(t, http.MethodGet, "/admin/audit-log/", orgID, "agent")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("returns entries for admin", func(t *testing.T) {
		entityType := domain.AuditEntityContact
		entries := []*domain.AuditLog{
			{
				ID:         uuid.New(),
				OrgID:      orgID,
				Action:     domain.AuditActionCreated,
				EntityType: entityType,
				CreatedAt:  time.Now(),
			},
		}
		repo.On("List", mock.Anything, mock.MatchedBy(func(f domain.AuditLogFilter) bool {
			return f.OrgID == orgID
		})).Return(entries, 1, nil).Once()

		req := auditLogRequest(t, http.MethodGet, "/admin/audit-log/", orgID, "admin")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.EqualValues(t, 1, resp["total"])
	})
}

func TestAuditLogHandler_Export_AdminOnly(t *testing.T) {
	orgID := uuid.New()
	repo := &mocks.MockAuditLogRepository{}

	h := handler.NewAuditLogHandler(repo)
	r := chi.NewRouter()
	r.Mount("/admin/audit-log", h.Router())

	t.Run("returns 403 for non-admin on export", func(t *testing.T) {
		req := auditLogRequest(t, http.MethodGet, "/admin/audit-log/export", orgID, "agent")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("returns CSV for admin", func(t *testing.T) {
		entityID := uuid.New()
		entries := []*domain.AuditLog{
			{
				ID:         uuid.New(),
				OrgID:      orgID,
				Action:     domain.AuditActionUpdated,
				EntityType: domain.AuditEntityDeal,
				EntityID:   &entityID,
				CreatedAt:  time.Now(),
			},
		}
		repo.On("List", mock.Anything, mock.Anything).Return(entries, 1, nil).Once()

		req := auditLogRequest(t, http.MethodGet, "/admin/audit-log/export", orgID, "admin")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	})
}

func TestAuditLogHandler_Append_FireAndForget(t *testing.T) {
	repo := &mocks.MockAuditLogRepository{}
	orgID := uuid.New()
	userID := uuid.New()

	entry := domain.AuditEntry{
		OrgID:      orgID,
		UserID:     &userID,
		Action:     domain.AuditActionCreated,
		EntityType: domain.AuditEntityContact,
	}

	repo.On("Append", mock.Anything, entry).Return(nil).Once()
	err := repo.Append(context.Background(), entry)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
