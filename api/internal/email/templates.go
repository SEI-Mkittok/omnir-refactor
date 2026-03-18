package email

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
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

var assignedText = `Hi {{.AssigneeName}},

Ticket #{{.TicketID}} has been assigned to you:

  {{.Subject}}

View ticket: {{.AppURL}}/tickets/{{.TicketID}}
`

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

var resolvedText = `Hi {{.ReporterName}},

Your ticket #{{.TicketID}} has been marked as {{.Status}}:

  {{.Subject}}

View ticket: {{.AppURL}}/portal/tickets/{{.TicketID}}
`

// RenderAssigned returns HTML and plaintext email bodies for the "assigned" event.
func RenderAssigned(data TicketAssignedData) (html, text string, err error) {
	html, err = renderHTML(assignedHTML, data)
	if err != nil {
		return
	}
	text = renderText(assignedText, map[string]string{
		"AssigneeName": data.AssigneeName,
		"TicketID":     data.TicketID,
		"Subject":      data.Subject,
		"AppURL":       data.AppURL,
	})
	return
}

// RenderResolved returns HTML and plaintext email bodies for the resolved/closed event.
func RenderResolved(data TicketResolvedData) (html, text string, err error) {
	html, err = renderHTML(resolvedHTML, data)
	if err != nil {
		return
	}
	text = renderText(resolvedText, map[string]string{
		"ReporterName": data.ReporterName,
		"TicketID":     data.TicketID,
		"Subject":      data.Subject,
		"Status":       data.Status,
		"AppURL":       data.AppURL,
	})
	return
}

func renderHTML(t *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render email template %s: %w", t.Name(), err)
	}
	return buf.String(), nil
}

func renderText(tmpl string, vars map[string]string) string {
	s := tmpl
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{."+k+"}}", v)
	}
	return s
}
