package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ModuleLayoutFieldSource string

const (
	ModuleLayoutFieldSourceStandard ModuleLayoutFieldSource = "standard"
	ModuleLayoutFieldSourceCustom   ModuleLayoutFieldSource = "custom"
)

type ModuleLayoutField struct {
	Source      ModuleLayoutFieldSource `json:"source"`
	FieldKey    string                  `json:"field_key"`
	Label       string                  `json:"label"`
	Visible     bool                    `json:"visible"`
	Required    bool                    `json:"required"`
	Order       int                     `json:"order"`
	QuickCreate bool                    `json:"quick_create"`
	MassEdit    bool                    `json:"mass_edit"`
	Header      bool                    `json:"header"`
	KeyField    bool                    `json:"key_field"`
}

type ModuleLayoutBlock struct {
	ID     string              `json:"id"`
	Label  string              `json:"label"`
	Order  int                 `json:"order"`
	Fields []ModuleLayoutField `json:"fields"`
}

type ModuleLayout struct {
	ID         uuid.UUID             `json:"id"`
	OrgID      uuid.UUID             `json:"org_id"`
	EntityType CustomFieldEntityType `json:"entity_type"`
	Blocks     []ModuleLayoutBlock   `json:"blocks"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

type ModuleStandardField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	QuickCreate bool   `json:"quick_create"`
	MassEdit    bool   `json:"mass_edit"`
	Header      bool   `json:"header"`
	KeyField    bool   `json:"key_field"`
}

var moduleStandardFields = map[CustomFieldEntityType][]ModuleStandardField{
	CustomFieldEntityAccount: {
		{Key: "name", Label: "Company name", Required: true, QuickCreate: true, MassEdit: true, Header: true, KeyField: true},
		{Key: "domain", Label: "Domain", QuickCreate: true, MassEdit: true, Header: true},
		{Key: "industry", Label: "Industry", QuickCreate: true, MassEdit: true},
		{Key: "size", Label: "Company size", QuickCreate: true, MassEdit: true},
		{Key: "owner_id", Label: "Owner", MassEdit: true},
		{Key: "tags", Label: "Tags", MassEdit: true},
	},
	CustomFieldEntityContact: {
		{Key: "first_name", Label: "First name", Required: true, QuickCreate: true, MassEdit: true, Header: true, KeyField: true},
		{Key: "last_name", Label: "Last name", Required: true, QuickCreate: true, MassEdit: true, Header: true, KeyField: true},
		{Key: "email", Label: "Email", QuickCreate: true, MassEdit: true, Header: true},
		{Key: "phone", Label: "Phone", QuickCreate: true, MassEdit: true},
		{Key: "account_id", Label: "Account", QuickCreate: true, MassEdit: true},
		{Key: "stage", Label: "Stage", QuickCreate: true, MassEdit: true},
		{Key: "lead_source", Label: "Lead source", QuickCreate: true, MassEdit: true},
		{Key: "owner_id", Label: "Owner", MassEdit: true},
		{Key: "tags", Label: "Tags", MassEdit: true},
	},
	CustomFieldEntityLead: {
		{Key: "first_name", Label: "First name", Required: true, QuickCreate: true, MassEdit: true, Header: true, KeyField: true},
		{Key: "last_name", Label: "Last name", Required: true, QuickCreate: true, MassEdit: true, Header: true, KeyField: true},
		{Key: "email", Label: "Email", QuickCreate: true, MassEdit: true, Header: true},
		{Key: "phone", Label: "Phone", QuickCreate: true, MassEdit: true},
		{Key: "company", Label: "Company", QuickCreate: true, MassEdit: true},
		{Key: "lead_source", Label: "Lead source", QuickCreate: true, MassEdit: true},
		{Key: "status", Label: "Status", QuickCreate: true, MassEdit: true},
		{Key: "owner_id", Label: "Owner", MassEdit: true},
	},
	CustomFieldEntityDeal: {
		{Key: "title", Label: "Deal title", Required: true, QuickCreate: true, MassEdit: true, Header: true, KeyField: true},
		{Key: "value_cents", Label: "Value", QuickCreate: true, MassEdit: true, Header: true},
		{Key: "stage", Label: "Stage", QuickCreate: true, MassEdit: true},
		{Key: "probability", Label: "Probability", MassEdit: true},
		{Key: "expected_close_date", Label: "Expected close date", QuickCreate: true, MassEdit: true},
		{Key: "account_id", Label: "Account", QuickCreate: true, MassEdit: true},
		{Key: "contact_id", Label: "Contact", QuickCreate: true, MassEdit: true},
		{Key: "owner_id", Label: "Owner", MassEdit: true},
		{Key: "pipeline_id", Label: "Pipeline", Required: true, QuickCreate: true, MassEdit: true},
		{Key: "tags", Label: "Tags", MassEdit: true},
	},
	CustomFieldEntityTicket: {
		{Key: "subject", Label: "Subject", Required: true, QuickCreate: true, MassEdit: true, Header: true, KeyField: true},
		{Key: "description", Label: "Description", QuickCreate: true, MassEdit: true},
		{Key: "status", Label: "Status", QuickCreate: true, MassEdit: true, Header: true},
		{Key: "priority", Label: "Priority", QuickCreate: true, MassEdit: true, Header: true},
		{Key: "assignee_id", Label: "Assignee", QuickCreate: true, MassEdit: true},
		{Key: "contact_id", Label: "Contact", QuickCreate: true, MassEdit: true},
		{Key: "account_id", Label: "Account", QuickCreate: true, MassEdit: true},
		{Key: "source", Label: "Source", MassEdit: true},
		{Key: "tags", Label: "Tags", MassEdit: true},
	},
}

func ModuleStandardFields(entityType CustomFieldEntityType) []ModuleStandardField {
	fields := moduleStandardFields[entityType]
	out := make([]ModuleStandardField, len(fields))
	copy(out, fields)
	return out
}

func DefaultModuleLayout(orgID uuid.UUID, entityType CustomFieldEntityType, customFields []*CustomFieldDefinition) *ModuleLayout {
	layout := &ModuleLayout{OrgID: orgID, EntityType: entityType}
	standard := ModuleLayoutBlock{ID: "main", Label: "Main", Order: 0}
	for i, field := range ModuleStandardFields(entityType) {
		standard.Fields = append(standard.Fields, ModuleLayoutField{
			Source:      ModuleLayoutFieldSourceStandard,
			FieldKey:    field.Key,
			Label:       field.Label,
			Visible:     true,
			Required:    field.Required,
			Order:       i,
			QuickCreate: field.QuickCreate,
			MassEdit:    field.MassEdit,
			Header:      field.Header,
			KeyField:    field.KeyField,
		})
	}
	layout.Blocks = append(layout.Blocks, standard)
	if len(customFields) > 0 {
		custom := ModuleLayoutBlock{ID: "custom", Label: "Custom Fields", Order: 1}
		for i, field := range customFields {
			if field == nil {
				continue
			}
			custom.Fields = append(custom.Fields, ModuleLayoutField{
				Source:      ModuleLayoutFieldSourceCustom,
				FieldKey:    field.Name,
				Label:       field.Label,
				Visible:     true,
				Required:    field.Required,
				Order:       i,
				QuickCreate: true,
				MassEdit:    true,
			})
		}
		layout.Blocks = append(layout.Blocks, custom)
	}
	return layout
}

func NormalizeModuleLayout(layout *ModuleLayout, customFields []*CustomFieldDefinition) *ModuleLayout {
	if layout == nil {
		return nil
	}
	out := *layout
	out.Blocks = append([]ModuleLayoutBlock(nil), layout.Blocks...)
	valid := validLayoutFields(out.EntityType, customFields)
	seen := map[string]bool{}

	for bIdx := range out.Blocks {
		block := out.Blocks[bIdx]
		block.Fields = append([]ModuleLayoutField(nil), block.Fields...)
		filtered := make([]ModuleLayoutField, 0, len(block.Fields))
		for _, field := range block.Fields {
			key := layoutQualifiedFieldKey(field.Source, field.FieldKey)
			if !valid[key] || seen[key] {
				continue
			}
			if field.Label == "" {
				field.Label = defaultLayoutFieldLabel(out.EntityType, customFields, field.Source, field.FieldKey)
			}
			seen[key] = true
			filtered = append(filtered, field)
		}
		block.Fields = filtered
		out.Blocks[bIdx] = block
	}

	if len(out.Blocks) == 0 {
		out.Blocks = []ModuleLayoutBlock{{ID: "main", Label: "Main", Order: 0}}
	}

	appendMissing := func(field ModuleLayoutField) {
		key := layoutQualifiedFieldKey(field.Source, field.FieldKey)
		if seen[key] {
			return
		}
		seen[key] = true
		blockIdx := 0
		if field.Source == ModuleLayoutFieldSourceCustom {
			blockIdx = ensureCustomBlock(&out)
		}
		field.Order = len(out.Blocks[blockIdx].Fields)
		out.Blocks[blockIdx].Fields = append(out.Blocks[blockIdx].Fields, field)
	}

	for _, field := range ModuleStandardFields(out.EntityType) {
		appendMissing(ModuleLayoutField{
			Source:      ModuleLayoutFieldSourceStandard,
			FieldKey:    field.Key,
			Label:       field.Label,
			Visible:     true,
			Required:    field.Required,
			QuickCreate: field.QuickCreate,
			MassEdit:    field.MassEdit,
			Header:      field.Header,
			KeyField:    field.KeyField,
		})
	}
	for _, field := range customFields {
		if field == nil {
			continue
		}
		appendMissing(ModuleLayoutField{
			Source:      ModuleLayoutFieldSourceCustom,
			FieldKey:    field.Name,
			Label:       field.Label,
			Visible:     true,
			Required:    field.Required,
			QuickCreate: true,
			MassEdit:    true,
		})
	}

	sort.SliceStable(out.Blocks, func(i, j int) bool { return out.Blocks[i].Order < out.Blocks[j].Order })
	for bIdx := range out.Blocks {
		sort.SliceStable(out.Blocks[bIdx].Fields, func(i, j int) bool {
			return out.Blocks[bIdx].Fields[i].Order < out.Blocks[bIdx].Fields[j].Order
		})
	}
	return &out
}

func ensureCustomBlock(layout *ModuleLayout) int {
	for i, block := range layout.Blocks {
		if block.ID == "custom" {
			return i
		}
	}
	layout.Blocks = append(layout.Blocks, ModuleLayoutBlock{ID: "custom", Label: "Custom Fields", Order: len(layout.Blocks)})
	return len(layout.Blocks) - 1
}

func ValidateModuleLayout(layout *ModuleLayout, customFields []*CustomFieldDefinition) error {
	if layout == nil {
		return fmt.Errorf("%w: layout is required", ErrValidation)
	}
	if !layout.EntityType.IsValid() {
		return fmt.Errorf("%w: entity_type is invalid", ErrValidation)
	}
	if len(layout.Blocks) == 0 {
		return fmt.Errorf("%w: at least one layout block is required", ErrValidation)
	}
	valid := validLayoutFields(layout.EntityType, customFields)
	hardRequired := hardRequiredLayoutFields(layout.EntityType)
	requiredCustom := requiredCustomLayoutFields(customFields)
	seenBlocks := map[string]bool{}
	seenFields := map[string]bool{}
	for _, block := range layout.Blocks {
		if strings.TrimSpace(block.ID) == "" {
			return fmt.Errorf("%w: block id is required", ErrValidation)
		}
		if strings.TrimSpace(block.Label) == "" {
			return fmt.Errorf("%w: block label is required", ErrValidation)
		}
		if seenBlocks[block.ID] {
			return fmt.Errorf("%w: duplicate block %q", ErrValidation, block.ID)
		}
		seenBlocks[block.ID] = true
		for _, field := range block.Fields {
			if field.Source != ModuleLayoutFieldSourceStandard && field.Source != ModuleLayoutFieldSourceCustom {
				return fmt.Errorf("%w: invalid field source %q", ErrValidation, field.Source)
			}
			if strings.TrimSpace(field.FieldKey) == "" {
				return fmt.Errorf("%w: field key is required", ErrValidation)
			}
			key := layoutQualifiedFieldKey(field.Source, field.FieldKey)
			if !valid[key] {
				return fmt.Errorf("%w: unknown layout field %q", ErrValidation, field.FieldKey)
			}
			if seenFields[key] {
				return fmt.Errorf("%w: duplicate layout field %q", ErrValidation, field.FieldKey)
			}
			seenFields[key] = true
			if field.Required && !field.Visible {
				return fmt.Errorf("%w: required field %q cannot be hidden", ErrValidation, field.FieldKey)
			}
			if hardRequired[key] || requiredCustom[key] {
				if !field.Visible {
					return fmt.Errorf("%w: required field %q cannot be hidden", ErrValidation, field.FieldKey)
				}
				if !field.Required {
					return fmt.Errorf("%w: required field %q cannot be relaxed", ErrValidation, field.FieldKey)
				}
				if !field.QuickCreate {
					return fmt.Errorf("%w: required field %q must remain in quick create", ErrValidation, field.FieldKey)
				}
			}
		}
	}
	for key := range hardRequired {
		if !seenFields[key] {
			return fmt.Errorf("%w: required system field %q is missing", ErrValidation, strings.TrimPrefix(key, string(ModuleLayoutFieldSourceStandard)+":"))
		}
	}
	for key := range requiredCustom {
		if !seenFields[key] {
			return fmt.Errorf("%w: required custom field %q is missing", ErrValidation, strings.TrimPrefix(key, string(ModuleLayoutFieldSourceCustom)+":"))
		}
	}
	return nil
}

func ValidateModuleLayoutRequiredValues(layout *ModuleLayout, raw map[string]json.RawMessage) error {
	if layout == nil {
		return nil
	}
	var custom map[string]json.RawMessage
	if rawCustom, ok := raw["custom_fields"]; ok && len(bytes.TrimSpace(rawCustom)) > 0 && !bytes.Equal(bytes.TrimSpace(rawCustom), []byte("null")) {
		_ = json.Unmarshal(rawCustom, &custom)
	}
	for _, block := range layout.Blocks {
		for _, field := range block.Fields {
			if !field.Visible || !field.Required {
				continue
			}
			if field.Source == ModuleLayoutFieldSourceCustom {
				if isJSONEmpty(custom[field.FieldKey]) {
					return fmt.Errorf("%w: %s is required", ErrValidation, field.FieldKey)
				}
				continue
			}
			if isJSONEmpty(raw[field.FieldKey]) {
				return fmt.Errorf("%w: %s is required", ErrValidation, field.FieldKey)
			}
		}
	}
	return nil
}

func validLayoutFields(entityType CustomFieldEntityType, customFields []*CustomFieldDefinition) map[string]bool {
	valid := map[string]bool{}
	for _, field := range ModuleStandardFields(entityType) {
		valid[layoutQualifiedFieldKey(ModuleLayoutFieldSourceStandard, field.Key)] = true
	}
	for _, field := range customFields {
		if field != nil {
			valid[layoutQualifiedFieldKey(ModuleLayoutFieldSourceCustom, field.Name)] = true
		}
	}
	return valid
}

func hardRequiredLayoutFields(entityType CustomFieldEntityType) map[string]bool {
	required := map[string]bool{}
	for fieldKey := range moduleBackendHardRequiredFields[entityType] {
		required[layoutQualifiedFieldKey(ModuleLayoutFieldSourceStandard, fieldKey)] = true
	}
	return required
}

var moduleBackendHardRequiredFields = map[CustomFieldEntityType]map[string]bool{
	CustomFieldEntityAccount: {
		"name": true,
	},
	CustomFieldEntityContact: {
		"first_name": true,
		"last_name":  true,
	},
	CustomFieldEntityLead: {
		"first_name": true,
		"last_name":  true,
	},
	CustomFieldEntityDeal: {
		"title":       true,
		"pipeline_id": true,
	},
	CustomFieldEntityTicket: {
		"subject": true,
	},
}

func requiredCustomLayoutFields(customFields []*CustomFieldDefinition) map[string]bool {
	required := map[string]bool{}
	for _, field := range customFields {
		if field != nil && field.Required {
			required[layoutQualifiedFieldKey(ModuleLayoutFieldSourceCustom, field.Name)] = true
		}
	}
	return required
}

func layoutQualifiedFieldKey(source ModuleLayoutFieldSource, key string) string {
	return string(source) + ":" + key
}

func defaultLayoutFieldLabel(entityType CustomFieldEntityType, customFields []*CustomFieldDefinition, source ModuleLayoutFieldSource, key string) string {
	if source == ModuleLayoutFieldSourceCustom {
		for _, field := range customFields {
			if field != nil && field.Name == key {
				return field.Label
			}
		}
		return key
	}
	for _, field := range ModuleStandardFields(entityType) {
		if field.Key == key {
			return field.Label
		}
	}
	return key
}

func isJSONEmpty(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) || bytes.Equal(trimmed, []byte("\"\"")) {
		return true
	}
	var s string
	if err := json.Unmarshal(trimmed, &s); err == nil {
		return strings.TrimSpace(s) == ""
	}
	var arr []any
	if err := json.Unmarshal(trimmed, &arr); err == nil {
		return len(arr) == 0
	}
	return false
}

type RelationshipCardinality string

const (
	RelationshipCardinalityOneToOne   RelationshipCardinality = "one_to_one"
	RelationshipCardinalityManyToOne  RelationshipCardinality = "many_to_one"
	RelationshipCardinalityOneToMany  RelationshipCardinality = "one_to_many"
	RelationshipCardinalityManyToMany RelationshipCardinality = "many_to_many"
)

func (c RelationshipCardinality) IsValid() bool {
	switch c {
	case RelationshipCardinalityOneToOne, RelationshipCardinalityManyToOne, RelationshipCardinalityOneToMany, RelationshipCardinalityManyToMany:
		return true
	default:
		return false
	}
}

type RelationshipStorageStrategy string

const (
	RelationshipStorageNative         RelationshipStorageStrategy = "native"
	RelationshipStorageCRMEntityLinks RelationshipStorageStrategy = "crm_entity_links"
)

func (s RelationshipStorageStrategy) IsValid() bool {
	return s == RelationshipStorageNative || s == RelationshipStorageCRMEntityLinks
}

type ModuleRelationshipDefinition struct {
	ID              uuid.UUID                   `json:"id"`
	OrgID           uuid.UUID                   `json:"org_id"`
	RelationshipKey string                      `json:"relationship_key"`
	FromEntityType  CustomFieldEntityType       `json:"from_entity_type"`
	ToEntityType    CustomFieldEntityType       `json:"to_entity_type"`
	Label           string                      `json:"label"`
	Cardinality     RelationshipCardinality     `json:"cardinality"`
	StorageStrategy RelationshipStorageStrategy `json:"storage_strategy"`
	IsEnabled       bool                        `json:"is_enabled"`
	SystemLocked    bool                        `json:"system_locked"`
	OrderIdx        int                         `json:"order_idx"`
	Metadata        json.RawMessage             `json:"metadata,omitempty"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
}

