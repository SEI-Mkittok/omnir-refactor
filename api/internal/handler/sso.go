package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

const ssoStateCookie = "sso_state"

// SSOHandler handles OIDC SSO login/callback and org SSO config management.
type SSOHandler struct {
	ssoConfigs     repository.SSOConfigRepository
	orgs           repository.OrgRepository
	users          repository.UserRepository
	jwtSvc         *auth.JWTService
	encryptKey     string
	callbackURL    string
	apiCallbackURL string
}

func NewSSOHandler(
	ssoConfigs repository.SSOConfigRepository,
	orgs repository.OrgRepository,
	users repository.UserRepository,
	jwtSvc *auth.JWTService,
	encryptKey, callbackURL string,
) *SSOHandler {
	return &SSOHandler{
		ssoConfigs:     ssoConfigs,
		orgs:           orgs,
		users:          users,
		jwtSvc:         jwtSvc,
		encryptKey:     encryptKey,
		callbackURL:    callbackURL,
		apiCallbackURL: callbackURL, // overridden by WithAPICallbackURL
	}
}

// WithAPICallbackURL sets the callback URL used by the API-style SSO routes.
func (h *SSOHandler) WithAPICallbackURL(url string) *SSOHandler {
	h.apiCallbackURL = url
	return h
}

// Router returns the public (browser-redirect) SSO routes.
// Mounted at /auth/sso
func (h *SSOHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/config", h.Config)
	r.Get("/{orgSlug}/login", h.Login)
	r.Get("/callback", h.Callback)
	return r
}

// ApiRouter returns the API-style SSO routes (POST initiate + GET callback).
// Mounted at /api/auth/sso
func (h *SSOHandler) ApiRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/microsoft", h.InitiateMicrosoft)
	r.Post("/google", h.InitiateGoogle)
	r.Get("/callback", h.ApiCallback)
	return r
}

// OrgSSORouter returns authenticated routes for managing org SSO config.
// Mounted at /api/v1/orgs/{orgId}/sso
func (h *SSOHandler) OrgSSORouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetOrgSSO)
	r.Patch("/", h.UpdateOrgSSO)
	return r
}

// Config returns the org's SSO provider info without requiring authentication.
// GET /auth/sso/config?orgSlug=...
func (h *SSOHandler) Config(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("orgSlug")
	var cfg *domain.SSOConfig
	var err error
	if slug != "" {
		cfg, err = h.ssoConfigs.GetByOrgSlug(r.Context(), slug)
	}
	if err != nil || cfg == nil {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled":  cfg.Enabled,
		"provider": cfg.Provider,
		"orgSlug":  slug,
	})
}

