package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/repository"
)

// TeamsNotifier sends Adaptive Card messages to Microsoft Teams via Incoming Webhooks.
type TeamsNotifier struct {
	repo   repository.TeamsConnectionRepository
	appURL string
	client *http.Client
	logger *slog.Logger
}

func NewTeamsNotifier(repo repository.TeamsConnectionRepository, appURL string, logger *slog.Logger) *TeamsNotifier {
	return &TeamsNotifier{
		repo:   repo,
		appURL: appURL,
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

// teamsMessage is the wrapper for Teams Incoming Webhook payloads.
type teamsMessage struct {
	Type        string            `json:"type"`
	Attachments []teamsAttachment `json:"attachments"`
}

type teamsAttachment struct {
	ContentType string          `json:"contentType"`
	Content     json.RawMessage `json:"content"`
}

// buildAdaptiveCard builds a minimal Adaptive Card JSON.
func buildAdaptiveCard(title, body, actionURL, actionTitle string) json.RawMessage {
	card := map[string]any{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": "1.4",
		"body": []map[string]any{
			{
				"type":   "TextBlock",
				"text":   title,
				"weight": "Bolder",
				"size":   "Medium",
				"wrap":   true,
			},
			{
				"type": "TextBlock",
				"text": body,
				"wrap": true,
			},
		},
		"actions": []map[string]any{
			{
				"type":  "Action.OpenUrl",
				"title": actionTitle,
				"url":   actionURL,
			},
		},
	}
	b, _ := json.Marshal(card)
	return b
}

func (n *TeamsNotifier) send(orgID uuid.UUID, title, body, actionURL, actionTitle string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		conn, err := n.repo.GetByOrgID(ctx, orgID)
		if err != nil {
			n.logger.Error("teams_notifier: failed to get connection", "org_id", orgID, "err", err)
			return
		}
		if conn == nil {
			return // Teams not configured for this org
		}

		card := buildAdaptiveCard(title, body, actionURL, actionTitle)
		msg := teamsMessage{
			Type: "message",
			Attachments: []teamsAttachment{
				{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content:     card,
				},
			},
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			n.logger.Error("teams_notifier: marshal failed", "err", err)
			return
		}

		resp, err := n.client.Post(conn.BotToken, "application/json", bytes.NewReader(payload))
		if err != nil {
			n.logger.Error("teams_notifier: POST failed", "org_id", orgID, "err", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			n.logger.Warn("teams_notifier: non-2xx response", "status", resp.StatusCode, "org_id", orgID)
		}
	}()
}

// NotifyDealStageChanged fires a fire-and-forget Teams message when a deal stage changes.
func (n *TeamsNotifier) NotifyDealStageChanged(orgID, dealID uuid.UUID, dealTitle, newStage string) {
	actionURL := fmt.Sprintf("%s/deals/%s", n.appURL, dealID)
	body := fmt.Sprintf("Deal **%s** moved to stage **%s**.", dealTitle, newStage)
	n.send(orgID, "Deal Stage Changed", body, actionURL, "View Deal")
}

// NotifyTicketAssigned fires a Teams message when a ticket is assigned.
func (n *TeamsNotifier) NotifyTicketAssigned(orgID, ticketID uuid.UUID, subject string) {
	actionURL := fmt.Sprintf("%s/tickets/%s", n.appURL, ticketID)
	body := fmt.Sprintf("Ticket **%s** has been assigned.", subject)
	n.send(orgID, "Ticket Assigned", body, actionURL, "View Ticket")
}

// NotifyTicketStatusChanged fires a Teams message when a ticket is resolved or closed.
func (n *TeamsNotifier) NotifyTicketStatusChanged(orgID, ticketID uuid.UUID, subject, status string) {
	actionURL := fmt.Sprintf("%s/tickets/%s", n.appURL, ticketID)
	body := fmt.Sprintf("Ticket **%s** status changed to **%s**.", subject, status)
	n.send(orgID, "Ticket Status Changed", body, actionURL, "View Ticket")
}
