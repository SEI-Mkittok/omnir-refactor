package worker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

const (
	gmailListURL    = "https://gmail.googleapis.com/gmail/v1/users/me/messages"
	gmailGetURL     = "https://gmail.googleapis.com/gmail/v1/users/me/messages/%s"
	gmailRefreshURL = "https://oauth2.googleapis.com/token" //nolint:gosec // OAuth endpoint, not a credential

	outlookMailURL    = "https://graph.microsoft.com/v1.0/me/mailFolders/inbox/messages"
	outlookRefreshURL = "https://login.microsoftonline.com/%s/oauth2/v2.0/token" //nolint:gosec // OAuth endpoint, not a credential
)

// EmailInboxSyncWorker polls all email connections every interval,
// fetches new messages, and links them to contacts by address.
type EmailInboxSyncWorker struct {
	connections repository.EmailConnectionRepository
	inbox       repository.EmailInboxRepository
	contacts    repository.ContactRepository
	interval    time.Duration
	log         *slog.Logger
	encKey      string

	// OAuth credentials for token refresh.
	googleClientID        string
	googleClientSecret    string
	microsoftClientID     string
	microsoftClientSecret string
	microsoftTenantID     string
}

// NewEmailInboxSyncWorker creates a new EmailInboxSyncWorker.
func NewEmailInboxSyncWorker(
	connections repository.EmailConnectionRepository,
	inbox repository.EmailInboxRepository,
	contacts repository.ContactRepository,
	interval time.Duration,
	log *slog.Logger,
	encKey string,
	googleClientID, googleClientSecret string,
	microsoftClientID, microsoftClientSecret, microsoftTenantID string,
) *EmailInboxSyncWorker {
	return &EmailInboxSyncWorker{
		connections:           connections,
		inbox:                 inbox,
		contacts:              contacts,
		interval:              interval,
		log:                   log,
		encKey:                encKey,
		googleClientID:        googleClientID,
		googleClientSecret:    googleClientSecret,
		microsoftClientID:     microsoftClientID,
		microsoftClientSecret: microsoftClientSecret,
		microsoftTenantID:     microsoftTenantID,
	}
}

// Start launches the sync loop in a background goroutine.
func (w *EmailInboxSyncWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		w.log.Info("email inbox sync worker started", "interval", w.interval)
		for {
			select {
			case <-ticker.C:
				w.runSync(ctx)
			case <-ctx.Done():
				w.log.Info("email inbox sync worker stopped")
				return
			}
		}
	}()
}

func (w *EmailInboxSyncWorker) runSync(ctx context.Context) {
	conns, err := w.connections.ListAllActive(ctx)
	if err != nil {
		w.log.Error("email inbox sync: failed to list connections", "err", err)
		return
	}
	for _, conn := range conns {
		if err := w.syncConnection(ctx, conn); err != nil {
			w.log.Warn("email inbox sync: connection failed",
				"connection_id", conn.ID,
				"provider", conn.Provider,
				"user_id", conn.UserID,
				"err", err,
			)
		}
	}
}

func (w *EmailInboxSyncWorker) syncConnection(ctx context.Context, conn *domain.EmailConnection) error {
	// Refresh token if it expires within 5 minutes.
	if time.Until(conn.TokenExpiry) < 5*time.Minute {
		if err := w.refreshToken(ctx, conn); err != nil {
			return fmt.Errorf("token refresh: %w", err)
		}
	}

	accessToken, err := auth.Decrypt(w.encKey, conn.AccessToken)
	if err != nil {
		return fmt.Errorf("decrypt access token: %w", err)
	}

	var msgs []inboxMessage
	var newCursor string

	switch conn.Provider {
	case domain.EmailProviderGmail:
		msgs, newCursor, err = w.fetchGmailMessages(accessToken, conn.SyncCursor)
	case domain.EmailProviderOutlook:
		msgs, newCursor, err = w.fetchOutlookMessages(accessToken, conn.SyncCursor)
	default:
		return fmt.Errorf("unknown provider: %s", conn.Provider)
	}
	if err != nil {
		return fmt.Errorf("fetch messages: %w", err)
	}

	for _, m := range msgs {
		msg := &domain.EmailInboxMessage{
			OrgID:        conn.OrgID,
			ConnectionID: conn.ID,
			MessageID:    m.MessageID,
			ThreadID:     m.ThreadID,
			FromAddr:     m.From,
			ToAddrs:      m.To,
			Subject:      m.Subject,
			Direction:    domain.EmailDirectionInbound,
			SentAt:       m.Date,
		}
		if m.BodyText != "" {
			msg.BodyText = &m.BodyText
		}
		if m.BodyHTML != "" {
			msg.BodyHTML = &m.BodyHTML
		}

		saved, err := w.inbox.Upsert(ctx, msg)
		if err != nil {
			w.log.Warn("email inbox sync: failed to upsert message",
				"message_id", m.MessageID,
				"err", err,
			)
			continue
		}

		// Try to link to a contact by from_addr.
		if saved.ContactID == nil {
			w.tryLinkContact(ctx, conn.OrgID, m.From, saved.ID)
		}
	}

	// Update cursor and last_synced_at.
	now := time.Now().UTC()
	patch := domain.EmailConnectionPatch{LastSyncedAt: &now}
	if newCursor != "" {
		patch.SyncCursor = &newCursor
	}
	if _, err := w.connections.Update(ctx, conn.ID, patch); err != nil {
		w.log.Warn("email inbox sync: failed to update cursor", "connection_id", conn.ID, "err", err)
	}
	return nil
}

