package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// allProviders is the canonical list of integrations shown in the UI.
var allProviders = []domain.IntegrationProvider{
	domain.IntegrationProviderGmail,
	domain.IntegrationProviderOutlook,
	domain.IntegrationProviderGoogleCal,
	domain.IntegrationProviderOutlookCal,
	domain.IntegrationProviderConfluence,
	domain.IntegrationProviderSlack,
	domain.IntegrationProviderTeams,
	domain.IntegrationProviderStripe,
	domain.IntegrationProviderSendGrid,
	domain.IntegrationProviderTwilio,
	domain.IntegrationProviderZapier,
}

// comingSoonProviders are stubbed — OAuth flows not yet implemented.
var comingSoonProviders = map[domain.IntegrationProvider]bool{
	domain.IntegrationProviderGoogleCal:  true,
	domain.IntegrationProviderOutlookCal: true,
	domain.IntegrationProviderConfluence: true,
	domain.IntegrationProviderSlack:      true,
	domain.IntegrationProviderTeams:      true,
	domain.IntegrationProviderStripe:     true,
	domain.IntegrationProviderSendGrid:   true,
	domain.IntegrationProviderTwilio:     true,
	domain.IntegrationProviderZapier:     true,
}

// IntegrationsHandler handles integration credentials and connection status.
type IntegrationsHandler struct {
	creds       repository.IntegrationCredentialRepository
	connections repository.EmailConnectionRepository
	encKey      string
}

func NewIntegrationsHandler(
	creds repository.IntegrationCredentialRepository,
	connections repository.EmailConnectionRepository,
	encKey string,
) *IntegrationsHandler {
	return &IntegrationsHandler{
		creds:       creds,
		connections: connections,
		encKey:      encKey,
	}
}

func (h *IntegrationsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAdmin)
	r.Get("/", h.List)
	r.Get("/{provider}", h.Get)
	r.Put("/{provider}/credentials", h.UpsertCredentials)
	r.Delete("/{provider}/credentials", h.DeleteCredentials)
	r.Delete("/{provider}/connection", h.DisconnectConnection)
	return r
}

