package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// OutboundWebhookHandler handles CRUD for org-registered outbound webhooks.
type OutboundWebhookHandler struct {
	repo repository.OutboundWebhookRepository
}

func NewOutboundWebhookHandler(repo repository.OutboundWebhookRepository) *OutboundWebhookHandler {
	return &OutboundWebhookHandler{repo: repo}
}

// NewWebhookMgmtHandler is an alias for NewOutboundWebhookHandler (used by main.go).
func NewWebhookMgmtHandler(repo repository.OutboundWebhookRepository) *OutboundWebhookHandler {
	return NewOutboundWebhookHandler(repo)
}

func (h *OutboundWebhookHandler) Router() chi.Router {
	r := chi.NewRouter()
	adminOnly := middleware.RequireRole(domain.UserRoleAdmin)
	r.With(adminOnly).Get("/", h.List)
	r.With(adminOnly).Post("/", h.Create)
	r.With(adminOnly).Get("/{id}", h.Get)
	r.With(adminOnly).Patch("/{id}", h.Update)
	r.With(adminOnly).Delete("/{id}", h.Delete)
	r.With(adminOnly).Post("/{id}/test", h.Test)
	r.With(adminOnly).Get("/{id}/deliveries", h.ListDeliveries)
	return r
}

// List returns all outbound webhooks for the org.
func (h *OutboundWebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	hooks, err := h.repo.List(r.Context())
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	if hooks == nil {
		hooks = []*domain.Webhook{}
	}
	// Omit secrets from list response.
	for _, hook := range hooks {
		hook.Secret = ""
	}
	writeJSON(w, http.StatusOK, hooks)
}

// Create registers a new outbound webhook.
func (h *OutboundWebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL    string                `json:"url"`
		Events []domain.WebhookEvent `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if body.URL == "" {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "url is required")
		return
	}
	if len(body.Events) == 0 {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "at least one event is required")
		return
	}
	for _, e := range body.Events {
		if !e.IsValid() {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "unknown event: "+string(e))
			return
		}
	}

	secret, err := generateSecret()
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "could not generate secret")
		return
	}

	hook, err := h.repo.Create(r.Context(), &domain.Webhook{
		URL:    body.URL,
		Events: body.Events,
		Secret: secret,
	})
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, hook)
}

// Get returns a single webhook by ID (includes secret for key rotation purposes).
func (h *OutboundWebhookHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid webhook id")
		return
	}
	hook, err := h.repo.GetByID(r.Context(), id)
	if err == domain.ErrNotFound {
		writeProblem(w, http.StatusNotFound, "Not Found", "webhook not found")
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, hook)
}

// Update patches a webhook's URL, events, or active status.
func (h *OutboundWebhookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid webhook id")
		return
	}

	var patch domain.WebhookPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	for _, e := range patch.Events {
		if !e.IsValid() {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "unknown event: "+string(e))
			return
		}
	}

	hook, err := h.repo.Update(r.Context(), id, patch)
	if err == domain.ErrNotFound {
		writeProblem(w, http.StatusNotFound, "Not Found", "webhook not found")
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	hook.Secret = ""
	writeJSON(w, http.StatusOK, hook)
}

// Delete removes a webhook registration.
func (h *OutboundWebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid webhook id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Test sends a synthetic test event to the webhook URL.
func (h *OutboundWebhookHandler) Test(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid webhook id")
		return
	}
	hook, err := h.repo.GetByID(r.Context(), id)
	if err == domain.ErrNotFound {
		writeProblem(w, http.StatusNotFound, "Not Found", "webhook not found")
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	payload := domain.WebhookEventPayload{
		Event:    "test",
		OrgID:    hook.OrgID,
		EntityID: uuid.Nil,
		Data:     map[string]string{"message": "This is a test event from Omnir CRM"},
	}
	body, _ := json.Marshal(payload)

	delivery := &domain.WebhookDelivery{
		WebhookID: hook.ID,
		Event:     "test",
		Payload:   body,
	}
	if _, err := h.repo.CreateDelivery(r.Context(), delivery); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "could not queue test delivery")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
}

// ListDeliveries returns recent delivery attempts for a webhook.
func (h *OutboundWebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid webhook id")
		return
	}
	// Verify ownership before returning deliveries.
	if _, err := h.repo.GetByID(r.Context(), id); err == domain.ErrNotFound {
		writeProblem(w, http.StatusNotFound, "Not Found", "webhook not found")
		return
	} else if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	deliveries, err := h.repo.ListDeliveries(r.Context(), id, 20)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	if deliveries == nil {
		deliveries = []*domain.WebhookDelivery{}
	}
	writeJSON(w, http.StatusOK, deliveries)
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
