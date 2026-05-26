package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	"github.com/omnir/crm-api/internal/repository"
)

// AutomationEvent carries trigger information to the automation engine.
type AutomationEvent struct {
	OrgID       uuid.UUID
	TriggerType domain.TriggerType
	EntityID    uuid.UUID
	EntityType  string
	// Data is a flat map of the entity's fields used for condition evaluation.
	Data map[string]interface{}
}

// AutomationWorker evaluates automation rules on entity events.
type AutomationWorker struct {
	repo         repository.AutomationRepository
	activities   repository.ActivityRepository
	contacts     repository.ContactRepository
	deals        repository.DealRepository
	tickets      repository.TicketRepository
	sequences    repository.SequenceRepository
	mailer       *email.Mailer
	interval     time.Duration
	log          *slog.Logger
	scheduler    *SchedulerRegistry
	schedulerKey string

	// Events is the channel through which CRM handlers submit entity events.
	Events chan AutomationEvent
}

// NewAutomationWorker creates a worker that polls every interval for overdue
// activities and processes incoming events from the Events channel.
func NewAutomationWorker(
	repo repository.AutomationRepository,
	activities repository.ActivityRepository,
	contacts repository.ContactRepository,
	deals repository.DealRepository,
	tickets repository.TicketRepository,
	sequences repository.SequenceRepository,
	mailer *email.Mailer,
	interval time.Duration,
	log *slog.Logger,
) *AutomationWorker {
	return &AutomationWorker{
		repo:       repo,
		activities: activities,
		contacts:   contacts,
		deals:      deals,
		tickets:    tickets,
		sequences:  sequences,
		mailer:     mailer,
		interval:   interval,
		log:        log,
		Events:     make(chan AutomationEvent, 512),
	}
}

func (w *AutomationWorker) WithScheduler(registry *SchedulerRegistry, key string) *AutomationWorker {
	w.scheduler = registry
	w.schedulerKey = key
	return w
}

// Start launches the worker in a background goroutine.
func (w *AutomationWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		w.log.Info("automation worker started", "interval", w.interval)
		for {
			select {
			case evt := <-w.Events:
				w.processEvent(evt)
			case <-ticker.C:
				if err := w.scheduler.TrackRun(w.schedulerKey, w.interval, func() error {
					return w.tickOverdue(ctx)
				}); err != nil {
					w.log.Error("automation worker: overdue scan failed", "err", err)
				}
			case <-ctx.Done():
				w.log.Info("automation worker stopped")
				return
			}
		}
	}()
}

// processEvent evaluates all active automations for the given trigger.
func (w *AutomationWorker) processEvent(evt AutomationEvent) {
	bgCtx := context.Background()
	automations, err := w.repo.ListActiveByTrigger(bgCtx, evt.OrgID, evt.TriggerType)
	if err != nil {
		w.log.Error("automation worker: list automations failed",
			"trigger", evt.TriggerType, "err", err)
		return
	}

	for _, a := range automations {
		if !EvaluateConditions(a.Conditions, evt.Data) {
			continue
		}
		if _, err := w.ExecuteAutomation(bgCtx, a, evt); err != nil {
			w.log.Warn("automation worker: execution failed", "automation_id", a.ID, "err", err)
		}
	}
}

// tickOverdue scans for overdue activities and fires activity_overdue automations.
func (w *AutomationWorker) tickOverdue(ctx context.Context) error {
	refs, err := w.repo.OverdueActivityIDs(ctx, time.Now().UTC(), 100)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		w.processEvent(AutomationEvent{
			OrgID:       ref.OrgID,
			TriggerType: domain.TriggerActivityOverdue,
			EntityID:    ref.ActivityID,
			EntityType:  "activity",
			Data: map[string]interface{}{
				"owner_id": ref.OwnerID.String(),
			},
		})
	}
	return nil
}

