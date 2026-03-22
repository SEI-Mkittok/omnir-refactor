package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	"github.com/omnir/crm-api/internal/repository"
)

// ReportSchedulerWorker checks scheduled dashboard reports every minute and
// emails recipients when a schedule is due.
type ReportSchedulerWorker struct {
	dashboardRepo repository.DashboardRepository
	reportsRepo   repository.ReportsRepository
	mailer        *email.Mailer
	interval      time.Duration
	logger        *slog.Logger
	parser        cron.Parser
}

// NewReportSchedulerWorker creates a worker that ticks on interval.
func NewReportSchedulerWorker(
	dashboardRepo repository.DashboardRepository,
	reportsRepo repository.ReportsRepository,
	mailer *email.Mailer,
	interval time.Duration,
	logger *slog.Logger,
) *ReportSchedulerWorker {
	return &ReportSchedulerWorker{
		dashboardRepo: dashboardRepo,
		reportsRepo:   reportsRepo,
		mailer:        mailer,
		interval:      interval,
		logger:        logger,
		parser:        cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// Start runs the worker in a background goroutine until ctx is cancelled.
func (w *ReportSchedulerWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		w.logger.Info("report scheduler worker started", "interval", w.interval)
		for {
			select {
			case <-ticker.C:
				w.tick(ctx, time.Now().UTC())
			case <-ctx.Done():
				w.logger.Info("report scheduler worker stopped")
				return
			}
		}
	}()
}

func (w *ReportSchedulerWorker) tick(ctx context.Context, now time.Time) {
	schedules, err := w.dashboardRepo.ListAllDueSchedules(ctx, now)
	if err != nil {
		w.logger.Error("report scheduler: failed to load schedules", "err", err)
		return
	}

	for _, s := range schedules {
		if !w.isDue(s, now) {
			continue
		}
		if err := w.deliver(ctx, s, now); err != nil {
			w.logger.Error("report scheduler: delivery failed",
				"schedule_id", s.ID, "dashboard_id", s.DashboardID, "err", err)
			continue
		}
		if err := w.dashboardRepo.MarkScheduleSent(ctx, s.ID, now); err != nil {
			w.logger.Error("report scheduler: failed to mark sent",
				"schedule_id", s.ID, "err", err)
		}
	}
}

// isDue returns true if the schedule's cron expression fired between
// last_sent_at (or the epoch) and now.
func (w *ReportSchedulerWorker) isDue(s *domain.ScheduledReport, now time.Time) bool {
	sched, err := w.parser.Parse(s.Schedule)
	if err != nil {
		w.logger.Warn("report scheduler: invalid cron expression",
			"schedule_id", s.ID, "expr", s.Schedule, "err", err)
		return false
	}

	from := time.Time{} // epoch — first run
	if s.LastSentAt != nil {
		from = *s.LastSentAt
	}

	// Due if the next occurrence after 'from' is on or before 'now'.
	next := sched.Next(from)
	return !next.After(now)
}

// deliver builds an HTML email report for a scheduled dashboard and sends it.
func (w *ReportSchedulerWorker) deliver(ctx context.Context, s *domain.ScheduledReport, now time.Time) error {
	orgCtx := domain.WithOrgID(ctx, s.OrgID)
	html := buildReportHTML(orgCtx, w.reportsRepo, s, now)
	subject := fmt.Sprintf("[Omnir] Scheduled dashboard report — %s", now.Format("2006-01-02"))

	for _, recipient := range s.Recipients {
		if err := w.mailer.SendDirect(recipient, subject, html); err != nil {
			return fmt.Errorf("send to %s: %w", recipient, err)
		}
	}
	return nil
}

// buildReportHTML constructs an HTML summary of key CRM metrics for email delivery.
func buildReportHTML(ctx context.Context, repo repository.ReportsRepository, s *domain.ScheduledReport, now time.Time) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<h2>Dashboard Report — %s</h2>\n", now.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("<p>Schedule: <code>%s</code> | Dashboard: %s</p>\n", s.Schedule, s.DashboardID))

	if deals, err := repo.DealsByStage(ctx); err == nil {
		sb.WriteString("<h3>Deals by Stage</h3><ul>")
		for _, d := range deals {
			sb.WriteString(fmt.Sprintf("<li>%s: %d deals, $%.2f</li>",
				d.Stage, d.Count, float64(d.TotalValueCents)/100))
		}
		sb.WriteString("</ul>\n")
	}

	if contacts, err := repo.ContactsMonthly(ctx); err == nil && len(contacts) > 0 {
		latest := contacts[len(contacts)-1]
		sb.WriteString(fmt.Sprintf("<p><strong>New contacts this month (%s):</strong> %d</p>\n",
			latest.Month, latest.Count))
	}

	if activities, err := repo.ActivitiesByType(ctx); err == nil {
		sb.WriteString("<h3>Activities by Type</h3><ul>")
		for _, a := range activities {
			sb.WriteString(fmt.Sprintf("<li>%s: %d</li>", a.Type, a.Count))
		}
		sb.WriteString("</ul>\n")
	}

	return sb.String()
}
