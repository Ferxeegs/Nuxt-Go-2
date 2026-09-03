// Package mailer sends plain-text notification emails via SMTP (e.g.
// forgot-password links). It is optional: if SMTP host is not configured,
// Send is a no-op so the boilerplate works without email infra out of the box.
package mailer

import (
	"fmt"
	"net/smtp"

	"go-cp-boilerplate/internal/config"
)

type Mailer struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) Enabled() bool {
	return m.cfg.SMTPHost != ""
}

// Send delivers a plain-text email. Callers should invoke this in a
// goroutine so a slow/unavailable SMTP server never blocks the request.
func (m *Mailer) Send(to, subject, body string) error {
	if !m.Enabled() {
		return nil
	}

	addr := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, m.cfg.SMTPPort)
	from := m.cfg.SMTPFrom
	if from == "" {
		from = m.cfg.SMTPUser
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body)

	var auth smtp.Auth
	if m.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPassword, m.cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
