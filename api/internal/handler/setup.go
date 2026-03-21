package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// SetupHandler handles the fresh-install onboarding endpoints.
type SetupHandler struct {
	users   repository.UserRepository
	orgs    repository.OrgRepository
	jwtSvc  *auth.JWTService
	orgMode config.OrgMode
}

func NewSetupHandler(users repository.UserRepository, orgs repository.OrgRepository, jwtSvc *auth.JWTService, orgMode config.OrgMode) *SetupHandler {
	return &SetupHandler{users: users, orgs: orgs, jwtSvc: jwtSvc, orgMode: orgMode}
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
	hasAdmin, err := h.users.HasAdminUser(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"setupRequired": !hasAdmin})
}

type setupRequest struct {
	AdminName   string `json:"adminName"`
	CompanyName string `json:"companyName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

// Setup creates the first admin user and seeds the default org.
// POST /api/setup → {token, user} or 409 if already set up.
// In multi-tenant mode this endpoint is disabled once any org exists; use POST /api/v1/auth/signup instead.
func (h *SetupHandler) Setup(w http.ResponseWriter, r *http.Request) {
	// In non-single mode: block setup once orgs have been provisioned (use signup instead).
	if h.orgMode != config.OrgModeSingle {
		hasOrgs, err := h.orgs.HasAny(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if hasOrgs {
			writeError(w, http.StatusNotFound, "setup is not available in multi-tenant mode")
			return
		}
	}

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

	// Block setup if an admin user already exists.
	hasAdmin, err := h.users.HasAdminUser(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if hasAdmin {
		writeError(w, http.StatusConflict, "setup already completed")
		return
	}

	// Hash the password.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Ensure the default org row exists. In self-hosted deployments the migration
	// seeds it, but we verify (and create if missing) here so that the JWT's
	// org_id is guaranteed to match a real row in the orgs table.
	org, err := h.orgs.GetByID(r.Context(), domain.DefaultOrgID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		// Org not seeded yet — create it now.
		org, err = h.orgs.Create(r.Context(), &domain.Organization{
			ID:   domain.DefaultOrgID,
			Name: "Default",
			Slug: "default",
			Plan: domain.OrgPlanSingle,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	// Create the first admin user under the default org.
	// Inject org_id into context so the user repo scopes the INSERT correctly.
	userCtx := domain.WithOrgID(r.Context(), org.ID)
	user := &domain.User{
		OrgID: org.ID,
		Email: req.Email,
		Name:  req.AdminName,
		Role:  domain.UserRoleAdmin,
	}
	created, err := h.users.Create(userCtx, user, string(hash))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Issue access and refresh JWTs via httpOnly cookies for immediate login.
	// org.ID is the authoritative org_id — it matches the row the user was
	// inserted into and what the OrgScopedPool will use for subsequent requests.
	claims := auth.Claims{
		UserID: created.ID,
		OrgID:  org.ID,
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
