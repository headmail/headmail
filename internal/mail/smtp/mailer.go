// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package smtp

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/mailer"
)

// Mailer sends mail using an SMTP server.
type Mailer struct {
	cfg config.SMTPConfig
}

// NewMailer constructs an Mailer with provided config.
func NewMailer(cfg config.SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

// Send implements Mailer.Send using net/smtp.
func (m *Mailer) Send(ctx context.Context, d *domain.Delivery) error {
	// Build email message
	fromHeader := fmt.Sprintf("%s <%s>", m.cfg.From.Name, m.cfg.From.Email)
	toHeader := d.Email
	subject := d.Subject

	var body string
	var contentType string
	if d.BodyText != "" && d.BodyHTML != "" {
		// send both text and html using multipart/alternative
		boundary := fmt.Sprintf("BOUNDARY_%s", d.ID)
		contentType = "multipart/alternative; boundary=" + boundary

		var b strings.Builder
		// first: plain text part
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(d.BodyText)
		b.WriteString("\r\n")
		// second: html part
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(d.BodyHTML)
		b.WriteString("\r\n")
		// end boundary
		b.WriteString("--" + boundary + "--\r\n")

		body = b.String()
	} else if d.BodyHTML != "" {
		contentType = "text/html; charset=\"utf-8\""
		body = d.BodyHTML
	} else {
		contentType = "text/plain; charset=\"utf-8\""
		body = d.BodyText
	}

	headers := make([]string, 0, 8)
	headers = append(headers, "From: "+fromHeader)
	headers = append(headers, "To: "+toHeader)
	headers = append(headers, "Subject: "+subject)
	headers = append(headers, mailer.HeadmailDeliveryHeaderName+": "+d.ID)
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