// ExecuteAutomation creates a run record, executes all actions, then updates run status.
func (w *AutomationWorker) ExecuteAutomation(ctx context.Context, a *domain.Automation, evt AutomationEvent) (*domain.AutomationRun, error) {
	if evt.OrgID == uuid.Nil {
		evt.OrgID = a.OrgID
	}
	if evt.TriggerType == "" {
		evt.TriggerType = domain.TriggerManual
	}
	if evt.EntityType == "" {
		evt.EntityType = "manual"
	}
	entityID := evt.EntityID
	run := &domain.AutomationRun{
		AutomationID: a.ID,
		OrgID:        evt.OrgID,
		EntityType:   evt.EntityType,
		EntityID:     &entityID,
	}
	created, err := w.repo.CreateRun(ctx, run)
	if err != nil {
		w.log.Error("automation worker: create run failed",
			"automation_id", a.ID, "err", err)
		return nil, err
	}

	var firstErr error
	for _, action := range a.Actions {
		if err := w.executeAction(ctx, action, a, evt); err != nil {
			w.log.Warn("automation worker: action failed",
				"automation_id", a.ID, "action", action.Type, "err", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	status := domain.RunStatusSucceeded
	errMsg := ""
	if firstErr != nil {
		status = domain.RunStatusFailed
		errMsg = firstErr.Error()
	}
	if err := w.repo.UpdateRun(ctx, created.ID, status, errMsg); err != nil {
		w.log.Error("automation worker: update run failed",
			"run_id", created.ID, "err", err)
		return created, err
	}
	created.Status = status
	created.ErrorMessage = errMsg
	now := time.Now().UTC()
	created.FinishedAt = &now
	return created, firstErr
}

// executeAction dispatches a single action.
func (w *AutomationWorker) executeAction(
	ctx context.Context,
	action domain.AutomationAction,
	_ *domain.Automation,
	evt AutomationEvent,
) error {
	orgCtx := domain.WithOrgID(ctx, evt.OrgID)

	switch action.Type {
	case domain.ActionAssignOwner:
		return w.execAssignOwner(orgCtx, action.Config, evt)
	case domain.ActionSendEmail:
		return w.execSendEmail(orgCtx, action.Config, evt)
	case domain.ActionEnrollInSequence:
		return w.execEnrollInSequence(orgCtx, action.Config, evt)
	case domain.ActionCreateActivity:
		return w.execCreateActivity(orgCtx, action.Config, evt)
	case domain.ActionWebhook:
		return w.execWebhook(orgCtx, action.Config, evt)
	default:
		w.log.Warn("automation worker: unknown action type", "type", action.Type)
		return nil
	}
}

// execAssignOwner sets the owner_id on the triggered entity.
// Config keys: owner_id (UUID string).
func (w *AutomationWorker) execAssignOwner(ctx context.Context, cfg map[string]interface{}, evt AutomationEvent) error {
	ownerStr, _ := cfg["owner_id"].(string)
	if ownerStr == "" {
		return fmt.Errorf("assign_owner: owner_id required in config")
	}
	ownerID, err := uuid.Parse(ownerStr)
	if err != nil {
		return fmt.Errorf("assign_owner: invalid owner_id: %w", err)
	}

	switch evt.EntityType {
	case "deal":
		_, err = w.deals.Update(ctx, evt.EntityID, domain.DealPatch{OwnerID: &ownerID})
	case "contact":
		_, err = w.contacts.Update(ctx, evt.EntityID, domain.ContactPatch{OwnerID: &ownerID})
	case "ticket":
		_, err = w.tickets.Update(ctx, evt.EntityID, domain.TicketPatch{AssigneeID: &ownerID})
	default:
		return fmt.Errorf("assign_owner: unsupported entity type %q", evt.EntityType)
	}
	return err
}

// execSendEmail sends a plain-text email.
// Config keys: to (string, required unless entity is contact), subject, body.
func (w *AutomationWorker) execSendEmail(ctx context.Context, cfg map[string]interface{}, evt AutomationEvent) error {
	if w.mailer == nil {
		return nil
	}

	to, _ := cfg["to"].(string)
	subject, _ := cfg["subject"].(string)
	body, _ := cfg["body"].(string)

	// If no explicit "to", try to resolve from the contact entity.
	if to == "" && evt.EntityType == "contact" {
		contact, err := w.contacts.GetByID(ctx, evt.EntityID)
		if err != nil {
			return fmt.Errorf("send_email: resolve contact: %w", err)
		}
		if contact.Email != nil {
			to = *contact.Email
		}
	}

	if to == "" {
		return fmt.Errorf("send_email: recipient address required")
	}
	if subject == "" {
		return fmt.Errorf("send_email: subject required")
	}

	return w.mailer.SendDirect(to, subject, body)
}

// execEnrollInSequence enrolls the entity (must be a contact) in a sequence.
// Config keys: sequence_id (UUID string).
func (w *AutomationWorker) execEnrollInSequence(ctx context.Context, cfg map[string]interface{}, evt AutomationEvent) error {
	seqStr, _ := cfg["sequence_id"].(string)
	if seqStr == "" {
		return fmt.Errorf("enroll_in_sequence: sequence_id required in config")
	}
	seqID, err := uuid.Parse(seqStr)
	if err != nil {
		return fmt.Errorf("enroll_in_sequence: invalid sequence_id: %w", err)
	}

	// Determine contact ID: either the entity itself or from config.
	contactID := evt.EntityID
	if evt.EntityType != "contact" {
		cidStr, _ := cfg["contact_id"].(string)
		if cidStr == "" {
			return fmt.Errorf("enroll_in_sequence: contact_id required for entity type %q", evt.EntityType)
		}
		contactID, err = uuid.Parse(cidStr)
		if err != nil {
			return fmt.Errorf("enroll_in_sequence: invalid contact_id: %w", err)
		}
	}

	_, err = w.sequences.Enroll(ctx, seqID, domain.EnrollRequest{
		ContactIDs: []uuid.UUID{contactID},
	})
	return err
}

// execCreateActivity creates a task/activity linked to the entity.
// Config keys: type (default "task"), subject (required), description, owner_id.
func (w *AutomationWorker) execCreateActivity(ctx context.Context, cfg map[string]interface{}, evt AutomationEvent) error {
	actType := domain.ActivityTypeTask
	if t, ok := cfg["type"].(string); ok && t != "" {
		actType = domain.ActivityType(t)
	}
	subject, _ := cfg["subject"].(string)
	if subject == "" {
		subject, _ = cfg["title"].(string)
	}
	if subject == "" {
		return fmt.Errorf("create_activity: subject required in config")
	}

	// Resolve owner: from config, or fall back to entity's owner.
	var ownerID uuid.UUID
	if ownerStr, ok := cfg["owner_id"].(string); ok && ownerStr != "" {
		var err error
		ownerID, err = uuid.Parse(ownerStr)
		if err != nil {
			return fmt.Errorf("create_activity: invalid owner_id: %w", err)
		}
	}
	if ownerID == uuid.Nil {
		if ownerStr, ok := evt.Data["owner_id"].(string); ok {
			ownerID, _ = uuid.Parse(ownerStr)
		}
	}
	if ownerID == uuid.Nil {
		return fmt.Errorf("create_activity: owner_id required")
	}

	var desc *string
	if d, ok := cfg["description"].(string); ok && d != "" {
		desc = &d
	}

	a := &domain.Activity{
		Type:    actType,
		Subject: subject,
		OrgID:   evt.OrgID,
		OwnerID: ownerID,
	}
	if desc != nil {
		a.Description = desc
	}

	// Link to the triggering entity.
	switch evt.EntityType {
	case "contact":
		a.ContactID = &evt.EntityID
	case "deal":
		a.DealID = &evt.EntityID
	}

	_, err := w.activities.Create(ctx, a)
	return err
}

// privateRanges contains CIDR blocks that must not be reachable from automation webhooks.
var privateRanges = func() []*net.IPNet {
	cidrs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16", // link-local / AWS metadata
		"100.64.0.0/10",  // shared address space (RFC 6598)
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}
	var nets []*net.IPNet
	for _, c := range cidrs {
		_, ipnet, _ := net.ParseCIDR(c)
		nets = append(nets, ipnet)
	}
	return nets
}()

func isPrivateIP(ip net.IP) bool {
	for _, r := range privateRanges {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}

// safeDialContext resolves the target hostname and rejects connections to
// private or link-local addresses to prevent SSRF.
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("webhook: invalid address %q: %w", addr, err)
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("webhook: DNS lookup %q: %w", host, err)
	}
	for _, ip := range ips {
		if isPrivateIP(ip.IP) {
			return nil, fmt.Errorf("webhook: target %q resolves to a private IP — blocked", host)
		}
	}
	return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

// execWebhook POSTs a JSON payload to a configured URL.
// Config keys: url (string, required), timeout_seconds (int, 1–30, default 10).
//
// Security controls:
//   - Only https:// scheme is accepted.
//   - The target hostname is resolved before connecting; private/link-local IPs are blocked.
//   - HTTP redirects are not followed to prevent SSRF via redirect chains.
//   - Timeout is clamped to [1, 30] seconds.
func (w *AutomationWorker) execWebhook(_ context.Context, cfg map[string]interface{}, evt AutomationEvent) error {
	webhookURL, _ := cfg["url"].(string)
	if webhookURL == "" {
		return fmt.Errorf("webhook: url required in config")
	}

	parsed, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("webhook: invalid url %q: %w", webhookURL, err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("webhook: only https:// URLs are allowed, got %q", parsed.Scheme)
	}

	timeoutSec := 10
	if t, ok := cfg["timeout_seconds"].(float64); ok {
		timeoutSec = int(t)
	}
	if timeoutSec < 1 {
		timeoutSec = 1
	} else if timeoutSec > 30 {
		timeoutSec = 30
	}

	transport := &http.Transport{
		DialContext: safeDialContext,
	}
	client := &http.Client{
		Timeout:   time.Duration(timeoutSec) * time.Second,
		Transport: transport,
		// Do not follow redirects — a redirect could point to a private address.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	payload := map[string]interface{}{
		"trigger":     evt.TriggerType,
		"entity_type": evt.EntityType,
		"entity_id":   evt.EntityID.String(),
		"org_id":      evt.OrgID.String(),
		"data":        evt.Data,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: POST %s: %w", webhookURL, err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: POST %s returned %d", webhookURL, resp.StatusCode)
	}
	return nil
}

// --- Condition evaluation (exported for tests) ---

// EvaluateConditions returns true when all conditions pass (AND logic).
// An empty conditions slice always returns true.
// Disabled rules must not be passed here — callers are responsible for
// filtering out automations whose status != active.
func EvaluateConditions(conditions []domain.AutomationCondition, data map[string]interface{}) bool {
	for _, c := range conditions {
		if !evaluateCondition(c, data) {
			return false
		}
	}
	return true
}

func evaluateCondition(c domain.AutomationCondition, data map[string]interface{}) bool {
	val, exists := data[c.Field]

	switch c.Operator {
	case domain.ConditionOpIsSet:
		return exists && val != nil && val != ""
	case domain.ConditionOpIsNotSet:
		return !exists || val == nil || val == ""
	}

	if !exists {
		return false
	}

	valStr := fmt.Sprintf("%v", val)
	condStr := fmt.Sprintf("%v", c.Value)

	switch c.Operator {
	case domain.ConditionOpEquals:
		return strings.EqualFold(valStr, condStr)
	case domain.ConditionOpNotEquals:
		return !strings.EqualFold(valStr, condStr)
	case domain.ConditionOpContains:
		return strings.Contains(strings.ToLower(valStr), strings.ToLower(condStr))
	case domain.ConditionOpNotContains:
		return !strings.Contains(strings.ToLower(valStr), strings.ToLower(condStr))
	case domain.ConditionOpGreaterThan:
		return compareNumeric(valStr, condStr) > 0
	case domain.ConditionOpLessThan:
		return compareNumeric(valStr, condStr) < 0
	}

	return false
}

// compareNumeric parses both values as float64 and returns -1, 0, or 1.
// Returns 0 on parse failure.
func compareNumeric(a, b string) int {
	var fa, fb float64
	if _, err := fmt.Sscanf(a, "%f", &fa); err != nil {
		return 0
	}
	if _, err := fmt.Sscanf(b, "%f", &fb); err != nil {
		return 0
	}
	if fa < fb {
		return -1
	}
	if fa > fb {
		return 1
	}
	return 0
}
