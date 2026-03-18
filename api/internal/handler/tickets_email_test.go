package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
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
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

// captureQueue captures enqueued EmailJobs for assertion.
type captureQueue struct {
	mu   sync.Mutex
	jobs []domain.EmailJob
}

func (q *captureQueue) Enqueue(job domain.EmailJob) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, job)
}

func (q *captureQueue) Jobs() []domain.EmailJob {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]domain.EmailJob, len(q.jobs))
	copy(out, q.jobs)
	return out
}

// buildTicketRequest constructs a PATCH request with admin JWT claims.
func buildTicketRequest(t *testing.T, ticketID uuid.UUID, body any) *http.Request {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPatch, "/tickets/"+ticketID.String(), bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")

	jwtSvc := auth.NewJWTService("test-secret")
	orgID := uuid.New()
	token, err := jwtSvc.Issue(auth.Claims{UserID: uuid.New(), OrgID: orgID, Role: "admin"}, time.Hour)
	require.NoError(t, err)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	r = r.WithContext(middleware.WithClaims(r.Context(), &auth.Claims{
		UserID: uuid.New(), OrgID: orgID, Role: "admin",
	}))
	// Inject chi URL params.
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ticketID.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	return r
}

func TestTicketHandler_Update_EnqueuesAssignedEmail(t *testing.T) {
	ticketID := uuid.New()
	assigneeID := uuid.New()
	orgID := uuid.New()

	oldTicket := &domain.Ticket{
		ID:      ticketID,
		OrgID:   orgID,
		Subject: "Cannot login",
		Status:  domain.TicketStatusOpen,
	}
	newStatus := domain.TicketStatusOpen
	newTicket := &domain.Ticket{
		ID:         ticketID,
		OrgID:      orgID,
		Subject:    "Cannot login",
		Status:     newStatus,
		AssigneeID: &assigneeID,
	}
	assignee := &domain.User{
		ID:    assigneeID,
		OrgID: orgID,
		Email: "agent@test.com",
		Name:  "Agent One",
		Role:  domain.UserRoleAgent,
	}
	pref := &domain.UserNotificationPref{
		UserID:          assigneeID,
		OrgID:           orgID,
		EmailOnAssigned: true,
	}

	ticketRepo := &mocks.MockTicketRepository{}
	ticketRepo.On("GetByID", mock.Anything, ticketID).Return(oldTicket, nil)
	ticketRepo.On("Update", mock.Anything, ticketID, mock.Anything).Return(newTicket, nil)

	userRepo := &mocks.MockUserRepository{}
	userRepo.On("GetByID", mock.Anything, assigneeID).Return(assignee, nil)

	notifPrefRepo := &mocks.MockNotificationPrefRepository{}
	notifPrefRepo.On("GetByUser", mock.Anything, assigneeID, orgID).Return(pref, nil)

	queue := &captureQueue{}

	h := handler.NewTicketHandler(
		ticketRepo,
		&mocks.MockTicketCommentRepository{},
		&mocks.MockTicketAttachmentRepository{},
		&mocks.MockSLAPolicyRepository{},
	).WithEmailNotifications(userRepo, notifPrefRepo, queue)

	patch := map[string]any{"assignee_id": assigneeID.String()}
	req := buildTicketRequest(t, ticketID, patch)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Patch("/tickets/{id}", h.Update)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Give the goroutine a moment to run.
	time.Sleep(50 * time.Millisecond)

	jobs := queue.Jobs()
	require.Len(t, jobs, 1)
	assert.Equal(t, domain.EmailEventAssigned, jobs[0].Kind)
	assert.Equal(t, "agent@test.com", jobs[0].ToEmail)
	assert.Equal(t, ticketID.String(), jobs[0].TicketID)
}

