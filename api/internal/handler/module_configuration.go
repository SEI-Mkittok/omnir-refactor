package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type ModuleConfigurationHandler struct {
	layouts       repository.ModuleLayoutRepository
	relationships repository.ModuleRelationshipDefinitionRepository
	customFields  repository.CustomFieldDefinitionRepository
	links         repository.CRMEntityLinkRepository
	recordAccess  repository.RecordAccessRepository
}

func NewModuleConfigurationHandler(
	layouts repository.ModuleLayoutRepository,
	relationships repository.ModuleRelationshipDefinitionRepository,
	customFields repository.CustomFieldDefinitionRepository,
	links repository.CRMEntityLinkRepository,
	recordAccess repository.RecordAccessRepository,
) *ModuleConfigurationHandler {
	return &ModuleConfigurationHandler{
		layouts:       layouts,
		relationships: relationships,
		customFields:  customFields,
		links:         links,
		recordAccess:  recordAccess,
	}
}

func (h *ModuleConfigurationHandler) LayoutSettingsRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/{entityType}", h.GetAdminLayout)
	r.Put("/{entityType}", h.UpsertLayout)
	r.Post("/{entityType}/reset", h.ResetLayout)
	return r
}

func (h *ModuleConfigurationHandler) RuntimeLayoutRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/{entityType}", h.GetRuntimeLayout)
	return r
}

func (h *ModuleConfigurationHandler) RelationshipSettingsRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListAdminRelationships)
	r.Put("/", h.UpsertRelationship)
	r.Delete("/{id}", h.DeleteRelationship)
	return r
}

func (h *ModuleConfigurationHandler) RuntimeRelationshipRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListRuntimeRelationships)
	return r
}

func (h *ModuleConfigurationHandler) EntityLinksRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/{entityType}/{entityID}", h.ListEntityLinks)
	r.Post("/", h.CreateEntityLink)
	r.Delete("/{id}", h.DeleteEntityLink)
	return r
}

func (h *ModuleConfigurationHandler) GetAdminLayout(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "settings admin access required")
		return
	}
	entityType, ok := parseLayoutEntityParam(w, r)
	if !ok {
		return
	}
	layout, err := h.effectiveLayout(r, entityType)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, layout)
}

func (h *ModuleConfigurationHandler) GetRuntimeLayout(w http.ResponseWriter, r *http.Request) {
	entityType, ok := parseLayoutEntityParam(w, r)
	if !ok {
		return
	}
	if !h.ensureModulePermission(w, r, entityType, domain.ACLActionRead) {
		return
	}
	layout, err := h.effectiveLayout(r, entityType)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, layout)
}

