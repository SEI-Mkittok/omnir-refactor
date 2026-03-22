package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

const ssoStateCookie = "sso_state"

// SSOHandler handles OIDC SSO login/callback.
type SSOHandler struct {
	ssoConfigs  repository.SSOConfigRepository
	orgs        repository.OrgRepository
	users       repository.UserRepository
	jwtSvc      *auth.JWTService
	encryptKey  string
	callbackURL string
}

func NewSSOHandler(
	ssoConfigs repository.SSOConfigRepository,
	orgs repository.OrgRepository,
	users repository.UserRepository,
	jwtSvc *auth.JWTService,
	encryptKey, callbackURL string,
) *SSOHandler {
	return &SSOHandler{
		ssoConfigs:  ssoConfigs,
		orgs:        orgs,
		users:       users,
		jwtSvc:      jwtSvc,
		encryptKey:  encryptKey,
		callbackURL: callbackURL,
	}
}

func (h *SSOHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/config", h.Config)
	r.Get("/{orgSlug}/login", h.Login)
	r.Get("/callback", h.Callback)
	return r
}

// Config returns the org's SSO provider info without requiring authentication.
// GET /auth/sso/config
func (h *SSOHandler) Config(w http.ResponseWriter, r *http.Request) {
	// Resolve the org slug from query param or fall back to any configured org.
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

	oauth2Cfg, err := h.buildOAuth2Config(r.Context(), cfg, secret)
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
	http.Redirect(w, r, oauth2Cfg.AuthCodeURL(state), http.StatusFound)
}

// Callback handles the OIDC redirect, provisions user, and issues JWTs.
// GET /auth/sso/callback
func (h *SSOHandler) Callback(w http.ResponseWriter, r *http.Request) {
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

	oauth2Cfg, err := h.buildOAuth2Config(r.Context(), cfg, secret)
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

	provider, err := oidc.NewProvider(r.Context(), cfg.IssuerURL)
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

	// Redirect to the SPA SSO landing page which will fetch the current user and
	// navigate to the dashboard.
	http.Redirect(w, r, "/auth/sso/done", http.StatusFound)
}

func (h *SSOHandler) buildOAuth2Config(ctx context.Context, cfg *domain.SSOConfig, clientSecret string) (*oauth2.Config, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC provider: %w", err)
	}
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: clientSecret,
		RedirectURL:  h.callbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}, nil
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
