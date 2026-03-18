package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

var nonAlphanumRE = regexp.MustCompile(`[^a-z0-9]+`)

// OrgHandler handles org-level endpoints (SaaS signup).
type OrgHandler struct {
	orgs    repository.OrgRepository
	users   repository.UserRepository
	jwtSvc  *auth.JWTService
	orgMode config.OrgMode
}

func NewOrgHandler(
	orgs repository.OrgRepository,
	users repository.UserRepository,
	jwtSvc *auth.JWTService,
	orgMode config.OrgMode,
) *OrgHandler {
	return &OrgHandler{orgs: orgs, users: users, jwtSvc: jwtSvc, orgMode: orgMode}
}

func (h *OrgHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/signup", h.Signup)
	return r
}

type orgSignupRequest struct {
	OrgName   string `json:"orgName"`
	AdminName string `json:"adminName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// Signup creates a new organization and its first admin user (SaaS mode only).
// POST /api/orgs/signup
// Returns 403 if ORG_MODE is not saas/multitenant/enterprise.
func (h *OrgHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if h.orgMode == config.OrgModeSingle {
		writeError(w, http.StatusForbidden, "org signup is not available in single-tenant mode")
		return
	}

	var req orgSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.OrgName = strings.TrimSpace(req.OrgName)
	req.AdminName = strings.TrimSpace(req.AdminName)
	req.Email = strings.TrimSpace(req.Email)

	if req.OrgName == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation error: orgName is required")
		return
	}
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

	// Generate a URL-safe slug from the org name, then ensure uniqueness.
	slug := generateSlug(req.OrgName)
	slug, err := h.uniqueSlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Create the org (no org context needed — inserts into global orgs table).
	org, err := h.orgs.Create(r.Context(), &domain.Organization{
		Name: req.OrgName,
		Slug: slug,
		Plan: domain.OrgPlanStarter,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Hash password.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Create the first admin user for this org.
	// Inject org into context so the user repo can scope to it.
	ctx := domain.WithOrgID(r.Context(), org.ID)
	user := &domain.User{
		OrgID: org.ID,
		Email: req.Email,
		Name:  req.AdminName,
		Role:  domain.UserRoleAdmin,
	}
	created, err := h.users.Create(ctx, user, string(hash))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Issue JWTs with the new org_id embedded.
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

	writeJSON(w, http.StatusCreated, map[string]any{
		"org":  org,
		"user": created,
	})
}

// generateSlug converts an org name to a lowercase, hyphenated URL slug.
func generateSlug(name string) string {
	s := strings.ToLower(name)
	s = nonAlphanumRE.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "org"
	}
	return s
}

// uniqueSlug returns the slug if available, or appends -2, -3, … until unique.
func (h *OrgHandler) uniqueSlug(ctx context.Context, base string) (string, error) {
	candidate := base
	for i := 2; ; i++ {
		exists, err := h.orgs.SlugExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = base + "-" + itoa(i)
	}
}

func itoa(n int) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{digits[n%10]}, buf...)
		n /= 10
	}
	return string(buf)
}