type ModuleRelationshipDefinitionFilter struct {
	OrgID      uuid.UUID
	EntityType *CustomFieldEntityType
}

func DefaultModuleRelationshipDefinitions(orgID uuid.UUID) []*ModuleRelationshipDefinition {
	rows := []*ModuleRelationshipDefinition{
		{OrgID: orgID, RelationshipKey: "account_contacts", FromEntityType: CustomFieldEntityAccount, ToEntityType: CustomFieldEntityContact, Label: "Account contacts", Cardinality: RelationshipCardinalityOneToMany, StorageStrategy: RelationshipStorageNative, IsEnabled: true, SystemLocked: true, OrderIdx: 10},
		{OrgID: orgID, RelationshipKey: "account_hierarchy", FromEntityType: CustomFieldEntityAccount, ToEntityType: CustomFieldEntityAccount, Label: "Account hierarchy", Cardinality: RelationshipCardinalityOneToMany, StorageStrategy: RelationshipStorageNative, IsEnabled: true, SystemLocked: true, OrderIdx: 20},
		{OrgID: orgID, RelationshipKey: "deal_contacts", FromEntityType: CustomFieldEntityDeal, ToEntityType: CustomFieldEntityContact, Label: "Deal contacts", Cardinality: RelationshipCardinalityManyToMany, StorageStrategy: RelationshipStorageNative, IsEnabled: true, SystemLocked: true, OrderIdx: 30},
		{OrgID: orgID, RelationshipKey: "deal_account", FromEntityType: CustomFieldEntityDeal, ToEntityType: CustomFieldEntityAccount, Label: "Deal account", Cardinality: RelationshipCardinalityManyToOne, StorageStrategy: RelationshipStorageNative, IsEnabled: true, SystemLocked: true, OrderIdx: 40},
		{OrgID: orgID, RelationshipKey: "ticket_contact", FromEntityType: CustomFieldEntityTicket, ToEntityType: CustomFieldEntityContact, Label: "Ticket contact", Cardinality: RelationshipCardinalityManyToOne, StorageStrategy: RelationshipStorageNative, IsEnabled: true, SystemLocked: true, OrderIdx: 50},
		{OrgID: orgID, RelationshipKey: "ticket_account", FromEntityType: CustomFieldEntityTicket, ToEntityType: CustomFieldEntityAccount, Label: "Ticket account", Cardinality: RelationshipCardinalityManyToOne, StorageStrategy: RelationshipStorageNative, IsEnabled: true, SystemLocked: true, OrderIdx: 60},
		{OrgID: orgID, RelationshipKey: "lead_conversion", FromEntityType: CustomFieldEntityLead, ToEntityType: CustomFieldEntityContact, Label: "Lead conversion", Cardinality: RelationshipCardinalityOneToOne, StorageStrategy: RelationshipStorageNative, IsEnabled: true, SystemLocked: true, OrderIdx: 70},
	}
	for _, row := range rows {
		row.Metadata = json.RawMessage(`{}`)
	}
	return rows
}

