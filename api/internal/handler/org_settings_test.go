package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
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
)

type fakeOrgSettingsRepository struct {
	settings     *domain.OrgSettings
	getCalled    bool
	updateCalled bool
	updatePatch  domain.OrgSettingsPatch
}

func (f *fakeOrgSettingsRepository) GetOrCreate(_ context.Context, orgID uuid.UUID) (*domain.OrgSettings, error) {
	f.getCalled = true
	if f.settings == nil {
		f.settings = makeOrgSettings(orgID)
	}
	f.settings.OrgID = orgID
	return f.settings, nil
}

func (f *fakeOrgSettingsRepository) Update(_ context.Context, orgID uuid.UUID, patch domain.OrgSettingsPatch) (*domain.OrgSettings, error) {
	f.updateCalled = true
	f.updatePatch = patch
	s := makeOrgSettings(orgID)
	if patch.QuoteNumberStart != nil {
		s.QuoteNumberStart = *patch.QuoteNumberStart
	}
	if patch.TicketNumberStart != nil {
		s.TicketNumberStart = *patch.TicketNumberStart
	}
	if patch.KBArticleNumberStart != nil {
		s.KBArticleNumberStart = *patch.KBArticleNumberStart
	}
	if patch.InvoiceNumberStart != nil {
		s.InvoiceNumberStart = *patch.InvoiceNumberStart
	}
	f.settings = s
	return s, nil
}

func makeOrgSettings(orgID uuid.UUID) *domain.OrgSettings {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	return &domain.OrgSettings{
		ID:                   uuid.New(),
		OrgID:                orgID,
		QuoteNumberStart:     1000,
		TicketNumberStart:    2000,
		KBArticleNumberStart: 3000,
		InvoiceNumberStart:   4000,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func TestOrgSettingsHandler_AllowsWorkspaceAdmins(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeOrgSettingsRepository{settings: makeOrgSettings(orgID)}
	h := handler.NewOrgSettingsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withClaims(req, &auth.Claims{
		UserID: uuid.New(),
		OrgID:  orgID,
		Role:   string(domain.UserRoleSuperAdmin),
	})
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, repo.getCalled)
	var resp domain.OrgSettings
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, orgID, resp.OrgID)
	assert.Equal(t, int64(1000), resp.QuoteNumberStart)
}

func TestOrgSettingsHandler_RejectsNonAdmins(t *testing.T) {
	repo := &fakeOrgSettingsRepository{}
	h := handler.NewOrgSettingsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withClaims(req, &auth.Claims{
		UserID: uuid.New(),
		OrgID:  uuid.New(),
		Role:   string(domain.UserRoleAgent),
	})
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, repo.getCalled)
}

func TestOrgSettingsHandler_UpdateValidatesPositiveNumbers(t *testing.T) {
	repo := &fakeOrgSettingsRepository{}
	h := handler.NewOrgSettingsHandler(repo)
	body := []byte(`{"quote_number_start":0}`)

	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	req = withClaims(req, &auth.Claims{
		UserID: uuid.New(),
		OrgID:  uuid.New(),
		Role:   string(domain.UserRoleAdmin),
	})
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.False(t, repo.updateCalled)
}

func TestOrgSettingsHandler_UpdatePersistsSnakeCaseNumbering(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeOrgSettingsRepository{}
	h := handler.NewOrgSettingsHandler(repo)
	body := []byte(`{
		"quote_number_start":101,
		"ticket_number_start":202,
		"kb_article_number_start":303,
		"invoice_number_start":404
	}`)

	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	req = withClaims(req, &auth.Claims{
		UserID: uuid.New(),
		OrgID:  orgID,
		Role:   string(domain.UserRoleAdmin),
	})
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, repo.updateCalled)
	assert.Equal(t, int64(101), *repo.updatePatch.QuoteNumberStart)
	assert.Equal(t, int64(202), *repo.updatePatch.TicketNumberStart)
	assert.Equal(t, int64(303), *repo.updatePatch.KBArticleNumberStart)
	assert.Equal(t, int64(404), *repo.updatePatch.InvoiceNumberStart)
}
