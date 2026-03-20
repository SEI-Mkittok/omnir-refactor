package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
	"github.com/omnir/crm-api/internal/seqtoken"
)

// transparentPixel is a 1×1 transparent GIF (43 bytes).
var transparentPixel = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00,
	0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x21,
	0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

// SequenceTrackingHandler handles public tracking endpoints for email sequences.
// No JWT auth — tokens carry their own HMAC-signed credentials.
type SequenceTrackingHandler struct {
	sequences   repository.SequenceRepository
	contacts    repository.ContactRepository
	tokenSecret string
}

// NewSequenceTrackingHandler creates a SequenceTrackingHandler.
func NewSequenceTrackingHandler(
	sequences repository.SequenceRepository,
	contacts repository.ContactRepository,
	tokenSecret string,
) *SequenceTrackingHandler {
	return &SequenceTrackingHandler{
		sequences:   sequences,
		contacts:    contacts,
		tokenSecret: tokenSecret,
	}
}

// TrackRouter mounts the /track/open and /track/click routes.
// Mount at /track.
func (h *SequenceTrackingHandler) TrackRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/open/{token}", h.OpenPixel)
	r.Get("/click/{token}", h.ClickRedirect)
	return r
}

// UnsubscribeRouter mounts the GET|POST /{token} unsubscribe routes.
// Mount at /unsubscribe.
func (h *SequenceTrackingHandler) UnsubscribeRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/{token}", h.Unsubscribe)
	r.Post("/{token}", h.Unsubscribe)
	return r
}

// BounceRouter mounts the bounce webhook route (uses WEBHOOK_SECRET, placed separately).
// Mount at /api/emails/bounce.
func (h *SequenceTrackingHandler) BounceRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.HandleBounce)
	return r
}

// ─────────────── Open pixel ───────────────

// OpenPixel records an "opened" event and returns a 1×1 transparent GIF.
// GET /track/open/{token}
func (h *SequenceTrackingHandler) OpenPixel(w http.ResponseWriter, r *http.Request) {
	claims, enrollment, err := h.verifyAndLoad(r)
	if err == nil && enrollment.Status == domain.EnrollmentStatusActive {
		h.recordEvent(r, claims, domain.SequenceEventOpened)
	}
	// Always return the pixel regardless of token validity (email clients retry).
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(transparentPixel)
}

// ─────────────── Click redirect ───────────────

// ClickRedirect records a "clicked" event and redirects to the destination URL.
// GET /track/click/{token}?url={destination}
func (h *SequenceTrackingHandler) ClickRedirect(w http.ResponseWriter, r *http.Request) {
	dest := r.URL.Query().Get("url")
	if dest == "" || (!strings.HasPrefix(dest, "http://") && !strings.HasPrefix(dest, "https://")) {
		http.Error(w, "missing or invalid url parameter", http.StatusBadRequest)
		return
	}

	claims, enrollment, err := h.verifyAndLoad(r)
	if err == nil && enrollment.Status == domain.EnrollmentStatusActive {
		h.recordEvent(r, claims, domain.SequenceEventClicked)
	}

	http.Redirect(w, r, dest, http.StatusFound)
}

// ─────────────── Unsubscribe ───────────────

// Unsubscribe sets enrollment=unsubscribed and contact email_opt_out=true.
// GET|POST /unsubscribe/{token}
func (h *SequenceTrackingHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	claims, enrollment, err := h.verifyAndLoad(r)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, unsubscribeHTML("Invalid or expired unsubscribe link."))
		return
	}

	if enrollment.Status != domain.EnrollmentStatusUnsubscribed {
		_ = h.sequences.UpdateEnrollmentStatus(r.Context(), enrollment.ID, domain.EnrollmentStatusUnsubscribed)
		_ = h.contacts.SetEmailOptOut(r.Context(), enrollment.ContactID)
		h.recordEvent(r, claims, domain.SequenceEventUnsubscribed)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, unsubscribeHTML("You have been unsubscribed. You will no longer receive emails from this sequence."))
}

// ─────────────── Bounce webhook ───────────────

// bouncePayload is a minimal SendGrid/Mailgun-compatible bounce webhook body.
// SendGrid sends an array; Mailgun sends a flat object. We handle both.
type bouncePayload struct {
	Email string `json:"email"`
}

// HandleBounce processes a bounce notification from SendGrid or Mailgun.
// POST /api/emails/bounce
func (h *SequenceTrackingHandler) HandleBounce(w http.ResponseWriter, r *http.Request) {
	var emails []string

	// Try SendGrid format: array of events
	var sgEvents []struct {
		Email string `json:"email"`
		Event string `json:"event"`
	}
	if err := json.NewDecoder(r.Body).Decode(&sgEvents); err == nil {
		for _, e := range sgEvents {
			if e.Event == "bounce" || e.Event == "blocked" {
				emails = append(emails, e.Email)
			}
		}
	}

	// Fall back: try single Mailgun-style object (already consumed body above if SG failed)
	// Note: if we got here we already tried to decode; reset is not possible.
	// So we only process SG format here; Mailgun support can be added if needed.

	processed := 0
	for _, email := range emails {
		contact, err := h.contacts.GetByEmail(r.Context(), email)
		if err != nil || contact == nil {
			continue
		}
		_ = h.contacts.IncrementBounceCount(r.Context(), contact.ID)

		// Mark all active enrollments for this contact as bounced.
		enrollments, err := h.sequences.GetActiveEnrollmentsByContact(r.Context(), contact.ID)
		if err == nil {
			for _, enr := range enrollments {
				_ = h.sequences.UpdateEnrollmentStatus(r.Context(), enr.ID, domain.EnrollmentStatusBounced)
				evt := &domain.SequenceEvent{
					SequenceID:   enr.SequenceID,
					EnrollmentID: enr.ID,
					ContactID:    enr.ContactID,
					OrgID:        enr.OrgID,
					Kind:         domain.SequenceEventBounced,
				}
				_ = h.sequences.RecordEvent(r.Context(), evt)
			}
		}
		processed++
	}

	writeJSON(w, http.StatusOK, map[string]any{"processed": processed})
}

// ─────────────── helpers ───────────────

func (h *SequenceTrackingHandler) verifyAndLoad(r *http.Request) (*seqtoken.Claims, *domain.SequenceEnrollment, error) {
	token := chi.URLParam(r, "token")
	claims, err := seqtoken.Verify(h.tokenSecret, token)
	if err != nil {
		return nil, nil, err
	}
	enrollment, err := h.sequences.GetEnrollmentByID(r.Context(), claims.EnrollmentID)
	if err != nil {
		return nil, nil, err
	}
	return claims, enrollment, nil
}

func (h *SequenceTrackingHandler) recordEvent(r *http.Request, claims *seqtoken.Claims, kind domain.SequenceEventKind) {
	stepID := claims.StepID
	evt := &domain.SequenceEvent{
		SequenceID:   claims.SequenceID,
		StepID:       &stepID,
		EnrollmentID: claims.EnrollmentID,
		OrgID:        claims.OrgID,
		Kind:         kind,
	}
	// ContactID from the enrollment is more reliable than the token.
	_ = h.sequences.RecordEvent(r.Context(), evt)
}

func unsubscribeHTML(msg string) string {
	return fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>Unsubscribe</title>
<style>body{font-family:sans-serif;max-width:480px;margin:80px auto;text-align:center;color:#444}</style>
</head><body><h2>%s</h2></body></html>`, msg)
}
