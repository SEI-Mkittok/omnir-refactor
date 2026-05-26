package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AutomationStatus is the lifecycle state of an automation rule.
type AutomationStatus string

const (
	AutomationStatusDraft  AutomationStatus = "draft"
	AutomationStatusActive AutomationStatus = "active"
	AutomationStatusPaused AutomationStatus = "paused"
)

// TriggerType identifies what event fires an automation.
type TriggerType string

const (
	TriggerContactCreated   TriggerType = "contact_created"
	TriggerContactUpdated   TriggerType = "contact_updated"
	TriggerDealCreated      TriggerType = "deal_created"
	TriggerDealStageChanged TriggerType = "deal_stage_changed"
	TriggerActivityOverdue  TriggerType = "activity_overdue"
	TriggerTicketCreated    TriggerType = "ticket_created"
	TriggerManual           TriggerType = "manual"
)

// ConditionOperator is the comparison operator used in a condition.
type ConditionOperator string

const (
	ConditionOpEquals      ConditionOperator = "equals"
	ConditionOpNotEquals   ConditionOperator = "not_equals"
	ConditionOpContains    ConditionOperator = "contains"
	ConditionOpNotContains ConditionOperator = "not_contains"
	ConditionOpGreaterThan ConditionOperator = "greater_than"
	ConditionOpLessThan    ConditionOperator = "less_than"
	ConditionOpIsSet       ConditionOperator = "is_set"
	ConditionOpIsNotSet    ConditionOperator = "is_not_set"
)

// ActionType identifies what an automation action does.
type ActionType string

const (
	ActionAssignOwner      ActionType = "assign_owner"
	ActionSendEmail        ActionType = "send_email"
	ActionEnrollInSequence ActionType = "enroll_in_sequence"
	ActionCreateActivity   ActionType = "create_activity"
	ActionWebhook          ActionType = "webhook"
)

// AutomationRunStatus is the execution state of a single automation run.
type AutomationRunStatus string

const (
	RunStatusPending   AutomationRunStatus = "pending"
	RunStatusRunning   AutomationRunStatus = "running"
	RunStatusSucceeded AutomationRunStatus = "succeeded"
	RunStatusFailed    AutomationRunStatus = "failed"
)

