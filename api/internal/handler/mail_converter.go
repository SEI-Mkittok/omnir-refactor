package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type MailConverterHandler struct {
	repo       repository.MailConverterRepository
	leads      repository.LeadRepository
	contacts   repository.ContactRepository
	tickets    repository.TicketRepository
	activities repository.ActivityRepository
}

func NewMailConverterHandler(
	repo repository.MailConverterRepository,
	leads repository.LeadRepository,
	contacts repository.ContactRepository,
	tickets repository.TicketRepository,
	activities repository.ActivityRepository,
) *MailConverterHandler {
	return &MailConverterHandler{repo: repo, leads: leads, contacts: contacts, tickets: tickets, activities: activities}
}

func (h *MailConverterHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/rules", h.ListRules)
	r.Post("/rules", h.CreateRule)
	r.Patch("/rules/{id}", h.UpdateRule)
	r.Delete("/rules/{id}", h.DeleteRule)
	r.Post("/rules/{id}/preview", h.PreviewRule)
	r.Post("/rules/{id}/scan", h.ScanRule)
	r.Get("/rules/{id}/runs", h.ListRuns)
	return r
}

func (h *MailConverterHandler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		return true
	}
	writeError(w, http.StatusForbidden, "admin access required")
	return false
}

func (h *MailConverterHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	filter := domain.MailConverterRuleFilter{Page: parseIntDefault(r.URL.Query().Get("page"), 1), Limit: parseIntDefault(r.URL.Query().Get("limit"), 50)}
	if v := r.URL.Query().Get("status"); v != "" {
		status := domain.MailConverterRuleStatus(v)
		filter.Status = &status
	}
	rules, total, err := h.repo.ListMailConverterRules(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(rules, total, filter.Page, filter.Limit))
}

