package mailer

import (
	"context"
	"fmt"
	"net/smtp"

	"go.uber.org/zap"
)

type Config struct {
	Host, Port, User, Password, From string
}

type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

type smtpMailer struct {
	cfg Config
	log *zap.Logger
}

type logMailer struct{ log *zap.Logger }

func New(cfg Config, log *zap.Logger) Mailer {
	if cfg.Host == "" {
		log.Warn("SMTP_HOST not configured — emails will be logged instead of sent")
		return &logMailer{log: log}
	}
	return &smtpMailer{cfg: cfg, log: log}
}

func (m *smtpMailer) Send(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Password, m.cfg.Host)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		m.cfg.From, to, subject, body)
	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg)); err != nil {
		m.log.Error("smtp send failed", zap.String("to", to), zap.Error(err))
		return err
	}
	return nil
}

func (m *logMailer) Send(ctx context.Context, to, subject, body string) error {
	m.log.Info("email (dev mode, not sent)", zap.String("to", to), zap.String("subject", subject), zap.String("body", body))
	return nil
}
