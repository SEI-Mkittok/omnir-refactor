package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
)

func TestEmailInboxRoutesMountUnderEmails(t *testing.T) {
	orgID := uuid.New()
	contactID := uuid.New()
	connectionID := uuid.New()
	repo := &fakeEmailInboxRepo{
		threads: []*domain.EmailInboxThreadSummary{
			{
				ThreadID:      "thread-1",
				OrgID:         orgID,
				ConnectionID:  connectionID,
				Subject:       "Account timeline thread",
				Participants:  []string{"sales@example.com"},
				Snippet:       "Hello from the account page",
				MessageCount:  1,
				LastMessageAt: time.Now(),
				ContactID:     &contactID,
			},
		},
		total: 1,
	}
	handler := NewEmailInboxHandler(nil, repo, config.EmailInboxConfig{})
	emailHandler := NewEmailHandler(nil, nil, nil, nil, nil, "sales@example.com")

	r := chi.NewRouter()
	r.Route("/emails", func(r chi.Router) {
		r.Mount("/", emailHandler.Router())
		handler.RegisterInboxRoutes(r)
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/emails/threads?contact_id="+contactID.String()+"&page=2&limit=25",
		nil,
	)
	req = req.WithContext(domain.WithOrgID(
		middleware.WithClaims(req.Context(), &auth.Claims{
			UserID: uuid.New(),
			OrgID:  orgID,
			Role:   string(domain.UserRoleAgent),
		}),
		orgID,
	))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, orgID, repo.filter.OrgID)
	require.Equal(t, contactID, *repo.filter.ContactID)
	require.Equal(t, 2, repo.filter.Page)
	require.Equal(t, 25, repo.filter.Limit)
}

type fakeEmailInboxRepo struct {
	filter  domain.EmailInboxFilter
	threads []*domain.EmailInboxThreadSummary
	total   int
}

func (r *fakeEmailInboxRepo) Upsert(context.Context, *domain.EmailInboxMessage) (*domain.EmailInboxMessage, error) {
	return nil, nil
}

func (r *fakeEmailInboxRepo) List(context.Context, domain.EmailInboxFilter) ([]*domain.EmailInboxMessage, int, error) {
	return nil, 0, nil
}

func (r *fakeEmailInboxRepo) ListThreads(_ context.Context, filter domain.EmailInboxFilter) ([]*domain.EmailInboxThreadSummary, int, error) {
	r.filter = filter
	return r.threads, r.total, nil
}

func (r *fakeEmailInboxRepo) GetThread(context.Context, uuid.UUID, string) ([]*domain.EmailInboxMessage, error) {
	return nil, nil
}

func (r *fakeEmailInboxRepo) LinkContact(context.Context, uuid.UUID, string, uuid.UUID) error {
	return nil
}

func (r *fakeEmailInboxRepo) MarkThreadRead(context.Context, uuid.UUID, string) error {
	return nil
}
