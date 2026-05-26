package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SchedulerJobStatus struct {
	Key          string     `json:"key"`
	Name         string     `json:"name"`
	Interval     string     `json:"interval"`
	Enabled      bool       `json:"enabled"`
	LastStarted  *time.Time `json:"last_started_at,omitempty"`
	LastFinished *time.Time `json:"last_finished_at,omitempty"`
	LastResult   *string    `json:"last_result,omitempty"`
	ErrorText    *string    `json:"error_text,omitempty"`
	NextRun      *time.Time `json:"next_run_at,omitempty"`
}

type SchedulerJobRun struct {
	ID         uuid.UUID  `json:"id"`
	JobKey     string     `json:"job_key"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	ErrorText  *string    `json:"error_text,omitempty"`
}

type WebformTargetModule string

const (
	WebformTargetLead    WebformTargetModule = "lead"
	WebformTargetContact WebformTargetModule = "contact"
	WebformTargetTicket  WebformTargetModule = "ticket"
)

func (m WebformTargetModule) IsValid() bool {
	switch m {
	case WebformTargetLead, WebformTargetContact, WebformTargetTicket:
		return true
	default:
		return false
	}
}

type WebformStatus string

const (
	WebformStatusActive   WebformStatus = "active"
	WebformStatusInactive WebformStatus = "inactive"
)

func (s WebformStatus) IsValid() bool {
	return s == WebformStatusActive || s == WebformStatusInactive
}

type WebformField struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	TargetField string   `json:"target_field,omitempty"`
	Options     []string `json:"options,omitempty"`
}

type Webform struct {
	ID             uuid.UUID           `json:"id"`
	OrgID          uuid.UUID           `json:"org_id"`
	Name           string              `json:"name"`
	PublicID       string              `json:"public_id"`
	Status         WebformStatus       `json:"status"`
	TargetModule   WebformTargetModule `json:"target_module"`
	CampaignID     *uuid.UUID          `json:"campaign_id,omitempty"`
	ReturnURL      *string             `json:"return_url,omitempty"`
	SuccessMessage string              `json:"success_message"`
	SpamTrapField  string              `json:"spam_trap_field"`
	Fields         []WebformField      `json:"fields"`
	CreatedBy      *uuid.UUID          `json:"created_by,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	DeletedAt      *time.Time          `json:"deleted_at,omitempty"`
}