func MergeDefaultRelationshipDefinitions(orgID uuid.UUID, stored []*ModuleRelationshipDefinition) []*ModuleRelationshipDefinition {
	byKey := map[string]*ModuleRelationshipDefinition{}
	for _, row := range DefaultModuleRelationshipDefinitions(orgID) {
		byKey[row.RelationshipKey] = row
	}
	for _, row := range stored {
		if row == nil {
			continue
		}
		if IsSystemRelationshipKey(row.RelationshipKey) {
			row.SystemLocked = true
			row.StorageStrategy = byKey[row.RelationshipKey].StorageStrategy
			row.FromEntityType = byKey[row.RelationshipKey].FromEntityType
			row.ToEntityType = byKey[row.RelationshipKey].ToEntityType
			row.Cardinality = byKey[row.RelationshipKey].Cardinality
		}
		byKey[row.RelationshipKey] = row
	}
	out := make([]*ModuleRelationshipDefinition, 0, len(byKey))
	for _, row := range byKey {
		out = append(out, row)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].OrderIdx == out[j].OrderIdx {
			return out[i].RelationshipKey < out[j].RelationshipKey
		}
		return out[i].OrderIdx < out[j].OrderIdx
	})
	return out
}

func IsSystemRelationshipKey(key string) bool {
	for _, row := range DefaultModuleRelationshipDefinitions(uuid.Nil) {
		if row.RelationshipKey == key {
			return true
		}
	}
	return false
}

