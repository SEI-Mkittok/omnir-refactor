package email

import "fmt"

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