// AutomationTrigger describes what fires the automation.
type AutomationTrigger struct {
	Type   TriggerType            `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// AutomationCondition is one logical filter applied before running actions.
type AutomationCondition struct {
	Field    string            `json:"field"`
	Operator ConditionOperator `json:"operator"`
	Value    interface{}       `json:"value,omitempty"`
}

// AutomationAction is one step performed when an automation fires.
type AutomationAction struct {
	Type   ActionType             `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// Automation is a workflow rule: trigger → conditions → actions.
type Automation struct {
	ID          uuid.UUID             `json:"id"`
	OrgID       uuid.UUID             `json:"org_id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Status      AutomationStatus      `json:"status"`
	Trigger     AutomationTrigger     `json:"trigger"`
	Conditions  []AutomationCondition `json:"conditions"`
	Actions     []AutomationAction    `json:"actions"`
	CreatedBy   *uuid.UUID            `json:"created_by,omitempty"`
	RunCount    int                   `json:"run_count"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// AutomationRun is a single execution of an automation.
type AutomationRun struct {
	ID           uuid.UUID           `json:"id"`
	AutomationID uuid.UUID           `json:"automation_id"`
	OrgID        uuid.UUID           `json:"org_id"`
	Status       AutomationRunStatus `json:"status"`
	EntityType   string              `json:"entity_type,omitempty"`
	EntityID     *uuid.UUID          `json:"entity_id,omitempty"`
	ErrorMessage string              `json:"error_message,omitempty"`
	StartedAt    *time.Time          `json:"started_at,omitempty"`
	FinishedAt   *time.Time          `json:"finished_at,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
}

// --- Request / Filter types ---

// CreateAutomationRequest is the payload to create a new automation.
type CreateAutomationRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Trigger     AutomationTrigger     `json:"trigger"`
	Conditions  []AutomationCondition `json:"conditions"`
	Actions     []AutomationAction    `json:"actions"`
}

func (r *CreateAutomationRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if len(r.Actions) == 0 {
		return fmt.Errorf("%w: at least one action is required", ErrValidation)
	}
	return ValidateAutomationConfig(&r.Trigger, r.Conditions, r.Actions)
}

// UpdateAutomationRequest is the payload to update an existing automation.
type UpdateAutomationRequest struct {
	Name        *string               `json:"name,omitempty"`
	Description *string               `json:"description,omitempty"`
	Status      *AutomationStatus     `json:"status,omitempty"`
	Trigger     *AutomationTrigger    `json:"trigger,omitempty"`
	Conditions  []AutomationCondition `json:"conditions,omitempty"`
	Actions     []AutomationAction    `json:"actions,omitempty"`
}

// AutomationFilter holds query parameters for listing automations.
type AutomationFilter struct {
	Status *AutomationStatus
	Page   int
	Limit  int
}

// AutomationRunFilter holds query parameters for listing runs.
type AutomationRunFilter struct {
	AutomationID uuid.UUID
	Page         int
	Limit        int
}

func (s AutomationStatus) IsValid() bool {
	return s == AutomationStatusDraft || s == AutomationStatusActive || s == AutomationStatusPaused
}

func (t TriggerType) IsValid() bool {
	switch t {
	case TriggerContactCreated, TriggerContactUpdated, TriggerDealCreated, TriggerDealStageChanged, TriggerActivityOverdue, TriggerTicketCreated, TriggerManual:
		return true
	default:
		return false
	}
}

func (op ConditionOperator) IsValid() bool {
	switch op {
	case ConditionOpEquals, ConditionOpNotEquals, ConditionOpContains, ConditionOpNotContains, ConditionOpGreaterThan, ConditionOpLessThan, ConditionOpIsSet, ConditionOpIsNotSet:
		return true
	default:
		return false
	}
}

func (a ActionType) IsValid() bool {
	switch a {
	case ActionAssignOwner, ActionSendEmail, ActionEnrollInSequence, ActionCreateActivity, ActionWebhook:
		return true
	default:
		return false
	}
}

func ValidateAutomationConfig(trigger *AutomationTrigger, conditions []AutomationCondition, actions []AutomationAction) error {
	if trigger == nil || !trigger.Type.IsValid() {
		return fmt.Errorf("%w: trigger type is invalid", ErrValidation)
	}
	for _, condition := range conditions {
		if strings.TrimSpace(condition.Field) == "" {
			return fmt.Errorf("%w: condition field is required", ErrValidation)
		}
		if !condition.Operator.IsValid() {
			return fmt.Errorf("%w: condition operator is invalid", ErrValidation)
		}
	}
	for _, action := range actions {
		if !action.Type.IsValid() {
			return fmt.Errorf("%w: action type is invalid", ErrValidation)
		}
		if err := ValidateAutomationAction(action); err != nil {
			return err
		}
	}
	return nil
}

func ValidateAutomationAction(action AutomationAction) error {
	cfg := action.Config
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	switch action.Type {
	case ActionAssignOwner:
		if strings.TrimSpace(configStringValue(cfg, "owner_id")) == "" {
			return fmt.Errorf("%w: assign_owner.owner_id is required", ErrValidation)
		}
	case ActionSendEmail:
		if strings.TrimSpace(configStringValue(cfg, "subject")) == "" {
			return fmt.Errorf("%w: send_email.subject is required", ErrValidation)
		}
	case ActionEnrollInSequence:
		if strings.TrimSpace(configStringValue(cfg, "sequence_id")) == "" {
			return fmt.Errorf("%w: enroll_in_sequence.sequence_id is required", ErrValidation)
		}
	case ActionCreateActivity:
		if strings.TrimSpace(configStringValue(cfg, "subject")) == "" && strings.TrimSpace(configStringValue(cfg, "title")) == "" {
			return fmt.Errorf("%w: create_activity.subject is required", ErrValidation)
		}
	case ActionWebhook:
		if strings.TrimSpace(configStringValue(cfg, "url")) == "" {
			return fmt.Errorf("%w: webhook.url is required", ErrValidation)
		}
	}
	return nil
}

func configStringValue(cfg map[string]interface{}, key string) string {
	v, ok := cfg[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
