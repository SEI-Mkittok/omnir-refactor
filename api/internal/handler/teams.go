package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// TeamsHandler manages Microsoft Teams Incoming Webhook connections.
type TeamsHandler struct {
	repo repository.TeamsConnectionRepository
}

func NewTeamsHandler(repo repository.TeamsConnectionRepository) *TeamsHandler {
	return &TeamsHandler{repo: repo}
}

func (h *TeamsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/connect", h.Connect)
	r.Delete("/disconnect", h.Disconnect)
	r.Get("/status", h.Status)
	return r
}

// Connect stores or replaces the Teams Incoming Webhook for the caller's org.
//
// POST /api/integrations/teams/connect
// Body: { "webhook_url": "https://...", "channel_name": "...", "tenant_id": "..." }
func (h *TeamsHandler) Connect(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	var input domain.TeamsConnectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if input.WebhookURL == "" {
		writeProblem(w, http.StatusUnprocessableEntity, "Unprocessable Entity", "webhook_url is required")
		return
	}

	conn := &domain.TeamsConnection{
		OrgID:       orgID,
		TenantID:    input.TenantID,
		BotToken:    input.WebhookURL,
		ChannelName: input.ChannelName,
	}

	saved, err := h.repo.Upsert(r.Context(), conn)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	// Mask the webhook URL in the response for security.
	saved.BotToken = "[configured]"
	writeJSON(w, http.StatusOK, saved)
}

// Disconnect removes the Teams connection for the caller's org.
//
// DELETE /api/integrations/teams/disconnect
func (h *TeamsHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	if err := h.repo.Delete(r.Context(), orgID); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Status returns whether Teams is connected for the caller's org.
//
// GET /api/integrations/teams/status
func (h *TeamsHandler) Status(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	conn, err := h.repo.GetByOrgID(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if conn == nil {
		writeJSON(w, http.StatusOK, map[string]any{"connected": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"connected":    true,
		"channel_name": conn.ChannelName,
		"tenant_id":    conn.TenantID,
		"created_at":   conn.CreatedAt,
		"updated_at":   conn.UpdatedAt,
	})
}
