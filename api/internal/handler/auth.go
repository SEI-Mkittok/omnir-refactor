package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

const (
	cookieAccessToken  = "access_token"
	cookieRefreshToken = "refresh_token"
	accessTokenTTL     = 24 * time.Hour
	refreshTokenTTL    = 7 * 24 * time.Hour
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	users    repository.UserRepository
	orgs     repository.OrgRepository
	jwtSvc   *auth.JWTService
	auditor  Auditor
	totpRepo repository.TOTPRepository
}

func NewAuthHandler(users repository.UserRepository, jwtSvc *auth.JWTService) *AuthHandler {
	return &AuthHandler{users: users, jwtSvc: jwtSvc}
}

// WithOrgs wires an org repository into the handler for registration support.
func (h *AuthHandler) WithOrgs(orgs repository.OrgRepository) *AuthHandler {
	h.orgs = orgs
	return h
}

// WithAuditLog wires an audit log repository into the handler.
func (h *AuthHandler) WithAuditLog(r repository.AuditLogRepository) *AuthHandler {
	h.auditor = newAuditor(r)
	return h
}

// WithTOTP wires a TOTP repository into the handler for 2FA support.
func (h *AuthHandler) WithTOTP(totpRepo repository.TOTPRepository) *AuthHandler {
	h.totpRepo = totpRepo
	return h
}

func (h *AuthHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	return r
}

// RegisterRouter returns the public registration route.
// Mount this OUTSIDE any auth middleware group.
//
//	POST /register — self-service org + admin user creation (multi-tenant)
func (h *AuthHandler) RegisterRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	return r
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	OrgName  string `json:"orgName"`
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func orgSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = uuid.New().String()[:8]
	}
	return s
}

// Register creates a new org and an admin user, then issues auth cookies.
// POST /api/v1/auth/register — must be mounted outside the auth middleware group.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if h.orgs == nil {
		writeError(w, http.StatusNotImplemented, "registration is not available in this deployment")
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.OrgName = strings.TrimSpace(req.OrgName)

	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation error: name is required")
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
	if req.OrgName == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation error: orgName is required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Create org with a unique slug derived from the org name.
	slug := orgSlug(req.OrgName)
	exists, err := h.orgs.SlugExists(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if exists {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	org, err := h.orgs.Create(r.Context(), &domain.Organization{
		Name: req.OrgName,
		Slug: slug,
		Plan: domain.OrgPlanStarter,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	userCtx := domain.WithOrgID(r.Context(), org.ID)
	user := &domain.User{
		OrgID: org.ID,
		Email: req.Email,
		Name:  req.Name,
		Role:  domain.UserRoleAdmin,
	}
	created, err := h.users.Create(userCtx, user, string(hash))
	if err != nil {
		handleDomainErr(w, err)
		return
	}

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

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates a user and sets httpOnly cookies for access + refresh tokens.
// POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation error: email and password are required")
		return
	}

	user, hash, err := h.users.FindByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check if TOTP 2FA is required for this user.
	if h.totpRepo != nil {
		enabled, err := h.totpRepo.IsEnabled(r.Context(), user.ID)
		if err == nil && enabled {
			// Issue a short-lived pre-auth token so the frontend can call /auth/2fa/verify.
			preClaims := auth.Claims{UserID: user.ID, OrgID: user.OrgID, Role: string(user.Role), PreAuth: true}
			preAuthToken, tokenErr := h.jwtSvc.Issue(preClaims, 5*time.Minute)
			if tokenErr != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"requires_2fa":   true,
				"pre_auth_token": preAuthToken,
			})
			return
		}
	}

	claims := auth.Claims{
		UserID: user.ID,
		OrgID:  user.OrgID,
		Role:   string(user.Role),
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

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// Refresh validates the refresh_token cookie and rotates both cookies.
// POST /api/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieRefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	claims, err := h.jwtSvc.Verify(cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	accessToken, err := h.jwtSvc.Issue(*claims, accessTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	refreshToken, err := h.jwtSvc.Issue(*claims, refreshTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	secure := r.TLS != nil
	setAccessCookie(w, accessToken, secure)
	setRefreshCookie(w, refreshToken, secure)

	w.WriteHeader(http.StatusNoContent)
}

// Logout clears the auth cookies.
// POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	secure := r.TLS != nil
	clearAuthCookie(w, cookieAccessToken, "/", secure)
	clearAuthCookie(w, cookieRefreshToken, "/api/auth/refresh", secure)
	w.WriteHeader(http.StatusNoContent)
}

// --- package-level cookie helpers (shared with SetupHandler) ---

func setAccessCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieAccessToken,
		Value:    token,
		Path:     "/",
		MaxAge:   int(accessTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func setRefreshCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieRefreshToken,
		Value:    token,
		Path:     "/api/auth/refresh",
		MaxAge:   int(refreshTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearAuthCookie(w http.ResponseWriter, name, path string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}