func (h *MailConverterHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var rule domain.MailConverterRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		rule.CreatedBy = &claims.UserID
	}
	created, err := h.repo.CreateMailConverterRule(r.Context(), &rule)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *MailConverterHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var patch domain.MailConverterRulePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	current, err := h.repo.GetMailConverterRule(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	merged := *current
	if patch.Name != nil {
		merged.Name = *patch.Name
	}
	if patch.Status != nil {
		merged.Status = *patch.Status
	}
	if patch.Conditions != nil {
		merged.Conditions = *patch.Conditions
	}
	if patch.Actions != nil {
		merged.Actions = *patch.Actions
	}
	if err := merged.Validate(); err != nil {
		handleDomainErr(w, err)
		return
	}
	updated, err := h.repo.UpdateMailConverterRule(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *MailConverterHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.DeleteMailConverterRule(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MailConverterHandler) PreviewRule(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	rule, matches, err := h.ruleMatches(r.Context(), chi.URLParam(r, "id"), parseIntDefault(r.URL.Query().Get("limit"), 50))
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, domain.MailConverterPreviewResponse{RuleID: rule.ID, Matches: matches, Total: len(matches)})
}

func (h *MailConverterHandler) ScanRule(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	rule, matches, err := h.ruleMatches(r.Context(), chi.URLParam(r, "id"), parseIntDefault(r.URL.Query().Get("limit"), 250))
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if rule.Status != domain.MailConverterRuleActive {
		writeError(w, http.StatusUnprocessableEntity, "rule must be active before scanning")
		return
	}
	run := &domain.MailConverterRun{OrgID: rule.OrgID, RuleID: rule.ID, MatchedCount: len(matches)}
	run, err = h.repo.CreateMailConverterRun(r.Context(), run)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	var firstErr error
	for _, msg := range matches {
		log, inserted, err := h.repo.TryCreateMailConverterLog(r.Context(), &domain.MailConverterLog{
			OrgID:     rule.OrgID,
			RuleID:    rule.ID,
			RunID:     run.ID,
			MessageID: msg.ID,
			Status:    "running",
		})
		if err != nil {
			firstErr = err
			continue
		}
		if !inserted {
			run.SkippedCount++
			continue
		}

		recordType, recordID, err := h.executeActions(r.Context(), rule, msg)
		if err != nil {
			run.ProcessedCount++
			if firstErr == nil {
				firstErr = err
			}
			text := err.Error()
			log.Status = "failed"
			log.ErrorText = &text
		} else {
			run.ProcessedCount++
			log.Status = "succeeded"
			log.CreatedRecordType = recordType
			log.CreatedRecordID = recordID
		}
		if err := h.repo.UpdateMailConverterLog(r.Context(), log); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	now := time.Now().UTC()
	run.FinishedAt = &now
	run.Status = "succeeded"
	if firstErr != nil {
		run.Status = "failed"
		text := firstErr.Error()
		run.ErrorText = &text
	}
	if err := h.repo.UpdateMailConverterRun(r.Context(), run); err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (h *MailConverterHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	runs, err := h.repo.ListMailConverterRuns(r.Context(), id, parseIntDefault(r.URL.Query().Get("limit"), 50))
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": runs})
}

func (h *MailConverterHandler) ruleMatches(ctx context.Context, rawID string, limit int) (*domain.MailConverterRule, []*domain.EmailInboxMessage, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: invalid id", domain.ErrValidation)
	}
	rule, err := h.repo.GetMailConverterRule(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	orgID, ok := domain.OrgIDFromContext(ctx)
	if !ok {
		orgID = rule.OrgID
	}
	candidates, err := h.repo.ListMailConverterCandidateMessages(ctx, orgID, limit)
	if err != nil {
		return nil, nil, err
	}
	matches := []*domain.EmailInboxMessage{}
	for _, msg := range candidates {
		if mailRuleMatches(rule, msg) {
			matches = append(matches, msg)
		}
	}
	return rule, matches, nil
}

func mailRuleMatches(rule *domain.MailConverterRule, msg *domain.EmailInboxMessage) bool {
	for _, cond := range rule.Conditions {
		if !mailConditionMatches(cond, msg) {
			return false
		}
	}
	return true
}

func mailConditionMatches(cond domain.MailConverterCondition, msg *domain.EmailInboxMessage) bool {
	value := mailFieldValue(cond.Field, msg)
	needle := strings.ToLower(strings.TrimSpace(cond.Value))
	haystack := strings.ToLower(strings.TrimSpace(value))
	switch cond.Operator {
	case "contains":
		return strings.Contains(haystack, needle)
	case "not_contains":
		return !strings.Contains(haystack, needle)
	case "equals":
		return haystack == needle
	case "starts_with":
		return strings.HasPrefix(haystack, needle)
	case "ends_with":
		return strings.HasSuffix(haystack, needle)
	case "is_set":
		return strings.TrimSpace(value) != ""
	case "is_not_set":
		return strings.TrimSpace(value) == ""
	default:
		return false
	}
}

func mailFieldValue(field string, msg *domain.EmailInboxMessage) string {
	switch field {
	case "from", "from_addr":
		return msg.FromAddr
	case "to", "to_addrs":
		return strings.Join(msg.ToAddrs, ", ")
	case "subject":
		return msg.Subject
	case "body", "body_text":
		if msg.BodyText != nil {
			return *msg.BodyText
		}
	case "body_html":
		if msg.BodyHTML != nil {
			return *msg.BodyHTML
		}
	case "thread_id":
		return msg.ThreadID
	}
	return ""
}

func (h *MailConverterHandler) executeActions(ctx context.Context, rule *domain.MailConverterRule, msg *domain.EmailInboxMessage) (*string, *uuid.UUID, error) {
	var recordType *string
	var recordID *uuid.UUID
	for _, action := range rule.Actions {
		t, id, err := h.executeAction(ctx, rule, msg, action)
		if err != nil {
			return recordType, recordID, err
		}
		if recordType == nil && t != nil {
			recordType = t
			recordID = id
		}
	}
	return recordType, recordID, nil
}

func (h *MailConverterHandler) executeAction(ctx context.Context, rule *domain.MailConverterRule, msg *domain.EmailInboxMessage, action domain.MailConverterAction) (*string, *uuid.UUID, error) {
	name, emailAddr := parseMailbox(msg.FromAddr)
	first, last := splitName(name, emailAddr)
	source := "mail_converter"
	switch action.Type {
	case "create_lead":
		lead := &domain.Lead{
			FirstName:  first,
			LastName:   last,
			Email:      stringPtrNonEmpty(emailAddr),
			LeadSource: &source,
			Status:     domain.LeadStatusNew,
			OwnerID:    actionOwner(action.Config, rule.CreatedBy),
		}
		created, err := h.leads.Create(ctx, lead)
		if err != nil {
			return nil, nil, err
		}
		t := "lead"
		return &t, &created.ID, nil
	case "create_contact":
		if emailAddr != "" {
			existing, err := h.contacts.GetByEmail(ctx, emailAddr)
			if err == nil && existing != nil {
				t := "contact"
				return &t, &existing.ID, nil
			}
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return nil, nil, err
			}
		}
		owner := actionOwner(action.Config, rule.CreatedBy)
		if owner == nil {
			return nil, nil, fmt.Errorf("%w: create_contact requires owner_id or rule creator", domain.ErrValidation)
		}
		contact := &domain.Contact{
			FirstName:  first,
			LastName:   last,
			Email:      stringPtrNonEmpty(emailAddr),
			OwnerID:    *owner,
			LeadSource: &source,
			Stage:      domain.ContactStageLead,
		}
		if err := contact.Validate(); err != nil {
			return nil, nil, err
		}
		created, err := h.contacts.Create(ctx, contact)
		if err != nil {
			return nil, nil, err
		}
		t := "contact"
		return &t, &created.ID, nil
	case "update_contact":
		if emailAddr == "" {
			return nil, nil, nil
		}
		existing, err := h.contacts.GetByEmail(ctx, emailAddr)
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		leadSource := source
		updated, err := h.contacts.Update(ctx, existing.ID, domain.ContactPatch{LeadSource: &leadSource})
		if err != nil {
			return nil, nil, err
		}
		t := "contact"
		return &t, &updated.ID, nil
	case "create_ticket":
		if msg.MessageID != "" {
			existing, err := h.tickets.GetByEmailMessageID(ctx, msg.MessageID)
			if err == nil && existing != nil {
				t := "ticket"
				return &t, &existing.ID, nil
			}
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return nil, nil, err
			}
		}
		description := ""
		if msg.BodyText != nil {
			description = *msg.BodyText
		}
		ticket := &domain.Ticket{
			Subject:        defaultString(msg.Subject, "Inbound email"),
			Description:    stringPtrNonEmpty(description),
			Status:         domain.TicketStatusOpen,
			Priority:       domain.TicketPriorityMedium,
			Source:         &source,
			EmailMessageID: &msg.MessageID,
			ContactID:      msg.ContactID,
		}
		created, err := h.tickets.Create(ctx, ticket)
		if err != nil {
			return nil, nil, err
		}
		t := "ticket"
		return &t, &created.ID, nil
	case "create_activity":
		owner := actionOwner(action.Config, rule.CreatedBy)
		if owner == nil {
			return nil, nil, fmt.Errorf("%w: create_activity requires owner_id or rule creator", domain.ErrValidation)
		}
		subject := configString(action.Config, "subject")
		if subject == "" {
			subject = "Review inbound email: " + defaultString(msg.Subject, msg.FromAddr)
		}
		description := ""
		if msg.BodyText != nil {
			description = *msg.BodyText
		}
		activity := &domain.Activity{
			Type:        domain.ActivityTypeEmail,
			Subject:     subject,
			Description: stringPtrNonEmpty(description),
			ContactID:   msg.ContactID,
			OwnerID:     *owner,
		}
		created, err := h.activities.Create(ctx, activity)
		if err != nil {
			return nil, nil, err
		}
		t := "activity"
		return &t, &created.ID, nil
	default:
		return nil, nil, fmt.Errorf("%w: unsupported action %q", domain.ErrValidation, action.Type)
	}
}

func parseMailbox(raw string) (string, string) {
	addr, err := mail.ParseAddress(raw)
	if err != nil {
		raw = strings.TrimSpace(raw)
		if strings.Contains(raw, "@") {
			return "", raw
		}
		return raw, ""
	}
	return strings.TrimSpace(addr.Name), strings.TrimSpace(addr.Address)
}

func splitName(name string, emailAddr string) (string, string) {
	parts := strings.Fields(name)
	if len(parts) == 0 && emailAddr != "" {
		local := emailAddr
		if at := strings.Index(local, "@"); at > 0 {
			local = local[:at]
		}
		parts = strings.Fields(strings.ReplaceAll(local, ".", " "))
	}
	if len(parts) == 0 {
		return "Email", "Sender"
	}
	if len(parts) == 1 {
		return parts[0], "Sender"
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func actionOwner(config map[string]interface{}, fallback *uuid.UUID) *uuid.UUID {
	if id := parseUUIDPtr(configString(config, "owner_id")); id != nil {
		return id
	}
	return fallback
}

func configString(config map[string]interface{}, key string) string {
	if config == nil {
		return ""
	}
	v, ok := config[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}
