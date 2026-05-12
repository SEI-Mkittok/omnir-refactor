package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type UserHandler struct {
	repo        repository.UserRepository
	assignments userAssignmentRepository
}

type userAssignmentRepository interface {
	ListRoles(ctx context.Context, orgID uuid.UUID) ([]*domain.ACLRole, error)
	ListProfiles(ctx context.Context, orgID uuid.UUID) ([]*domain.ACLProfile, error)
}

func NewUserHandler(repo repository.UserRepository, assignments ...userAssignmentRepository) *UserHandler {
	h := &UserHandler{repo: repo}
	if len(assignments) > 0 {
		h.assignments = assignments[0]
	}
	return h
}

func (h *UserHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.requireAdmin(h.List))
	r.Get("/assignment-options", h.requireAdmin(h.AssignmentOptions))
	r.Post("/", h.requireAdmin(h.Create))
	r.Get("/me", h.GetMe)
	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.requireAdmin(h.Delete))
	return r
}

// requireAdmin returns 403 if the caller is not an admin.
func (h *UserHandler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.hasAdminAccess(r) {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next(w, r)
	}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.UserFilter{
		Q:     q.Get("q"),
		Sort:  q.Get("sort"),
		Order: q.Get("order"),
	}

	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n <= 200 {
			filter.Limit = n
		}
	}
	if v := q.Get("role"); v != "" {
		role := domain.UserRole(v)
		filter.Role = &role
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	users, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(users, total, filter.Page, filter.Limit))
}

func (h *UserHandler) AssignmentOptions(w http.ResponseWriter, r *http.Request) {
	if h.assignments == nil {
		writeError(w, http.StatusInternalServerError, "assignment repository unavailable")
		return
	}
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	roles, err := h.assignments.ListRoles(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	profiles, err := h.assignments.ListProfiles(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"roles":    roles,
		"profiles": profiles,
	})
}

