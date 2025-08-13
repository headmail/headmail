package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/domain"
)

// SMTPMailer sends mail using an SMTP server.
type SMTPMailer struct {
	cfg config.SMTPConfig
}

// NewSMTPMailer constructs an SMTPMailer with provided config.
func NewSMTPMailer(cfg config.SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

// Send implements Mailer.Send using net/smtp.
func (m *SMTPMailer) Send(ctx context.Context, d *domain.Delivery) error {
	// Build email message
	fromHeader := fmt.Sprintf("%s <%s>", m.cfg.From.Name, m.cfg.From.Email)
	toHeader := d.Email
	subject := d.Subject

	var body string
	var contentType string
	if d.BodyText != "" && d.BodyHTML != "" {
		// prefer HTML
		contentType = "text/html; charset=\"utf-8\""
		body = d.BodyHTML
	} else if d.BodyHTML != "" {
		contentType = "text/html; charset=\"utf-8\""
		body = d.BodyHTML
	} else {
		contentType = "text/plain; charset=\"utf-8\""
		body = d.BodyText
	}

	headers := make([]string, 0, 6)
	headers = append(headers, "From: "+fromHeader)
	headers = append(headers, "To: "+toHeader)
	headers = append(headers, "Subject: "+subject)
	headers = append(headers, "MIME-Version: 1.0")
	headers = append(headers, "Content-Type: "+contentType)

	msg := strings.Join(headers, "\r\n") + "\r\n\r\n" + body

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	var auth smtp.Auth
	if m.cfg.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}

	// SendMail is blocking; caller is expected to run in a worker goroutine.
	if err := smtp.SendMail(addr, auth, m.cfg.From.Email, []string{d.Email}, []byte(msg)); err != nil {
		return err
	}
	return nil
}
