package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

const (
	gmailAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	gmailTokenURL = "https://oauth2.googleapis.com/token" //nolint:gosec // OAuth endpoint, not a credential

	outlookAuthURLFmt  = "https://login.microsoftonline.com/%s/oauth2/v2.0/authorize"
	outlookTokenURLFmt = "https://login.microsoftonline.com/%s/oauth2/v2.0/token" //nolint:gosec // OAuth endpoint, not a credential

	gmailUserInfoURL   = "https://www.googleapis.com/oauth2/v3/userinfo"
	outlookUserInfoURL = "https://graph.microsoft.com/v1.0/me"

	gmailSendURL   = "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"
	outlookSendURL = "https://graph.microsoft.com/v1.0/me/sendMail"
)

// emailInboxOAuthState tracks in-flight OAuth state values.
type emailInboxOAuthState struct {
	userID    uuid.UUID
	orgID     uuid.UUID
	provider  domain.EmailProvider
	expiresAt time.Time
}

// EmailInboxHandler handles OAuth flows, connection management,
// and inbox list/thread/send for Gmail and Outlook.
type EmailInboxHandler struct {
	connections repository.EmailConnectionRepository
	inbox       repository.EmailInboxRepository
	cfg         config.EmailInboxConfig

	mu     sync.Mutex
	states map[string]emailInboxOAuthState
}

func NewEmailInboxHandler(
	connections repository.EmailConnectionRepository,
	inbox repository.EmailInboxRepository,
	cfg config.EmailInboxConfig,
) *EmailInboxHandler {
	return &EmailInboxHandler{
		connections: connections,
		inbox:       inbox,
		cfg:         cfg,
		states:      make(map[string]emailInboxOAuthState),
	}
}

// OAuthRouter returns routes mounted under /api/integrations/email.
// Callers are expected to apply Authenticate + OrgScope middleware before mounting.
func (h *EmailInboxHandler) OAuthRouter() chi.Router {
	r := chi.NewRouter()

	agentOnly := middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent)

	r.With(agentOnly).Get("/auth/google", h.InitiateGoogle)
	r.With(agentOnly).Get("/auth/microsoft", h.InitiateMicrosoft)
	// Callback — provider redirects here; state contains the user identity (CSRF-safe).
	r.Get("/callback", h.Callback)
	r.With(agentOnly).Get("/connections", h.ListConnections)
	r.With(agentOnly).Delete("/connections/{id}", h.Disconnect)

	return r
}

// InboxRouter returns routes to be mounted under /api/v1/emails for inbox features.
func (h *EmailInboxHandler) InboxRouter() chi.Router {
	r := chi.NewRouter()
	h.RegisterInboxRoutes(r)
	return r
}

// RegisterInboxRoutes adds inbox routes to an existing /emails-style router.
func (h *EmailInboxHandler) RegisterInboxRoutes(r chi.Router) {
	agentOnly := middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent)

	r.With(agentOnly).Get("/threads", h.ListThreads)
	r.With(agentOnly).Get("/threads/{threadId}", h.GetThread)
	r.With(agentOnly).Get("/", h.ListInbox)
	r.With(agentOnly).Post("/send", h.SendViaConnection)
	r.With(agentOnly).Patch("/{threadId}/read", h.MarkThreadRead)
}

// ─── OAuth state helpers ──────────────────────────────────────────────────────

func (h *EmailInboxHandler) generateState(userID, orgID uuid.UUID, provider domain.EmailProvider) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	h.mu.Lock()
	h.states[state] = emailInboxOAuthState{
		userID:    userID,
		orgID:     orgID,
		provider:  provider,
		expiresAt: time.Now().Add(10 * time.Minute),
	}
	h.mu.Unlock()
	return state, nil
}

func (h *EmailInboxHandler) consumeState(state string) (emailInboxOAuthState, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.states[state]
	if !ok {
		return emailInboxOAuthState{}, false
	}
	delete(h.states, state)
	if time.Now().After(s.expiresAt) {
		return emailInboxOAuthState{}, false
	}
	return s, true
}

// ─── Google OAuth ─────────────────────────────────────────────────────────────

