package email

import (
	"bytes"
	"fmt"
	"html/template"
	texttemplate "text/template"
)

// TicketAssignedData is passed to the assigned email templates.
type TicketAssignedData struct {
	AssigneeName string
	TicketID     string
	Subject      string
	AppURL       string
}

// TicketResolvedData is passed to the resolved/closed email templates.
type TicketResolvedData struct {
	ReporterName string
	TicketID     string
	Subject      string
	Status       string // "resolved" or "closed"
	AppURL       string
}

var assignedHTML = template.Must(template.New("assigned_html").Parse(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;color:#1e293b;max-width:600px;margin:0 auto;padding:24px">
  <h2 style="color:#4f46e5">Ticket assigned to you</h2>
  <p>Hi {{.AssigneeName}},</p>
  <p>Ticket <strong>#{{.TicketID}}</strong> has been assigned to you:</p>
  <blockquote style="border-left:3px solid #4f46e5;padding:8px 16px;margin:16px 0;color:#475569">
    {{.Subject}}
  </blockquote>
  <p>
    <a href="{{.AppURL}}/tickets/{{.TicketID}}"
       style="background:#4f46e5;color:#fff;padding:10px 20px;border-radius:6px;text-decoration:none;display:inline-block">
      View ticket
    </a>
  </p>
  <p style="color:#94a3b8;font-size:12px">You are receiving this because you were assigned to this ticket.</p>
</body>
</html>`))

var assignedText = texttemplate.Must(texttemplate.New("assigned_text").Parse(`Hi {{.AssigneeName}},

Ticket #{{.TicketID}} has been assigned to you:

  {{.Subject}}

View ticket: {{.AppURL}}/tickets/{{.TicketID}}
`))

var resolvedHTML = template.Must(template.New("resolved_html").Parse(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;color:#1e293b;max-width:600px;margin:0 auto;padding:24px">
  <h2 style="color:#4f46e5">Your ticket has been {{.Status}}</h2>
  <p>Hi {{.ReporterName}},</p>
  <p>Ticket <strong>#{{.TicketID}}</strong> has been marked as <strong>{{.Status}}</strong>:</p>
  <blockquote style="border-left:3px solid #4f46e5;padding:8px 16px;margin:16px 0;color:#475569">
    {{.Subject}}
  </blockquote>
  <p>
    <a href="{{.AppURL}}/portal/tickets/{{.TicketID}}"
       style="background:#4f46e5;color:#fff;padding:10px 20px;border-radius:6px;text-decoration:none;display:inline-block">
      View ticket
    </a>
  </p>
  <p style="color:#94a3b8;font-size:12px">You are receiving this because you submitted this ticket.</p>
</body>
</html>`))

var resolvedText = texttemplate.Must(texttemplate.New("resolved_text").Parse(`Hi {{.ReporterName}},

Your ticket #{{.TicketID}} has been marked as {{.Status}}:

  {{.Subject}}

View ticket: {{.AppURL}}/portal/tickets/{{.TicketID}}
`))

// RenderAssigned returns HTML and plaintext email bodies for the "assigned" event.
func RenderAssigned(data TicketAssignedData) (html, text string, err error) {
	html, err = renderHTML(assignedHTML, data)
	if err != nil {
		return
	}
	text, err = renderText(assignedText, data)
	return
}

// TicketCommentData is passed to the comment notification email templates.
type TicketCommentData struct {
	RecipientName string
	TicketID      string
	Subject       string
	CommentBody   string
	AppURL        string
}

var commentHTML = template.Must(template.New("comment_html").Parse(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;color:#1e293b;max-width:600px;margin:0 auto;padding:24px">
  <h2 style="color:#4f46e5">New comment on ticket #{{.TicketID}}</h2>
  <p>Hi {{.RecipientName}},</p>
  <p>A new comment has been added to ticket <strong>#{{.TicketID}}: {{.Subject}}</strong>:</p>
  <blockquote style="border-left:3px solid #4f46e5;padding:8px 16px;margin:16px 0;color:#475569">
    {{.CommentBody}}
  </blockquote>
  <p>
    <a href="{{.AppURL}}/tickets/{{.TicketID}}"
       style="background:#4f46e5;color:#fff;padding:10px 20px;border-radius:6px;text-decoration:none;display:inline-block">
      View ticket
    </a>
  </p>
  <p style="color:#94a3b8;font-size:12px">You are receiving this because you are associated with this ticket.</p>
</body>
</html>`))

var commentText = texttemplate.Must(texttemplate.New("comment_text").Parse(`Hi {{.RecipientName}},

A new comment has been added to ticket #{{.TicketID}}: {{.Subject}}

  {{.CommentBody}}

View ticket: {{.AppURL}}/tickets/{{.TicketID}}
`))

// RenderComment returns HTML and plaintext email bodies for a new comment notification.
func RenderComment(data TicketCommentData) (html, text string, err error) {
	html, err = renderHTML(commentHTML, data)
	if err != nil {
		return
	}
	text, err = renderText(commentText, data)
	return
}

// RenderResolved returns HTML and plaintext email bodies for the resolved/closed event.
func RenderResolved(data TicketResolvedData) (html, text string, err error) {
	html, err = renderHTML(resolvedHTML, data)
	if err != nil {
		return
	}
	text, err = renderText(resolvedText, data)
	return
}

func renderHTML(t *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render email template %s: %w", t.Name(), err)
	}
	return buf.String(), nil
}

func renderText(t *texttemplate.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render email template %s: %w", t.Name(), err)
	}
	return buf.String(), nil
}
