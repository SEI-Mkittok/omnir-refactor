package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// WebhookHandler handles inbound email webhooks from Postmark and Mailgun.
// Routes are mounted outside the /api/v1 auth group so they are publicly
// reachable by the email provider. Requests are authenticated via HMAC
// signature verification (Mailgun) or HTTP Basic Auth (Postmark) using
// WEBHOOK_SECRET.
type WebhookHandler struct {
	tickets  repository.TicketRepository
	comments repository.TicketCommentRepository
	contacts repository.ContactRepository
	secret   string
	orgMode  config.OrgMode
	logger   *slog.Logger
}

func NewWebhookHandler(
	tickets repository.TicketRepository,
	comments repository.TicketCommentRepository,
	contacts repository.ContactRepository,
	secret string,
	orgMode config.OrgMode,
	logger *slog.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		tickets:  tickets,
		comments: comments,
		contacts: contacts,
		secret:   secret,
		orgMode:  orgMode,
		logger:   logger,
	}
}

func (h *WebhookHandler) Router() http.Handler {
	r := chi.NewRouter()
	r.Post("/postmark", h.HandlePostmark)
	r.Post("/mailgun", h.HandleMailgun)
	return r
}

// ───────────────────────── Postmark ─────────────────────────

// postmarkPayload is the inbound email JSON body sent by Postmark.
type postmarkPayload struct {
	MessageID string           `json:"MessageID"`
	From      string           `json:"From"`
	Subject   string           `json:"Subject"`
	TextBody  string           `json:"TextBody"`
	HtmlBody  string           `json:"HtmlBody"`
	Headers   []postmarkHeader `json:"Headers"`
}

type postmarkHeader struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

func (p *postmarkPayload) inReplyTo() string {
	for _, h := range p.Headers {
		if strings.EqualFold(h.Name, "In-Reply-To") {
			return strings.TrimSpace(h.Value)
		}
	}
	return ""
}

func (p *postmarkPayload) body() string {
	if p.TextBody != "" {
		return p.TextBody
	}
	return p.HtmlBody
}

