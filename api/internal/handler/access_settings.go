package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

type AccessSettingsHandler struct {
	repo repository.AccessRepository
}

func NewAccessSettingsHandler(repo repository.AccessRepository) *AccessSettingsHandler {
	return &AccessSettingsHandler{repo: repo}
}

func (h *AccessSettingsHandler) RolesRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListRoles)
	r.Post("/", h.CreateRole)
	r.Patch("/{id}", h.UpdateRole)
	r.Delete("/{id}", h.DeleteRole)
	r.Patch("/{id}/parent", h.MoveRole)
	return r
}

func (h *AccessSettingsHandler) ProfilesRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListProfiles)
	r.Post("/", h.CreateProfile)
	r.Get("/catalog", h.PermissionCatalog)
	r.Patch("/{id}", h.UpdateProfile)
	r.Delete("/{id}", h.DeleteProfile)
	r.Get("/{id}/permissions", h.GetProfilePermissions)
	r.Put("/{id}/permissions", h.ReplaceProfilePermissions)
	return r
}

func (h *AccessSettingsHandler) GroupsRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListGroups)
	r.Post("/", h.CreateGroup)
	r.Patch("/{id}", h.UpdateGroup)
	r.Delete("/{id}", h.DeleteGroup)
	r.Put("/{id}/members", h.ReplaceGroupMembers)
	return r
}

func (h *AccessSettingsHandler) SharingRulesRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetSharingRules)
	r.Patch("/", h.ReplaceSharingRules)
	return r
}

func (h *AccessSettingsHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	roles, err := h.repo.ListRoles(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": roles})
}

func (h *AccessSettingsHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	var req domain.ACLRole
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	req.OrgID = orgID
	created, err := h.repo.CreateRole(r.Context(), &req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AccessSettingsHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var patch domain.ACLRolePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.UpdateRole(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *AccessSettingsHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	if err := h.repo.DeleteRole(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccessSettingsHandler) MoveRole(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var req struct {
		ParentID *uuid.UUID `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.MoveRole(r.Context(), id, req.ParentID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *AccessSettingsHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	profiles, err := h.repo.ListProfiles(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": profiles})
}

func (h *AccessSettingsHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	var req domain.ACLProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	req.OrgID = orgID
	created, err := h.repo.CreateProfile(r.Context(), &req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AccessSettingsHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var patch domain.ACLProfilePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.UpdateProfile(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *AccessSettingsHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	if err := h.repo.DeleteProfile(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccessSettingsHandler) GetProfilePermissions(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	permissions, fieldPermissions, err := h.repo.ListProfilePermissions(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"permissions":       permissions,
		"field_permissions": fieldPermissions,
	})
}

func (h *AccessSettingsHandler) ReplaceProfilePermissions(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var req struct {
		Permissions      []domain.ACLProfilePermission      `json:"permissions"`
		FieldPermissions []domain.ACLProfileFieldPermission `json:"field_permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.repo.ReplaceProfilePermissions(r.Context(), id, req.Permissions, req.FieldPermissions); err != nil {
		handleDomainErr(w, err)
		return
	}
	h.GetProfilePermissions(w, r)
}

func (h *AccessSettingsHandler) PermissionCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"modules": domain.PermissionModules,
		"actions": domain.PermissionActions,
	})
}

func (h *AccessSettingsHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	groups, err := h.repo.ListGroups(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": groups})
}

func (h *AccessSettingsHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	var req domain.ACLGroup
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	req.OrgID = orgID
	created, err := h.repo.CreateGroup(r.Context(), &req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AccessSettingsHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var patch domain.ACLGroupPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.UpdateGroup(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *AccessSettingsHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	if err := h.repo.DeleteGroup(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccessSettingsHandler) ReplaceGroupMembers(w http.ResponseWriter, r *http.Request) {
	id, ok := parseACLID(w, r)
	if !ok {
		return
	}
	var req struct {
		UserIDs []uuid.UUID `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.repo.ReplaceGroupMembers(r.Context(), id, req.UserIDs); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccessSettingsHandler) GetSharingRules(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	rules, err := h.repo.GetSharingRules(r.Context(), orgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (h *AccessSettingsHandler) ReplaceSharingRules(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org_id required")
		return
	}
	var req domain.ACLSharingRules
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rules, err := h.repo.ReplaceSharingRules(r.Context(), orgID, &req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func parseACLID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return uuid.Nil, false
	}
	return id, true
}