// tryLinkContact looks up a contact by email address and links it to messages.
func (w *EmailInboxSyncWorker) tryLinkContact(ctx context.Context, orgID uuid.UUID, addr string, _ uuid.UUID) {
	contacts, _, err := w.contacts.List(ctx, domain.ContactFilter{OrgID: orgID})
	if err != nil {
		return
	}
	for _, c := range contacts {
		if c.Email != nil && strings.EqualFold(*c.Email, addr) {
			_ = w.inbox.LinkContact(ctx, orgID, addr, c.ID)
			return
		}
	}
}

// ─── Token refresh ────────────────────────────────────────────────────────────

func (w *EmailInboxSyncWorker) refreshToken(ctx context.Context, conn *domain.EmailConnection) error {
	if conn.RefreshToken == nil {
		return fmt.Errorf("no refresh token available for connection %s", conn.ID)
	}
	refreshToken, err := auth.Decrypt(w.encKey, *conn.RefreshToken)
	if err != nil {
		return fmt.Errorf("decrypt refresh token: %w", err)
	}

	var vals url.Values
	var tokenURL string

	switch conn.Provider {
	case domain.EmailProviderGmail:
		tokenURL = gmailRefreshURL
		vals = url.Values{
			"client_id":     {w.googleClientID},
			"client_secret": {w.googleClientSecret},
			"refresh_token": {refreshToken},
			"grant_type":    {"refresh_token"},
		}
	case domain.EmailProviderOutlook:
		tokenURL = fmt.Sprintf(outlookRefreshURL, w.microsoftTenantID)
		vals = url.Values{
			"client_id":     {w.microsoftClientID},
			"client_secret": {w.microsoftClientSecret},
			"refresh_token": {refreshToken},
			"grant_type":    {"refresh_token"},
		}
	default:
		return fmt.Errorf("unknown provider: %s", conn.Provider)
	}

	resp, err := http.PostForm(tokenURL, vals) //#nosec G107 -- tokenURL is constructed from config, not user input
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var t struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &t); err != nil {
		return err
	}

	encAccess, err := auth.Encrypt(w.encKey, t.AccessToken)
	if err != nil {
		return err
	}
	expiry := time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
	patch := domain.EmailConnectionPatch{
		AccessToken: &encAccess,
		TokenExpiry: &expiry,
	}
	conn.AccessToken = encAccess
	conn.TokenExpiry = expiry

	if t.RefreshToken != "" {
		encRefresh, err := auth.Encrypt(w.encKey, t.RefreshToken)
		if err != nil {
			return err
		}
		patch.RefreshToken = &encRefresh
		conn.RefreshToken = &encRefresh
	}

	_, err = w.connections.Update(ctx, conn.ID, patch)
	return err
}

// ─── Message fetch helpers ────────────────────────────────────────────────────

type inboxMessage struct {
	MessageID string
	ThreadID  string
	From      string
	To        []string
	Subject   string
	BodyText  string
	BodyHTML  string
	Date      time.Time
}

// ─── Gmail ────────────────────────────────────────────────────────────────────

