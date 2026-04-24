package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/rekurt/relax-hub/internal/logger"
)

// EmailSender sends email notifications.
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SMTPConfig holds SMTP connection settings.
//
// Port 587 triggers STARTTLS (explicit TLS upgrade), port 465 triggers implicit
// TLS on connect, port 25 is plain. Both 465 and 587 authenticate via PLAIN.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	// From may be a bare address ("partners@relax-hub.ru") or include a display
	// name ("RelaxHUB <partners@relax-hub.ru>"). Display name is preserved in
	// the message header but the envelope uses the bare address.
	From string
}

type smtpEmailSender struct {
	cfg    SMTPConfig
	logger *logger.Logger
}

// NewSMTPEmailSender creates an EmailSender that sends emails via SMTP.
func NewSMTPEmailSender(cfg SMTPConfig, log *logger.Logger) EmailSender {
	return &smtpEmailSender{cfg: cfg, logger: log}
}

func (s *smtpEmailSender) Send(_ context.Context, to, subject, body string) error {
	fromHeader, fromEnvelope, err := parseFrom(s.cfg.From)
	if err != nil {
		return fmt.Errorf("invalid From address %q: %w", s.cfg.From, err)
	}

	// RFC 2047 encode subject so non-ASCII (Russian) renders correctly in
	// MUAs. The plain-text body itself is UTF-8 so we only protect the header.
	encodedSubject := mime.QEncoding.Encode("utf-8", subject)

	msg := strings.Join([]string{
		"From: " + fromHeader,
		"To: " + to,
		"Subject: " + encodedSubject,
		"MIME-Version: 1.0",
		`Content-Type: text/plain; charset="utf-8"`,
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	if err := s.deliver(fromEnvelope, to, auth, []byte(msg)); err != nil {
		s.logger.Error("failed to send email", "to", to, "subject", subject, "host", s.cfg.Host, "port", s.cfg.Port, "error", err)
		return fmt.Errorf("send email: %w", err)
	}

	s.logger.Info("email sent", "to", to, "subject", subject)
	return nil
}

// deliver picks the right transport (implicit TLS vs STARTTLS vs plain) based
// on the configured port.
func (s *smtpEmailSender) deliver(fromEnvelope, to string, auth smtp.Auth, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	// Implicit TLS (SMTPS) — standard for port 465. stdlib smtp.SendMail does
	// not handle this, so we wrap the dial in tls.Dial.
	if s.cfg.Port == 465 {
		return sendImplicitTLS(addr, s.cfg.Host, fromEnvelope, []string{to}, auth, msg)
	}

	// STARTTLS on port 587 (and fallback for 25 if server advertises it).
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = c.Quit() }()

	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("auth: %w", err)
			}
		}
	}

	if err := c.Mail(fromEnvelope); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return nil
}

// sendImplicitTLS replicates smtp.SendMail over a TLS-wrapped conn for port 465.
func sendImplicitTLS(addr, host, from string, to []string, auth smtp.Auth, msg []byte) error {
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	rawConn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return fmt.Errorf("dial tls %s: %w", addr, err)
	}
	c, err := smtp.NewClient(rawConn, host)
	if err != nil {
		_ = rawConn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = c.Quit() }()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("auth: %w", err)
			}
		}
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("rcpt %s: %w", rcpt, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return nil
}

// parseFrom accepts either a bare address ("foo@bar") or a mailbox expression
// ("Display Name <foo@bar>") and returns a header-ready form plus the bare
// address to use for the SMTP MAIL FROM envelope.
func parseFrom(raw string) (header string, envelope string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("empty From")
	}
	addr, err := mail.ParseAddress(raw)
	if err != nil {
		return "", "", err
	}
	return addr.String(), addr.Address, nil
}
