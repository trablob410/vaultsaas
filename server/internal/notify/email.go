package notify

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
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

func buildApprovalEmail(secretName, requester, reason, approveURL, rejectURL string) string {
	return fmt.Sprintf(`Secret access request requires your approval.

Secret:    %s
Requester: %s
Reason:    %s

Approve: %s
Reject:  %s

This link expires in 72 hours and can only be used once.
`, secretName, requester, reason, approveURL, rejectURL)
}

// sanitizeHeader strips CRLF sequences from email header values to prevent header injection.
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// Send sends a plaintext email.
func (e *EmailSender) Send(_ context.Context, to, subject, body string) error {
	addr := net.JoinHostPort(e.host, strconv.Itoa(e.port))

	// Sanitize header values to prevent CRLF injection (BCC exfiltration, header spoofing)
	to = sanitizeHeader(to)
	subject = sanitizeHeader(subject)

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