func (w *EmailInboxSyncWorker) fetchGmailMessages(accessToken string, cursor *string) ([]inboxMessage, string, error) {
	params := url.Values{"maxResults": {"50"}, "labelIds": {"INBOX"}}
	if cursor != nil && *cursor != "" {
		params.Set("pageToken", *cursor)
	}

	req, _ := http.NewRequest(http.MethodGet, gmailListURL+"?"+params.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("gmail list failed (%d): %s", resp.StatusCode, string(b))
	}

	var list struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
		NextPageToken string `json:"nextPageToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, "", err
	}

	var out []inboxMessage
	for _, m := range list.Messages {
		msg, err := w.fetchGmailMessage(accessToken, m.ID)
		if err != nil {
			w.log.Warn("email inbox sync: failed to fetch gmail message", "id", m.ID, "err", err)
			continue
		}
		out = append(out, *msg)
	}
	return out, list.NextPageToken, nil
}

func (w *EmailInboxSyncWorker) fetchGmailMessage(accessToken, id string) (*inboxMessage, error) {
	reqURL := fmt.Sprintf(gmailGetURL, id) + "?format=full"
	req, _ := http.NewRequest(http.MethodGet, reqURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		ID       string `json:"id"`
		ThreadID string `json:"threadId"`
		Payload  struct {
			Headers []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"headers"`
			Parts []struct {
				MimeType string `json:"mimeType"`
				Body     struct {
					Data string `json:"data"`
				} `json:"body"`
			} `json:"parts"`
			Body struct {
				Data string `json:"data"`
			} `json:"body"`
			MimeType string `json:"mimeType"`
		} `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	msg := &inboxMessage{
		MessageID: raw.ID,
		ThreadID:  raw.ThreadID,
		Date:      time.Now().UTC(),
	}
	for _, h := range raw.Payload.Headers {
		switch strings.ToLower(h.Name) {
		case "from":
			msg.From = h.Value
		case "to":
			msg.To = splitAddrs(h.Value)
		case "subject":
			msg.Subject = h.Value
		case "date":
			if t, err := parseRFC2822Date(h.Value); err == nil {
				msg.Date = t
			}
		}
	}

	// Extract body from parts or direct body.
	for _, part := range raw.Payload.Parts {
		decoded, _ := base64.URLEncoding.DecodeString(part.Body.Data)
		switch part.MimeType {
		case "text/plain":
			msg.BodyText = string(decoded)
		case "text/html":
			msg.BodyHTML = string(decoded)
		}
	}
	if msg.BodyText == "" && raw.Payload.MimeType == "text/plain" {
		decoded, _ := base64.URLEncoding.DecodeString(raw.Payload.Body.Data)
		msg.BodyText = string(decoded)
	}
	return msg, nil
}

// ─── Outlook ──────────────────────────────────────────────────────────────────

func (w *EmailInboxSyncWorker) fetchOutlookMessages(accessToken string, cursor *string) ([]inboxMessage, string, error) {
	reqURL := outlookMailURL + "?$top=50&$orderby=receivedDateTime desc"
	if cursor != nil && *cursor != "" {
		// Outlook uses $skipToken embedded in the @odata.nextLink URL.
		reqURL = *cursor
	}

	req, _ := http.NewRequest(http.MethodGet, reqURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("outlook list failed (%d): %s", resp.StatusCode, string(b))
	}

	var list struct {
		Value []struct {
			ID               string `json:"id"`
			ConversationID   string `json:"conversationId"`
			Subject          string `json:"subject"`
			ReceivedDateTime string `json:"receivedDateTime"`
			From             struct {
				EmailAddress struct {
					Address string `json:"address"`
				} `json:"emailAddress"`
			} `json:"from"`
			ToRecipients []struct {
				EmailAddress struct {
					Address string `json:"address"`
				} `json:"emailAddress"`
			} `json:"toRecipients"`
			Body struct {
				ContentType string `json:"contentType"`
				Content     string `json:"content"`
			} `json:"body"`
		} `json:"value"`
		NextLink string `json:"@odata.nextLink"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, "", err
	}

	var out []inboxMessage
	for _, m := range list.Value {
		msg := inboxMessage{
			MessageID: m.ID,
			ThreadID:  m.ConversationID,
			From:      m.From.EmailAddress.Address,
			Subject:   m.Subject,
			Date:      time.Now().UTC(),
		}
		if t, err := time.Parse(time.RFC3339, m.ReceivedDateTime); err == nil {
			msg.Date = t
		}
		for _, r := range m.ToRecipients {
			msg.To = append(msg.To, r.EmailAddress.Address)
		}
		switch strings.ToLower(m.Body.ContentType) {
		case "html":
			msg.BodyHTML = m.Body.Content
		default:
			msg.BodyText = m.Body.Content
		}
		out = append(out, msg)
	}
	return out, list.NextLink, nil
}

// ─── Utility ──────────────────────────────────────────────────────────────────

func splitAddrs(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parseRFC2822Date(s string) (time.Time, error) {
	layouts := []string{
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2 Jan 2006 15:04:05 -0700",
		time.RFC1123Z,
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable date: %s", s)
}