func (h *IntegrationsHandler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.ClaimsFromContext(r)
		if !ok || claims.Role != string(domain.UserRoleAdmin) {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// List returns the aggregated status for all integrations.
// GET /api/v1/integrations
func (h *IntegrationsHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r)

	statuses, err := h.buildStatusList(r, claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, statuses)
}

// Get returns the status for a single provider.
// GET /api/v1/integrations/{provider}
func (h *IntegrationsHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r)
	provider := domain.IntegrationProvider(chi.URLParam(r, "provider"))

	if !isKnownProvider(provider) {
		writeError(w, http.StatusNotFound, "unknown provider")
		return
	}

	statuses, err := h.buildStatusList(r, claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	for i := range statuses {
		if statuses[i].Provider == provider {
			writeJSON(w, http.StatusOK, statuses[i])
			return
		}
	}
	writeError(w, http.StatusNotFound, "provider not found")
}

type upsertCredentialsRequest struct {
	ClientID      *string `json:"client_id"`
	ClientSecret  *string `json:"client_secret"`
	APIKey        *string `json:"api_key"`
	WebhookSecret *string `json:"webhook_secret"`
}

// UpsertCredentials stores custom OAuth app credentials or API key for a provider.
// PUT /api/v1/integrations/{provider}/credentials
func (h *IntegrationsHandler) UpsertCredentials(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r)
	provider := domain.IntegrationProvider(chi.URLParam(r, "provider"))

	if !isKnownProvider(provider) {
		writeError(w, http.StatusNotFound, "unknown provider")
		return
	}

	var req upsertCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cred := &domain.IntegrationCredential{
		OrgID:    claims.OrgID,
		Provider: provider,
		ClientID: req.ClientID,
	}

	if req.ClientSecret != nil {
		enc, err := auth.Encrypt(h.encKey, *req.ClientSecret)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		cred.ClientSecretEnc = &enc
	}
	if req.APIKey != nil {
		enc, err := auth.Encrypt(h.encKey, *req.APIKey)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		cred.APIKeyEnc = &enc
	}
	if req.WebhookSecret != nil {
		cred.WebhookSecret = req.WebhookSecret
	}

	saved, err := h.creds.Upsert(r.Context(), cred)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// DeleteCredentials removes custom credentials, reverting to platform defaults.
// DELETE /api/v1/integrations/{provider}/credentials
func (h *IntegrationsHandler) DeleteCredentials(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r)
	provider := domain.IntegrationProvider(chi.URLParam(r, "provider"))

	if !isKnownProvider(provider) {
		writeError(w, http.StatusNotFound, "unknown provider")
		return
	}

	if err := h.creds.Delete(r.Context(), claims.OrgID, provider); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DisconnectConnection revokes the OAuth connection for email providers.
// DELETE /api/v1/integrations/{provider}/connection
func (h *IntegrationsHandler) DisconnectConnection(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r)
	provider := domain.IntegrationProvider(chi.URLParam(r, "provider"))

	if !isKnownProvider(provider) {
		writeError(w, http.StatusNotFound, "unknown provider")
		return
	}
	if comingSoonProviders[provider] {
		writeError(w, http.StatusUnprocessableEntity, "disconnect not supported for this provider yet")
		return
	}

	// Map integration provider to email provider.
	var emailProvider domain.EmailProvider
	switch provider {
	case domain.IntegrationProviderGmail:
		emailProvider = domain.EmailProviderGmail
	case domain.IntegrationProviderOutlook:
		emailProvider = domain.EmailProviderOutlook
	default:
		writeError(w, http.StatusUnprocessableEntity, "disconnect not supported for this provider")
		return
	}

	conns, err := h.connections.List(r.Context(), domain.EmailConnectionFilter{
		OrgID:    claims.OrgID,
		Provider: &emailProvider,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	for _, c := range conns {
		if err := h.connections.Delete(r.Context(), c.ID); err != nil && !errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// buildStatusList aggregates email connections and custom creds into a status list.
func (h *IntegrationsHandler) buildStatusList(r *http.Request, orgID uuid.UUID) ([]domain.IntegrationStatus, error) {
	// Fetch email connections for gmail/outlook status.
	emailConns, err := h.connections.List(r.Context(), domain.EmailConnectionFilter{
		OrgID: orgID,
	})
	if err != nil {
		return nil, err
	}

	emailConnMap := map[domain.EmailProvider]*domain.EmailConnection{}
	for _, c := range emailConns {
		emailConnMap[c.Provider] = c
	}

	// Fetch custom creds.
	credList, err := h.creds.ListByOrg(r.Context(), orgID)
	if err != nil {
		return nil, err
	}
	credMap := map[domain.IntegrationProvider]*domain.IntegrationCredential{}
	for _, c := range credList {
		credMap[c.Provider] = c
	}

	statuses := make([]domain.IntegrationStatus, 0, len(allProviders))
	for _, p := range allProviders {
		s := domain.IntegrationStatus{
			Provider:       p,
			Status:         domain.IntegrationStatusDisconnected,
			HasCustomCreds: credMap[p] != nil && (credMap[p].ClientID != nil || credMap[p].APIKeyEnc != nil),
		}

		if comingSoonProviders[p] {
			s.Status = domain.IntegrationStatusComingSoon
		}

		// Check live email connections for gmail/outlook.
		switch p {
		case domain.IntegrationProviderGmail:
			if c, ok := emailConnMap[domain.EmailProviderGmail]; ok {
				s.Status = domain.IntegrationStatusConnected
				s.EmailAddress = &c.EmailAddress
				s.LastSyncedAt = c.LastSyncedAt
			}
		case domain.IntegrationProviderOutlook:
			if c, ok := emailConnMap[domain.EmailProviderOutlook]; ok {
				s.Status = domain.IntegrationStatusConnected
				s.EmailAddress = &c.EmailAddress
				s.LastSyncedAt = c.LastSyncedAt
			}
		}

		statuses = append(statuses, s)
	}
	return statuses, nil
}

func isKnownProvider(p domain.IntegrationProvider) bool {
	for _, known := range allProviders {
		if p == known {
			return true
		}
	}
	return false
}

