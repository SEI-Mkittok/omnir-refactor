package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type WebformRepo struct {
	db *pgxpool.Pool
}

func NewWebformRepo(db *pgxpool.Pool) *WebformRepo {
	return &WebformRepo{db: db}
}

const webformCols = `
	id, org_id, name, public_id, status, target_module, campaign_id,
	return_url, success_message, spam_trap_field, fields, created_by,
	created_at, updated_at, deleted_at
`

func scanWebform(row pgx.Row) (*domain.Webform, error) {
	var f domain.Webform
	var fieldsJSON []byte
	err := row.Scan(
		&f.ID, &f.OrgID, &f.Name, &f.PublicID, &f.Status, &f.TargetModule, &f.CampaignID,
		&f.ReturnURL, &f.SuccessMessage, &f.SpamTrapField, &fieldsJSON, &f.CreatedBy,
		&f.CreatedAt, &f.UpdatedAt, &f.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(fieldsJSON) > 0 {
		if err := json.Unmarshal(fieldsJSON, &f.Fields); err != nil {
			return nil, fmt.Errorf("unmarshal webform fields: %w", err)
		}
	}
	if f.Fields == nil {
		f.Fields = []domain.WebformField{}
	}
	return &f, nil
}

func (r *WebformRepo) ListWebforms(ctx context.Context, filter domain.WebformFilter) ([]*domain.Webform, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}

	where := []string{"deleted_at IS NULL"}
	args := []any{}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		where = append(where, "org_id = $"+itoa(len(args)))
	}
	if filter.Status != nil {
		args = append(args, *filter.Status)
		where = append(where, "status = $"+itoa(len(args)))
	}
	if filter.TargetModule != nil {
		args = append(args, *filter.TargetModule)
		where = append(where, "target_module = $"+itoa(len(args)))
	}
	if filter.CampaignID != nil {
		args = append(args, *filter.CampaignID)
		where = append(where, "campaign_id = $"+itoa(len(args)))
	}
	if filter.Q != "" {
		args = append(args, "%"+strings.ToLower(filter.Q)+"%")
		where = append(where, "LOWER(name) LIKE $"+itoa(len(args)))
	}
	whereClause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM webforms WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*domain.Webform{}, 0, nil
	}

	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)
	rows, err := r.db.Query(ctx,
		`SELECT `+webformCols+` FROM webforms WHERE `+whereClause+
			` ORDER BY updated_at DESC LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []*domain.Webform{}
	for rows.Next() {
		f, err := scanWebform(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, f)
	}
	return out, total, rows.Err()
}

func (r *WebformRepo) CreateWebform(ctx context.Context, form *domain.Webform) (*domain.Webform, error) {
	if err := form.Validate(); err != nil {
		return nil, err
	}
	if form.ID == uuid.Nil {
		form.ID = uuid.New()
	}
	if form.PublicID == "" {
		form.PublicID = strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		form.OrgID = orgID
	}
	if form.Status == "" {
		form.Status = domain.WebformStatusInactive
	}
	if form.SuccessMessage == "" {
		form.SuccessMessage = "Thanks. Your submission has been received."
	}
	if form.SpamTrapField == "" {
		form.SpamTrapField = "website"
	}
	fieldsJSON, err := json.Marshal(form.Fields)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	form.CreatedAt = now
	form.UpdatedAt = now

	return scanWebform(r.db.QueryRow(ctx, `
		INSERT INTO webforms
			(id, org_id, name, public_id, status, target_module, campaign_id,
			 return_url, success_message, spam_trap_field, fields, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING `+webformCols,
		form.ID, form.OrgID, form.Name, form.PublicID, form.Status, form.TargetModule, form.CampaignID,
		form.ReturnURL, form.SuccessMessage, form.SpamTrapField, fieldsJSON, form.CreatedBy, form.CreatedAt, form.UpdatedAt,
	))
}

func (r *WebformRepo) GetWebform(ctx context.Context, id uuid.UUID) (*domain.Webform, error) {
	q := `SELECT ` + webformCols + ` FROM webforms WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		q += ` AND org_id=$` + itoa(len(args))
	}
	return scanWebform(r.db.QueryRow(ctx, q, args...))
}