func (f *Webform) Validate() error {
	if strings.TrimSpace(f.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if !f.TargetModule.IsValid() {
		return fmt.Errorf("%w: target_module is invalid", ErrValidation)
	}
	if f.Status != "" && !f.Status.IsValid() {
		return fmt.Errorf("%w: status is invalid", ErrValidation)
	}
	if len(f.Fields) == 0 {
		return fmt.Errorf("%w: at least one field is required", ErrValidation)
	}
	seen := map[string]bool{}
	for _, field := range f.Fields {
		key := strings.TrimSpace(field.Key)
		if key == "" {
			return fmt.Errorf("%w: field key is required", ErrValidation)
		}
		if seen[key] {
			return fmt.Errorf("%w: duplicate field key %q", ErrValidation, key)
		}
		seen[key] = true
		if strings.TrimSpace(field.Label) == "" {
			return fmt.Errorf("%w: field label is required", ErrValidation)
		}
	}
	return nil
}

type WebformPatch struct {
	Name            *string              `json:"name,omitempty"`
	Status          *WebformStatus       `json:"status,omitempty"`
	TargetModule    *WebformTargetModule `json:"target_module,omitempty"`
	CampaignID      *uuid.UUID           `json:"campaign_id,omitempty"`
	ClearCampaignID bool                 `json:"-"`
	ReturnURL       *string              `json:"return_url,omitempty"`
	ClearReturnURL  bool                 `json:"-"`
	SuccessMessage  *string              `json:"success_message,omitempty"`
	SpamTrapField   *string              `json:"spam_trap_field,omitempty"`
	Fields          *[]WebformField      `json:"fields,omitempty"`
}

type WebformFilter struct {
	OrgID        uuid.UUID
	Status       *WebformStatus
	TargetModule *WebformTargetModule
	CampaignID   *uuid.UUID
	Q            string
	Page         int
	Limit        int
}

type WebformSubmission struct {
	ID                uuid.UUID           `json:"id"`
	OrgID             uuid.UUID           `json:"org_id"`
	WebformID         uuid.UUID           `json:"webform_id"`
	CampaignID        *uuid.UUID          `json:"campaign_id,omitempty"`
	TargetModule      WebformTargetModule `json:"target_module"`
	Payload           json.RawMessage     `json:"payload"`
	CreatedRecordType *string             `json:"created_record_type,omitempty"`
	CreatedRecordID   *uuid.UUID          `json:"created_record_id,omitempty"`
	IPAddress         *string             `json:"ip_address,omitempty"`
	UserAgent         *string             `json:"user_agent,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
}

type WebformSubmissionFilter struct {
	OrgID      uuid.UUID
	WebformID  *uuid.UUID
	CampaignID *uuid.UUID
	Page       int
	Limit      int
}

type PublicWebformSubmissionRequest struct {
	Payload map[string]interface{} `json:"payload"`
}

type WebformSubmissionResult struct {
	Success           bool       `json:"success"`
	ReturnURL         *string    `json:"return_url,omitempty"`
	SuccessMessage    string     `json:"success_message"`
	SubmissionID      *uuid.UUID `json:"submission_id,omitempty"`
	CreatedRecordType *string    `json:"created_record_type,omitempty"`
	CreatedRecordID   *uuid.UUID `json:"created_record_id,omitempty"`
}

type WebformPreviewRequest struct {
	Payload map[string]interface{} `json:"payload"`
}

type WebformPreviewResponse struct {
	TargetModule  WebformTargetModule    `json:"target_module"`
	MappedFields  map[string]interface{} `json:"mapped_fields"`
	MissingFields []string               `json:"missing_fields"`
}

type MailConverterRuleStatus string

const (
	MailConverterRuleActive   MailConverterRuleStatus = "active"
	MailConverterRuleInactive MailConverterRuleStatus = "inactive"
)

func (s MailConverterRuleStatus) IsValid() bool {
	return s == MailConverterRuleActive || s == MailConverterRuleInactive
}

type MailConverterCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
}

type MailConverterAction struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type MailConverterRule struct {
	ID         uuid.UUID                `json:"id"`
	OrgID      uuid.UUID                `json:"org_id"`
	Name       string                   `json:"name"`
	Status     MailConverterRuleStatus  `json:"status"`
	Conditions []MailConverterCondition `json:"conditions"`
	Actions    []MailConverterAction    `json:"actions"`
	CreatedBy  *uuid.UUID               `json:"created_by,omitempty"`
	LastRunAt  *time.Time               `json:"last_run_at,omitempty"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
	DeletedAt  *time.Time               `json:"deleted_at,omitempty"`
}

func (r *MailConverterRule) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if r.Status != "" && !r.Status.IsValid() {
		return fmt.Errorf("%w: status is invalid", ErrValidation)
	}
	if len(r.Conditions) == 0 {
		return fmt.Errorf("%w: at least one condition is required", ErrValidation)
	}
	if len(r.Actions) == 0 {
		return fmt.Errorf("%w: at least one action is required", ErrValidation)
	}
	for _, c := range r.Conditions {
		if strings.TrimSpace(c.Field) == "" {
			return fmt.Errorf("%w: condition field is required", ErrValidation)
		}
		if strings.TrimSpace(c.Operator) == "" {
			return fmt.Errorf("%w: condition operator is required", ErrValidation)
		}
	}
	for _, a := range r.Actions {
		if strings.TrimSpace(a.Type) == "" {
			return fmt.Errorf("%w: action type is required", ErrValidation)
		}
	}
	return nil
}