func ValidateModuleRelationshipDefinition(def *ModuleRelationshipDefinition) error {
	if def == nil {
		return fmt.Errorf("%w: relationship definition is required", ErrValidation)
	}
	def.RelationshipKey = strings.TrimSpace(def.RelationshipKey)
	def.Label = strings.TrimSpace(def.Label)
	if def.RelationshipKey == "" {
		return fmt.Errorf("%w: relationship_key is required", ErrValidation)
	}
	if def.Label == "" {
		return fmt.Errorf("%w: label is required", ErrValidation)
	}
	if _, ok := CRMEntityTypeForCustomFieldEntity(def.FromEntityType); !ok {
		return fmt.Errorf("%w: relationship entity types must be account, contact, lead, deal, ticket, or quote", ErrValidation)
	}
	if _, ok := CRMEntityTypeForCustomFieldEntity(def.ToEntityType); !ok {
		return fmt.Errorf("%w: relationship entity types must be account, contact, lead, deal, ticket, or quote", ErrValidation)
	}
	if !def.Cardinality.IsValid() {
		return fmt.Errorf("%w: cardinality is invalid", ErrValidation)
	}
	if !def.StorageStrategy.IsValid() {
		return fmt.Errorf("%w: storage_strategy is invalid", ErrValidation)
	}
	if len(def.Metadata) == 0 {
		def.Metadata = json.RawMessage(`{}`)
	}
	if def.StorageStrategy == RelationshipStorageCRMEntityLinks && def.FromEntityType == def.ToEntityType && def.Cardinality != RelationshipCardinalityManyToMany {
		return fmt.Errorf("%w: custom same-module relationships must be many_to_many", ErrValidation)
	}
	if IsSystemRelationshipKey(def.RelationshipKey) {
		def.SystemLocked = true
	}
	return nil
}