func (r *WebformRepo) GetWebformByPublicID(ctx context.Context, publicID string) (*domain.Webform, error) {
	return scanWebform(r.db.QueryRow(ctx,
		`SELECT `+webformCols+` FROM webforms WHERE public_id=$1 AND deleted_at IS NULL`, publicID))
}

func (r *WebformRepo) UpdateWebform(ctx context.Context, id uuid.UUID, patch domain.WebformPatch) (*domain.Webform, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	add := func(col string, val any) {
		args = append(args, val)
		sets = append(sets, col+" = $"+itoa(len(args)))
	}

	if patch.Name != nil {
		add("name", *patch.Name)
	}
	if patch.Status != nil {
		if !patch.Status.IsValid() {
			return nil, fmt.Errorf("%w: status is invalid", domain.ErrValidation)
		}
		add("status", *patch.Status)
	}
	if patch.TargetModule != nil {
		if !patch.TargetModule.IsValid() {
			return nil, fmt.Errorf("%w: target_module is invalid", domain.ErrValidation)
		}
		add("target_module", *patch.TargetModule)
	}
	if patch.ClearCampaignID {
		sets = append(sets, "campaign_id = NULL")
	} else if patch.CampaignID != nil {
		add("campaign_id", *patch.CampaignID)
	}
	if patch.ClearReturnURL {
		sets = append(sets, "return_url = NULL")
	} else if patch.ReturnURL != nil {
		add("return_url", *patch.ReturnURL)
	}
	if patch.SuccessMessage != nil {
		add("success_message", *patch.SuccessMessage)
	}
	if patch.SpamTrapField != nil {
		add("spam_trap_field", *patch.SpamTrapField)
	}
	if patch.Fields != nil {
		if len(*patch.Fields) == 0 {
			return nil, fmt.Errorf("%w: at least one field is required", domain.ErrValidation)
		}
		b, err := json.Marshal(*patch.Fields)
		if err != nil {
			return nil, err
		}
		add("fields", b)
	}

	args = append(args, id)
	where := "id = $" + itoa(len(args)) + " AND deleted_at IS NULL"
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		where += " AND org_id = $" + itoa(len(args))
	}
	_, err := r.db.Exec(ctx, `UPDATE webforms SET `+strings.Join(sets, ", ")+` WHERE `+where, args...)
	if err != nil {
		return nil, err
	}
	return r.GetWebform(ctx, id)
}

func (r *WebformRepo) DeleteWebform(ctx context.Context, id uuid.UUID) error {
	args := []any{id}
	q := `UPDATE webforms SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		q += ` AND org_id=$` + itoa(len(args))
	}
	tag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

const webformSubmissionCols = `
	id, org_id, webform_id, campaign_id, target_module, payload,
	created_record_type, created_record_id, ip_address, user_agent, created_at
`

func scanWebformSubmission(row pgx.Row) (*domain.WebformSubmission, error) {
	var s domain.WebformSubmission
	var payload []byte
	err := row.Scan(
		&s.ID, &s.OrgID, &s.WebformID, &s.CampaignID, &s.TargetModule, &payload,
		&s.CreatedRecordType, &s.CreatedRecordID, &s.IPAddress, &s.UserAgent, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	s.Payload = payload
	return &s, nil
}

func (r *WebformRepo) CreateWebformSubmission(ctx context.Context, submission *domain.WebformSubmission) (*domain.WebformSubmission, error) {
	if submission.ID == uuid.Nil {
		submission.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		submission.OrgID = orgID
	}
	submission.CreatedAt = time.Now().UTC()
	return scanWebformSubmission(r.db.QueryRow(ctx, `
		INSERT INTO webform_submissions
			(id, org_id, webform_id, campaign_id, target_module, payload,
			 created_record_type, created_record_id, ip_address, user_agent, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+webformSubmissionCols,
		submission.ID, submission.OrgID, submission.WebformID, submission.CampaignID, submission.TargetModule,
		[]byte(submission.Payload), submission.CreatedRecordType, submission.CreatedRecordID,
		submission.IPAddress, submission.UserAgent, submission.CreatedAt,
	))
}

