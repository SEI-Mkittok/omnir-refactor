package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

func validateModuleLayoutCreate(
	ctx context.Context,
	layouts repository.ModuleLayoutRepository,
	cfDefs repository.CustomFieldDefinitionRepository,
	entityType domain.CustomFieldEntityType,
	payload any,
) error {
	if layouts == nil {
		return nil
	}
	defs, err := customFieldDefinitionsForLayout(ctx, cfDefs, entityType)
	if err != nil {
		return err
	}
	rawBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &raw); err != nil {
		return err
	}
	if rawCustom := raw["custom_fields"]; cfDefs != nil {
		if err := domain.ValidateCustomFields(rawCustom, defs); err != nil {
			return fmt.Errorf("%w: %s", domain.ErrValidation, err.Error())
		}
	}
	orgID, _ := domain.OrgIDFromContext(ctx)
	layout, err := layouts.GetByEntity(ctx, orgID, entityType)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			layout = domain.DefaultModuleLayout(orgID, entityType, defs)
		} else {
			return err
		}
	}
	layout = domain.NormalizeModuleLayout(layout, defs)
	return domain.ValidateModuleLayoutRequiredValues(layout, raw)
}

func customFieldDefinitionsForLayout(ctx context.Context, cfDefs repository.CustomFieldDefinitionRepository, entityType domain.CustomFieldEntityType) ([]*domain.CustomFieldDefinition, error) {
	if cfDefs == nil {
		return nil, nil
	}
	return cfDefs.List(ctx, domain.CustomFieldDefinitionFilter{EntityType: &entityType})
}
