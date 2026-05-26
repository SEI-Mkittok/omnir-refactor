package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

type WebformRepository interface {
	ListWebforms(ctx context.Context, filter domain.WebformFilter) ([]*domain.Webform, int, error)
	CreateWebform(ctx context.Context, form *domain.Webform) (*domain.Webform, error)
	GetWebform(ctx context.Context, id uuid.UUID) (*domain.Webform, error)
	GetWebformByPublicID(ctx context.Context, publicID string) (*domain.Webform, error)
	UpdateWebform(ctx context.Context, id uuid.UUID, patch domain.WebformPatch) (*domain.Webform, error)
	DeleteWebform(ctx context.Context, id uuid.UUID) error
	CreateWebformSubmission(ctx context.Context, submission *domain.WebformSubmission) (*domain.WebformSubmission, error)
	ListWebformSubmissions(ctx context.Context, filter domain.WebformSubmissionFilter) ([]*domain.WebformSubmission, int, error)
}

type MailConverterRepository interface {
	ListMailConverterRules(ctx context.Context, filter domain.MailConverterRuleFilter) ([]*domain.MailConverterRule, int, error)
	CreateMailConverterRule(ctx context.Context, rule *domain.MailConverterRule) (*domain.MailConverterRule, error)
	GetMailConverterRule(ctx context.Context, id uuid.UUID) (*domain.MailConverterRule, error)
	UpdateMailConverterRule(ctx context.Context, id uuid.UUID, patch domain.MailConverterRulePatch) (*domain.MailConverterRule, error)
	DeleteMailConverterRule(ctx context.Context, id uuid.UUID) error
	ListMailConverterCandidateMessages(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.EmailInboxMessage, error)
	CreateMailConverterRun(ctx context.Context, run *domain.MailConverterRun) (*domain.MailConverterRun, error)
	UpdateMailConverterRun(ctx context.Context, run *domain.MailConverterRun) error
	ListMailConverterRuns(ctx context.Context, ruleID uuid.UUID, limit int) ([]*domain.MailConverterRun, error)
	TryCreateMailConverterLog(ctx context.Context, log *domain.MailConverterLog) (*domain.MailConverterLog, bool, error)
	UpdateMailConverterLog(ctx context.Context, log *domain.MailConverterLog) error
}

type CampaignRepository interface {
	ListCampaigns(ctx context.Context, filter domain.CampaignFilter) ([]*domain.Campaign, int, error)
	CreateCampaign(ctx context.Context, campaign *domain.Campaign) (*domain.Campaign, error)
	GetCampaign(ctx context.Context, id uuid.UUID) (*domain.Campaign, error)
	UpdateCampaign(ctx context.Context, id uuid.UUID, patch domain.CampaignPatch) (*domain.Campaign, error)
	DeleteCampaign(ctx context.Context, id uuid.UUID) error
	ListCampaignMembers(ctx context.Context, campaignID uuid.UUID, limit int) ([]*domain.CampaignMember, error)
	AddCampaignMembers(ctx context.Context, campaignID uuid.UUID, members []domain.CampaignMemberInput) ([]*domain.CampaignMember, error)
}