func (h *ModuleConfigurationHandler) UpsertLayout(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "settings admin access required")
		return
	}
	entityType, ok := parseLayoutEntityParam(w, r)
	if !ok {
		return
	}
	var req struct {
		Blocks []domain.ModuleLayoutBlock `json:"blocks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}
	customFields, err := h.customFieldDefs(r, entityType)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	layout := &domain.ModuleLayout{OrgID: orgID, EntityType: entityType, Blocks: req.Blocks}
	if err := domain.ValidateModuleLayout(layout, customFields); err != nil {
		handleDomainErr(w, err)
		return
	}
	layout = domain.NormalizeModuleLayout(layout, customFields)
	if err := domain.ValidateModuleLayout(layout, customFields); err != nil {
		handleDomainErr(w, err)
		return
	}
	updated, err := h.layouts.Upsert(r.Context(), layout)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	updated = domain.NormalizeModuleLayout(updated, customFields)
	writeJSON(w, http.StatusOK, updated)
}

func (h *ModuleConfigurationHandler) ResetLayout(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "settings admin access required")
		return
	}
	entityType, ok := parseLayoutEntityParam(w, r)
	if !ok {
		return
	}
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}
	if err := h.layouts.Delete(r.Context(), orgID, entityType); err != nil && !strings.Contains(err.Error(), domain.ErrNotFound.Error()) {
		handleDomainErr(w, err)
		return
	}
	layout, err := h.effectiveLayout(r, entityType)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, layout)
}

func (h *ModuleConfigurationHandler) ListAdminRelationships(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "settings admin access required")
		return
	}
	defs, err := h.mergedRelationshipDefinitions(r, false)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, defs)
}

func (h *ModuleConfigurationHandler) ListRuntimeRelationships(w http.ResponseWriter, r *http.Request) {
	entityType, ok := parseQueryEntityType(w, r)
	if !ok {
		return
	}
	if !h.ensureModulePermission(w, r, entityType, domain.ACLActionRead) {
		return
	}
	defs, err := h.mergedRelationshipDefinitions(r, true)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, defs)
}

func (h *ModuleConfigurationHandler) UpsertRelationship(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "settings admin access required")
		return
	}
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}
	var req domain.ModuleRelationshipDefinition
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	req.OrgID = orgID
	if req.RelationshipKey == "" {
		req.RelationshipKey = generatedRelationshipKey(req.FromEntityType, req.ToEntityType)
	}
	if domain.IsSystemRelationshipKey(req.RelationshipKey) {
		base := systemRelationshipDefinition(req.RelationshipKey, orgID)
		if base == nil {
			writeError(w, http.StatusUnprocessableEntity, "unknown system relationship")
			return
		}
		base.ID = req.ID
		base.Label = strings.TrimSpace(req.Label)
		if base.Label == "" {
			base.Label = systemRelationshipDefinition(req.RelationshipKey, orgID).Label
		}
		base.IsEnabled = req.IsEnabled
		base.OrderIdx = req.OrderIdx
		base.Metadata = req.Metadata
		req = *base
	} else {
		req.StorageStrategy = domain.RelationshipStorageCRMEntityLinks
		req.SystemLocked = false
	}
	if err := domain.ValidateModuleRelationshipDefinition(&req); err != nil {
		handleDomainErr(w, err)
		return
	}
	if existing, err := h.relationships.GetByKey(r.Context(), orgID, req.RelationshipKey); err == nil && existing != nil {
		if hasLinks, err := h.links.HasForRelationshipDefinition(r.Context(), existing.ID); err != nil {
			handleDomainErr(w, err)
			return
		} else if hasLinks && relationshipDefinitionShapeChanged(existing, &req) {
			writeError(w, http.StatusConflict, "relationship definitions with existing links cannot change entity types, cardinality, or storage strategy")
			return
		}
	} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
		handleDomainErr(w, err)
		return
	}
	updated, err := h.relationships.Upsert(r.Context(), &req)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ModuleConfigurationHandler) DeleteRelationship(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "settings admin access required")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	def, err := h.relationships.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if def.SystemLocked || domain.IsSystemRelationshipKey(def.RelationshipKey) {
		writeError(w, http.StatusUnprocessableEntity, "system relationships cannot be deleted")
		return
	}
	hasLinks, err := h.links.HasForRelationshipDefinition(r.Context(), def.ID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if hasLinks {
		writeError(w, http.StatusConflict, "relationship definitions with existing links cannot be deleted")
		return
	}
	if err := h.relationships.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ModuleConfigurationHandler) ListEntityLinks(w http.ResponseWriter, r *http.Request) {
	entityType, entityID, ok := parseEntityLinkTarget(w, r)
	if !ok {
		return
	}
	if !h.ensureRecordAccess(w, r, entityType, entityID, domain.SharingAccessRead) {
		return
	}
	crmType, _ := domain.CRMEntityTypeForCustomFieldEntity(entityType)
	filter := domain.CRMEntityLinkFilter{EntityType: crmType, EntityID: entityID}
	if raw := r.URL.Query().Get("relationship_definition_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid relationship_definition_id")
			return
		}
		filter.RelationshipDefinitionID = &id
	}
	links, err := h.links.ListForEntity(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	definitionCache := map[uuid.UUID]bool{}
	filtered := make([]*domain.CRMEntityLink, 0, len(links))
	for _, link := range links {
		enabledDefinition, err := h.linkDefinitionEnabled(r, link, definitionCache)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		if !enabledDefinition {
			continue
		}
		allowed, err := h.canReadEntityLinkOpposite(r, entityType, entityID, link)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		if allowed {
			filtered = append(filtered, link)
		}
	}
	writeJSON(w, http.StatusOK, filtered)
}

func (h *ModuleConfigurationHandler) CreateEntityLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RelationshipDefinitionID uuid.UUID                    `json:"relationship_definition_id"`
		FromEntityType           domain.CustomFieldEntityType `json:"from_entity_type"`
		FromEntityID             uuid.UUID                    `json:"from_entity_id"`
		ToEntityType             domain.CustomFieldEntityType `json:"to_entity_type"`
		ToEntityID               uuid.UUID                    `json:"to_entity_id"`
		Metadata                 json.RawMessage              `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.RelationshipDefinitionID == uuid.Nil {
		writeError(w, http.StatusUnprocessableEntity, "relationship_definition_id is required")
		return
	}
	def, err := h.relationships.GetByID(r.Context(), req.RelationshipDefinitionID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if !def.IsEnabled || def.StorageStrategy != domain.RelationshipStorageCRMEntityLinks || def.SystemLocked {
		writeError(w, http.StatusUnprocessableEntity, "relationship definition cannot create generic links")
		return
	}
	if req.FromEntityType != def.FromEntityType || req.ToEntityType != def.ToEntityType {
		writeError(w, http.StatusUnprocessableEntity, "entity types do not match relationship definition")
		return
	}
	if req.FromEntityID == uuid.Nil || req.ToEntityID == uuid.Nil {
		writeError(w, http.StatusUnprocessableEntity, "from_entity_id and to_entity_id are required")
		return
	}
	if !h.ensureRecordAccess(w, r, req.FromEntityType, req.FromEntityID, domain.SharingAccessWrite) {
		return
	}
	if !h.ensureRecordAccess(w, r, req.ToEntityType, req.ToEntityID, domain.SharingAccessWrite) {
		return
	}
	fromCRM, _ := domain.CRMEntityTypeForCustomFieldEntity(req.FromEntityType)
	toCRM, _ := domain.CRMEntityTypeForCustomFieldEntity(req.ToEntityType)
	if err := h.validateCardinality(r, def, fromCRM, req.FromEntityID, toCRM, req.ToEntityID); err != nil {
		handleDomainErr(w, err)
		return
	}
	if len(req.Metadata) == 0 {
		req.Metadata = json.RawMessage(`{}`)
	}
	link := &domain.CRMEntityLink{
		RelationshipDefinitionID: &def.ID,
		FromEntityType:           fromCRM,
		FromEntityID:             req.FromEntityID,
		ToEntityType:             toCRM,
		ToEntityID:               req.ToEntityID,
		LinkType:                 def.RelationshipKey,
		Metadata:                 req.Metadata,
	}
	created, err := h.links.Create(r.Context(), link)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *ModuleConfigurationHandler) DeleteEntityLink(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	link, err := h.links.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	fromEntity, ok := customFieldEntityForCRMEntity(link.FromEntityType)
	if !ok || !h.ensureRecordAccess(w, r, fromEntity, link.FromEntityID, domain.SharingAccessWrite) {
		return
	}
	toEntity, ok := customFieldEntityForCRMEntity(link.ToEntityType)
	if !ok || !h.ensureRecordAccess(w, r, toEntity, link.ToEntityID, domain.SharingAccessWrite) {
		return
	}
	if err := h.links.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ModuleConfigurationHandler) effectiveLayout(r *http.Request, entityType domain.CustomFieldEntityType) (*domain.ModuleLayout, error) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		return nil, fmt.Errorf("%w: org context is required", domain.ErrValidation)
	}
	customFields, err := h.customFieldDefs(r, entityType)
	if err != nil {
		return nil, err
	}
	layout, err := h.layouts.GetByEntity(r.Context(), orgID, entityType)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrNotFound.Error()) {
			return domain.DefaultModuleLayout(orgID, entityType, customFields), nil
		}
		return nil, err
	}
	return domain.NormalizeModuleLayout(layout, customFields), nil
}

