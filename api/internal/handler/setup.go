package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// SetupHandler handles the fresh-install onboarding endpoints.
type SetupHandler struct {
	users  repository.UserRepository
	jwtSvc *auth.JWTService
}

func NewSetupHandler(users repository.UserRepository, jwtSvc *auth.JWTService) *SetupHandler {
	return &SetupHandler{users: users, jwtSvc: jwtSvc}
}

func (h *SetupHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/status", h.Status)
	r.Post("/", h.Setup)
	return r
}

// Status returns whether the app needs initial setup.
// GET /api/setup/status → {"setup_required": true|false}
func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	count, err := h.users.CountAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"setupRequired": count == 0})
}

type setupRequest struct {
	AdminName   string `json:"adminName"`
	CompanyName string `json:"companyName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

// Setup creates the first admin user and seeds the default org.
// POST /api/setup → {token, user} or 409 if already set up.
func (h *SetupHandler) Setup(w http.ResponseWriter, r *http.Request) {
	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields.
	req.AdminName = strings.TrimSpace(req.AdminName)
	req.Email = strings.TrimSpace(req.Email)
	if req.AdminName == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation error: adminName is required")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation error: email is required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusUnprocessableEntity, "validation error: password must be at least 8 characters")
		return
	}

	// Ensure no users exist yet.
	count, err := h.users.CountAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if count > 0 {
		writeError(w, http.StatusConflict, "setup already completed")
		return
	}

	// Hash the password.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Create the first admin user under the default org.
	user := &domain.User{
		OrgID: domain.DefaultOrgID,
		Email: req.Email,
		Name:  req.AdminName,
		Role:  domain.UserRoleAdmin,
	}
	created, err := h.users.Create(r.Context(), user, string(hash))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Issue access and refresh JWTs via httpOnly cookies for immediate login.
	claims := auth.Claims{
		UserID: created.ID,
		OrgID:  created.OrgID,
		Role:   string(created.Role),
	}
	accessToken, err := h.jwtSvc.Issue(claims, accessTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	refreshToken, err := h.jwtSvc.Issue(claims, refreshTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	secure := r.TLS != nil
	setAccessCookie(w, accessToken, secure)
	setRefreshCookie(w, refreshToken, secure)

	writeJSON(w, http.StatusCreated, map[string]any{"user": created})
}
