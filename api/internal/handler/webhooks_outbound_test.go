package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

func TestOutboundWebhookHandler_AdminGuard(t *testing.T) {
	adminID := uuid.New()
	hookID := uuid.New()

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/"},
		{http.MethodPost, "/"},
		{http.MethodGet, "/" + hookID.String()},
		{http.MethodPatch, "/" + hookID.String()},
		{http.MethodDelete, "/" + hookID.String()},
		{http.MethodPost, "/" + hookID.String() + "/test"},
		{http.MethodGet, "/" + hookID.String() + "/deliveries"},
	}

	for _, route := range routes {
		route := route

		t.Run(route.method+" "+route.path+" agent gets 403", func(t *testing.T) {
			repo := new(mocks.MockOutboundWebhookRepository)
			h := handler.NewOutboundWebhookHandler(repo)

			req := httptest.NewRequest(route.method, route.path, nil)
			req = withClaims(req, userClaims(uuid.New()))
			w := httptest.NewRecorder()

			h.Router().ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code)
			repo.AssertNotCalled(t, mock.Anything)
		})

		t.Run(route.method+" "+route.path+" no claims gets 401", func(t *testing.T) {
			repo := new(mocks.MockOutboundWebhookRepository)
			h := handler.NewOutboundWebhookHandler(repo)

			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			h.Router().ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			repo.AssertNotCalled(t, mock.Anything)
		})
	}

	t.Run("admin can list webhooks", func(t *testing.T) {
		repo := new(mocks.MockOutboundWebhookRepository)
		repo.On("List", mock.Anything).Return([]*domain.Webhook{}, nil)

		h := handler.NewOutboundWebhookHandler(repo)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = withClaims(req, adminClaims(adminID))
		w := httptest.NewRecorder()

		h.Router().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		repo.AssertExpectations(t)
	})
}
