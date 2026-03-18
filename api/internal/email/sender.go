// Package email provides SMTP-based email delivery for transactional notifications.
package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/omnir/crm-api/internal/config"
)

// Sender delivers email via SMTP.
type Sender struct {
	cfg config.SMTPConfig
}

// NewSender constructs a Sender from config.
func NewSender(cfg config.SMTPConfig) *Sender {
	return &Sender{cfg: cfg}
}

// Message is a fully rendered email ready to send.
type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

// Send delivers msg via SMTP. It is a no-op when SMTP is disabled.
func (s *Sender) Send(msg Message) error {
	if !s.cfg.Enabled {
		return nil
	}

	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	body := buildMIME(s.cfg.From, msg)

	// Try STARTTLS first; fall back to plain TCP for local/test servers.
	tlsCfg := &tls.Config{ServerName: s.cfg.Host} //nolint:gosec
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if auth != nil {
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := c.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := c.Rcpt(msg.To); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := fmt.Fprint(wc, body); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return c.Quit()
}

// buildMIME assembles a multipart/alternative MIME message.
func buildMIME(from string, msg Message) string {
	boundary := "omnir_email_boundary_42"
	var b strings.Builder

	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", msg.To))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	b.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q\r\n", boundary))
	b.WriteString("\r\n")

	// Plaintext part
	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Text)
	b.WriteString("\r\n")

	// HTML part
	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.HTML)
	b.WriteString("\r\n")

	b.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return b.String()
}