// HandlePostmark processes an inbound email webhook from Postmark.
// Postmark authenticates via HTTP Basic Auth — configure the webhook password
// in the Postmark dashboard and set WEBHOOK_SECRET to the same value.
func (h *WebhookHandler) HandlePostmark(w http.ResponseWriter, r *http.Request) {
	if h.secret != "" {
		_, password, ok := r.BasicAuth()
		if !ok || !hmac.Equal([]byte(password), []byte(h.secret)) {
			writeError(w, http.StatusUnauthorized, "invalid webhook credentials")
			return
		}
	}

	var payload postmarkPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "malformed JSON")
		return
	}

	parsed := &parsedEmail{
		MessageID: payload.MessageID,
		From:      payload.From,
		Subject:   payload.Subject,
		Body:      payload.body(),
		InReplyTo: payload.inReplyTo(),
	}

	ctx := h.scopedContext(r)
	if err := h.ingest(ctx, parsed); err != nil {
		h.logger.Error("postmark ingest failed", "err", err)
		writeError(w, http.StatusInternalServerError, "ingestion error")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ───────────────────────── Mailgun ─────────────────────────

// HandleMailgun processes an inbound email webhook from Mailgun.
// Mailgun verifies authenticity by signing `timestamp + token` with your
// webhook signing key. Set WEBHOOK_SECRET to that key.
func (h *WebhookHandler) HandleMailgun(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "malformed form data")
		return
	}

	if h.secret != "" {
		timestamp := r.FormValue("timestamp")
		token := r.FormValue("token")
		signature := r.FormValue("signature")
		if !verifyMailgunSignature(h.secret, timestamp, token, signature) {
			writeError(w, http.StatusUnauthorized, "invalid webhook signature")
			return
		}
	}

	parsed := &parsedEmail{
		MessageID: r.FormValue("Message-Id"),
		From:      r.FormValue("sender"),
		Subject:   r.FormValue("subject"),
		Body:      r.FormValue("body-plain"),
		InReplyTo: strings.TrimSpace(r.FormValue("In-Reply-To")),
	}

	ctx := h.scopedContext(r)
	if err := h.ingest(ctx, parsed); err != nil {
		h.logger.Error("mailgun ingest failed", "err", err)
		writeError(w, http.StatusInternalServerError, "ingestion error")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// verifyMailgunSignature validates a Mailgun webhook signature.
// https://documentation.mailgun.com/docs/mailgun/user-manual/webhooks/
func verifyMailgunSignature(signingKey, timestamp, token, signature string) bool {
	mac := hmac.New(sha256.New, []byte(signingKey))
	_, _ = fmt.Fprintf(mac, "%s%s", timestamp, token)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ───────────────────────── Core ingestion ─────────────────────────

// parsedEmail normalises the fields from either provider into one struct.
type parsedEmail struct {
	MessageID string
	From      string
	Subject   string
	Body      string
	InReplyTo string
}

// ingest either appends a comment to an existing threaded ticket or creates
// a new ticket with the email body as its first comment.
func (h *WebhookHandler) ingest(ctx context.Context, e *parsedEmail) error {
	// Normalise subject.
	subject := strings.TrimSpace(e.Subject)
	if subject == "" {
		subject = "(no subject)"
	}

	// Resolve contact from the From address.
	var contactID *uuid.UUID
	if fromEmail := parseEmailAddress(e.From); fromEmail != "" {
		c, err := h.contacts.GetByEmail(ctx, fromEmail)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("contact lookup: %w", err)
		}
		if c != nil {
			contactID = &c.ID
		}
	}

	// Reply threading: if In-Reply-To matches an existing ticket, add a comment.
	if e.InReplyTo != "" {
		ticket, err := h.tickets.GetByEmailMessageID(ctx, e.InReplyTo)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("thread lookup: %w", err)
		}
		if ticket != nil {
			comment := &domain.TicketComment{
				TicketID:   ticket.ID,
				Body:       e.Body,
				IsInternal: false,
			}
			if _, err := h.comments.Create(ctx, comment); err != nil {
				return fmt.Errorf("create reply comment: %w", err)
			}
			h.logger.Info("email threaded to existing ticket",
				"ticket_id", ticket.ID,
				"in_reply_to", e.InReplyTo,
			)
			return nil
		}
	}

	// No matching thread — create a new ticket.
	source := "email"
	ticket := &domain.Ticket{
		Subject:   subject,
		Source:    &source,
		ContactID: contactID,
	}
	if e.MessageID != "" {
		ticket.EmailMessageID = &e.MessageID
	}

	created, err := h.tickets.Create(ctx, ticket)
	if err != nil {
		return fmt.Errorf("create ticket: %w", err)
	}

	// Attach the email body as the first (public) comment.
	if body := strings.TrimSpace(e.Body); body != "" {
		comment := &domain.TicketComment{
			TicketID:   created.ID,
			Body:       body,
			IsInternal: false,
		}
		if _, err := h.comments.Create(ctx, comment); err != nil {
			// Log but don't fail — the ticket was already created.
			h.logger.Warn("failed to attach email body as comment",
				"ticket_id", created.ID,
				"err", err,
			)
		}
	}

	h.logger.Info("email ticket created",
		"ticket_id", created.ID,
		"subject", created.Subject,
		"contact_id", contactID,
	)
	return nil
}

// ───────────────────────── Helpers ─────────────────────────

// scopedContext returns a context with org_id set for DB calls.
// In single-tenant mode the default org is used.
// In multi/enterprise mode the caller passes ?org_id=<uuid>.
func (h *WebhookHandler) scopedContext(r *http.Request) context.Context {
	if h.orgMode == config.OrgModeSingle {
		return domain.WithOrgID(r.Context(), domain.DefaultOrgID)
	}
	rawID := r.URL.Query().Get("org_id")
	if orgID, err := uuid.Parse(rawID); err == nil {
		return domain.WithOrgID(r.Context(), orgID)
	}
	// No org_id provided in multi-tenant mode — DB calls will fail or
	// silently omit the org scope. The handler still proceeds; the repo
	// will return an appropriate error.
	return r.Context()
}

// parseEmailAddress extracts the bare email address from a From header value
// that may be formatted as "Name <address>" or just "address".
func parseEmailAddress(from string) string {
	addr, err := mail.ParseAddress(from)
	if err != nil {
		// Fallback: treat the whole string as an address.
		trimmed := strings.TrimSpace(from)
		if strings.Contains(trimmed, "@") {
			return strings.ToLower(trimmed)
		}
		return ""
	}
	return strings.ToLower(addr.Address)
}
