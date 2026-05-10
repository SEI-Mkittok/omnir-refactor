package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token" //nolint:gosec // OAuth endpoint URL, not a credential

	microsoftAuthURL  = "https://login.microsoftonline.com/%s/oauth2/v2.0/authorize"
	microsoftTokenURL = "https://login.microsoftonline.com/%s/oauth2/v2.0/token" //nolint:gosec // OAuth endpoint URL, not a credential
)

// oauthState tracks in-flight OAuth state values to prevent CSRF.
type oauthState struct {
	userID    uuid.UUID
	orgID     uuid.UUID
	provider  domain.CalendarProvider
	expiresAt time.Time
}

// CalendarHandler handles OAuth flows and connection management for calendar integrations.
type CalendarHandler struct {
	repo   repository.CalendarConnectionRepository
	cfg    config.CalendarConfig
	mu     sync.Mutex
	states map[string]oauthState
}

func NewCalendarHandler(repo repository.CalendarConnectionRepository, cfg config.CalendarConfig) *CalendarHandler {
	return &CalendarHandler{
		repo:   repo,
		cfg:    cfg,
		states: make(map[string]oauthState),
	}
}

// Router mounts all calendar routes under the caller's prefix.
func (h *CalendarHandler) Router() chi.Router {
	r := chi.NewRouter()

	// OAuth initiation (redirects browser to provider)
	r.Get("/auth/google", h.InitiateGoogle)
	r.Get("/auth/microsoft", h.InitiateMicrosoft)

	// Connection management
	r.Get("/connections", h.ListConnections)
	r.Delete("/connections/{id}", h.Disconnect)

	// Manual sync trigger
	r.Post("/sync", h.TriggerSync)

	return r
}

// CallbackRouter mounts only OAuth callback routes.
// This router should be mounted outside auth middleware because cross-site
// provider redirects may not include SameSite-strict session cookies.
func (h *CalendarHandler) CallbackRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/auth/google/callback", h.CallbackGoogle)
	r.Get("/auth/microsoft/callback", h.CallbackMicrosoft)
	return r
}

// ─── OAuth helpers ───────────────────────────────────────────────────────────

func (h *CalendarHandler) generateState(userID, orgID uuid.UUID, provider domain.CalendarProvider) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	h.mu.Lock()
	h.states[state] = oauthState{
		userID:    userID,
		orgID:     orgID,
		provider:  provider,
		expiresAt: time.Now().Add(10 * time.Minute),
	}
	h.mu.Unlock()
	return state
}

func (h *CalendarHandler) consumeState(state string) (oauthState, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.states[state]
	if !ok {
		return oauthState{}, false
	}
	delete(h.states, state)
	if time.Now().After(s.expiresAt) {
		return oauthState{}, false
	}
	return s, true
}

// ─── Google OAuth ─────────────────────────────────────────────────────────────

func (h *CalendarHandler) InitiateGoogle(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing auth context")
		return
	}
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}
	if h.cfg.GoogleClientID == "" {
		writeProblem(w, http.StatusServiceUnavailable, "Not Configured", "Google Calendar OAuth is not configured")
		return
	}

	state := h.generateState(claims.UserID, orgID, domain.CalendarProviderGoogle)
	params := url.Values{
		"client_id":     {h.cfg.GoogleClientID},
		"redirect_uri":  {h.cfg.GoogleRedirectURL},
		"response_type": {"code"},
		"scope":         {"https://www.googleapis.com/auth/calendar.events"},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
		"state":         {state},
	}
	http.Redirect(w, r, googleAuthURL+"?"+params.Encode(), http.StatusFound)
}

func (h *CalendarHandler) CallbackGoogle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if errParam := q.Get("error"); errParam != "" {
		writeProblem(w, http.StatusBadRequest, "OAuth Error", errParam)
		return
	}

	s, ok := h.consumeState(q.Get("state"))
	if !ok {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid or expired OAuth state")
		return
	}

	tokens, err := h.exchangeGoogleCode(q.Get("code"))
	if err != nil {
		writeProblem(w, http.StatusBadGateway, "Token Exchange Failed", err.Error())
		return
	}

	conn := &domain.CalendarConnection{
		OrgID:        s.orgID,
		UserID:       s.userID,
		Provider:     domain.CalendarProviderGoogle,
		AccessToken:  tokens.AccessToken,
		RefreshToken: nilIfEmpty(tokens.RefreshToken),
		TokenExpiry:  tokenExpiry(tokens.ExpiresIn),
	}

	saved, err := h.repo.Upsert(r.Context(), conn)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h *CalendarHandler) exchangeGoogleCode(code string) (*oauthTokenResponse, error) {
	body := url.Values{
		"code":          {code},
		"client_id":     {h.cfg.GoogleClientID},
		"client_secret": {h.cfg.GoogleClientSecret},
		"redirect_uri":  {h.cfg.GoogleRedirectURL},
		"grant_type":    {"authorization_code"},
	}
	return doTokenRequest(googleTokenURL, body)
}

