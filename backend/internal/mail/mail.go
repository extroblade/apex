// Package mail is a thin, fail-open SMTP client for transactional emails
// (password reset, email verification). It uses net/smtp with STARTTLS, which
// works with every common provider on port 587 (Gmail, SES, Mailgun, Postmark,
// local MailHog/Mailpit). When SMTP_HOST is empty the mailer is "disabled": Send
// is a no-op and Enabled() returns false, so the app keeps working without an
// SMTP server — the auth flows just won't actually deliver mail.
package mail

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// Config holds the SMTP connection settings.
type Config struct {
	Host string // SMTP server hostname (empty disables the mailer)
	Port string // SMTP server port (default "587")
	User string // SMTP username (empty = no auth, for local MailHog/Mailpit)
	Pass string // SMTP password
	From string // From: address (e.g. "ContentPilot <no-reply@contentpilot.app>")
}

// Mailer sends transactional email. Construct with New; a nil *Mailer or a
// disabled Mailer both no-op, so callers don't need to branch on config.
type Mailer struct {
	cfg     Config
	enabled bool
}

// New returns a Mailer backed by cfg. An empty Host yields a disabled mailer.
func New(cfg Config) *Mailer {
	return &Mailer{cfg: cfg, enabled: cfg.Host != ""}
}

// Enabled reports whether SMTP is configured. Send is a no-op when false.
func (m *Mailer) Enabled() bool { return m != nil && m.enabled }

// ErrDisabled is returned by Send when the mailer is not configured AND the
// caller asked for an error. Send itself is a no-op (returns nil) so callers
// don't have to branch — this is only for tests that want to assert "would
// have sent".
var ErrDisabled = errors.New("mail: SMTP not configured")

// Send delivers a plain-text message to `to`. The message body is `body`; the
// Subject header is set from `subject`. Returns nil when disabled (no-op) so
// callers can always call Send without checking Enabled first.
func (m *Mailer) Send(ctx context.Context, to, subject, body string) error {
	if !m.Enabled() {
		return nil
	}
	port := m.cfg.Port
	if port == "" {
		port = "587"
	}
	addr := m.cfg.Host + ":" + port

	from := m.cfg.From
	if from == "" {
		from = "no-reply@contentpilot.app"
	}
	// Pull the bare address out of "Name <addr>" for the envelope (SMTP MAIL FROM).
	fromAddr := stripAngle(from)

	msg := buildMessage(from, to, subject, body)

	// net/smtp doesn't take a context, so we cap the dial+send with a timeout
	// goroutine. 10s is plenty for a single recipient; if the server is slow we
	// fail rather than hanging the request.
	done := make(chan error, 1)
	go func() {
		var auth smtp.Auth
		if m.cfg.User != "" {
			auth = smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.Host)
		}
		done <- smtp.SendMail(addr, auth, fromAddr, []string{to}, msg)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		return errors.New("mail: send timed out after 10s")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// buildMessage assembles the RFC 5322 message: headers + blank line + body.
func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: text/plain; charset=utf-8\r\n")
	fmt.Fprintf(&b, "Content-Transfer-Encoding: 8bit\r\n")
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Auto-Submitted: auto-generated\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

// stripAngle pulls the bare address out of a "Name <addr>" From header, for the
// SMTP envelope (MAIL FROM:<addr>). If there's no <> it returns the input as-is.
func stripAngle(s string) string {
	i := strings.IndexByte(s, '<')
	if i < 0 {
		return s
	}
	j := strings.IndexByte(s, '>')
	if j < 0 || j < i {
		return s
	}
	return s[i+1 : j]
}
