package email_test

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/email"
)

// ---- Template rendering tests ----

func TestRenderAssigned(t *testing.T) {
	html, text, err := email.RenderAssigned(email.TicketAssignedData{
		AssigneeName: "Alice Smith",
		TicketID:     "abc-123",
		Subject:      "Login broken",
		AppURL:       "https://app.example.com",
	})
	require.NoError(t, err)
	assert.Contains(t, html, "Alice Smith")
	assert.Contains(t, html, "abc-123")
	assert.Contains(t, html, "Login broken")
	assert.Contains(t, html, "https://app.example.com/tickets/abc-123")
	assert.Contains(t, text, "Alice Smith")
	assert.Contains(t, text, "abc-123")
}

func TestRenderResolved(t *testing.T) {
	html, text, err := email.RenderResolved(email.TicketResolvedData{
		ReporterName: "Bob Jones",
		TicketID:     "xyz-456",
		Subject:      "Cannot export report",
		Status:       "resolved",
		AppURL:       "https://app.example.com",
	})
	require.NoError(t, err)
	assert.Contains(t, html, "Bob Jones")
	assert.Contains(t, html, "resolved")
	assert.Contains(t, text, "xyz-456")
}

func TestRenderComment(t *testing.T) {
	html, text, err := email.RenderComment(email.TicketCommentData{
		RecipientName: "Carol Davis",
		TicketID:      "cmt-789",
		Subject:       "API not responding",
		CommentBody:   "We are investigating the issue.",
		AppURL:        "https://app.example.com",
	})
	require.NoError(t, err)
	assert.Contains(t, html, "Carol Davis")
	assert.Contains(t, html, "cmt-789")
	assert.Contains(t, html, "API not responding")
	assert.Contains(t, html, "We are investigating the issue.")
	assert.Contains(t, html, "https://app.example.com/tickets/cmt-789")
	assert.Contains(t, text, "Carol Davis")
	assert.Contains(t, text, "cmt-789")
}

// ---- SMTP sender with mock server ----

// mockSMTPServer starts a minimal SMTP server that records received DATA payloads.
func mockSMTPServer(t *testing.T) (addr string, received func() []string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	var mu sync.Mutex
	var msgs []string

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(5 * time.Second))
				rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
				fmt.Fprintf(rw, "220 mock SMTP ready\r\n")
				rw.Flush()

				var body strings.Builder
				inData := false
				for {
					line, err := rw.ReadString('\n')
					if err != nil {
						return
					}
					line = strings.TrimRight(line, "\r\n")

					if inData {
						if line == "." {
							mu.Lock()
							msgs = append(msgs, body.String())
							mu.Unlock()
							body.Reset()
							inData = false
							fmt.Fprintf(rw, "250 OK\r\n")
							rw.Flush()
						} else {
							body.WriteString(line + "\n")
						}
						continue
					}

					upper := strings.ToUpper(line)
					switch {
					case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
						fmt.Fprintf(rw, "250 mock\r\n")
					case strings.HasPrefix(upper, "MAIL FROM"):
						fmt.Fprintf(rw, "250 OK\r\n")
					case strings.HasPrefix(upper, "RCPT TO"):
						fmt.Fprintf(rw, "250 OK\r\n")
					case upper == "DATA":
						fmt.Fprintf(rw, "354 Start input\r\n")
						inData = true
					case upper == "QUIT":
						fmt.Fprintf(rw, "221 Bye\r\n")
						rw.Flush()
						return
					default:
						fmt.Fprintf(rw, "500 Unknown\r\n")
					}
					rw.Flush()
				}
			}(conn)
		}
	}()

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	return ln.Addr().String(), func() []string {
		_ = host
		_ = port
		mu.Lock()
		defer mu.Unlock()
		out := make([]string, len(msgs))
		copy(out, msgs)
		return out
	}
}

func TestSender_Send(t *testing.T) {
	addr, received := mockSMTPServer(t)
	host, port, _ := net.SplitHostPort(addr)

	cfg := config.SMTPConfig{
		Enabled: true,
		Host:    host,
		Port:    port,
		From:    "test@omnir.io",
	}
	s := email.NewSender(cfg)
	err := s.Send(email.Message{
		To:      "user@example.com",
		Subject: "Test email",
		HTML:    "<p>Hello</p>",
		Text:    "Hello",
	})
	require.NoError(t, err)

	// Give a tiny window for the goroutine to process.
	time.Sleep(50 * time.Millisecond)
	msgs := received()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0], "Test email")
}

func TestSender_Disabled(t *testing.T) {
	cfg := config.SMTPConfig{Enabled: false}
	s := email.NewSender(cfg)
	err := s.Send(email.Message{To: "x@y.com", Subject: "noop", HTML: "h", Text: "t"})
	require.NoError(t, err) // must not try to connect
}

// ---- Mailer tests ----

func TestMailer_SendAssigned(t *testing.T) {
	addr, received := mockSMTPServer(t)
	host, port, _ := net.SplitHostPort(addr)
	cfg := config.SMTPConfig{Enabled: true, Host: host, Port: port, From: "noreply@omnir.io"}
	m := email.NewMailer(email.NewSender(cfg), "https://app.omnir.io")

	err := m.SendAssigned("agent@example.com", "Agent One", "ticket-uuid-001", "Cannot login")
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	msgs := received()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0], "ticket-uuid-001")
}

func TestMailer_SendResolved(t *testing.T) {
	addr, received := mockSMTPServer(t)
	host, port, _ := net.SplitHostPort(addr)
	cfg := config.SMTPConfig{Enabled: true, Host: host, Port: port, From: "noreply@omnir.io"}
	m := email.NewMailer(email.NewSender(cfg), "https://app.omnir.io")

	err := m.SendResolved("client@example.com", "Client User", "ticket-uuid-002", "Bug in export", "resolved")
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	msgs := received()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0], "resolved")
}

func TestMailer_SendComment(t *testing.T) {
	addr, received := mockSMTPServer(t)
	host, port, _ := net.SplitHostPort(addr)
	cfg := config.SMTPConfig{Enabled: true, Host: host, Port: port, From: "noreply@omnir.io"}
	m := email.NewMailer(email.NewSender(cfg), "https://app.omnir.io")

	err := m.SendComment("user@example.com", "Alice", "ticket-uuid-003", "Slow dashboard", "We are working on it.")
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	msgs := received()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0], "ticket-uuid-003")
}
