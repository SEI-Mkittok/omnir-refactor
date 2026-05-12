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

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func TestReportsHandler_ManagerDashboard(t *testing.T) {
	t.Run("returns manager dashboard for valid range", func(t *testing.T) {
		repo := &mocks.MockReportsRepository{}
		expected := &domain.ManagerDashboardReport{
			CRM:      domain.DashboardCRMMetrics{PipelineValueCents: 12345, WonCount: 2, LostCount: 1},
			HelpDesk: domain.DashboardHelpDeskMetrics{OpenCount: 5, BacklogCount: 2},
		}
		repo.On("ManagerDashboard", mock.Anything, mock.MatchedBy(func(f domain.ReportFilter) bool {
			if f.From == nil || f.To == nil {
				return false
			}
			return f.From.Format("2006-01-02") == "2026-04-01" &&
				f.To.Format("2006-01-02") == "2026-04-07"
		})).Return(expected, nil)

		h := handler.NewReportsHandler(repo)
		req := httptest.NewRequest(http.MethodGet, "/manager-dashboard?from=2026-04-01&to=2026-04-07", nil)
		rec := httptest.NewRecorder()

		h.Router().ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var got domain.ManagerDashboardReport
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, expected.CRM.PipelineValueCents, got.CRM.PipelineValueCents)
		repo.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid from date", func(t *testing.T) {
		repo := &mocks.MockReportsRepository{}
		h := handler.NewReportsHandler(repo)
		req := httptest.NewRequest(http.MethodGet, "/manager-dashboard?from=not-a-date", nil)
		rec := httptest.NewRecorder()

		h.Router().ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		repo.AssertNotCalled(t, "ManagerDashboard", mock.Anything, mock.Anything)
	})

	t.Run("allows org override for super admin only", func(t *testing.T) {
		repo := &mocks.MockReportsRepository{}
		orgID := uuid.New()
		repo.On("ManagerDashboard", mock.Anything, mock.MatchedBy(func(f domain.ReportFilter) bool {
			return f.OrgID != nil && *f.OrgID == orgID
		})).Return(&domain.ManagerDashboardReport{}, nil)

		h := handler.NewReportsHandler(repo)
		req := httptest.NewRequest(
			http.MethodGet,
			"/manager-dashboard?org_id="+orgID.String()+"&from="+time.Now().UTC().Format(time.RFC3339),
			nil,
		)
		req = withClaims(req, &auth.Claims{UserID: uuid.New(), Role: string(domain.UserRoleSuperAdmin)})
		rec := httptest.NewRecorder()

		h.Router().ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		repo.AssertExpectations(t)
	})

	t.Run("ignores org override for tenant admin", func(t *testing.T) {
		repo := &mocks.MockReportsRepository{}
		orgID := uuid.New()
		repo.On("ManagerDashboard", mock.Anything, mock.MatchedBy(func(f domain.ReportFilter) bool {
			return f.OrgID == nil
		})).Return(&domain.ManagerDashboardReport{}, nil)

		h := handler.NewReportsHandler(repo)
		req := httptest.NewRequest(
			http.MethodGet,
			"/manager-dashboard?org_id="+orgID.String()+"&from="+time.Now().UTC().Format(time.RFC3339),
			nil,
		)
		req = withClaims(req, &auth.Claims{UserID: uuid.New(), Role: string(domain.UserRoleAdmin)})
		rec := httptest.NewRecorder()

		h.Router().ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		repo.AssertExpectations(t)
	})
}