func (h *ModuleConfigurationHandler) customFieldDefs(r *http.Request, entityType domain.CustomFieldEntityType) ([]*domain.CustomFieldDefinition, error) {
	if h.customFields == nil {
		return nil, nil
	}
	return h.customFields.List(r.Context(), domain.CustomFieldDefinitionFilter{EntityType: &entityType})
}

func (h *ModuleConfigurationHandler) mergedRelationshipDefinitions(r *http.Request, enabledOnly bool) ([]*domain.ModuleRelationshipDefinition, error) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		return nil, fmt.Errorf("%w: org context is required", domain.ErrValidation)
	}
	filter := domain.ModuleRelationshipDefinitionFilter{OrgID: orgID}
	if raw := r.URL.Query().Get("entity_type"); raw != "" {
		entityType := domain.CustomFieldEntityType(raw)
		if !entityType.IsValid() {
			return nil, fmt.Errorf("%w: invalid entity_type", domain.ErrValidation)
		}
		filter.EntityType = &entityType
	}
	stored, err := h.relationships.List(r.Context(), filter)
	if err != nil {
		return nil, err
	}
	defs := domain.MergeDefaultRelationshipDefinitions(orgID, stored)
	out := defs[:0]
	for _, def := range defs {
		if filter.EntityType != nil && def.FromEntityType != *filter.EntityType && def.ToEntityType != *filter.EntityType {
			continue
		}
		if enabledOnly && !def.IsEnabled {
			continue
		}
		out = append(out, def)
	}
	return out, nil
}

