package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

type fakePortalLinkRepository struct {
	link      *domain.PortalLink
	revokedID uuid.UUID
}

func (f *fakePortalLinkRepository) Create(context.Context, *domain.PortalLink) (*domain.PortalLink, error) {
	panic("unexpected Create call")
}

func (f *fakePortalLinkRepository) GetByID(_ context.Context, id uuid.UUID) (*domain.PortalLink, error) {
	if f.link == nil || f.link.ID != id {
		return nil, domain.ErrNotFound
	}
	return f.link, nil
}

func (f *fakePortalLinkRepository) GetByToken(context.Context, string) (*domain.PortalLink, error) {
	panic("unexpected GetByToken call")
}

func (f *fakePortalLinkRepository) ListByDeal(context.Context, uuid.UUID) ([]*domain.PortalLink, error) {
	panic("unexpected ListByDeal call")
}

func (f *fakePortalLinkRepository) Revoke(_ context.Context, id, _ uuid.UUID) error {
	f.revokedID = id
	return nil
}

func (f *fakePortalLinkRepository) IncrementView(context.Context, uuid.UUID) error {
	panic("unexpected IncrementView call")
}

func TestDealPortalLinksHandler_RevokeRequiresDealWriteAccess(t *testing.T) {
	linkID := uuid.New()
	dealID := uuid.New()
	orgID := domain.DefaultOrgID
	links := &fakePortalLinkRepository{
		link: &domain.PortalLink{ID: linkID, OrgID: orgID, DealID: dealID},
	}
	deals := new(mocks.MockDealRepository)
	deals.On("CanAccess", mock.Anything, dealID, domain.SharingAccessWrite).Return(false, nil).Once()
	h := handler.NewDealPortalLinksHandler(links, deals, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/portal-links/"+linkID.String(), nil)
	req = withURLParam(req, "id", linkID.String())
	req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: uuid.New(),
		OrgID:  orgID,
	}))
	w := httptest.NewRecorder()

	h.RevokeLink(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Equal(t, uuid.Nil, links.revokedID)
	deals.AssertExpectations(t)
}

func TestDealPortalLinksHandler_RevokeAllowsVisibleDeal(t *testing.T) {
	linkID := uuid.New()
	dealID := uuid.New()
	orgID := domain.DefaultOrgID
	links := &fakePortalLinkRepository{
		link: &domain.PortalLink{ID: linkID, OrgID: orgID, DealID: dealID},
	}
	deals := new(mocks.MockDealRepository)
	deals.On("CanAccess", mock.Anything, dealID, domain.SharingAccessWrite).Return(true, nil).Once()
	h := handler.NewDealPortalLinksHandler(links, deals, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/portal-links/"+linkID.String(), nil)
	req = withURLParam(req, "id", linkID.String())
	req = req.WithContext(domain.WithOrgID(req.Context(), orgID))
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		UserID: uuid.New(),
		OrgID:  orgID,
	}))
	w := httptest.NewRecorder()

	h.RevokeLink(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	require.Equal(t, linkID, links.revokedID)
	deals.AssertExpectations(t)
}
