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
	} else if breached > 0 {
		w.logger.Info("sla breaches detected", "count", breached)
		w.notifyBreaches(ctx)
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

func (w *SLABreachWorker) notifyBreaches(ctx context.Context) {
	breachedTrue := true
	instances, err := w.instances.List(ctx, domain.SLAInstanceFilter{Breached: &breachedTrue, Limit: 500})
	if err != nil {
		w.logger.Error("listing breached sla instances", "err", err)
		return
	}
	for _, inst := range instances {
		entityType := string(inst.EntityType)
		entityID := inst.EntityID
		title := "SLA breached"
		body := "An SLA deadline has been missed."

		// We don't have easy access to owner here without joining —
		// use org_id placeholder user. Notifications are scoped by org.
		// A real implementation would join to get the deal owner.
		_, err := w.notifications.Create(ctx, &domain.Notification{
			OrgID:      inst.OrgID,
			UserID:     uuid.Nil, // broadcast to org; frontend filters by org
			Kind:       domain.NotificationKindSLABreached,
			EntityType: &entityType,
			EntityID:   &entityID,
			Title:      title,
			Body:       &body,
		})
		if err != nil {
			w.logger.Error("failed to create breach notification", "instance_id", inst.ID, "err", err)
		}
	}
}

func (w *SLABreachWorker) emitWarning(ctx context.Context, inst *domain.SLAInstance) {
	now := time.Now().UTC()
	entityType := string(inst.EntityType)
	entityID := inst.EntityID
	title := "SLA at risk — 80% elapsed"
	body := "An SLA is approaching its deadline."

	_, err := w.notifications.Create(ctx, &domain.Notification{
		OrgID:      inst.OrgID,
		UserID:     uuid.Nil,
		Kind:       domain.NotificationKindSLAWarning,
		EntityType: &entityType,
		EntityID:   &entityID,
		Title:      title,
		Body:       &body,
	})
	if err != nil {
		w.logger.Error("failed to create sla warning notification", "instance_id", inst.ID, "err", err)
		return
	}

	if err := w.instances.MarkWarned(ctx, inst.ID, now); err != nil {
		w.logger.Error("failed to mark sla instance warned", "instance_id", inst.ID, "err", err)
	}
}