func (r *WebformRepo) ListWebformSubmissions(ctx context.Context, filter domain.WebformSubmissionFilter) ([]*domain.WebformSubmission, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	where := []string{"1=1"}
	args := []any{}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		where = append(where, "org_id = $"+itoa(len(args)))
	}
	if filter.WebformID != nil {
		args = append(args, *filter.WebformID)
		where = append(where, "webform_id = $"+itoa(len(args)))
	}
	if filter.CampaignID != nil {
		args = append(args, *filter.CampaignID)
		where = append(where, "campaign_id = $"+itoa(len(args)))
	}
	whereClause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM webform_submissions WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*domain.WebformSubmission{}, 0, nil
	}
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)
	rows, err := r.db.Query(ctx,
		`SELECT `+webformSubmissionCols+` FROM webform_submissions WHERE `+whereClause+
			` ORDER BY created_at DESC LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*domain.WebformSubmission{}
	for rows.Next() {
		s, err := scanWebformSubmission(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

type MailConverterRepo struct {
	db *pgxpool.Pool
}

func NewMailConverterRepo(db *pgxpool.Pool) *MailConverterRepo {
	return &MailConverterRepo{db: db}
}

const mailConverterRuleCols = `
	id, org_id, name, status, conditions, actions, created_by,
	last_run_at, created_at, updated_at, deleted_at
`

func scanMailConverterRule(row pgx.Row) (*domain.MailConverterRule, error) {
	var rule domain.MailConverterRule
	var conditionsJSON, actionsJSON []byte
	err := row.Scan(
		&rule.ID, &rule.OrgID, &rule.Name, &rule.Status, &conditionsJSON, &actionsJSON,
		&rule.CreatedBy, &rule.LastRunAt, &rule.CreatedAt, &rule.UpdatedAt, &rule.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(conditionsJSON) > 0 {
		if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
			return nil, err
		}
	}
	if len(actionsJSON) > 0 {
		if err := json.Unmarshal(actionsJSON, &rule.Actions); err != nil {
			return nil, err
		}
	}
	if rule.Conditions == nil {
		rule.Conditions = []domain.MailConverterCondition{}
	}
	if rule.Actions == nil {
		rule.Actions = []domain.MailConverterAction{}
	}
	return &rule, nil
}

func (r *MailConverterRepo) ListMailConverterRules(ctx context.Context, filter domain.MailConverterRuleFilter) ([]*domain.MailConverterRule, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		where = append(where, "org_id = $"+itoa(len(args)))
	}
	if filter.Status != nil {
		args = append(args, *filter.Status)
		where = append(where, "status = $"+itoa(len(args)))
	}
	whereClause := strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM mail_converter_rules WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*domain.MailConverterRule{}, 0, nil
	}
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)
	rows, err := r.db.Query(ctx,
		`SELECT `+mailConverterRuleCols+` FROM mail_converter_rules WHERE `+whereClause+
			` ORDER BY updated_at DESC LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*domain.MailConverterRule{}
	for rows.Next() {
		rule, err := scanMailConverterRule(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, rule)
	}
	return out, total, rows.Err()
}

func (r *MailConverterRepo) CreateMailConverterRule(ctx context.Context, rule *domain.MailConverterRule) (*domain.MailConverterRule, error) {
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		rule.OrgID = orgID
	}
	if rule.Status == "" {
		rule.Status = domain.MailConverterRuleInactive
	}
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return nil, err
	}
	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	return scanMailConverterRule(r.db.QueryRow(ctx, `
		INSERT INTO mail_converter_rules
			(id, org_id, name, status, conditions, actions, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING `+mailConverterRuleCols,
		rule.ID, rule.OrgID, rule.Name, rule.Status, conditionsJSON, actionsJSON,
		rule.CreatedBy, rule.CreatedAt, rule.UpdatedAt,
	))
}

