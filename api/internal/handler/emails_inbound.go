package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// InboundEmailHandler handles inbound email webhooks from SendGrid/Mailgun and
// stores parsed emails in the contact_emails table.
// Mounted unauthenticated under /api/emails/inbound — protected by a shared
// webhook secret verified per provider.
type InboundEmailHandler struct {
	emails     repository.EmailRepository
	contacts   repository.ContactRepository
	activities repository.ActivityRepository
	users      repository.UserRepository
	secret     string
	orgMode    config.OrgMode
}

func NewInboundEmailHandler(
	emails repository.EmailRepository,
	contacts repository.ContactRepository,
	activities repository.ActivityRepository,
	users repository.UserRepository,
	secret string,
	orgMode config.OrgMode,
) *InboundEmailHandler {
	return &InboundEmailHandler{emails: emails, contacts: contacts, activities: activities, users: users, secret: secret, orgMode: orgMode}
}

// Router returns routes for provider-specific inbound webhook paths.
//
//	POST /sendgrid — SendGrid inbound parse
//	POST /mailgun  — Mailgun inbound parse (form)
func (h *InboundEmailHandler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /sendgrid", h.HandleSendGrid)
	mux.HandleFunc("POST /mailgun", h.HandleMailgun)
	return mux
}

// HandleSendGrid handles the SendGrid Inbound Parse webhook.
// Payload is multipart/form-data with fields: from, to, subject, text, html, headers, Message-Id.
// Authenticated via X-Webhook-Secret header when secret is set.
func (h *InboundEmailHandler) HandleSendGrid(w http.ResponseWriter, r *http.Request) {
	if h.secret != "" {
		if r.Header.Get("X-Webhook-Secret") != h.secret {
			writeError(w, http.StatusUnauthorized, "invalid webhook secret")
			return
		}
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	msgID := firstNonEmpty(r.FormValue("Message-Id"), r.FormValue("message-id"))
	inReplyTo := extractInReplyTo(r.FormValue("headers"))

	parsed := &parsedInbound{
		MessageID: strings.TrimSpace(msgID),
		From:      r.FormValue("from"),
		To:        r.FormValue("to"),
		Subject:   r.FormValue("subject"),
		Body:      firstNonEmpty(r.FormValue("text"), r.FormValue("html")),
		InReplyTo: inReplyTo,
	}

	ctx := h.orgContext(r)
	if err := h.store(ctx, parsed); err != nil {
		writeError(w, http.StatusInternalServerError, "ingestion error: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleMailgun handles the Mailgun inbound parse webhook.
// Authenticated via HMAC-SHA256 signature (timestamp + token).
func (h *InboundEmailHandler) HandleMailgun(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	if h.secret != "" {
		timestamp := r.FormValue("timestamp")
		token := r.FormValue("token")
		signature := r.FormValue("signature")
		if !verifyInboundMailgunSignature(h.secret, timestamp, token, signature) {
			writeError(w, http.StatusUnauthorized, "invalid webhook signature")
			return
		}
	}

	parsed := &parsedInbound{
		MessageID: strings.TrimSpace(r.FormValue("Message-Id")),
		From:      r.FormValue("sender"),
		To:        r.FormValue("recipient"),
		Subject:   r.FormValue("subject"),
		Body:      firstNonEmpty(r.FormValue("body-plain"), r.FormValue("body-html")),
		InReplyTo: strings.TrimSpace(r.FormValue("In-Reply-To")),
	}

	ctx := h.orgContext(r)
	if err := h.store(ctx, parsed); err != nil {
		writeError(w, http.StatusInternalServerError, "ingestion error: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// store resolves the contact, threads the email, and persists it.
func (h *InboundEmailHandler) store(ctx context.Context, e *parsedInbound) error {
	subject := strings.TrimSpace(e.Subject)
	if subject == "" {
		subject = "(no subject)"
	}

	// Resolve contact from the From address.
	var contactID *uuid.UUID
	fromEmail := inboundParseAddr(e.From)
	if fromEmail != "" {
		c, err := h.contacts.GetByEmail(ctx, fromEmail)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("contact lookup: %w", err)
		}
		if c != nil {
			contactID = &c.ID
		}
		// No match → contactID stays nil (stored unassigned, flagged by absence).
	}

	// Threading: use In-Reply-To message ID as thread_id when available.
	threadID := ""
	if e.InReplyTo != "" {
		threadID = e.InReplyTo
	}

	var msgID *string
	if e.MessageID != "" {
		msgID = &e.MessageID
	}

	record := &domain.ContactEmail{
		ContactID: contactID,
		Direction: domain.EmailDirectionInbound,
		FromAddr:  e.From,
		ToAddr:    e.To,
		Subject:   subject,
		Body:      e.Body,
		ThreadID:  threadID,
		MessageID: msgID,
		SentAt:    time.Now().UTC(),
	}

	if _, err := h.emails.Create(ctx, record); err != nil {
		return fmt.Errorf("store email: %w", err)
	}

	// Auto-log activity on matched contact (best-effort).
	if contactID != nil {
		users, _, err := h.users.List(ctx, domain.UserFilter{Limit: 1})
		if err == nil && len(users) > 0 {
			snippet := bodySnippet(e.Body)
			act := &domain.Activity{
				Type:        domain.ActivityTypeEmail,
				Subject:     subject,
				Description: &snippet,
				ContactID:   contactID,
				OwnerID:     users[0].ID,
			}
			_, _ = h.activities.Create(ctx, act)
		}
	}

	return nil
}

// orgContext returns a context with the org_id set. In single-tenant mode
// the default org is used; in multi-tenant mode the ?org_id query param is read.
func (h *InboundEmailHandler) orgContext(r *http.Request) context.Context {
	if h.orgMode == config.OrgModeSingle {
		return domain.WithOrgID(r.Context(), domain.DefaultOrgID)
	}
	if rawID := r.URL.Query().Get("org_id"); rawID != "" {
		if orgID, err := uuid.Parse(rawID); err == nil {
			return domain.WithOrgID(r.Context(), orgID)
		}
	}
	return r.Context()
}

// ─── types ───────────────────────────────────────────────────────────────────

type parsedInbound struct {
	MessageID string
	From      string
	To        string
	Subject   string
	Body      string
	InReplyTo string
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// bodySnippet truncates body text to a preview snippet of up to 500 characters.
func bodySnippet(body string) string {
	const maxLen = 500
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen]
}

// extractInReplyTo pulls the In-Reply-To value out of a raw multi-line headers string.
func extractInReplyTo(rawHeaders string) string {
	for _, line := range strings.Split(rawHeaders, "\n") {
		if strings.HasPrefix(strings.ToLower(line), "in-reply-to:") {
			return strings.TrimSpace(line[len("in-reply-to:"):])
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func inboundParseAddr(from string) string {
	addr, err := mail.ParseAddress(from)
	if err != nil {
		trimmed := strings.TrimSpace(from)
		if strings.Contains(trimmed, "@") {
			return strings.ToLower(trimmed)
		}
		return ""
	}
	return strings.ToLower(addr.Address)
}

func verifyInboundMailgunSignature(signingKey, timestamp, token, signature string) bool {
	mac := hmac.New(sha256.New, []byte(signingKey))
	_, _ = fmt.Fprintf(mac, "%s%s", timestamp, token)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
