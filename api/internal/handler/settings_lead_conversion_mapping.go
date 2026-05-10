package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type LeadConversionMappingSettingsHandler struct {
	repo repository.LeadConversionMappingRepository
	defs repository.CustomFieldDefinitionRepository
}

func NewLeadConversionMappingSettingsHandler(
	repo repository.LeadConversionMappingRepository,
	defs repository.CustomFieldDefinitionRepository,
) *LeadConversionMappingSettingsHandler {
	return &LeadConversionMappingSettingsHandler{repo: repo, defs: defs}
}

func (h *LeadConversionMappingSettingsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.With(requireAdminRole).Get("/", h.List)
	r.With(requireAdminRole).Put("/", h.Replace)
	return r
}

func (h *LeadConversionMappingSettingsHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rows, err := h.repo.List(r.Context(), claims.OrgID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": rows})
}

func (h *LeadConversionMappingSettingsHandler) Replace(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Mappings []domain.LeadConversionMapping `json:"mappings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	for i := range req.Mappings {
		req.Mappings[i].LeadField = strings.TrimSpace(req.Mappings[i].LeadField)
		req.Mappings[i].TargetField = strings.TrimSpace(req.Mappings[i].TargetField)
		if req.Mappings[i].LeadField == "" || req.Mappings[i].TargetField == "" || !req.Mappings[i].TargetEntity.IsValid() {
			writeError(w, http.StatusUnprocessableEntity, "each mapping requires lead_field, target_entity, and target_field")
			return
		}
	}

	if h.defs != nil {
		byEntity, err := h.loadDefinitionsByEntity(r.Context())
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		for _, row := range req.Mappings {
			if row.LeadField == "" || row.TargetField == "" {
				continue
			}
			if err := validateLeadFieldMapping(row, byEntity); err != nil {
				writeError(w, http.StatusUnprocessableEntity, err.Error())
				return
			}
		}
	}

	rows, err := h.repo.Replace(r.Context(), claims.OrgID, req.Mappings)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": rows})
}

func (h *LeadConversionMappingSettingsHandler) loadDefinitionsByEntity(ctx context.Context) (map[domain.CustomFieldEntityType]map[string]struct{}, error) {
	entityTypes := []domain.CustomFieldEntityType{
		domain.CustomFieldEntityLead,
		domain.CustomFieldEntityContact,
		domain.CustomFieldEntityAccount,
		domain.CustomFieldEntityDeal,
	}
	out := map[domain.CustomFieldEntityType]map[string]struct{}{}
	for _, entityType := range entityTypes {
		et := entityType
		rows, err := h.defs.List(ctx, domain.CustomFieldDefinitionFilter{EntityType: &et})
		if err != nil {
			return nil, err
		}
		set := map[string]struct{}{}
		for _, row := range rows {
			if row == nil {
				continue
			}
			set[row.Name] = struct{}{}
		}
		out[entityType] = set
	}
	return out, nil
}

func validateLeadFieldMapping(
	row domain.LeadConversionMapping,
	customFields map[domain.CustomFieldEntityType]map[string]struct{},
) error {
	leadStandard := []string{"first_name", "last_name", "email", "phone", "company", "lead_source", "lead_score", "status"}
	contactStandard := []string{"first_name", "last_name", "email", "phone", "lead_source"}
	accountStandard := []string{"name", "domain"}
	dealStandard := []string{"title", "value_cents", "currency"}

	if err := validateMappingField("lead_field", row.LeadField, leadStandard, customFields[domain.CustomFieldEntityLead]); err != nil {
		return err
	}

	switch row.TargetEntity {
	case domain.LeadConversionTargetContact:
		return validateMappingField("target_field", row.TargetField, contactStandard, customFields[domain.CustomFieldEntityContact])
	case domain.LeadConversionTargetAccount:
		return validateMappingField("target_field", row.TargetField, accountStandard, customFields[domain.CustomFieldEntityAccount])
	case domain.LeadConversionTargetDeal:
		return validateMappingField("target_field", row.TargetField, dealStandard, customFields[domain.CustomFieldEntityDeal])
	default:
		return fmt.Errorf("target_entity must be one of contact, account, deal")
	}
}

func validateMappingField(fieldName, value string, standard []string, custom map[string]struct{}) error {
	if slices.Contains(standard, value) {
		return nil
	}
	if strings.HasPrefix(value, "custom:") {
		key := strings.TrimSpace(strings.TrimPrefix(value, "custom:"))
		if key == "" {
			return fmt.Errorf("%s custom field key is empty", fieldName)
		}
		if _, ok := custom[key]; !ok {
			return fmt.Errorf("%s %q is not a valid custom field", fieldName, value)
		}
		return nil
	}
	if _, ok := custom[value]; ok {
		return nil
	}
	return fmt.Errorf("%s %q is not supported", fieldName, value)
}
