package handler

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

var customFieldNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

const maxCustomFieldsPerEntityType = 50

// CustomFieldHandler serves the /custom-fields resource (admin only).
type CustomFieldHandler struct {
	defs repository.CustomFieldDefinitionRepository
}

func NewCustomFieldHandler(defs repository.CustomFieldDefinitionRepository) *CustomFieldHandler {
	return &CustomFieldHandler{defs: defs}
}

func (h *CustomFieldHandler) Router() chi.Router {
	r := chi.NewRouter()
	adminOnly := middleware.RequireRole(domain.UserRoleAdmin)

	r.With(adminOnly).Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.With(adminOnly).Patch("/{id}", h.Update)
	r.With(adminOnly).Delete("/{id}", h.Delete)
	return r
}

func (h *CustomFieldHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EntityType string   `json:"entity_type"`
		Name       string   `json:"name"`
		Label      string   `json:"label"`
		FieldType  string   `json:"field_type"`
		Options    []string `json:"options"`
		Required   bool     `json:"required"`
		OrderIdx   int      `json:"order_idx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.Name == "" || req.Label == "" {
		writeError(w, http.StatusUnprocessableEntity, "name and label are required")
		return
	}
	if !customFieldNameRe.MatchString(req.Name) {
		writeError(w, http.StatusUnprocessableEntity, "name must match ^[a-z][a-z0-9_]*$")
		return
	}
	et := domain.CustomFieldEntityType(req.EntityType)
	if !et.IsValid() {
		writeError(w, http.StatusUnprocessableEntity, "invalid entity_type: must be ticket, contact, lead, deal, or account")
		return
	}
	ft := domain.CustomFieldType(req.FieldType)
	if !ft.IsValid() {
		writeError(w, http.StatusUnprocessableEntity, "invalid field_type: must be text, number, date, url, checkbox, select, or multiselect")
		return
	}
	if (ft == domain.CustomFieldTypeSelect || ft == domain.CustomFieldTypeMultiSelect) && len(req.Options) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "options are required for select and multiselect field types")
		return
	}

	existing, err := h.defs.List(r.Context(), domain.CustomFieldDefinitionFilter{EntityType: &et})
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if len(existing) >= maxCustomFieldsPerEntityType {
		writeError(w, http.StatusUnprocessableEntity, "maximum of 50 custom fields per entity type reached")
		return
	}

	def := &domain.CustomFieldDefinition{
		EntityType: et,
		Name:       req.Name,
		Label:      req.Label,
		FieldType:  ft,
		Options:    req.Options,
		Required:   req.Required,
		OrderIdx:   req.OrderIdx,
	}
	created, err := h.defs.Create(r.Context(), def)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *CustomFieldHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := domain.CustomFieldDefinitionFilter{}
	if v := r.URL.Query().Get("entity_type"); v != "" {
		et := domain.CustomFieldEntityType(v)
		filter.EntityType = &et
	}
	defs, err := h.defs.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if defs == nil {
		defs = []*domain.CustomFieldDefinition{}
	}
	writeJSON(w, http.StatusOK, defs)
}

func (h *CustomFieldHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	def, err := h.defs.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, def)
}

func (h *CustomFieldHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.CustomFieldDefinitionPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	updated, err := h.defs.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *CustomFieldHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	if err := h.defs.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
