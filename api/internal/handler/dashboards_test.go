package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

type dashboardRepoStub struct {
	dashboardErr error
	created      *domain.ScheduledReport
	updated      domain.ScheduledReportPatch
}

func (s *dashboardRepoStub) CreateDashboard(context.Context, *domain.CustomDashboard) (*domain.CustomDashboard, error) {
	return nil, nil
}

func (s *dashboardRepoStub) GetDashboardByID(_ context.Context, id uuid.UUID) (*domain.CustomDashboard, error) {
	if s.dashboardErr != nil {
		return nil, s.dashboardErr
	}
	return &domain.CustomDashboard{ID: id}, nil
}

func (s *dashboardRepoStub) UpdateDashboard(context.Context, uuid.UUID, domain.CustomDashboardPatch) (*domain.CustomDashboard, error) {
	return nil, nil
}

func (s *dashboardRepoStub) DeleteDashboard(context.Context, uuid.UUID) error { return nil }

func (s *dashboardRepoStub) ListDashboards(context.Context) ([]*domain.CustomDashboard, error) {
	return nil, nil
}

func (s *dashboardRepoStub) CreateSchedule(_ context.Context, sched *domain.ScheduledReport) (*domain.ScheduledReport, error) {
	s.created = sched
	return sched, nil
}

func (s *dashboardRepoStub) GetScheduleByID(context.Context, uuid.UUID) (*domain.ScheduledReport, error) {
	return nil, nil
}

func (s *dashboardRepoStub) UpdateSchedule(_ context.Context, _ uuid.UUID, patch domain.ScheduledReportPatch) (*domain.ScheduledReport, error) {
	s.updated = patch
	return &domain.ScheduledReport{ID: uuid.New()}, nil
}

func (s *dashboardRepoStub) DeleteSchedule(context.Context, uuid.UUID) error { return nil }

func (s *dashboardRepoStub) ListSchedules(context.Context) ([]*domain.ScheduledReport, error) {
	return nil, nil
}

func (s *dashboardRepoStub) ListAllDueSchedules(context.Context, time.Time) ([]*domain.ScheduledReport, error) {
	return nil, nil
}

func (s *dashboardRepoStub) MarkScheduleSent(context.Context, uuid.UUID, time.Time) error {
	return nil
}

func scheduleAdminRequest(method, path string, body []byte, orgID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req = withClaims(req, &auth.Claims{UserID: uuid.New(), OrgID: orgID, Role: string(domain.UserRoleAdmin)})
	req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID:           uuid.New(),
		OrgID:            orgID,
		SuperAdminBypass: true,
	}))
	return req
}

func TestDashboardHandlerScheduleDashboardOwnership(t *testing.T) {
	t.Run("create rejects dashboard outside current org", func(t *testing.T) {
		orgID := uuid.New()
		dashboardID := uuid.New()
		repo := &dashboardRepoStub{dashboardErr: domain.ErrNotFound}
		h := handler.NewDashboardHandler(repo, new(mocks.MockReportsRepository))
		req := scheduleAdminRequest(http.MethodPost, "/", []byte(`{
			"dashboard_id":"`+dashboardID.String()+`",
			"schedule":"0 9 * * *",
			"recipients":["ops@example.com"]
		}`), orgID)
		rr := httptest.NewRecorder()

		h.ScheduleRouter().ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
		assert.Nil(t, repo.created)
	})

	t.Run("create accepts dashboard scoped to current org", func(t *testing.T) {
		orgID := uuid.New()
		dashboardID := uuid.New()
		repo := &dashboardRepoStub{}
		h := handler.NewDashboardHandler(repo, new(mocks.MockReportsRepository))
		req := scheduleAdminRequest(http.MethodPost, "/", []byte(`{
			"dashboard_id":"`+dashboardID.String()+`",
			"schedule":"0 9 * * *",
			"recipients":["ops@example.com"]
		}`), orgID)
		rr := httptest.NewRecorder()

		h.ScheduleRouter().ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
		require.NotNil(t, repo.created)
		assert.Equal(t, orgID, repo.created.OrgID)
		assert.Equal(t, dashboardID, repo.created.DashboardID)
	})

	t.Run("update rejects a replacement dashboard outside current org", func(t *testing.T) {
		orgID := uuid.New()
		scheduleID := uuid.New()
		dashboardID := uuid.New()
		repo := &dashboardRepoStub{dashboardErr: domain.ErrNotFound}
		h := handler.NewDashboardHandler(repo, new(mocks.MockReportsRepository))
		req := scheduleAdminRequest(http.MethodPatch, "/"+scheduleID.String(), []byte(`{
			"dashboard_id":"`+dashboardID.String()+`"
		}`), orgID)
		rr := httptest.NewRecorder()

		h.ScheduleRouter().ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
		assert.Nil(t, repo.updated.DashboardID)
	})
}
