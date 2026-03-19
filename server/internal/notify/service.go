package notify

import (
	"context"
	"fmt"
	"time"
)

// Notifier sends notifications.
type Notifier interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Service routes notifications through available notifiers.
type Service struct {
	email   Notifier
	tokens  *ActionTokenStore
	baseURL string
}

// NewService creates a notification Service. Pass nil email for no-op mode.
func NewService(email Notifier, tokens *ActionTokenStore, baseURL string) *Service {
	return &Service{email: email, tokens: tokens, baseURL: baseURL}
}

// NotifyApprovalNeeded sends an approval email with one-click approve/reject links.
func (s *Service) NotifyApprovalNeeded(ctx context.Context, to, requestID, secretName, requester, reason string) error {
	if s.email == nil || to == "" {
		return nil
	}

	approveToken, err := s.tokens.Create(ctx, requestID, "approve", 72*time.Hour)
	if err != nil {
		return fmt.Errorf("creating approve token: %w", err)
	}
	rejectToken, err := s.tokens.Create(ctx, requestID, "reject", 72*time.Hour)
	if err != nil {
		return fmt.Errorf("creating reject token: %w", err)
	}

	approveURL := s.baseURL + "/api/v1/action-tokens/" + approveToken + "/redeem?action=approve"
	rejectURL := s.baseURL + "/api/v1/action-tokens/" + rejectToken + "/redeem?action=reject"

	subject := "Valt: Access Request Needs Your Approval"
	body := buildApprovalEmail(secretName, requester, reason, approveURL, rejectURL)
	return s.email.Send(ctx, to, subject, body)
}

// NotifyAccessGranted sends an access-granted notification.
func (s *Service) NotifyAccessGranted(ctx context.Context, to, secretName string, durationMin int) error {
	if s.email == nil {
		return nil
	}
	subject := "Valt: Access Granted"
	body := "Access to secret '" + secretName + "' has been granted.\nDuration: " + intToStr(durationMin) + " minutes."
	return s.email.Send(ctx, to, subject, body)
}

func intToStr(n int) string {
	if n <= 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
