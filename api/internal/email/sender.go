// Package email provides SMTP-based email delivery for transactional notifications.
package email

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"net/textproto"

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

// buildMIME assembles a multipart/alternative MIME message with proper
// quoted-printable encoding and a cryptographically random boundary.
func buildMIME(from string, msg Message) string {
	// Random boundary satisfying RFC 2046.
	randBytes := make([]byte, 12)
	_, _ = rand.Read(randBytes)
	boundary := "omnir_" + hex.EncodeToString(randBytes)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.SetBoundary(boundary)

	// Plaintext part — QP-encoded.
	ph := make(textproto.MIMEHeader)
	ph.Set("Content-Type", "text/plain; charset=UTF-8")
	ph.Set("Content-Transfer-Encoding", "quoted-printable")
	pw, _ := mw.CreatePart(ph)
	qpw := quotedprintable.NewWriter(pw)
	_, _ = qpw.Write([]byte(msg.Text))
	_ = qpw.Close()

	// HTML part — QP-encoded.
	hh := make(textproto.MIMEHeader)
	hh.Set("Content-Type", "text/html; charset=UTF-8")
	hh.Set("Content-Transfer-Encoding", "quoted-printable")
	hw, _ := mw.CreatePart(hh)
	qpwh := quotedprintable.NewWriter(hw)
	_, _ = qpwh.Write([]byte(msg.HTML))
	_ = qpwh.Close()

	_ = mw.Close()

	var out bytes.Buffer
	fmt.Fprintf(&out, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&out, "From: %s\r\n", from)
	fmt.Fprintf(&out, "To: %s\r\n", msg.To)
	fmt.Fprintf(&out, "Subject: %s\r\n", msg.Subject)
	fmt.Fprintf(&out, "Content-Type: multipart/alternative; boundary=%q\r\n", boundary)
	fmt.Fprintf(&out, "\r\n")
	out.Write(body.Bytes())

	return out.String()
}