func ACLModuleForCustomFieldEntity(entityType CustomFieldEntityType) (ACLModule, bool) {
	switch entityType {
	case CustomFieldEntityAccount:
		return ACLModuleAccounts, true
	case CustomFieldEntityContact:
		return ACLModuleContacts, true
	case CustomFieldEntityLead:
		return ACLModuleLeads, true
	case CustomFieldEntityDeal:
		return ACLModuleDeals, true
	case CustomFieldEntityTicket:
		return ACLModuleTickets, true
	case CustomFieldEntityQuote:
		return ACLModuleQuotes, true
	case CustomFieldEntityKBArticle:
		return ACLModuleKB, true
	default:
		return "", false
	}
}

func CRMEntityTypeForCustomFieldEntity(entityType CustomFieldEntityType) (CRMEntityType, bool) {
	switch entityType {
	case CustomFieldEntityAccount:
		return CRMEntityTypeAccount, true
	case CustomFieldEntityContact:
		return CRMEntityTypeContact, true
	case CustomFieldEntityLead:
		return CRMEntityTypeLead, true
	case CustomFieldEntityDeal:
		return CRMEntityTypeDeal, true
	case CustomFieldEntityTicket:
		return CRMEntityTypeTicket, true
	case CustomFieldEntityQuote:
		return CRMEntityTypeQuote, true
	default:
		return "", false
	}
}
