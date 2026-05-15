package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// APIKeyHandler handles CRUD for API keys.
type APIKeyHandler struct {
	repo  repository.APIKeyRepository
	audit Auditor
}

func NewAPIKeyHandler(repo repository.APIKeyRepository) *APIKeyHandler {
	return &APIKeyHandler{repo: repo}
}

func (h *APIKeyHandler) WithAuditLog(r repository.AuditLogRepository) *APIKeyHandler {
	h.audit = newAuditor(r)
	return h
}

func (h *APIKeyHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.requireAdmin(h.List))
	r.Post("/", h.requireAdmin(h.Create))
	r.Delete("/{id}", h.requireAdmin(h.Revoke))
	return r
}

func (h *APIKeyHandler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !hasModuleAdminAccess(r, domain.ACLModuleAPIKeys) {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next(w, r)
	}
}

// List returns all API keys for the org (prefix, name, last_used only — no hashes).
// GET /api/v1/api-keys
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	keys, err := h.repo.List(r.Context(), claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if keys == nil {
		keys = []*domain.APIKey{}
	}
	writeJSON(w, http.StatusOK, keys)
}

type createAPIKeyRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Scopes    []string   `json:"scopes,omitempty"`
}

type createAPIKeyResponse struct {
	*domain.APIKey
	// Key is the full plaintext key — returned once only.
	Key string `json:"key"`
}

// Create generates a new API key and returns the plaintext once.
// POST /api/v1/api-keys
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	plaintext, keyHash, keyPrefix, err := generateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	scopes := req.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	validScopes := map[string]bool{"read": true, "write": true}
	for _, s := range scopes {
		if !validScopes[s] {
			writeError(w, http.StatusUnprocessableEntity, "invalid scope: "+s+"; allowed values are read, write")
			return
		}
	}

	k := &domain.APIKey{
		OrgID:     claims.OrgID,
		CreatedBy: claims.UserID,
		Name:      req.Name,
		KeyPrefix: keyPrefix,
		Scopes:    scopes,
		ExpiresAt: req.ExpiresAt,
	}

	created, err := h.repo.Create(r.Context(), k, keyHash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logAPIKeyMutation(r, domain.AuditActionCreated, created, domain.AuditChanges{
		"name":       {From: nil, To: created.Name},
		"key_prefix": {From: nil, To: created.KeyPrefix},
		"scopes":     {From: nil, To: created.Scopes},
		"expires_at": {From: nil, To: created.ExpiresAt},
	})
	writeJSON(w, http.StatusCreated, createAPIKeyResponse{APIKey: created, Key: plaintext})
}

// Revoke marks an API key as revoked.
// DELETE /api/v1/api-keys/:id
func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}

	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.repo.Revoke(r.Context(), id, claims.OrgID); err != nil {
		handleDomainErr(w, err)
		return
	}
	h.logAPIKeyMutation(r, domain.AuditActionDeleted, &domain.APIKey{ID: id, OrgID: claims.OrgID}, domain.AuditChanges{
		"revoked_at": {From: nil, To: time.Now().UTC()},
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIKeyHandler) logAPIKeyMutation(r *http.Request, action domain.AuditAction, key *domain.APIKey, changes domain.AuditChanges) {
	if h.audit.repo == nil || key == nil {
		return
	}
	entry := domain.AuditEntry{
		OrgID:      key.OrgID,
		Action:     action,
		EntityType: domain.AuditEntityAPIKey,
		EntityID:   &key.ID,
		Changes:    changes,
	}
	if key.Name != "" {
		name := key.Name
		entry.EntityName = &name
	}
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		userID := claims.UserID
		entry.UserID = &userID
		if entry.OrgID == uuid.Nil {
			entry.OrgID = claims.OrgID
		}
	}
	if ip := realClientIP(r); ip != "" {
		entry.IPAddress = &ip
	}
	if ua := r.UserAgent(); ua != "" {
		entry.UserAgent = &ua
	}
	go func() {
		if err := h.audit.repo.Append(context.Background(), entry); err != nil {
			slog.Error("audit log write failed", "entity", "api_key", "error", err)
		}
	}()
}

// generateAPIKey produces a cryptographically random API key.
// Returns: plaintext (ak_<random>), SHA-256 hex hash, display prefix (first 10 chars).
func generateAPIKey() (plaintext, keyHash, keyPrefix string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plaintext = "ak_" + base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(plaintext))
	keyHash = hex.EncodeToString(h[:])
	// Prefix is "ak_" + first 7 chars of the base64 part — enough to identify, not enough to brute-force.
	if len(plaintext) >= 10 {
		keyPrefix = plaintext[:10]
	} else {
		keyPrefix = plaintext
	}
	return
}
