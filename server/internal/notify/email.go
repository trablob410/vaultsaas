package notify

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
)

// EmailSender sends email via SMTP.
type EmailSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

// NewEmailSender creates an EmailSender. Returns nil if host is empty (no-op mode).
func NewEmailSender(host string, port int, user, password, from string) *EmailSender {
	if host == "" {
		return nil
	}
	return &EmailSender{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}
}

// Send sends a plaintext email.
func (e *EmailSender) Send(_ context.Context, to, subject, body string) error {
	addr := net.JoinHostPort(e.host, strconv.Itoa(e.port))

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		e.from, to, subject, body)

	var auth smtp.Auth
	if e.user != "" {
		auth = smtp.PlainAuth("", e.user, e.password, e.host)
	}

	if err := smtp.SendMail(addr, auth, e.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	return nil
}