func relationshipDefinitionShapeChanged(existing, next *domain.ModuleRelationshipDefinition) bool {
	return existing.FromEntityType != next.FromEntityType ||
		existing.ToEntityType != next.ToEntityType ||
		existing.Cardinality != next.Cardinality ||
		existing.StorageStrategy != next.StorageStrategy
}

func (h *ModuleConfigurationHandler) linkDefinitionEnabled(r *http.Request, link *domain.CRMEntityLink, cache map[uuid.UUID]bool) (bool, error) {
	if link == nil || link.RelationshipDefinitionID == nil {
		return true, nil
	}
	if enabled, ok := cache[*link.RelationshipDefinitionID]; ok {
		return enabled, nil
	}
	def, err := h.relationships.GetByID(r.Context(), *link.RelationshipDefinitionID)
	if errors.Is(err, domain.ErrNotFound) {
		cache[*link.RelationshipDefinitionID] = false
		return false, nil
	}
	if err != nil {
		return false, err
	}
	enabled := def.IsEnabled && def.StorageStrategy == domain.RelationshipStorageCRMEntityLinks && !def.SystemLocked
	cache[*link.RelationshipDefinitionID] = enabled
	return enabled, nil
}

func (h *ModuleConfigurationHandler) canReadEntityLinkOpposite(r *http.Request, anchorType domain.CustomFieldEntityType, anchorID uuid.UUID, link *domain.CRMEntityLink) (bool, error) {
	if link == nil {
		return false, nil
	}
	anchorCRM, ok := domain.CRMEntityTypeForCustomFieldEntity(anchorType)
	if !ok {
		return false, fmt.Errorf("%w: invalid entity_type", domain.ErrValidation)
	}
	oppositeType := link.FromEntityType
	oppositeID := link.FromEntityID
	if link.FromEntityType == anchorCRM && link.FromEntityID == anchorID {
		oppositeType = link.ToEntityType
		oppositeID = link.ToEntityID
	} else if link.ToEntityType != anchorCRM || link.ToEntityID != anchorID {
		return false, nil
	}
	oppositeCustom, ok := customFieldEntityForCRMEntityType(oppositeType)
	if !ok {
		return false, nil
	}
	return h.canAccessRecordSilently(r, oppositeCustom, oppositeID, domain.SharingAccessRead)
}