// GetMe returns the currently authenticated user.
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	u, err := h.repo.GetByID(r.Context(), claims.UserID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	attachPermissionsFromAccess(r, u)
	writeJSON(w, http.StatusOK, u)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}

	// Users can only fetch their own record unless admin.
	if !h.isAdminOrSelf(r, id) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	u, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r)
	var req struct {
		Name      string          `json:"name"`
		Email     string          `json:"email"`
		Password  string          `json:"password"`
		Role      domain.UserRole `json:"role"`
		RoleID    *uuid.UUID      `json:"role_id"`
		ProfileID *uuid.UUID      `json:"profile_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusUnprocessableEntity, "name, email, and password are required")
		return
	}
	if req.Role == "" {
		req.Role = domain.UserRoleAgent
	}
	if !domain.IsValidUserRole(req.Role) {
		writeError(w, http.StatusUnprocessableEntity, "invalid role")
		return
	}
	if req.Role == domain.UserRoleSuperAdmin && claims.Role != string(domain.UserRoleSuperAdmin) {
		writeError(w, http.StatusForbidden, "only super admins can assign the super_admin role")
		return
	}
	if req.Role == domain.UserRoleAdmin && !isPlatformAdminClaims(claims) {
		writeError(w, http.StatusForbidden, "only platform admins can assign the admin role")
		return
	}
	if (req.RoleID != nil || req.ProfileID != nil) && !h.canManageACLAssignments(r) {
		writeError(w, http.StatusForbidden, "settings admin access required to assign roles or profiles")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	u := &domain.User{
		Email:     req.Email,
		Name:      req.Name,
		Role:      req.Role,
		RoleID:    req.RoleID,
		ProfileID: req.ProfileID,
	}
	created, err := h.repo.Create(r.Context(), u, string(hash))
	if err != nil {
		if h.handleUserMutationErr(w, err) {
			return
		}
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}

	// Users can update themselves; only admins can update role or other users.
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var patch domain.UserPatch
	raw, err := decodeJSONPatch(r, &patch)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	patch.ClearRoleID = patchFieldIsNull(raw, "role_id")
	patch.ClearProfileID = patchFieldIsNull(raw, "profile_id")
	if patch.Role != nil && !domain.IsValidUserRole(*patch.Role) {
		writeError(w, http.StatusUnprocessableEntity, "invalid role")
		return
	}

	isAdmin := h.hasAdminAccess(r)
	isSelf := claims.UserID == id

	if !isAdmin && !isSelf {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	// Non-admins cannot change platform role or Bundle 4 assignments.
	if !isAdmin && (patch.Role != nil || patch.RoleID != nil || patch.ProfileID != nil || patch.ClearRoleID || patch.ClearProfileID) {
		writeError(w, http.StatusForbidden, "only admins can change roles or profiles")
		return
	}
	if isSelf &&
		(patch.Role != nil || patch.RoleID != nil || patch.ProfileID != nil || patch.ClearRoleID || patch.ClearProfileID) &&
		claims.Role != string(domain.UserRoleSuperAdmin) {
		writeError(w, http.StatusForbidden, "cannot change your own roles or profiles")
		return
	}
	if (patch.RoleID != nil || patch.ProfileID != nil || patch.ClearRoleID || patch.ClearProfileID) && !h.canManageACLAssignments(r) {
		writeError(w, http.StatusForbidden, "settings admin access required to assign roles or profiles")
		return
	}
	if patch.Role != nil {
		callerIsSuperAdmin := claims.Role == string(domain.UserRoleSuperAdmin)
		callerIsPlatformAdmin := isPlatformAdminClaims(claims)
		if *patch.Role == domain.UserRoleAdmin && !callerIsPlatformAdmin {
			writeError(w, http.StatusForbidden, "only platform admins can assign the admin role")
			return
		}
		if *patch.Role == domain.UserRoleSuperAdmin && !callerIsSuperAdmin {
			writeError(w, http.StatusForbidden, "only super admins can assign the super_admin role")
			return
		}
		if !callerIsSuperAdmin {
			current, err := h.repo.GetByID(r.Context(), id)
			if err != nil {
				handleDomainErr(w, err)
				return
			}
			if current.Role == domain.UserRoleAdmin && !callerIsPlatformAdmin {
				writeError(w, http.StatusForbidden, "only platform admins can change the admin role")
				return
			}
			if current.Role == domain.UserRoleSuperAdmin {
				writeError(w, http.StatusForbidden, "only super admins can change the super_admin role")
				return
			}
		}
	}

	u, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		if h.handleUserMutationErr(w, err) {
			return
		}
		handleDomainErr(w, err)
		return
	}
	if claims.UserID == id {
		attachPermissionsFromAccess(r, u)
	}
	writeJSON(w, http.StatusOK, u)
}

func isPlatformAdminClaims(claims *auth.Claims) bool {
	return claims != nil && domain.IsAdminRole(claims.Role)
}

func attachPermissionsFromAccess(r *http.Request, u *domain.User) {
	if u == nil {
		return
	}
	if access, ok := domain.AccessContextFromContext(r.Context()); ok {
		u.Permissions = access.Permissions
	}
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}

	// Prevent self-deletion.
	if claims, ok := middleware.ClaimsFromContext(r); ok && claims.UserID == id {
		writeError(w, http.StatusUnprocessableEntity, "cannot delete your own account")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// isAdminOrSelf returns true if the caller is an admin or the target user.
func (h *UserHandler) isAdminOrSelf(r *http.Request, targetID uuid.UUID) bool {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		return false
	}
	return h.hasAdminAccess(r) || claims.UserID == targetID
}

func (h *UserHandler) hasAdminAccess(r *http.Request) bool {
	return hasModuleAdminAccess(r, domain.ACLModuleUsers)
}

func (h *UserHandler) canManageACLAssignments(r *http.Request) bool {
	return hasModuleAdminAccess(r, domain.ACLModuleSettings)
}

// handleUserMutationErr maps low-level Postgres constraint/type errors to
// stable API-level 4xx responses so user profile edits never surface as 500s.
func (h *UserHandler) handleUserMutationErr(w http.ResponseWriter, err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	switch pgErr.Code {
	case "23505": // unique_violation
		writeError(w, http.StatusConflict, "email already exists")
		return true
	case "23502": // not_null_violation
		writeError(w, http.StatusUnprocessableEntity, "name and email are required")
		return true
	case "23514", "22P02": // check_violation / invalid_text_representation
		writeError(w, http.StatusUnprocessableEntity, "invalid role")
		return true
	default:
		return false
	}
}
