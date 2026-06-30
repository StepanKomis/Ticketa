package mailer

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime/quotedprintable"
	"net/smtp"
	"strings"
	"sync"

	"github.com/StepanKomis/Ticketa/src/cmd/server/env"
	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

type Mailer struct {
	mu       sync.RWMutex
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

// NewFromDB načte SMTP nastavení z databáze. Vrátí nil pokud smtp_host není nastaven.
func NewFromDB(ctx context.Context, queries *db.Queries, logger *logs.Logger) *Mailer {
	settings, err := queries.GetAllServerSettings(ctx)
	if err != nil {
		return nil
	}
	m := map[string]string{}
	for _, s := range settings {
		m[s.Key] = s.Value
	}
	host := m["smtp_host"]
	if host == "" {
		return nil
	}
	port := m["smtp_port"]
	if port == "" {
		port = "587"
	}
	return &Mailer{
		host:     host,
		port:     port,
		username: m["smtp_user"],
		password: m["smtp_password"],
		from:     m["smtp_from"],
		logger:   logger,
	}
}

// Reload přenastaví SMTP připojení za běhu bez restartu serveru.
func (m *Mailer) Reload(host, port, username, password, from string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.host = host
	m.port = port
	m.username = username
	m.password = password
	m.from = from
}

// TestCredentials otestuje SMTP připojení s explicitně zadanými údaji.
// Vrací kategorickou chybovou zprávu, ne jen generické selhání.
func TestCredentials(host, port, username, password string) error {
	addr := host + ":" + port
	c, err := smtp.Dial(addr)
	if err != nil {
		return categorizeSMTPDialError(host, port, err)
	}
	defer c.Quit() //nolint:errcheck

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("Chyba TLS/SSL - server nepodporuje šifrované připojení na portu %s: %w", port, err)
		}
	}

	if username != "" || password != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("Přihlášení selhalo - zkontrolujte uživatelské jméno a heslo")
		}
	}
	return nil
}

func categorizeSMTPDialError(host, port string, err error) error {
	msg := err.Error()
	if strings.Contains(msg, "no such host") || strings.Contains(msg, "lookup") {
		return fmt.Errorf("Hostname '%s' nelze přeložit (DNS) - zkontrolujte adresu serveru", host)
	}
	if strings.Contains(msg, "connection refused") {
		return fmt.Errorf("Připojení odmítnuto na portu %s - ověřte port (587 pro STARTTLS, 465 pro SSL, 25 pro nešifrované)", port)
	}
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "i/o timeout") {
		return fmt.Errorf("Server '%s:%s' není dostupný nebo je port blokován firewallem", host, port)
	}
	return fmt.Errorf("Nepodařilo se připojit k '%s:%s': %s", host, port, msg)
}

// Ping ověří dostupnost SMTP serveru a správnost přihlašovacích údajů bez odeslání zprávy.
func (m *Mailer) Ping() error {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	host, port, username, password := m.host, m.port, m.username, m.password
	m.mu.RUnlock()
	return TestCredentials(host, port, username, password)
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
	m.mu.RLock()
	host, port, username, password, from := m.host, m.port, m.username, m.password, m.from
	m.mu.RUnlock()

	msg, err := buildMessage(from, to, subject, body)
	if err != nil {
		m.logger.Debugf("mailer: buildMessage selhalo: %s", err)
		return
	}
	addr := host + ":" + port
	auth := smtp.PlainAuth("", username, password, host)
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
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
