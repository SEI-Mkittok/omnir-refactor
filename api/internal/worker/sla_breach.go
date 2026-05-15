package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// SLABreachWorker scans open SLA instances every interval, marks breaches,
// and emits warning / breach notifications to deal owners.
type SLABreachWorker struct {
	instances     repository.SLAInstanceRepository
	notifications repository.NotificationRepository
	interval      time.Duration
	logger        *slog.Logger
	stop          chan struct{}
}

type slaNotificationRecipientResolver interface {
	NotificationRecipients(ctx context.Context, inst *domain.SLAInstance) ([]uuid.UUID, error)
}

func NewSLABreachWorker(
	instances repository.SLAInstanceRepository,
	notifications repository.NotificationRepository,
	interval time.Duration,
	logger *slog.Logger,
) *SLABreachWorker {
	return &SLABreachWorker{
		instances:     instances,
		notifications: notifications,
		interval:      interval,
		logger:        logger,
		stop:          make(chan struct{}),
	}
}

func (w *SLABreachWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		w.logger.Info("sla breach worker started", "interval", w.interval)
		for {
			select {
			case <-ticker.C:
				w.tick(ctx)
			case <-w.stop:
				w.logger.Info("sla breach worker stopped")
				return
			case <-ctx.Done():
				w.logger.Info("sla breach worker context cancelled")
				return
			}
		}
	}()
}

func (w *SLABreachWorker) Stop() {
	close(w.stop)
}

func (w *SLABreachWorker) tick(ctx context.Context) {
	// Scan and mark breached instances
	breached, err := w.instances.ScanBreaches(ctx)
	if err != nil {
		w.logger.Error("sla breach scan failed", "err", err)
	} else if len(breached) > 0 {
		w.logger.Info("sla breaches detected", "count", len(breached))
		for _, inst := range breached {
			w.emitBreach(ctx, inst)
		}
	}

	// Scan and emit 80% warnings
	warnings, err := w.instances.ScanWarnings(ctx)
	if err != nil {
		w.logger.Error("sla warning scan failed", "err", err)
		return
	}
	for _, inst := range warnings {
		w.emitWarning(ctx, inst)
	}
}

func (w *SLABreachWorker) emitWarning(ctx context.Context, inst *domain.SLAInstance) {
	now := time.Now().UTC()
	entityType := string(inst.EntityType)
	entityID := inst.EntityID
	title := "SLA at risk — 80% elapsed"
	body := "An SLA is approaching its deadline."

	if !w.emitToRecipients(ctx, inst, domain.NotificationKindSLAWarning, entityType, entityID, title, body) {
		return
	}

	if err := w.instances.MarkWarned(ctx, inst.ID, now); err != nil {
		w.logger.Error("failed to mark sla instance warned", "instance_id", inst.ID, "err", err)
	}
}

func (w *SLABreachWorker) emitBreach(ctx context.Context, inst *domain.SLAInstance) {
	entityType := string(inst.EntityType)
	entityID := inst.EntityID
	title := "SLA breached"
	body := "An SLA deadline has been missed."

	w.emitToRecipients(ctx, inst, domain.NotificationKindSLABreached, entityType, entityID, title, body)
}

func (w *SLABreachWorker) emitToRecipients(ctx context.Context, inst *domain.SLAInstance, kind domain.NotificationKind, entityType string, entityID uuid.UUID, title, body string) bool {
	resolver, ok := w.instances.(slaNotificationRecipientResolver)
	if !ok {
		w.logger.Warn("sla notification recipients unavailable", "instance_id", inst.ID)
		return false
	}
	recipients, err := resolver.NotificationRecipients(ctx, inst)
	if err != nil {
		w.logger.Error("failed to resolve sla notification recipients", "instance_id", inst.ID, "err", err)
		return false
	}
	if len(recipients) == 0 {
		w.logger.Warn("sla notification skipped with no recipients", "instance_id", inst.ID, "entity_type", entityType)
		return false
	}

	ok = true
	for _, userID := range recipients {
		_, err := w.notifications.Create(ctx, &domain.Notification{
			OrgID:      inst.OrgID,
			UserID:     userID,
			Kind:       kind,
			EntityType: &entityType,
			EntityID:   &entityID,
			Title:      title,
			Body:       &body,
		})
		if err != nil {
			ok = false
			w.logger.Error("failed to create sla notification", "instance_id", inst.ID, "user_id", userID, "err", err)
		}
	}
	return ok
}
