package mailer

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime/quotedprintable"
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
	msg, err := buildMessage(m.from, to, subject, body)
	if err != nil {
		m.logger.Debugf("mailer: buildMessage selhalo: %s", err)
		return
	}
	addr := m.host + ":" + m.port
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	if err := smtp.SendMail(addr, auth, m.from, []string{to}, msg); err != nil {
		m.logger.Debugf("mailer: odeslání selhalo (to=%s): %s", to, err)
	}
}

// buildMessage sestaví email jako multipart/alternative (plain text + HTML).
func buildMessage(from, to, subject, plainText string) ([]byte, error) {
	var buf bytes.Buffer

	// Náhodná MIME boundary
	b := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, fmt.Errorf("buildMessage: rand: %w", err)
	}
	boundary := hex.EncodeToString(b)

	// Hlavičky
	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	fmt.Fprintf(&buf, "Subject: %s\r\n", encodeSubject(subject))
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=%q\r\n", boundary)
	fmt.Fprintf(&buf, "\r\n")

	// -- Plain text část
	fmt.Fprintf(&buf, "--%s\r\n", boundary)
	fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: quoted-printable\r\n")
	fmt.Fprintf(&buf, "\r\n")
	qwText := quotedprintable.NewWriter(&buf)
	if _, err := qwText.Write([]byte(plainText)); err != nil {
		return nil, fmt.Errorf("buildMessage: qp text: %w", err)
	}
	qwText.Close()
	fmt.Fprintf(&buf, "\r\n")

	// -- HTML část
	htmlBytes, err := renderHTML(plainText)
	if err != nil {
		return nil, fmt.Errorf("buildMessage: renderHTML: %w", err)
	}
	fmt.Fprintf(&buf, "--%s\r\n", boundary)
	fmt.Fprintf(&buf, "Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: quoted-printable\r\n")
	fmt.Fprintf(&buf, "\r\n")
	qwHTML := quotedprintable.NewWriter(&buf)
	if _, err := qwHTML.Write(htmlBytes); err != nil {
		return nil, fmt.Errorf("buildMessage: qp html: %w", err)
	}
	qwHTML.Close()
	fmt.Fprintf(&buf, "\r\n")

	// Uzavírací boundary
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)

	return buf.Bytes(), nil
}

// encodeSubject zakóduje subject do RFC 2047 formátu pokud obsahuje non-ASCII znaky (čeština).
func encodeSubject(s string) string {
	for _, r := range s {
		if r > 127 {
			return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
		}
	}
	return s
}