func (h *ModuleConfigurationHandler) canAccessRecordSilently(r *http.Request, entityType domain.CustomFieldEntityType, id uuid.UUID, accessLevel domain.SharingAccessLevel) (bool, error) {
	module, ok := domain.ACLModuleForCustomFieldEntity(entityType)
	if !ok {
		return false, fmt.Errorf("%w: invalid entity_type", domain.ErrValidation)
	}
	if claims, ok := middleware.ClaimsFromContext(r); ok && domain.IsAdminRole(claims.Role) {
		return true, nil
	}
	access, ok := domain.AccessContextFromContext(r.Context())
	if !ok || !access.HasPermission(module, moduleActionForSharing(accessLevel)) {
		return false, nil
	}
	if h.recordAccess == nil {
		return true, nil
	}
	allowed, err := h.recordAccess.CanAccessRecord(r.Context(), module, id, accessLevel)
	if errors.Is(err, domain.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return allowed, nil
}

func customFieldEntityForCRMEntityType(entityType domain.CRMEntityType) (domain.CustomFieldEntityType, bool) {
	switch entityType {
	case domain.CRMEntityTypeAccount:
		return domain.CustomFieldEntityAccount, true
	case domain.CRMEntityTypeContact:
		return domain.CustomFieldEntityContact, true
	case domain.CRMEntityTypeLead:
		return domain.CustomFieldEntityLead, true
	case domain.CRMEntityTypeDeal:
		return domain.CustomFieldEntityDeal, true
	case domain.CRMEntityTypeTicket:
		return domain.CustomFieldEntityTicket, true
	case domain.CRMEntityTypeQuote:
		return domain.CustomFieldEntityQuote, true
	default:
		return "", false
	}
}

func (h *ModuleConfigurationHandler) validateCardinality(r *http.Request, def *domain.ModuleRelationshipDefinition, fromType domain.CRMEntityType, fromID uuid.UUID, toType domain.CRMEntityType, toID uuid.UUID) error {
	links, err := h.links.ListForRelationshipDefinition(r.Context(), def.ID)
	if err != nil {
		return err
	}
	for _, link := range links {
		samePair := link.FromEntityType == fromType && link.FromEntityID == fromID && link.ToEntityType == toType && link.ToEntityID == toID
		if samePair {
			return fmt.Errorf("%w: relationship link already exists", domain.ErrConflict)
		}
		sameFrom := link.FromEntityType == fromType && link.FromEntityID == fromID
		sameTo := link.ToEntityType == toType && link.ToEntityID == toID
		switch def.Cardinality {
		case domain.RelationshipCardinalityOneToOne:
			if sameFrom || sameTo {
				return fmt.Errorf("%w: one_to_one relationship already has a link", domain.ErrConflict)
			}
		case domain.RelationshipCardinalityManyToOne:
			if sameFrom {
				return fmt.Errorf("%w: many_to_one relationship already has a target for this source", domain.ErrConflict)
			}
		case domain.RelationshipCardinalityOneToMany:
			if sameTo {
				return fmt.Errorf("%w: one_to_many relationship already has a source for this target", domain.ErrConflict)
			}
		}
	}
	return nil
}

func (h *ModuleConfigurationHandler) ensureModulePermission(w http.ResponseWriter, r *http.Request, entityType domain.CustomFieldEntityType, action domain.ACLAction) bool {
	module, ok := domain.ACLModuleForCustomFieldEntity(entityType)
	if !ok {
		writeError(w, http.StatusUnprocessableEntity, "invalid entity_type")
		return false
	}
	if claims, ok := middleware.ClaimsFromContext(r); ok && domain.IsAdminRole(claims.Role) {
		return true
	}
	access, ok := domain.AccessContextFromContext(r.Context())
	if !ok || !access.HasPermission(module, action) {
		writeError(w, http.StatusForbidden, "permission denied")
		return false
	}
	return true
}

func (h *ModuleConfigurationHandler) ensureRecordAccess(w http.ResponseWriter, r *http.Request, entityType domain.CustomFieldEntityType, id uuid.UUID, accessLevel domain.SharingAccessLevel) bool {
	module, ok := domain.ACLModuleForCustomFieldEntity(entityType)
	if !ok {
		writeError(w, http.StatusUnprocessableEntity, "invalid entity_type")
		return false
	}
	if !h.ensureModulePermission(w, r, entityType, moduleActionForSharing(accessLevel)) {
		return false
	}
	if h.recordAccess == nil {
		return true
	}
	allowed, err := h.recordAccess.CanAccessRecord(r.Context(), module, id, accessLevel)
	if err != nil {
		handleDomainErr(w, err)
		return false
	}
	if !allowed {
		handleDomainErr(w, domain.ErrNotFound)
		return false
	}
	return true
}

func moduleActionForSharing(accessLevel domain.SharingAccessLevel) domain.ACLAction {
	if accessLevel == domain.SharingAccessWrite {
		return domain.ACLActionUpdate
	}
	return domain.ACLActionRead
}

func parseLayoutEntityParam(w http.ResponseWriter, r *http.Request) (domain.CustomFieldEntityType, bool) {
	entityType := domain.CustomFieldEntityType(chi.URLParam(r, "entityType"))
	if !entityType.IsValid() {
		writeError(w, http.StatusUnprocessableEntity, "invalid entity_type: must be ticket, contact, lead, deal, account, quote, or kb_article")
		return "", false
	}
	return entityType, true
}

func parseQueryEntityType(w http.ResponseWriter, r *http.Request) (domain.CustomFieldEntityType, bool) {
	entityType := domain.CustomFieldEntityType(r.URL.Query().Get("entity_type"))
	if !entityType.IsValid() {
		writeError(w, http.StatusUnprocessableEntity, "entity_type is required and must be ticket, contact, lead, deal, account, quote, or kb_article")
		return "", false
	}
	return entityType, true
}

func parseEntityLinkTarget(w http.ResponseWriter, r *http.Request) (domain.CustomFieldEntityType, uuid.UUID, bool) {
	entityType := domain.CustomFieldEntityType(chi.URLParam(r, "entityType"))
	if !entityType.IsValid() {
		writeError(w, http.StatusUnprocessableEntity, "invalid entity_type")
		return "", uuid.Nil, false
	}
	id, err := uuid.Parse(chi.URLParam(r, "entityID"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid entity id")
		return "", uuid.Nil, false
	}
	return entityType, id, true
}

func generatedRelationshipKey(from, to domain.CustomFieldEntityType) string {
	return fmt.Sprintf("custom_%s_%s_%s", from, to, strings.ReplaceAll(uuid.NewString()[:8], "-", ""))
}

func systemRelationshipDefinition(key string, orgID uuid.UUID) *domain.ModuleRelationshipDefinition {
	for _, def := range domain.DefaultModuleRelationshipDefinitions(orgID) {
		if def.RelationshipKey == key {
			return def
		}
	}
	return nil
}

func customFieldEntityForCRMEntity(entityType domain.CRMEntityType) (domain.CustomFieldEntityType, bool) {
	switch entityType {
	case domain.CRMEntityTypeAccount:
		return domain.CustomFieldEntityAccount, true
	case domain.CRMEntityTypeContact:
		return domain.CustomFieldEntityContact, true
	case domain.CRMEntityTypeLead:
		return domain.CustomFieldEntityLead, true
	case domain.CRMEntityTypeDeal:
		return domain.CustomFieldEntityDeal, true
	case domain.CRMEntityTypeTicket:
		return domain.CustomFieldEntityTicket, true
	default:
		return "", false
	}
}