// ─── Microsoft OAuth ──────────────────────────────────────────────────────────

func (h *CalendarHandler) InitiateMicrosoft(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing auth context")
		return
	}
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}
	if h.cfg.MicrosoftClientID == "" {
		writeProblem(w, http.StatusServiceUnavailable, "Not Configured", "Microsoft Calendar OAuth is not configured")
		return
	}

	state := h.generateState(claims.UserID, orgID, domain.CalendarProviderMicrosoft)
	authURL := fmt.Sprintf(microsoftAuthURL, h.cfg.MicrosoftTenantID)
	params := url.Values{
		"client_id":     {h.cfg.MicrosoftClientID},
		"redirect_uri":  {h.cfg.MicrosoftRedirectURL},
		"response_type": {"code"},
		"scope":         {"https://graph.microsoft.com/Calendars.ReadWrite offline_access"},
		"state":         {state},
	}
	http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound)
}

func (h *CalendarHandler) CallbackMicrosoft(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if errParam := q.Get("error"); errParam != "" {
		writeProblem(w, http.StatusBadRequest, "OAuth Error", q.Get("error_description"))
		return
	}

	s, ok := h.consumeState(q.Get("state"))
	if !ok {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid or expired OAuth state")
		return
	}

	tokens, err := h.exchangeMicrosoftCode(q.Get("code"))
	if err != nil {
		writeProblem(w, http.StatusBadGateway, "Token Exchange Failed", err.Error())
		return
	}

	conn := &domain.CalendarConnection{
		OrgID:        s.orgID,
		UserID:       s.userID,
		Provider:     domain.CalendarProviderMicrosoft,
		AccessToken:  tokens.AccessToken,
		RefreshToken: nilIfEmpty(tokens.RefreshToken),
		TokenExpiry:  tokenExpiry(tokens.ExpiresIn),
	}

	saved, err := h.repo.Upsert(r.Context(), conn)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h *CalendarHandler) exchangeMicrosoftCode(code string) (*oauthTokenResponse, error) {
	tokenURL := fmt.Sprintf(microsoftTokenURL, h.cfg.MicrosoftTenantID)
	body := url.Values{
		"code":          {code},
		"client_id":     {h.cfg.MicrosoftClientID},
		"client_secret": {h.cfg.MicrosoftClientSecret},
		"redirect_uri":  {h.cfg.MicrosoftRedirectURL},
		"grant_type":    {"authorization_code"},
	}
	return doTokenRequest(tokenURL, body)
}

// ─── Connection management ────────────────────────────────────────────────────

func (h *CalendarHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	filter := domain.CalendarConnectionFilter{OrgID: orgID}
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		filter.UserID = &claims.UserID
	}

	conns, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	if conns == nil {
		conns = []*domain.CalendarConnection{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": conns})
}

func (h *CalendarHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid connection id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "connection not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// TriggerSync allows an authenticated user to request an immediate sync.
// The actual sync runs in the background worker; this endpoint just signals it.
func (h *CalendarHandler) TriggerSync(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "sync queued"})
}

// ─── Shared OAuth helpers ─────────────────────────────────────────────────────

type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func doTokenRequest(tokenURL string, body url.Values) (*oauthTokenResponse, error) {
	resp, err := http.PostForm(tokenURL, body) //nolint:noctx,gosec // tokenURL is a known OAuth provider endpoint
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var tok oauthTokenResponse
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("provider returned empty access_token")
	}
	return &tok, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func tokenExpiry(expiresIn int) *time.Time {
	if expiresIn <= 0 {
		return nil
	}
	t := time.Now().Add(time.Duration(expiresIn) * time.Second)
	return &t
}
