package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// PushHandler serves /api/v1/push/subscribe.
type PushHandler struct {
	repo           repository.PushSubscriptionRepository
	vapidPublicKey string
}

func NewPushHandler(repo repository.PushSubscriptionRepository, vapidPublicKey string) *PushHandler {
	return &PushHandler{repo: repo, vapidPublicKey: vapidPublicKey}
}

func (h *PushHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/vapid-public-key", h.VAPIDPublicKey)
	r.Post("/subscribe", h.Subscribe)
	r.Delete("/subscribe", h.Unsubscribe)
	return r
}

// VAPIDPublicKey returns the server's VAPID public key for client-side subscription setup.
func (h *PushHandler) VAPIDPublicKey(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"vapid_public_key": h.vapidPublicKey})
}

// Subscribe stores a push subscription for the authenticated user.
func (h *PushHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Endpoint string `json:"endpoint"`
		P256dh   string `json:"p256dh"`
		Auth     string `json:"auth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.Endpoint == "" || req.P256dh == "" || req.Auth == "" {
		writeProblem(w, http.StatusUnprocessableEntity, "Unprocessable Entity", "endpoint, p256dh and auth are required")
		return
	}

	sub, err := h.repo.Upsert(r.Context(), &domain.PushSubscription{
		UserID:   claims.UserID,
		OrgID:    claims.OrgID,
		Endpoint: req.Endpoint,
		P256dh:   req.P256dh,
		Auth:     req.Auth,
	})
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sub)
}

// Unsubscribe removes a push subscription for the authenticated user.
func (h *PushHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.Endpoint == "" {
		writeProblem(w, http.StatusUnprocessableEntity, "Unprocessable Entity", "endpoint is required")
		return
	}

	if err := h.repo.DeleteByEndpoint(r.Context(), claims.UserID, req.Endpoint); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
