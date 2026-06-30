package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/StepanKomis/Ticketa/src/cmd/server/env"
	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
)

type Mailer struct {
	host     string
	port     string
	username string
	password string
	from     string
	logger   *logs.Logger
}

// New vytvoří Mailer z env proměnných. Pokud SMTP_HOST není nastaven, vrátí nil —
// volání Send na nil Maileru je no-op.
func New(logger *logs.Logger) *Mailer {
	host := env.Get("SMTP_HOST", "")
	if host == "" {
		return nil
	}
	return &Mailer{
		host:     host,
		port:     env.Get("SMTP_PORT", "587"),
		username: env.Get("SMTP_USER", ""),
		password: env.Get("SMTP_PASSWORD", ""),
		from:     env.Get("SMTP_FROM", ""),
		logger:   logger,
	}
}

// Ping ověří dostupnost SMTP serveru a správnost přihlašovacích údajů bez odeslání zprávy.
func (m *Mailer) Ping() error {
	if m == nil {
		return nil
	}
	addr := m.host + ":" + m.port
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("nelze se připojit k SMTP serveru %s: %w", addr, err)
	}
	defer c.Quit() //nolint:errcheck

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
			return fmt.Errorf("STARTTLS selhalo: %w", err)
		}
	}

	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("autentizace selhala: %w", err)
	}

	return nil
}

// Ping vytvoří dočasný Mailer z env proměnných a otestuje SMTP připojení.
// Pokud SMTP_HOST není nastaven, vrátí nil (no-op).
func Ping(logger *logs.Logger) error {
	return New(logger).Ping()
}

func (m *Mailer) Send(to, subject, body string) {
	if m == nil {
		return
	}
	addr := m.host + ":" + m.port
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.from, to, subject, body))
	if err := smtp.SendMail(addr, auth, m.from, []string{to}, msg); err != nil {
		m.logger.Debugf("mailer: odeslání selhalo (to=%s): %s", to, err)
	}
}