func (r *MailConverterRepo) GetMailConverterRule(ctx context.Context, id uuid.UUID) (*domain.MailConverterRule, error) {
	q := `SELECT ` + mailConverterRuleCols + ` FROM mail_converter_rules WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		q += ` AND org_id=$` + itoa(len(args))
	}
	return scanMailConverterRule(r.db.QueryRow(ctx, q, args...))
}

func (r *MailConverterRepo) UpdateMailConverterRule(ctx context.Context, id uuid.UUID, patch domain.MailConverterRulePatch) (*domain.MailConverterRule, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	add := func(col string, val any) {
		args = append(args, val)
		sets = append(sets, col+" = $"+itoa(len(args)))
	}
	if patch.Name != nil {
		add("name", *patch.Name)
	}
	if patch.Status != nil {
		if !patch.Status.IsValid() {
			return nil, fmt.Errorf("%w: status is invalid", domain.ErrValidation)
		}
		add("status", *patch.Status)
	}
	if patch.Conditions != nil {
		b, err := json.Marshal(*patch.Conditions)
		if err != nil {
			return nil, err
		}
		add("conditions", b)
	}
	if patch.Actions != nil {
		b, err := json.Marshal(*patch.Actions)
		if err != nil {
			return nil, err
		}
		add("actions", b)
	}
	args = append(args, id)
	where := "id = $" + itoa(len(args)) + " AND deleted_at IS NULL"
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		where += " AND org_id = $" + itoa(len(args))
	}
	_, err := r.db.Exec(ctx, `UPDATE mail_converter_rules SET `+strings.Join(sets, ", ")+` WHERE `+where, args...)
	if err != nil {
		return nil, err
	}
	return r.GetMailConverterRule(ctx, id)
}

func (r *MailConverterRepo) DeleteMailConverterRule(ctx context.Context, id uuid.UUID) error {
	args := []any{id}
	q := `UPDATE mail_converter_rules SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		q += ` AND org_id=$` + itoa(len(args))
	}
	tag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *MailConverterRepo) ListMailConverterCandidateMessages(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.EmailInboxMessage, error) {
	if limit <= 0 || limit > 1000 {
		limit = 250
	}
	rows, err := r.db.Query(ctx,
		`SELECT `+inboxMsgCols+`
		 FROM email_inbox_messages
		 WHERE org_id=$1 AND direction='inbound'
		 ORDER BY sent_at DESC
		 LIMIT $2`, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.EmailInboxMessage{}
	for rows.Next() {
		msg, err := scanInboxMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	return out, rows.Err()
}

const mailConverterRunCols = `
	id, org_id, rule_id, status, matched_count, processed_count,
	skipped_count, error_text, started_at, finished_at
`

func scanMailConverterRun(row pgx.Row) (*domain.MailConverterRun, error) {
	var run domain.MailConverterRun
	err := row.Scan(
		&run.ID, &run.OrgID, &run.RuleID, &run.Status, &run.MatchedCount,
		&run.ProcessedCount, &run.SkippedCount, &run.ErrorText, &run.StartedAt, &run.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

func (r *MailConverterRepo) CreateMailConverterRun(ctx context.Context, run *domain.MailConverterRun) (*domain.MailConverterRun, error) {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		run.OrgID = orgID
	}
	if run.Status == "" {
		run.Status = "running"
	}
	run.StartedAt = time.Now().UTC()
	return scanMailConverterRun(r.db.QueryRow(ctx, `
		INSERT INTO mail_converter_runs
			(id, org_id, rule_id, status, matched_count, processed_count, skipped_count, error_text, started_at, finished_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+mailConverterRunCols,
		run.ID, run.OrgID, run.RuleID, run.Status, run.MatchedCount,
		run.ProcessedCount, run.SkippedCount, run.ErrorText, run.StartedAt, run.FinishedAt,
	))
}

func (r *MailConverterRepo) UpdateMailConverterRun(ctx context.Context, run *domain.MailConverterRun) error {
	_, err := r.db.Exec(ctx, `
		UPDATE mail_converter_runs
		SET status=$2, matched_count=$3, processed_count=$4, skipped_count=$5, error_text=$6, finished_at=$7
		WHERE id=$1 AND org_id=$8`,
		run.ID, run.Status, run.MatchedCount, run.ProcessedCount, run.SkippedCount, run.ErrorText, run.FinishedAt, run.OrgID,
	)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `UPDATE mail_converter_rules SET last_run_at=NOW(), updated_at=NOW() WHERE id=$1 AND org_id=$2`, run.RuleID, run.OrgID)
	return err
}

func (r *MailConverterRepo) ListMailConverterRuns(ctx context.Context, ruleID uuid.UUID, limit int) ([]*domain.MailConverterRun, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args := []any{ruleID, limit}
	q := `SELECT ` + mailConverterRunCols + ` FROM mail_converter_runs WHERE rule_id=$1`
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		q += ` AND org_id=$3`
	}
	q += ` ORDER BY started_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.MailConverterRun{}
	for rows.Next() {
		run, err := scanMailConverterRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, rows.Err()
}

const mailConverterLogCols = `
	id, org_id, rule_id, run_id, message_id, status,
	created_record_type, created_record_id, error_text, created_at
`

func scanMailConverterLog(row pgx.Row) (*domain.MailConverterLog, error) {
	var log domain.MailConverterLog
	err := row.Scan(
		&log.ID, &log.OrgID, &log.RuleID, &log.RunID, &log.MessageID, &log.Status,
		&log.CreatedRecordType, &log.CreatedRecordID, &log.ErrorText, &log.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &log, nil
}

func (r *MailConverterRepo) TryCreateMailConverterLog(ctx context.Context, log *domain.MailConverterLog) (*domain.MailConverterLog, bool, error) {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		log.OrgID = orgID
	}
	if log.Status == "" {
		log.Status = "running"
	}
	log.CreatedAt = time.Now().UTC()
	created, err := scanMailConverterLog(r.db.QueryRow(ctx, `
		INSERT INTO mail_converter_logs
			(id, org_id, rule_id, run_id, message_id, status,
			 created_record_type, created_record_id, error_text, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (rule_id, message_id) DO NOTHING
		RETURNING `+mailConverterLogCols,
		log.ID, log.OrgID, log.RuleID, log.RunID, log.MessageID, log.Status,
		log.CreatedRecordType, log.CreatedRecordID, log.ErrorText, log.CreatedAt,
	))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return created, true, nil
}

func (r *MailConverterRepo) UpdateMailConverterLog(ctx context.Context, log *domain.MailConverterLog) error {
	_, err := r.db.Exec(ctx, `
		UPDATE mail_converter_logs
		SET status=$2, created_record_type=$3, created_record_id=$4, error_text=$5
		WHERE id=$1 AND org_id=$6`,
		log.ID, log.Status, log.CreatedRecordType, log.CreatedRecordID, log.ErrorText, log.OrgID,
	)
	return err
}

type CampaignRepo struct {
	db *pgxpool.Pool
}

func NewCampaignRepo(db *pgxpool.Pool) *CampaignRepo {
	return &CampaignRepo{db: db}
}

const campaignCols = `
	id, org_id, name, type, status, description, sequence_id, automation_id,
	metadata, created_by, created_at, updated_at, deleted_at
`

func scanCampaign(row pgx.Row) (*domain.Campaign, error) {
	var c domain.Campaign
	var metadata []byte
	err := row.Scan(
		&c.ID, &c.OrgID, &c.Name, &c.Type, &c.Status, &c.Description, &c.SequenceID, &c.AutomationID,
		&metadata, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	c.Metadata = metadata
	return &c, nil
}

func (r *CampaignRepo) ListCampaigns(ctx context.Context, filter domain.CampaignFilter) ([]*domain.Campaign, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		where = append(where, "org_id = $"+itoa(len(args)))
	}
	if filter.Status != nil {
		args = append(args, *filter.Status)
		where = append(where, "status = $"+itoa(len(args)))
	}
	if filter.Q != "" {
		args = append(args, "%"+strings.ToLower(filter.Q)+"%")
		where = append(where, "LOWER(name) LIKE $"+itoa(len(args)))
	}
	whereClause := strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM campaigns WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*domain.Campaign{}, 0, nil
	}
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)
	rows, err := r.db.Query(ctx,
		`SELECT `+campaignCols+` FROM campaigns WHERE `+whereClause+
			` ORDER BY updated_at DESC LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*domain.Campaign{}
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

func (r *CampaignRepo) CreateCampaign(ctx context.Context, campaign *domain.Campaign) (*domain.Campaign, error) {
	if err := campaign.Validate(); err != nil {
		return nil, err
	}
	if campaign.ID == uuid.Nil {
		campaign.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		campaign.OrgID = orgID
	}
	if campaign.Status == "" {
		campaign.Status = domain.CampaignStatusDraft
	}
	if campaign.Metadata == nil {
		campaign.Metadata = json.RawMessage(`{}`)
	}
	now := time.Now().UTC()
	campaign.CreatedAt = now
	campaign.UpdatedAt = now
	return scanCampaign(r.db.QueryRow(ctx, `
		INSERT INTO campaigns
			(id, org_id, name, type, status, description, sequence_id, automation_id,
			 metadata, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING `+campaignCols,
		campaign.ID, campaign.OrgID, campaign.Name, campaign.Type, campaign.Status, campaign.Description,
		campaign.SequenceID, campaign.AutomationID, []byte(campaign.Metadata), campaign.CreatedBy,
		campaign.CreatedAt, campaign.UpdatedAt,
	))
}

func (r *CampaignRepo) GetCampaign(ctx context.Context, id uuid.UUID) (*domain.Campaign, error) {
	q := `SELECT ` + campaignCols + ` FROM campaigns WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		q += ` AND org_id=$` + itoa(len(args))
	}
	return scanCampaign(r.db.QueryRow(ctx, q, args...))
}

func (r *CampaignRepo) UpdateCampaign(ctx context.Context, id uuid.UUID, patch domain.CampaignPatch) (*domain.Campaign, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	add := func(col string, val any) {
		args = append(args, val)
		sets = append(sets, col+" = $"+itoa(len(args)))
	}
	if patch.Name != nil {
		add("name", *patch.Name)
	}
	if patch.Type != nil {
		add("type", *patch.Type)
	}
	if patch.Status != nil {
		if !patch.Status.IsValid() {
			return nil, fmt.Errorf("%w: status is invalid", domain.ErrValidation)
		}
		add("status", *patch.Status)
	}
	if patch.Description != nil {
		add("description", *patch.Description)
	}
	if patch.ClearSequenceID {
		sets = append(sets, "sequence_id = NULL")
	} else if patch.SequenceID != nil {
		add("sequence_id", *patch.SequenceID)
	}
	if patch.ClearAutomationID {
		sets = append(sets, "automation_id = NULL")
	} else if patch.AutomationID != nil {
		add("automation_id", *patch.AutomationID)
	}
	if patch.Metadata != nil {
		add("metadata", []byte(patch.Metadata))
	}

	args = append(args, id)
	where := "id = $" + itoa(len(args)) + " AND deleted_at IS NULL"
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		where += " AND org_id = $" + itoa(len(args))
	}
	_, err := r.db.Exec(ctx, `UPDATE campaigns SET `+strings.Join(sets, ", ")+` WHERE `+where, args...)
	if err != nil {
		return nil, err
	}
	return r.GetCampaign(ctx, id)
}

func (r *CampaignRepo) DeleteCampaign(ctx context.Context, id uuid.UUID) error {
	args := []any{id}
	q := `UPDATE campaigns SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		q += ` AND org_id=$` + itoa(len(args))
	}
	tag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

const campaignMemberCols = `
	id, org_id, campaign_id, member_type, member_id, source, created_at
`

func scanCampaignMember(row pgx.Row) (*domain.CampaignMember, error) {
	var m domain.CampaignMember
	err := row.Scan(&m.ID, &m.OrgID, &m.CampaignID, &m.MemberType, &m.MemberID, &m.Source, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *CampaignRepo) ListCampaignMembers(ctx context.Context, campaignID uuid.UUID, limit int) ([]*domain.CampaignMember, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []any{campaignID, limit}
	q := `SELECT ` + campaignMemberCols + ` FROM campaign_members WHERE campaign_id=$1`
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		args = append(args, orgID)
		q += ` AND org_id=$3`
	}
	q += ` ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.CampaignMember{}
	for rows.Next() {
		m, err := scanCampaignMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *CampaignRepo) AddCampaignMembers(ctx context.Context, campaignID uuid.UUID, members []domain.CampaignMemberInput) ([]*domain.CampaignMember, error) {
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("%w: org_id is required", domain.ErrValidation)
	}
	for _, input := range members {
		if input.MemberID == uuid.Nil {
			return nil, fmt.Errorf("%w: member_id is required", domain.ErrValidation)
		}
		switch input.MemberType {
		case "lead", "contact", "account":
		default:
			return nil, fmt.Errorf("%w: invalid member_type %q", domain.ErrValidation, input.MemberType)
		}
		source := input.Source
		if source == "" {
			source = "manual"
		}
		_, err := r.db.Exec(ctx, `
			INSERT INTO campaign_members
				(id, org_id, campaign_id, member_type, member_id, source, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,NOW())
			ON CONFLICT (campaign_id, member_type, member_id) DO NOTHING`,
			uuid.New(), orgID, campaignID, input.MemberType, input.MemberID, source,
		)
		if err != nil {
			return nil, err
		}
	}
	return r.ListCampaignMembers(ctx, campaignID, 500)
}
