package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	testmocks "github.com/omnir/crm-api/internal/testutil/mocks"
)

type stubLeadStore struct {
	getByIDFn           func(ctx context.Context, id uuid.UUID) (*domain.Lead, error)
	defaultPipelineIDFn func(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error)
	convertToContactFn  func(ctx context.Context, leadID, contactID uuid.UUID) (*domain.Lead, error)
}

func (s *stubLeadStore) Create(_ context.Context, _ *domain.Lead) (*domain.Lead, error) {
	panic("unexpected Create call")
}

func (s *stubLeadStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	if s.getByIDFn == nil {
		panic("unexpected GetByID call")
	}
	return s.getByIDFn(ctx, id)
}

func (s *stubLeadStore) Update(_ context.Context, _ uuid.UUID, _ domain.LeadPatch) (*domain.Lead, error) {
	panic("unexpected Update call")
}

func (s *stubLeadStore) Delete(_ context.Context, _ uuid.UUID) error {
	panic("unexpected Delete call")
}

func (s *stubLeadStore) List(_ context.Context, _ domain.LeadFilter) ([]*domain.Lead, int, error) {
	panic("unexpected List call")
}

func (s *stubLeadStore) ListSources(_ context.Context) ([]string, error) {
	panic("unexpected ListSources call")
}

func (s *stubLeadStore) ConvertToContact(ctx context.Context, leadID, contactID uuid.UUID) (*domain.Lead, error) {
	if s.convertToContactFn != nil {
		return s.convertToContactFn(ctx, leadID, contactID)
	}
	panic("unexpected ConvertToContact call")
}

func (s *stubLeadStore) DefaultPipelineID(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error) {
	if s.defaultPipelineIDFn == nil {
		panic("unexpected DefaultPipelineID call")
	}
	return s.defaultPipelineIDFn(ctx, orgID)
}

func TestLeadHandlerConvert_NoPipeline_DoesNotCreatePartialRecords(t *testing.T) {
	leadID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	leads := &stubLeadStore{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
			return &domain.Lead{
				ID:        leadID,
				OrgID:     orgID,
				FirstName: "Alex",
				LastName:  "Prospect",
				Status:    domain.LeadStatusNew,
			}, nil
		},
		defaultPipelineIDFn: func(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, errors.New("no pipeline")
		},
		convertToContactFn: func(ctx context.Context, leadID, contactID uuid.UUID) (*domain.Lead, error) {
			t.Fatalf("ConvertToContact should not be called when no pipeline is available")
			return nil, nil
		},
	}

	contacts := &testmocks.MockContactRepository{}
	accounts := &testmocks.MockAccountRepository{}
	deals := &testmocks.MockDealRepository{}
	h := handler.NewLeadHandler(leads, contacts, accounts, deals, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/leads/"+leadID.String()+"/convert", nil)
	req = withURLParam(req, "id", leadID.String())
	req = withClaims(req, &auth.Claims{
		UserID: userID,
		OrgID:  orgID,
		Role:   string(domain.UserRoleAdmin),
	})
	req = req.WithContext(domain.WithOrgID(req.Context(), orgID))

	rr := httptest.NewRecorder()
	h.Convert(rr, req)

	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "no pipeline available for lead conversion")
	contacts.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	accounts.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	deals.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}
