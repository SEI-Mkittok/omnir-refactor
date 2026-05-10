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
	if patch.CompanyName != nil {
		s.CompanyName = patch.CompanyName
	}
	if patch.CompanyLogoURL != nil {
		s.CompanyLogoURL = patch.CompanyLogoURL
	}
	if patch.CompanyWebsite != nil {
		s.CompanyWebsite = patch.CompanyWebsite
	}
	if patch.CompanyEmail != nil {
		s.CompanyEmail = patch.CompanyEmail
	}
	if patch.CompanyPhone != nil {
		s.CompanyPhone = patch.CompanyPhone
	}
	if patch.CompanyAddressLine1 != nil {
		s.CompanyAddressLine1 = patch.CompanyAddressLine1
	}
	if patch.CompanyAddressLine2 != nil {
		s.CompanyAddressLine2 = patch.CompanyAddressLine2
	}
	if patch.CompanyCity != nil {
		s.CompanyCity = patch.CompanyCity
	}
	if patch.CompanyState != nil {
		s.CompanyState = patch.CompanyState
	}
	if patch.CompanyPostalCode != nil {
		s.CompanyPostalCode = patch.CompanyPostalCode
	}
	if patch.CompanyCountry != nil {
		s.CompanyCountry = patch.CompanyCountry
	}
	if patch.PortalEnabled != nil {
		s.PortalEnabled = *patch.PortalEnabled
	}
	if patch.PortalDisplayName != nil {
		s.PortalDisplayName = patch.PortalDisplayName
	}
	if patch.PortalAnnouncement != nil {
		s.PortalAnnouncement = patch.PortalAnnouncement
	}
	if patch.PortalDefaultAssigneeID != nil {
		s.PortalDefaultAssigneeID = patch.PortalDefaultAssigneeID
	}
	if patch.PortalMenu != nil {
		s.PortalMenu = *patch.PortalMenu
	}
	if patch.PortalShortcuts != nil {
		s.PortalShortcuts = *patch.PortalShortcuts
	}
	if patch.PortalRecentWidgetLimit != nil {
		s.PortalRecentWidgetLimit = *patch.PortalRecentWidgetLimit
	}
	if patch.SMTPHost != nil {
		s.SMTPHost = patch.SMTPHost
	}
	if patch.SMTPPort != nil {
		s.SMTPPort = patch.SMTPPort
	}
	if patch.SMTPUsername != nil {
		s.SMTPUsername = patch.SMTPUsername
	}
	if patch.SMTPFromEmail != nil {
		s.SMTPFromEmail = patch.SMTPFromEmail
	}
	if patch.SMTPFromName != nil {
		s.SMTPFromName = patch.SMTPFromName
	}
	if patch.SMTPSecurity != nil {
		s.SMTPSecurity = patch.SMTPSecurity
	}
	if patch.SMTPAuthType != nil {
		s.SMTPAuthType = patch.SMTPAuthType
	}
	if patch.SMTPPasswordEnc != nil {
		s.SMTPPasswordEnc = patch.SMTPPasswordEnc
		s.SMTPPasswordSet = *patch.SMTPPasswordEnc != ""
	}
	if patch.ConfigSupportEmail != nil {
		s.ConfigSupportEmail = patch.ConfigSupportEmail
	}
	if patch.ConfigUploadMaxMB != nil {
		s.ConfigUploadMaxMB = *patch.ConfigUploadMaxMB
	}
	if patch.ConfigDefaultPageSize != nil {
		s.ConfigDefaultPageSize = *patch.ConfigDefaultPageSize
	}
	if patch.ConfigListPreviewChars != nil {
		s.ConfigListPreviewChars = *patch.ConfigListPreviewChars
	}
	if patch.MenuConfig != nil {
		s.MenuConfig = *patch.MenuConfig
	}
	f.settings = s
	return s, nil
}

func makeOrgSettings(orgID uuid.UUID) *domain.OrgSettings {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	return &domain.OrgSettings{
		ID:                      uuid.New(),
		OrgID:                   orgID,
		QuoteNumberStart:        1000,
		TicketNumberStart:       2000,
		KBArticleNumberStart:    3000,
		InvoiceNumberStart:      4000,
		PortalMenu:              []string{},
		PortalShortcuts:         []string{},
		PortalRecentWidgetLimit: 5,
		ConfigUploadMaxMB:       25,
		ConfigDefaultPageSize:   25,
		ConfigListPreviewChars:  120,
		MenuConfig:              map[string]bool{},
		CreatedAt:               now,
		UpdatedAt:               now,
	}
}

func TestOrgSettingsHandler_AllowsWorkspaceAdmins(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeOrgSettingsRepository{settings: makeOrgSettings(orgID)}
	h := handler.NewOrgSettingsHandler(repo, "test-key")

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
	h := handler.NewOrgSettingsHandler(repo, "test-key")

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
	h := handler.NewOrgSettingsHandler(repo, "test-key")
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
	h := handler.NewOrgSettingsHandler(repo, "test-key")
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

func TestOrgSettingsHandler_UpdateCompanyPersistsFields(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeOrgSettingsRepository{}
	h := handler.NewOrgSettingsHandler(repo, "test-key")

	body := []byte(`{
		"company_name":"Acme CRM",
		"company_email":"ops@acme.test",
		"company_phone":"+1-555-0100"
	}`)
	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	req = withClaims(req, &auth.Claims{
		UserID: uuid.New(),
		OrgID:  orgID,
		Role:   string(domain.UserRoleAdmin),
	})
	w := httptest.NewRecorder()

	h.CompanyRouter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, repo.updatePatch.CompanyName)
	require.NotNil(t, repo.updatePatch.CompanyEmail)
	assert.Equal(t, "Acme CRM", *repo.updatePatch.CompanyName)
	assert.Equal(t, "ops@acme.test", *repo.updatePatch.CompanyEmail)
}

func TestOrgSettingsHandler_UpdateOutgoingServerEncryptsPassword(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeOrgSettingsRepository{}
	h := handler.NewOrgSettingsHandler(repo, "test-key")

	body := []byte(`{
		"smtp_host":"smtp.acme.test",
		"smtp_port":587,
		"smtp_auth_type":"password",
		"smtp_password":"super-secret"
	}`)
	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	req = withClaims(req, &auth.Claims{
		UserID: uuid.New(),
		OrgID:  orgID,
		Role:   string(domain.UserRoleSuperAdmin),
	})
	w := httptest.NewRecorder()

	h.OutgoingServerRouter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, repo.updatePatch.SMTPPasswordEnc)
	assert.NotEqual(t, "super-secret", *repo.updatePatch.SMTPPasswordEnc)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	_, hasPassword := resp["smtp_password"]
	assert.False(t, hasPassword)
	assert.Equal(t, true, resp["smtp_password_set"])
}

func TestOrgSettingsHandler_UpdatePortalValidatesWidgetLimit(t *testing.T) {
	repo := &fakeOrgSettingsRepository{}
	h := handler.NewOrgSettingsHandler(repo, "test-key")
	body := []byte(`{"portal_recent_widget_limit":99}`)

	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	req = withClaims(req, &auth.Claims{
		UserID: uuid.New(),
		OrgID:  uuid.New(),
		Role:   string(domain.UserRoleAdmin),
	})
	w := httptest.NewRecorder()

	h.PortalRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.False(t, repo.updateCalled)
}
