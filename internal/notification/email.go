package notification

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/nikitaaldaev/bani/internal/logger"
)

// EmailSender sends email notifications.
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SMTPConfig holds SMTP connection settings.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
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
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s",
		s.cfg.From, to, subject, body)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	if err := smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg)); err != nil {
		s.logger.Error("failed to send email", "to", to, "error", err)
		return fmt.Errorf("send email: %w", err)
	}

	s.logger.Debug("email sent", "to", to, "subject", subject)
	return nil
}

// NoopEmailSender is a no-op implementation used when email is not configured.
type NoopEmailSender struct{}

func NewNoopEmailSender() EmailSender {
	return &NoopEmailSender{}
}

func (n *NoopEmailSender) Send(_ context.Context, _, _, _ string) error {
	return nil
}
