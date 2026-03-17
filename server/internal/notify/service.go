package notify

import "context"

// Notifier sends notifications.
type Notifier interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Service routes notifications through available notifiers.
type Service struct {
	email Notifier
}

// NewService creates a notification Service. Pass nil for no-op mode.
func NewService(email Notifier) *Service {
	return &Service{email: email}
}

// NotifyApprovalNeeded sends an approval-needed notification.
func (s *Service) NotifyApprovalNeeded(ctx context.Context, to, secretName, requester, reason string) error {
	if s.email == nil {
		return nil
	}
	subject := "Valt: Access Request Needs Approval"
	body := "Secret: " + secretName + "\nRequester: " + requester + "\nReason: " + reason + "\n\nPlease log in to approve or reject this request."
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