func (h *EmailInboxHandler) InitiateGoogle(w http.ResponseWriter, r *http.Request) {
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

	state, err := h.generateState(claims.UserID, orgID, domain.EmailProviderGmail)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "failed to generate state token")
		return
	}

	params := url.Values{
		"client_id":     {h.cfg.GoogleClientID},
		"redirect_uri":  {h.cfg.GoogleRedirectURL},
		"response_type": {"code"},
		"scope":         {"https://www.googleapis.com/auth/gmail.modify https://www.googleapis.com/auth/userinfo.email"},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
		"state":         {state},
	}
	http.Redirect(w, r, gmailAuthURL+"?"+params.Encode(), http.StatusFound)
}

// ─── Microsoft OAuth ──────────────────────────────────────────────────────────

func (h *EmailInboxHandler) InitiateMicrosoft(w http.ResponseWriter, r *http.Request) {
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

	state, err := h.generateState(claims.UserID, orgID, domain.EmailProviderOutlook)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "failed to generate state token")
		return
	}

	authURL := fmt.Sprintf(outlookAuthURLFmt, h.cfg.MicrosoftTenantID)
	params := url.Values{
		"client_id":     {h.cfg.MicrosoftClientID},
		"redirect_uri":  {h.cfg.MicrosoftRedirectURL},
		"response_type": {"code"},
		"scope":         {"offline_access Mail.ReadWrite Mail.Send User.Read"},
		"state":         {state},
	}
	http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound)
}

// ─── Callback (shared) ────────────────────────────────────────────────────────

func (h *EmailInboxHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	stateParam := r.URL.Query().Get("state")
	if code == "" || stateParam == "" {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "missing code or state")
		return
	}

	s, ok := h.consumeState(stateParam)
	if !ok {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid or expired state")
		return
	}

	var tokens *emailOAuthTokenResponse
	var emailAddr string
	var err error

	switch s.provider {
	case domain.EmailProviderGmail:
		tokens, err = h.exchangeGoogleCode(code)
		if err != nil {
			writeProblem(w, http.StatusBadGateway, "OAuth Error", err.Error())
			return
		}
		emailAddr, err = h.fetchGoogleEmail(tokens.AccessToken)
	case domain.EmailProviderOutlook:
		tokens, err = h.exchangeMicrosoftCode(code)
		if err != nil {
			writeProblem(w, http.StatusBadGateway, "OAuth Error", err.Error())
			return
		}
		emailAddr, err = h.fetchOutlookEmail(tokens.AccessToken)
	default:
		writeProblem(w, http.StatusBadRequest, "Bad Request", "unknown provider")
		return
	}

	if err != nil {
		writeProblem(w, http.StatusBadGateway, "OAuth Error", "failed to fetch email address: "+err.Error())
		return
	}

	encAccess, err := auth.Encrypt(h.cfg.EncryptionKey, tokens.AccessToken)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", "failed to encrypt access token")
		return
	}

	conn := &domain.EmailConnection{
		OrgID:        s.orgID,
		UserID:       s.userID,
		Provider:     s.provider,
		EmailAddress: emailAddr,
		AccessToken:  encAccess,
		TokenExpiry:  tokenExpiryFromSeconds(tokens.ExpiresIn),
	}
	if tokens.RefreshToken != "" {
		encRefresh, err := auth.Encrypt(h.cfg.EncryptionKey, tokens.RefreshToken)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Internal Error", "failed to encrypt refresh token")
			return
		}
		conn.RefreshToken = &encRefresh
	}

	saved, err := h.connections.Upsert(r.Context(), conn)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	// Don't expose tokens.
	saved.AccessToken = ""
	saved.RefreshToken = nil
	writeJSON(w, http.StatusOK, saved)
}

// ─── Token exchange ───────────────────────────────────────────────────────────

type emailOAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (h *EmailInboxHandler) exchangeGoogleCode(code string) (*emailOAuthTokenResponse, error) {
	resp, err := http.PostForm(gmailTokenURL, url.Values{
		"code":          {code},
		"client_id":     {h.cfg.GoogleClientID},
		"client_secret": {h.cfg.GoogleClientSecret},
		"redirect_uri":  {h.cfg.GoogleRedirectURL},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeEmailTokenResponse(resp)
}

func (h *EmailInboxHandler) exchangeMicrosoftCode(code string) (*emailOAuthTokenResponse, error) {
	tokenURL := fmt.Sprintf(outlookTokenURLFmt, h.cfg.MicrosoftTenantID)
	resp, err := http.PostForm(tokenURL, url.Values{ //#nosec G107 -- tokenURL is constructed from config, not user input
		"code":          {code},
		"client_id":     {h.cfg.MicrosoftClientID},
		"client_secret": {h.cfg.MicrosoftClientSecret},
		"redirect_uri":  {h.cfg.MicrosoftRedirectURL},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeEmailTokenResponse(resp)
}

func decodeEmailTokenResponse(resp *http.Response) (*emailOAuthTokenResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(body))
	}
	var t emailOAuthTokenResponse
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func tokenExpiryFromSeconds(secs int) time.Time {
	if secs <= 0 {
		return time.Now().Add(time.Hour)
	}
	return time.Now().Add(time.Duration(secs) * time.Second)
}

// ─── User info helpers ────────────────────────────────────────────────────────

func (h *EmailInboxHandler) fetchGoogleEmail(accessToken string) (string, error) {
	req, _ := http.NewRequest(http.MethodGet, gmailUserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var info struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	return info.Email, nil
}

func (h *EmailInboxHandler) fetchOutlookEmail(accessToken string) (string, error) {
	req, _ := http.NewRequest(http.MethodGet, outlookUserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var info struct {
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	if info.Mail != "" {
		return info.Mail, nil
	}
	return info.UserPrincipalName, nil
}

// ─── Connection management ────────────────────────────────────────────────────

func (h *EmailInboxHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	conns, err := h.connections.List(r.Context(), domain.EmailConnectionFilter{OrgID: orgID})
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	if conns == nil {
		conns = []*domain.EmailConnection{}
	}
	writeJSON(w, http.StatusOK, conns)
}

func (h *EmailInboxHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid connection id")
		return
	}

	if err := h.connections.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Inbox routes ─────────────────────────────────────────────────────────────

func (h *EmailInboxHandler) ListThreads(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	filter := domain.EmailInboxFilter{OrgID: orgID}
	if s := r.URL.Query().Get("connection_id"); s != "" {
		id, err := uuid.Parse(s)
		if err == nil {
			filter.ConnectionID = &id
		}
	}
	if s := r.URL.Query().Get("contact_id"); s != "" {
		id, err := uuid.Parse(s)
		if err == nil {
			filter.ContactID = &id
		}
	}
	filter.UnreadOnly = r.URL.Query().Get("unread_only") == "true"
	filter.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if filter.Page < 1 {
		filter.Page = 1
	}
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if filter.Limit < 1 || filter.Limit > 200 {
		filter.Limit = 50
	}

	threads, total, err := h.inbox.ListThreads(r.Context(), filter)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, paginated(threads, total, filter.Page, filter.Limit))
}

func (h *EmailInboxHandler) ListInbox(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	filter := domain.EmailInboxFilter{OrgID: orgID}
	if s := r.URL.Query().Get("connection_id"); s != "" {
		id, err := uuid.Parse(s)
		if err == nil {
			filter.ConnectionID = &id
		}
	}
	if s := r.URL.Query().Get("contact_id"); s != "" {
		id, err := uuid.Parse(s)
		if err == nil {
			filter.ContactID = &id
		}
	}
	if s := r.URL.Query().Get("thread_id"); s != "" {
		filter.ThreadID = &s
	}
	filter.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if filter.Page < 1 {
		filter.Page = 1
	}
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if filter.Limit < 1 || filter.Limit > 200 {
		filter.Limit = 50
	}

	msgs, total, err := h.inbox.List(r.Context(), filter)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	if msgs == nil {
		msgs = []*domain.EmailInboxMessage{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": msgs, "total": total})
}

func (h *EmailInboxHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	threadID := chi.URLParam(r, "threadId")
	if threadID == "" {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "threadId is required")
		return
	}

	msgs, err := h.inbox.GetThread(r.Context(), orgID, threadID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if msgs == nil {
		msgs = []*domain.EmailInboxMessage{}
	}
	summary, err := h.threadSummaryFromMessages(orgID, msgs)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, domain.EmailInboxThreadDetail{
		ThreadSummary: summary,
		Messages:      msgs,
	})
}

// MarkThreadRead marks all messages in a thread as read.
// PATCH /api/v1/emails/{threadId}/read
func (h *EmailInboxHandler) MarkThreadRead(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	threadID := chi.URLParam(r, "threadId")
	if threadID == "" {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "threadId is required")
		return
	}

	if err := h.inbox.MarkThreadRead(r.Context(), orgID, threadID); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *EmailInboxHandler) SendViaConnection(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	var req domain.SendInboxEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if err := req.Validate(); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}

	conn, err := h.connections.GetByID(r.Context(), req.ConnectionID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	accessToken, mockToken, err := h.resolveConnectionAccessToken(r.Context(), conn)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "email connection token is invalid or expired; reconnect this inbox account")
		return
	}

	var sendErr error
	bodyText := htmlStripper.ReplaceAllString(req.BodyHTML, "")
	toCSV := strings.Join(req.To, ", ")
	if !mockToken {
		switch conn.Provider {
		case domain.EmailProviderGmail:
			sendErr = h.sendViaGmail(accessToken, conn.EmailAddress, req.To, req.CC, req.BCC, req.Subject, bodyText, req.ThreadID)
		case domain.EmailProviderOutlook:
			sendErr = h.sendViaOutlook(accessToken, req.To, req.CC, req.BCC, req.Subject, req.BodyHTML)
		default:
			writeProblem(w, http.StatusBadRequest, "Bad Request", "unsupported provider")
			return
		}
	}

	if sendErr != nil {
		writeProblem(w, http.StatusBadGateway, "Send Error", sendErr.Error())
		return
	}

	// Persist a record of the sent message.
	threadID := uuid.New().String()
	if req.ThreadID != nil && *req.ThreadID != "" {
		threadID = *req.ThreadID
	}
	msg := &domain.EmailInboxMessage{
		OrgID:        orgID,
		ConnectionID: conn.ID,
		MessageID:    uuid.New().String(),
		ThreadID:     threadID,
		FromAddr:     conn.EmailAddress,
		ToAddrs:      req.To,
		Subject:      req.Subject,
		Direction:    domain.EmailDirectionOutbound,
		ContactID:    req.ContactID,
		SentAt:       time.Now().UTC(),
	}
	msg.BodyText = &bodyText
	msg.BodyHTML = &req.BodyHTML

	saved, err := h.inbox.Upsert(r.Context(), msg)
	if err != nil {
		// Non-fatal: message was sent, just record-keeping failed.
		_ = err
		writeJSON(w, http.StatusCreated, msg)
		return
	}
	if saved != nil && len(saved.ToAddrs) == 0 && toCSV != "" {
		saved.ToAddrs = req.To
	}
	writeJSON(w, http.StatusCreated, saved)
}

func (h *EmailInboxHandler) resolveConnectionAccessToken(ctx context.Context, conn *domain.EmailConnection) (string, bool, error) {
	accessToken, err := auth.Decrypt(h.cfg.EncryptionKey, conn.AccessToken)
	if err == nil {
		return accessToken, strings.HasPrefix(accessToken, "mock-"), nil
	}

	// Backward-compatibility path:
	// older/staging records may still hold plaintext tokens (including mock seed
	// tokens). Accept them and opportunistically re-encrypt.
	if !looksLikeLegacyPlaintextToken(conn.AccessToken) {
		return "", false, err
	}
	legacyToken := strings.TrimSpace(conn.AccessToken)
	if enc, encErr := auth.Encrypt(h.cfg.EncryptionKey, legacyToken); encErr == nil {
		_, _ = h.connections.Update(ctx, conn.ID, domain.EmailConnectionPatch{AccessToken: &enc})
	}
	return legacyToken, strings.HasPrefix(legacyToken, "mock-"), nil
}

func looksLikeLegacyPlaintextToken(token string) bool {
	t := strings.TrimSpace(token)
	if t == "" {
		return false
	}
	return strings.HasPrefix(t, "mock-") ||
		strings.HasPrefix(t, "ya29.") || // Google OAuth token prefix
		strings.HasPrefix(t, "eyJ") || // JWT style token
		strings.Count(t, ".") >= 2 // generic JWT shape
}

// ─── Provider send helpers ────────────────────────────────────────────────────

var htmlStripper = regexp.MustCompile(`<[^>]+>`)

func (h *EmailInboxHandler) sendViaGmail(accessToken, from string, to, cc, bcc []string, subject, body string, threadID *string) error {
	// Build a minimal RFC2822 message.
	var raw strings.Builder
	raw.WriteString("From: " + from + "\r\n")
	raw.WriteString("To: " + strings.Join(to, ", ") + "\r\n")
	if len(cc) > 0 {
		raw.WriteString("Cc: " + strings.Join(cc, ", ") + "\r\n")
	}
	if len(bcc) > 0 {
		raw.WriteString("Bcc: " + strings.Join(bcc, ", ") + "\r\n")
	}
	raw.WriteString("Subject: " + subject + "\r\n")
	raw.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	raw.WriteString("\r\n")
	raw.WriteString(body)

	encoded := base64.URLEncoding.EncodeToString([]byte(raw.String()))
	payload := map[string]any{"raw": encoded}
	if threadID != nil && *threadID != "" {
		payload["threadId"] = *threadID
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, gmailSendURL, strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gmail send failed (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

func (h *EmailInboxHandler) sendViaOutlook(accessToken string, to, cc, bcc []string, subject, bodyHTML string) error {
	toRecipients := make([]map[string]any, 0, len(to))
	for _, addr := range to {
		toRecipients = append(toRecipients, map[string]any{"emailAddress": map[string]string{"address": addr}})
	}
	ccRecipients := make([]map[string]any, 0, len(cc))
	for _, addr := range cc {
		ccRecipients = append(ccRecipients, map[string]any{"emailAddress": map[string]string{"address": addr}})
	}
	bccRecipients := make([]map[string]any, 0, len(bcc))
	for _, addr := range bcc {
		bccRecipients = append(bccRecipients, map[string]any{"emailAddress": map[string]string{"address": addr}})
	}
	payload := map[string]any{
		"message": map[string]any{
			"subject": subject,
			"body": map[string]string{
				"contentType": "HTML",
				"content":     bodyHTML,
			},
			"toRecipients": toRecipients,
		},
		"saveToSentItems": "true",
	}
	if len(ccRecipients) > 0 {
		payload["message"].(map[string]any)["ccRecipients"] = ccRecipients
	}
	if len(bccRecipients) > 0 {
		payload["message"].(map[string]any)["bccRecipients"] = bccRecipients
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, outlookSendURL, strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("outlook send failed (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

func (h *EmailInboxHandler) threadSummaryFromMessages(orgID uuid.UUID, msgs []*domain.EmailInboxMessage) (domain.EmailInboxThreadSummary, error) {
	if len(msgs) == 0 {
		return domain.EmailInboxThreadSummary{}, domain.ErrNotFound
	}

	last := msgs[len(msgs)-1]
	participantsMap := make(map[string]struct{})
	unread := false
	contactID := last.ContactID
	for _, msg := range msgs {
		if msg.FromAddr != "" {
			participantsMap[msg.FromAddr] = struct{}{}
		}
		for _, addr := range msg.ToAddrs {
			participantsMap[addr] = struct{}{}
		}
		if msg.Direction == domain.EmailDirectionInbound && msg.ReadAt == nil {
			unread = true
		}
		if contactID == nil && msg.ContactID != nil {
			contactID = msg.ContactID
		}
	}

	participants := make([]string, 0, len(participantsMap))
	for participant := range participantsMap {
		participants = append(participants, participant)
	}

	snippet := ""
	if last.BodyText != nil {
		snippet = *last.BodyText
	}
	if len(snippet) > 140 {
		snippet = snippet[:140]
	}

	return domain.EmailInboxThreadSummary{
		ThreadID:      last.ThreadID,
		OrgID:         orgID,
		ConnectionID:  last.ConnectionID,
		Subject:       last.Subject,
		Participants:  participants,
		Snippet:       snippet,
		Unread:        unread,
		MessageCount:  len(msgs),
		LastMessageAt: last.SentAt,
		ContactID:     contactID,
	}, nil
}
