package handler

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestEmailInboxSendViaConnection_AllowsLegacyMockPlaintextToken(t *testing.T) {
	orgID := uuid.New()
	connectionID := uuid.New()

	inboxRepo := &fakeEmailInboxRepo{}
	connRepo := &fakeEmailConnectionRepo{
		byID: map[uuid.UUID]*domain.EmailConnection{
			connectionID: {
				ID:           connectionID,
				OrgID:        orgID,
				UserID:       uuid.New(),
				Provider:     domain.EmailProviderGmail,
				EmailAddress: "admin@omnir.test",
				AccessToken:  "mock-access-token-qa",
				TokenExpiry:  time.Now().Add(1 * time.Hour),
			},
		},
	}
	h := NewEmailInboxHandler(connRepo, inboxRepo, config.EmailInboxConfig{EncryptionKey: "test-key"})

	body, err := json.Marshal(map[string]any{
		"connection_id": connectionID,
		"to":            []string{"buyer@example.com"},
		"subject":       "Re: Pricing",
		"body_html":     "<p>Hello</p>",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/emails/send", bytes.NewReader(body))
	req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
	rr := httptest.NewRecorder()

	h.SendViaConnection(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.NotNil(t, inboxRepo.upserted)
	require.Equal(t, domain.EmailDirectionOutbound, inboxRepo.upserted.Direction)
	require.Equal(t, []string{"buyer@example.com"}, inboxRepo.upserted.ToAddrs)
}

func TestEmailInboxSendViaConnection_InvalidOpaqueTokenReturns422(t *testing.T) {
	orgID := uuid.New()
	connectionID := uuid.New()

	inboxRepo := &fakeEmailInboxRepo{}
	connRepo := &fakeEmailConnectionRepo{
		byID: map[uuid.UUID]*domain.EmailConnection{
			connectionID: {
				ID:           connectionID,
				OrgID:        orgID,
				UserID:       uuid.New(),
				Provider:     domain.EmailProviderGmail,
				EmailAddress: "admin@omnir.test",
				AccessToken:  "not-a-valid-ciphertext-or-plaintext-token",
				TokenExpiry:  time.Now().Add(1 * time.Hour),
			},
		},
	}
	h := NewEmailInboxHandler(connRepo, inboxRepo, config.EmailInboxConfig{EncryptionKey: "test-key"})

	body, err := json.Marshal(map[string]any{
		"connection_id": connectionID,
		"to":            []string{"buyer@example.com"},
		"subject":       "Re: Pricing",
		"body_html":     "<p>Hello</p>",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/emails/send", bytes.NewReader(body))
	req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
	rr := httptest.NewRecorder()

	h.SendViaConnection(rr, req)

	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	require.Nil(t, inboxRepo.upserted)
}

type fakeEmailInboxRepo struct {
	filter   domain.EmailInboxFilter
	threads  []*domain.EmailInboxThreadSummary
	total    int
	upserted *domain.EmailInboxMessage
}

func (r *fakeEmailInboxRepo) Upsert(_ context.Context, msg *domain.EmailInboxMessage) (*domain.EmailInboxMessage, error) {
	r.upserted = msg
	return msg, nil
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

type fakeEmailConnectionRepo struct {
	byID map[uuid.UUID]*domain.EmailConnection
}

func (r *fakeEmailConnectionRepo) Upsert(_ context.Context, c *domain.EmailConnection) (*domain.EmailConnection, error) {
	r.byID[c.ID] = c
	return c, nil
}

func (r *fakeEmailConnectionRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.EmailConnection, error) {
	c, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return c, nil
}

func (r *fakeEmailConnectionRepo) GetByUserAndProvider(context.Context, uuid.UUID, uuid.UUID, domain.EmailProvider) (*domain.EmailConnection, error) {
	return nil, domain.ErrNotFound
}

func (r *fakeEmailConnectionRepo) List(context.Context, domain.EmailConnectionFilter) ([]*domain.EmailConnection, error) {
	return nil, nil
}

func (r *fakeEmailConnectionRepo) Update(context.Context, uuid.UUID, domain.EmailConnectionPatch) (*domain.EmailConnection, error) {
	return nil, nil
}

func (r *fakeEmailConnectionRepo) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r *fakeEmailConnectionRepo) ListAllActive(context.Context) ([]*domain.EmailConnection, error) {
	return nil, nil
}
