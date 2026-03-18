package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// NotificationPrefHandler serves the /users/me/notification-prefs resource.
type NotificationPrefHandler struct {
	prefs repository.NotificationPrefRepository
}

func NewNotificationPrefHandler(prefs repository.NotificationPrefRepository) *NotificationPrefHandler {
	return &NotificationPrefHandler{prefs: prefs}
}

func (h *NotificationPrefHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Get)
	r.Patch("/", h.Update)
	return r
}

// Get returns the current user's notification preferences.
func (h *NotificationPrefHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing auth context")
		return
	}
	pref, err := h.prefs.GetByUser(r.Context(), claims.UserID, claims.OrgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pref)
}

// Update applies a partial patch to the current user's notification preferences.
func (h *NotificationPrefHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing auth context")
		return
	}

	var patch domain.UserNotificationPrefPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	// Fetch current (or default) prefs and apply the patch.
	current, err := h.prefs.GetByUser(r.Context(), claims.UserID, claims.OrgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	if patch.EmailOnAssigned != nil {
		current.EmailOnAssigned = *patch.EmailOnAssigned
	}
	if patch.EmailOnResolved != nil {
		current.EmailOnResolved = *patch.EmailOnResolved
	}
	if patch.EmailOnClosed != nil {
		current.EmailOnClosed = *patch.EmailOnClosed
	}

	updated, err := h.prefs.Upsert(r.Context(), current)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
