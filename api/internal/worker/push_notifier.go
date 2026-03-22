package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/repository"
)

// PushNotifier sends Web Push notifications to subscribed users.
type PushNotifier struct {
	repo           repository.PushSubscriptionRepository
	vapidPublicKey  string
	vapidPrivateKey string
	subject        string // VAPID subject — typically "mailto:..." or a URL
	logger         *slog.Logger
}

// NewPushNotifier constructs a PushNotifier.
// vapidPublicKey and vapidPrivateKey are the VAPID key pair (base64url-encoded).
// subject is sent in the VAPID Authorization header (e.g. "mailto:support@example.com").
func NewPushNotifier(
	repo repository.PushSubscriptionRepository,
	vapidPublicKey, vapidPrivateKey, subject string,
	logger *slog.Logger,
) *PushNotifier {
	return &PushNotifier{
		repo:           repo,
		vapidPublicKey:  vapidPublicKey,
		vapidPrivateKey: vapidPrivateKey,
		subject:        subject,
		logger:         logger,
	}
}

// Enabled reports whether VAPID keys are configured.
func (n *PushNotifier) Enabled() bool {
	return n.vapidPublicKey != "" && n.vapidPrivateKey != ""
}

// pushPayload is the JSON body delivered to the browser.
type pushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url,omitempty"`
}

func (n *PushNotifier) send(userID, orgID uuid.UUID, payload pushPayload) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		subs, err := n.repo.ListByUser(ctx, userID, orgID)
		if err != nil {
			n.logger.Error("push_notifier: list subscriptions failed", "user_id", userID, "err", err)
			return
		}
		if len(subs) == 0 {
			return
		}

		body, err := json.Marshal(payload)
		if err != nil {
			n.logger.Error("push_notifier: marshal payload failed", "err", err)
			return
		}

		for _, sub := range subs {
			s := &webpush.Subscription{
				Endpoint: sub.Endpoint,
				Keys: webpush.Keys{
					P256dh: sub.P256dh,
					Auth:   sub.Auth,
				},
			}
			resp, err := webpush.SendNotification(body, s, &webpush.Options{
				VAPIDPublicKey:  n.vapidPublicKey,
				VAPIDPrivateKey: n.vapidPrivateKey,
				Subscriber:      n.subject,
				TTL:             30,
			})
			if err != nil {
				n.logger.Warn("push_notifier: send failed", "endpoint", sub.Endpoint, "err", err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == 410 || resp.StatusCode == 404 {
				// Subscription gone — remove it.
				_ = n.repo.DeleteByEndpoint(ctx, userID, sub.Endpoint)
			}
		}
	}()
}

// NotifyTicketAssigned fires a push to the newly assigned user.
func (n *PushNotifier) NotifyTicketAssigned(orgID, userID, ticketID uuid.UUID, subject string) {
	n.send(userID, orgID, pushPayload{
		Title: "Ticket assigned to you",
		Body:  fmt.Sprintf("%s", subject),
		URL:   fmt.Sprintf("/tickets/%s", ticketID),
	})
}

// NotifyTicketCommented fires a push to the ticket assignee when an agent comments.
func (n *PushNotifier) NotifyTicketCommented(orgID, assigneeID, ticketID uuid.UUID, ticketSubject string) {
	n.send(assigneeID, orgID, pushPayload{
		Title: "New comment on your ticket",
		Body:  ticketSubject,
		URL:   fmt.Sprintf("/tickets/%s", ticketID),
	})
}

// NotifyDealStageChanged fires a push to the deal owner when the stage changes.
func (n *PushNotifier) NotifyDealStageChanged(orgID, ownerID, dealID uuid.UUID, dealTitle, newStage string) {
	n.send(ownerID, orgID, pushPayload{
		Title: "Deal stage changed",
		Body:  fmt.Sprintf("%s → %s", dealTitle, newStage),
		URL:   fmt.Sprintf("/deals/%s", dealID),
	})
}