// Login redirects the browser to the OIDC provider for the given org.
// GET /auth/sso/{orgSlug}/login
func (h *SSOHandler) Login(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "orgSlug")
	cfg, err := h.ssoConfigs.GetByOrgSlug(r.Context(), slug)
	if err != nil || cfg == nil {
		writeError(w, http.StatusNotFound, "SSO not configured for this org")
		return
	}

	secret, err := auth.Decrypt(h.encryptKey, cfg.ClientSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	oauth2Cfg, err := h.buildOAuth2Config(r.Context(), cfg, secret, h.callbackURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	state := randomState()
	http.SetCookie(w, &http.Cookie{
		Name:     ssoStateCookie,
		Value:    state + "|" + slug,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, buildAuthURL(oauth2Cfg, cfg, state), http.StatusFound)
}

// InitiateMicrosoft starts a Microsoft Entra ID OIDC flow.
// POST /api/auth/sso/microsoft
// Body: {"org_slug": "acme"}
// Returns: {"redirect_url": "https://login.microsoftonline.com/..."}
func (h *SSOHandler) InitiateMicrosoft(w http.ResponseWriter, r *http.Request) {
	h.initiateProviderSSO(w, r, "microsoft")
}

// InitiateGoogle starts a Google Workspace OIDC flow.
// POST /api/auth/sso/google
// Body: {"org_slug": "acme"}
// Returns: {"redirect_url": "https://accounts.google.com/..."}
func (h *SSOHandler) InitiateGoogle(w http.ResponseWriter, r *http.Request) {
	h.initiateProviderSSO(w, r, "google")
}

func (h *SSOHandler) initiateProviderSSO(w http.ResponseWriter, r *http.Request, provider string) {
	var body struct {
		OrgSlug string `json:"org_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OrgSlug == "" {
		writeError(w, http.StatusBadRequest, "org_slug is required")
		return
	}

	cfg, err := h.ssoConfigs.GetByOrgSlug(r.Context(), body.OrgSlug)
	if err != nil || cfg == nil {
		writeError(w, http.StatusNotFound, "SSO not configured for this org")
		return
	}
	if cfg.Provider != provider {
		writeError(w, http.StatusBadRequest, "provider mismatch: org is not configured for "+provider+" SSO")
		return
	}

	secret, err := auth.Decrypt(h.encryptKey, cfg.ClientSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	oauth2Cfg, err := h.buildOAuth2Config(r.Context(), cfg, secret, h.apiCallbackURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "OIDC provider error: "+err.Error())
		return
	}

	state := randomState()
	http.SetCookie(w, &http.Cookie{
		Name:     ssoStateCookie,
		Value:    state + "|" + body.OrgSlug,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	redirectURL := buildAuthURL(oauth2Cfg, cfg, state)
	writeJSON(w, http.StatusOK, map[string]any{"redirect_url": redirectURL})
}

// Callback handles the OIDC redirect for the legacy browser-redirect flow.
// GET /auth/sso/callback
func (h *SSOHandler) Callback(w http.ResponseWriter, r *http.Request) {
	h.handleCallback(w, r, h.callbackURL)
}

// ApiCallback handles the OIDC redirect for the API-style flow.
// GET /api/auth/sso/callback
func (h *SSOHandler) ApiCallback(w http.ResponseWriter, r *http.Request) {
	h.handleCallback(w, r, h.apiCallbackURL)
}

func (h *SSOHandler) handleCallback(w http.ResponseWriter, r *http.Request, callbackURL string) {
	cookie, err := r.Cookie(ssoStateCookie)
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing SSO state")
		return
	}

	// Extract state and slug from cookie value: "state|slug"
	cookieVal := cookie.Value
	pipe := len(cookieVal) - 1
	for i, c := range cookieVal {
		if c == '|' {
			pipe = i
		}
	}
	cookieState := cookieVal[:pipe]
	orgSlug := cookieVal[pipe+1:]

	if r.URL.Query().Get("state") != cookieState {
		writeError(w, http.StatusBadRequest, "invalid state")
		return
	}

	// Clear state cookie.
	http.SetCookie(w, &http.Cookie{Name: ssoStateCookie, Value: "", Path: "/", MaxAge: -1})

	cfg, err := h.ssoConfigs.GetByOrgSlug(r.Context(), orgSlug)
	if err != nil || cfg == nil {
		writeError(w, http.StatusBadRequest, "SSO config not found")
		return
	}

	secret, err := auth.Decrypt(h.encryptKey, cfg.ClientSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	oauth2Cfg, err := h.buildOAuth2Config(r.Context(), cfg, secret, callbackURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	code := r.URL.Query().Get("code")
	token, err := oauth2Cfg.Exchange(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "token exchange failed")
		return
	}

	issuerURL := resolveIssuerURL(cfg)
	provider, err := oidc.NewProvider(r.Context(), issuerURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "OIDC provider error")
		return
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "no id_token in response")
		return
	}
	idToken, err := verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid id_token")
		return
	}

	var claims struct {
		Email  string   `json:"email"`
		Name   string   `json:"name"`
		Groups []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		writeError(w, http.StatusInternalServerError, "claim extraction failed")
		return
	}

	// Map OIDC groups → Omnir role.
	role := mapRole(cfg.AttributeMapping, claims.Groups)

	// JIT provision: find or create user.
	user, _, err := h.users.FindByEmail(r.Context(), claims.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if user == nil {
		org, err := h.orgs.GetBySlug(r.Context(), orgSlug)
		if err != nil || org == nil {
			writeError(w, http.StatusInternalServerError, "org not found")
			return
		}
		ctx := domain.WithOrgID(r.Context(), org.ID)
		name := claims.Name
		if name == "" {
			name = claims.Email
		}
		user, err = h.users.Create(ctx, &domain.User{
			OrgID: org.ID,
			Email: claims.Email,
			Name:  name,
			Role:  domain.UserRole(role),
		}, "") // no password for SSO users
		if err != nil {
			writeError(w, http.StatusInternalServerError, "user provisioning failed")
			return
		}
	}

	jwtClaims := auth.Claims{
		UserID: user.ID,
		OrgID:  user.OrgID,
		Role:   string(user.Role),
	}
	accessToken, err := h.jwtSvc.Issue(jwtClaims, accessTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	refreshToken, err := h.jwtSvc.Issue(jwtClaims, refreshTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	secure := r.TLS != nil
	setAccessCookie(w, accessToken, secure)
	setRefreshCookie(w, refreshToken, secure)

	// Redirect to the SPA SSO landing page.
	http.Redirect(w, r, "/auth/sso/done", http.StatusFound)
}

// GetOrgSSO returns the org's SSO configuration (secret redacted).
// GET /api/v1/orgs/{orgId}/sso
// Requires admin or super_admin role.
func (h *SSOHandler) GetOrgSSO(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !isAdminOrAbove(claims.Role) && claims.OrgID != orgID {
		writeError(w, http.StatusForbidden, "admin role required")
		return
	}

	cfg, err := h.ssoConfigs.GetByOrgID(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if cfg == nil {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

// UpdateOrgSSO saves or updates the org's SSO configuration.
// PATCH /api/v1/orgs/{orgId}/sso
// Requires admin or super_admin role.
func (h *SSOHandler) UpdateOrgSSO(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !isAdminOrAbove(claims.Role) {
		writeError(w, http.StatusForbidden, "admin role required")
		return
	}

	var req struct {
		Provider         string            `json:"provider"`
		ClientID         string            `json:"client_id"`
		ClientSecret     string            `json:"client_secret"`
		TenantID         string            `json:"tenant_id"`
		Hd               string            `json:"hd"`
		AttributeMapping map[string]string `json:"attribute_mapping"`
		Enabled          bool              `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch req.Provider {
	case "google", "oidc", "microsoft":
		// valid
	default:
		writeError(w, http.StatusUnprocessableEntity, "provider must be one of: google, microsoft, oidc")
		return
	}
	if req.ClientID == "" {
		writeError(w, http.StatusUnprocessableEntity, "client_id is required")
		return
	}

	// Derive issuer URL for known providers if not explicitly set.
	issuerURL := deriveIssuerURL(req.Provider, req.TenantID)

	// Encrypt the client secret if provided; otherwise preserve existing.
	encryptedSecret := ""
	if req.ClientSecret != "" {
		encryptedSecret, err = auth.Encrypt(h.encryptKey, req.ClientSecret)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	} else {
		// Preserve existing secret.
		existing, _ := h.ssoConfigs.GetByOrgID(r.Context(), orgID)
		if existing != nil {
			encryptedSecret = existing.ClientSecret
		}
	}

	cfg := &domain.SSOConfig{
		OrgID:            orgID,
		Provider:         req.Provider,
		ClientID:         req.ClientID,
		ClientSecret:     encryptedSecret,
		IssuerURL:        issuerURL,
		TenantID:         req.TenantID,
		Hd:               req.Hd,
		AttributeMapping: req.AttributeMapping,
		Enabled:          req.Enabled,
	}

	saved, err := h.ssoConfigs.Upsert(r.Context(), cfg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// buildOAuth2Config constructs an oauth2.Config for the given SSO config.
func (h *SSOHandler) buildOAuth2Config(ctx context.Context, cfg *domain.SSOConfig, clientSecret, callbackURL string) (*oauth2.Config, error) {
	issuerURL := resolveIssuerURL(cfg)
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC provider: %w", err)
	}
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}, nil
}

// resolveIssuerURL returns the effective OIDC issuer URL for the config.
// For microsoft, it derives from tenant_id if IssuerURL is empty.
// For google, it uses the well-known Google OIDC endpoint.
func resolveIssuerURL(cfg *domain.SSOConfig) string {
	if cfg.IssuerURL != "" {
		return cfg.IssuerURL
	}
	return deriveIssuerURL(cfg.Provider, cfg.TenantID)
}

// deriveIssuerURL returns the standard OIDC issuer URL for known providers.
func deriveIssuerURL(provider, tenantID string) string {
	switch provider {
	case "microsoft":
		tid := tenantID
		if tid == "" {
			tid = "common"
		}
		return fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tid)
	case "google":
		return "https://accounts.google.com"
	default:
		return ""
	}
}

// buildAuthURL generates the OAuth2 authorization URL, adding provider-specific params.
func buildAuthURL(oauth2Cfg *oauth2.Config, cfg *domain.SSOConfig, state string) string {
	opts := []oauth2.AuthCodeOption{}
	if cfg.Provider == "google" && cfg.Hd != "" {
		opts = append(opts, oauth2.SetAuthURLParam("hd", cfg.Hd))
	}
	return oauth2Cfg.AuthCodeURL(state, opts...)
}

func mapRole(mapping map[string]string, groups []string) string {
	for _, g := range groups {
		if r, ok := mapping[g]; ok {
			return r
		}
	}
	if r, ok := mapping["default"]; ok {
		return r
	}
	return string(domain.UserRoleAgent)
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func isAdminOrAbove(role string) bool {
	return strings.EqualFold(role, string(domain.UserRoleAdmin)) ||
		strings.EqualFold(role, string(domain.UserRoleSuperAdmin))
}