type MailConverterRulePatch struct {
	Name       *string                   `json:"name,omitempty"`
	Status     *MailConverterRuleStatus  `json:"status,omitempty"`
	Conditions *[]MailConverterCondition `json:"conditions,omitempty"`
	Actions    *[]MailConverterAction    `json:"actions,omitempty"`
}

type MailConverterRuleFilter struct {
	OrgID  uuid.UUID
	Status *MailConverterRuleStatus
	Page   int
	Limit  int
}

type MailConverterRun struct {
	ID             uuid.UUID  `json:"id"`
	OrgID          uuid.UUID  `json:"org_id"`
	RuleID         uuid.UUID  `json:"rule_id"`
	Status         string     `json:"status"`
	MatchedCount   int        `json:"matched_count"`
	ProcessedCount int        `json:"processed_count"`
	SkippedCount   int        `json:"skipped_count"`
	ErrorText      *string    `json:"error_text,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}

type MailConverterLog struct {
	ID                uuid.UUID  `json:"id"`
	OrgID             uuid.UUID  `json:"org_id"`
	RuleID            uuid.UUID  `json:"rule_id"`
	RunID             uuid.UUID  `json:"run_id"`
	MessageID         uuid.UUID  `json:"message_id"`
	Status            string     `json:"status"`
	CreatedRecordType *string    `json:"created_record_type,omitempty"`
	CreatedRecordID   *uuid.UUID `json:"created_record_id,omitempty"`
	ErrorText         *string    `json:"error_text,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type MailConverterPreviewResponse struct {
	RuleID  uuid.UUID            `json:"rule_id"`
	Matches []*EmailInboxMessage `json:"matches"`
	Total   int                  `json:"total"`
}

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusArchived  CampaignStatus = "archived"
)

func (s CampaignStatus) IsValid() bool {
	switch s {
	case CampaignStatusDraft, CampaignStatusActive, CampaignStatusPaused, CampaignStatusCompleted, CampaignStatusArchived:
		return true
	default:
		return false
	}
}

type Campaign struct {
	ID           uuid.UUID       `json:"id"`
	OrgID        uuid.UUID       `json:"org_id"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Status       CampaignStatus  `json:"status"`
	Description  string          `json:"description"`
	SequenceID   *uuid.UUID      `json:"sequence_id,omitempty"`
	AutomationID *uuid.UUID      `json:"automation_id,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedBy    *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    *time.Time      `json:"deleted_at,omitempty"`
}

func (c *Campaign) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if c.Type == "" {
		c.Type = "marketing"
	}
	if c.Status != "" && !c.Status.IsValid() {
		return fmt.Errorf("%w: status is invalid", ErrValidation)
	}
	return nil
}

type CampaignPatch struct {
	Name              *string         `json:"name,omitempty"`
	Type              *string         `json:"type,omitempty"`
	Status            *CampaignStatus `json:"status,omitempty"`
	Description       *string         `json:"description,omitempty"`
	SequenceID        *uuid.UUID      `json:"sequence_id,omitempty"`
	ClearSequenceID   bool            `json:"-"`
	AutomationID      *uuid.UUID      `json:"automation_id,omitempty"`
	ClearAutomationID bool            `json:"-"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

type CampaignFilter struct {
	OrgID  uuid.UUID
	Status *CampaignStatus
	Q      string
	Page   int
	Limit  int
}

type CampaignMember struct {
	ID         uuid.UUID `json:"id"`
	OrgID      uuid.UUID `json:"org_id"`
	CampaignID uuid.UUID `json:"campaign_id"`
	MemberType string    `json:"member_type"`
	MemberID   uuid.UUID `json:"member_id"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"created_at"`
}

type CampaignMemberRequest struct {
	Members []CampaignMemberInput `json:"members"`
}

type CampaignMemberInput struct {
	MemberType string    `json:"member_type"`
	MemberID   uuid.UUID `json:"member_id"`
	Source     string    `json:"source,omitempty"`
}
