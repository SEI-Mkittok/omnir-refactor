package email

import (
	"fmt"
	"regexp"
	"strings"
)

// Mailer renders and sends typed email events.
type Mailer struct {
	sender *Sender
	appURL string
}

// NewMailer constructs a Mailer.
func NewMailer(sender *Sender, appURL string) *Mailer {
	return &Mailer{sender: sender, appURL: appURL}
}

// SendAssigned sends an email notifying assigneeName (at toEmail) that ticket was assigned.
func (m *Mailer) SendAssigned(toEmail, assigneeName, ticketID, subject string) error {
	html, text, err := RenderAssigned(TicketAssignedData{
		AssigneeName: assigneeName,
		TicketID:     ticketID,
		Subject:      subject,
		AppURL:       m.appURL,
	})
	if err != nil {
		return err
	}
	return m.sender.Send(Message{
		To:      toEmail,
		Subject: fmt.Sprintf("[Omnir] Ticket #%s assigned to you", ticketID),
		HTML:    html,
		Text:    text,
	})
}

// SendDirect sends a plain-text/HTML email directly from the CRM compose UI.
// If SMTP is disabled the call is a no-op and returns nil.
func (m *Mailer) SendDirect(toEmail, subject, body string) error {
	return m.sender.Send(Message{
		To:      toEmail,
		Subject: subject,
		HTML:    fmt.Sprintf("<p>%s</p>", body),
		Text:    body,
	})
}

// SequenceTokens bundles the pre-generated tracking tokens for one sequence email send.
type SequenceTokens struct {
	OpenToken        string
	ClickToken       string
	UnsubscribeToken string
}

// SendSequenceEmail sends a sequence step email with tracking pixel, click wrapping,
// and unsubscribe footer injected into the HTML body.
func (m *Mailer) SendSequenceEmail(toEmail, subject, htmlBody string, tokens SequenceTokens) error {
	baseURL := strings.TrimRight(m.appURL, "/")

	// Inject tracking pixel just before </body> (or at end).
	pixel := fmt.Sprintf(`<img src="%s/track/open/%s" width="1" height="1" style="display:none" alt="">`,
		baseURL, tokens.OpenToken)
	if idx := strings.LastIndex(htmlBody, "</body>"); idx != -1 {
		htmlBody = htmlBody[:idx] + pixel + htmlBody[idx:]
	} else {
		htmlBody += pixel
	}

	// Rewrite <a href="..."> links to click-tracking redirect.
	htmlBody = rewriteLinks(htmlBody, baseURL, tokens.ClickToken)

	// Append unsubscribe footer.
	footer := fmt.Sprintf(
		`<p style="font-size:11px;color:#999;margin-top:24px">Don't want these emails? `+
			`<a href="%s/unsubscribe/%s">Unsubscribe</a></p>`,
		baseURL, tokens.UnsubscribeToken)
	if idx := strings.LastIndex(htmlBody, "</body>"); idx != -1 {
		htmlBody = htmlBody[:idx] + footer + htmlBody[idx:]
	} else {
		htmlBody += footer
	}

	return m.sender.Send(Message{
		To:      toEmail,
		Subject: subject,
		HTML:    htmlBody,
		Text:    "",
	})
}

var hrefRe = regexp.MustCompile(`(?i)<a\s[^>]*href="(https?://[^"]+)"`)

// rewriteLinks wraps all http(s) href values in a click-tracking redirect URL.
func rewriteLinks(html, baseURL, clickToken string) string {
	return hrefRe.ReplaceAllStringFunc(html, func(match string) string {
		sub := hrefRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		original := sub[1]
		wrapped := fmt.Sprintf(`%s/track/click/%s?url=%s`, baseURL, clickToken, original)
		return strings.Replace(match, `"`+original+`"`, `"`+wrapped+`"`, 1)
	})
}

// SendComment sends an email notifying a ticket participant about a new comment.
func (m *Mailer) SendComment(toEmail, recipientName, ticketID, subject, commentBody string) error {
	html, text, err := RenderComment(TicketCommentData{
		RecipientName: recipientName,
		TicketID:      ticketID,
		Subject:       subject,
		CommentBody:   commentBody,
		AppURL:        m.appURL,
	})
	if err != nil {
		return err
	}
	return m.sender.Send(Message{
		To:      toEmail,
		Subject: fmt.Sprintf("[Omnir] New comment on ticket #%s", ticketID),
		HTML:    html,
		Text:    text,
	})
}

// SendResolved sends an email notifying the reporter that ticket was resolved or closed.
func (m *Mailer) SendResolved(toEmail, reporterName, ticketID, subject, status string) error {
	html, text, err := RenderResolved(TicketResolvedData{
		ReporterName: reporterName,
		TicketID:     ticketID,
		Subject:      subject,
		Status:       status,
		AppURL:       m.appURL,
	})
	if err != nil {
		return err
	}
	return m.sender.Send(Message{
		To:      toEmail,
		Subject: fmt.Sprintf("[Omnir] Your ticket #%s has been %s", ticketID, status),
		HTML:    html,
		Text:    text,
	})
}
