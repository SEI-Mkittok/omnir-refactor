package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// retryDelays defines the exponential backoff schedule for webhook deliveries.
var retryDelays = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	8 * time.Hour,
}

const maxAttempts = 5
const deliveryTimeout = 10 * time.Second

// WebhookDispatcher polls for pending webhook deliveries and dispatches them.
type WebhookDispatcher struct {
	repo     repository.OutboundWebhookRepository
	client   *http.Client
	interval time.Duration
	logger   *slog.Logger

	// Dispatch is the channel for submitting new events from CRM handlers.
	Dispatch chan WebhookEvent
}

// WebhookEvent carries an event and its payload to be fanned out to registered webhooks.
type WebhookEvent struct {
	OrgID    uuid.UUID
	EntityID uuid.UUID
	Event    domain.WebhookEvent
	Data     any
}

// NewWebhookDispatcher creates a dispatcher that polls every interval.
func NewWebhookDispatcher(repo repository.OutboundWebhookRepository, interval time.Duration, logger *slog.Logger) *WebhookDispatcher {
	return &WebhookDispatcher{
		repo:     repo,
		client:   &http.Client{Timeout: deliveryTimeout},
		interval: interval,
		logger:   logger,
		Dispatch: make(chan WebhookEvent, 256),
	}
}

// Start launches the dispatch loop. It processes the Dispatch channel and polls for retries.
func (d *WebhookDispatcher) Start(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	go func() {
		defer ticker.Stop()
		d.logger.Info("webhook dispatcher started", "interval", d.interval)
		for {
			select {
			case evt := <-d.Dispatch:
				d.fanOut(ctx, evt)
			case <-ticker.C:
				d.processRetries(ctx)
			case <-ctx.Done():
				d.logger.Info("webhook dispatcher stopped")
				return
			}
		}
	}()
}

// fanOut looks up webhooks subscribed to evt.Event and queues a delivery for each.
func (d *WebhookDispatcher) fanOut(ctx context.Context, evt WebhookEvent) {
	payload := domain.WebhookEventPayload{
		Event:    evt.Event,
		OrgID:    evt.OrgID,
		EntityID: evt.EntityID,
		Data:     evt.Data,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		d.logger.Error("webhook: failed to marshal payload", "event", evt.Event, "err", err)
		return
	}

	// Use a background context so delivery queue is not bound to the request lifecycle.
	bgCtx := context.Background()

	webhooks, err := d.repo.ListByEvent(bgCtx, evt.OrgID, evt.Event)
	if err != nil {
		d.logger.Error("webhook: failed to list webhooks", "event", evt.Event, "err", err)
		return
	}

	for _, wh := range webhooks {
		delivery := &domain.WebhookDelivery{
			WebhookID: wh.ID,
			Event:     evt.Event,
			Payload:   body,
		}
		saved, err := d.repo.CreateDelivery(bgCtx, delivery)
		if err != nil {
			d.logger.Error("webhook: failed to create delivery", "webhook_id", wh.ID, "err", err)
			continue
		}
		// Attempt immediate delivery in a goroutine.
		go d.deliver(bgCtx, wh, saved)
	}
}

// processRetries picks up pending deliveries with elapsed next_retry_at.
func (d *WebhookDispatcher) processRetries(ctx context.Context) {
	deliveries, err := d.repo.PendingDeliveries(ctx)
	if err != nil {
		d.logger.Error("webhook: failed to fetch pending deliveries", "err", err)
		return
	}
	for _, del := range deliveries {
		wh, err := d.repo.GetByID(ctx, del.WebhookID)
		if err != nil {
			d.logger.Error("webhook: webhook not found for delivery", "delivery_id", del.ID, "err", err)
			continue
		}
		go d.deliver(ctx, wh, del)
	}
}

// deliver sends a single delivery attempt to the webhook URL.
func (d *WebhookDispatcher) deliver(ctx context.Context, wh *domain.Webhook, del *domain.WebhookDelivery) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(del.Payload))
	if err != nil {
		d.markFailed(ctx, del, "failed to build request: "+err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Omnir-Event", string(del.Event))
	req.Header.Set("X-Omnir-Delivery", del.ID.String())
	req.Header.Set("X-Omnir-Signature", computeSignature(wh.Secret, del.Payload))

	resp, err := d.client.Do(req)
	del.Attempts++
	if err != nil {
		d.markFailed(ctx, del, "HTTP error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		now := time.Now().UTC()
		del.Status = domain.DeliveryStatusDelivered
		del.DeliveredAt = &now
		del.NextRetryAt = nil
		if err := d.repo.UpdateDelivery(ctx, del); err != nil {
			d.logger.Error("webhook: failed to mark delivered", "delivery_id", del.ID, "err", err)
		}
		d.logger.Info("webhook: delivered", "delivery_id", del.ID, "url", wh.URL, "status", resp.StatusCode)
		return
	}

	d.markFailed(ctx, del, fmt.Sprintf("HTTP %d", resp.StatusCode))
}

func (d *WebhookDispatcher) markFailed(ctx context.Context, del *domain.WebhookDelivery, reason string) {
	d.logger.Warn("webhook: delivery failed", "delivery_id", del.ID, "attempt", del.Attempts, "reason", reason)
	del.LastError = &reason

	if del.Attempts >= maxAttempts {
		del.Status = domain.DeliveryStatusFailed
		del.NextRetryAt = nil
	} else {
		next := time.Now().UTC().Add(retryDelays[del.Attempts-1])
		del.NextRetryAt = &next
	}

	if err := d.repo.UpdateDelivery(ctx, del); err != nil {
		d.logger.Error("webhook: failed to update delivery after failure", "delivery_id", del.ID, "err", err)
	}
}

// computeSignature returns an HMAC-SHA256 hex signature of the payload.
func computeSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