func TestTicketHandler_Update_EnqueuesResolvedEmail(t *testing.T) {
	ticketID := uuid.New()
	reporterID := uuid.New()
	orgID := uuid.New()

	oldTicket := &domain.Ticket{
		ID:                ticketID,
		OrgID:             orgID,
		Subject:           "Export fails",
		Status:            domain.TicketStatusInProgress,
		SubmittedByUserID: &reporterID,
	}
	newTicket := &domain.Ticket{
		ID:                ticketID,
		OrgID:             orgID,
		Subject:           "Export fails",
		Status:            domain.TicketStatusResolved,
		SubmittedByUserID: &reporterID,
	}
	reporter := &domain.User{
		ID:    reporterID,
		OrgID: orgID,
		Email: "client@test.com",
		Name:  "Client User",
		Role:  domain.UserRoleClient,
	}
	pref := &domain.UserNotificationPref{
		UserID:          reporterID,
		OrgID:           orgID,
		EmailOnResolved: true,
	}

	ticketRepo := &mocks.MockTicketRepository{}
	ticketRepo.On("GetByID", mock.Anything, ticketID).Return(oldTicket, nil)
	newStatus := domain.TicketStatusResolved
	ticketRepo.On("Update", mock.Anything, ticketID, mock.MatchedBy(func(p domain.TicketPatch) bool {
		return p.Status != nil && *p.Status == newStatus
	})).Return(newTicket, nil)

	userRepo := &mocks.MockUserRepository{}
	userRepo.On("GetByID", mock.Anything, reporterID).Return(reporter, nil)

	notifPrefRepo := &mocks.MockNotificationPrefRepository{}
	notifPrefRepo.On("GetByUser", mock.Anything, reporterID, orgID).Return(pref, nil)

	queue := &captureQueue{}

	h := handler.NewTicketHandler(
		ticketRepo,
		&mocks.MockTicketCommentRepository{},
		&mocks.MockTicketAttachmentRepository{},
		&mocks.MockSLAPolicyRepository{},
	).WithEmailNotifications(userRepo, notifPrefRepo, queue)

	patch := map[string]any{"status": "resolved"}
	req := buildTicketRequest(t, ticketID, patch)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Patch("/tickets/{id}", h.Update)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	time.Sleep(50 * time.Millisecond)

	jobs := queue.Jobs()
	require.Len(t, jobs, 1)
	assert.Equal(t, domain.EmailEventResolved, jobs[0].Kind)
	assert.Equal(t, "client@test.com", jobs[0].ToEmail)
}

func TestTicketHandler_Update_RespectsOptOut(t *testing.T) {
	ticketID := uuid.New()
	assigneeID := uuid.New()
	orgID := uuid.New()

	oldTicket := &domain.Ticket{ID: ticketID, OrgID: orgID, Subject: "S", Status: domain.TicketStatusOpen}
	newTicket := &domain.Ticket{ID: ticketID, OrgID: orgID, Subject: "S", Status: domain.TicketStatusOpen, AssigneeID: &assigneeID}
	assignee := &domain.User{ID: assigneeID, OrgID: orgID, Email: "a@t.com", Name: "Agent"}
	pref := &domain.UserNotificationPref{UserID: assigneeID, OrgID: orgID, EmailOnAssigned: false} // opted out

	ticketRepo := &mocks.MockTicketRepository{}
	ticketRepo.On("GetByID", mock.Anything, ticketID).Return(oldTicket, nil)
	ticketRepo.On("Update", mock.Anything, ticketID, mock.Anything).Return(newTicket, nil)

	userRepo := &mocks.MockUserRepository{}
	userRepo.On("GetByID", mock.Anything, assigneeID).Return(assignee, nil)

	notifPrefRepo := &mocks.MockNotificationPrefRepository{}
	notifPrefRepo.On("GetByUser", mock.Anything, assigneeID, orgID).Return(pref, nil)

	queue := &captureQueue{}

	h := handler.NewTicketHandler(
		ticketRepo,
		&mocks.MockTicketCommentRepository{},
		&mocks.MockTicketAttachmentRepository{},
		&mocks.MockSLAPolicyRepository{},
	).WithEmailNotifications(userRepo, notifPrefRepo, queue)

	patch := map[string]any{"assignee_id": assigneeID.String()}
	req := buildTicketRequest(t, ticketID, patch)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Patch("/tickets/{id}", h.Update)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	time.Sleep(50 * time.Millisecond)

	jobs := queue.Jobs()
	assert.Empty(t, jobs, "should not enqueue when user opted out of assigned emails")
}
