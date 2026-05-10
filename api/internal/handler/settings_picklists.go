package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type PicklistSettingsHandler struct {
	defs repository.CustomFieldDefinitionRepository
	repo repository.PicklistRepository
}

func NewPicklistSettingsHandler(defs repository.CustomFieldDefinitionRepository, repo repository.PicklistRepository) *PicklistSettingsHandler {
	return &PicklistSettingsHandler{defs: defs, repo: repo}
}

func (h *PicklistSettingsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.With(requireAdminRole).Get("/", h.ListFields)
	r.With(requireAdminRole).Get("/{fieldID}/values", h.ListValues)
	r.With(requireAdminRole).Put("/{fieldID}/values", h.UpsertValues)
	r.With(requireAdminRole).Post("/{fieldID}/remap-delete", h.RemapAndDelete)
	return r
}

func (h *PicklistSettingsHandler) ListFields(w http.ResponseWriter, r *http.Request) {
	entityTypeParam := r.URL.Query().Get("entity_type")
	var entityType *domain.CustomFieldEntityType
	if entityTypeParam != "" {
		parsed := domain.CustomFieldEntityType(entityTypeParam)
		if !parsed.IsValid() {
			writeError(w, http.StatusUnprocessableEntity, "invalid entity_type")
			return
		}
		entityType = &parsed
	}

	fields, err := h.defs.List(r.Context(), domain.CustomFieldDefinitionFilter{EntityType: entityType})
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	type responseItem struct {
		Field  *domain.CustomFieldDefinition `json:"field"`
		Values []*domain.PicklistValue       `json:"values"`
	}
	out := []responseItem{}
	for _, field := range fields {
		if field == nil {
			continue
		}
		if field.FieldType != domain.CustomFieldTypeSelect && field.FieldType != domain.CustomFieldTypeMultiSelect {
			continue
		}
		values, err := h.repo.ListValues(r.Context(), field.OrgID, field.ID)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		out = append(out, responseItem{Field: field, Values: values})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}

func parseFieldID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "fieldID"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid field id")
		return uuid.Nil, false
	}
	return id, true
}

func (h *PicklistSettingsHandler) ListValues(w http.ResponseWriter, r *http.Request) {
	fieldID, ok := parseFieldID(w, r)
	if !ok {
		return
	}
	claims, hasClaims := middleware.ClaimsFromContext(r)
	if !hasClaims {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	values, err := h.repo.ListValues(r.Context(), claims.OrgID, fieldID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"values": values})
}

func (h *PicklistSettingsHandler) UpsertValues(w http.ResponseWriter, r *http.Request) {
	fieldID, ok := parseFieldID(w, r)
	if !ok {
		return
	}
	claims, hasClaims := middleware.ClaimsFromContext(r)
	if !hasClaims {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Values []domain.PicklistValueInput `json:"values"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	values, err := h.repo.UpsertValues(r.Context(), claims.OrgID, fieldID, req.Values)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"values": values})
}

func (h *PicklistSettingsHandler) RemapAndDelete(w http.ResponseWriter, r *http.Request) {
	fieldID, ok := parseFieldID(w, r)
	if !ok {
		return
	}
	claims, hasClaims := middleware.ClaimsFromContext(r)
	if !hasClaims {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		FromValue string  `json:"from_value"`
		ToValue   *string `json:"to_value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.FromValue == "" {
		writeError(w, http.StatusUnprocessableEntity, "from_value is required")
		return
	}
	if err := h.repo.RemapAndDeleteValue(r.Context(), claims.OrgID, fieldID, req.FromValue, req.ToValue); err != nil {
		handleDomainErr(w, err)
		return
	}
	values, err := h.repo.ListValues(r.Context(), claims.OrgID, fieldID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"values": values})
}
