package auth

import (
	"context"
	"testing"
)

func TestWithUserID_Roundtrip(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user-123")
	if got := UserIDFromContext(ctx); got != "user-123" {
		t.Errorf("expected user-123, got %s", got)
	}
}

func TestWithUserID_Overwrite(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user-first")
	ctx = WithUserID(ctx, "user-second")
	if got := UserIDFromContext(ctx); got != "user-second" {
		t.Errorf("expected user-second, got %s", got)
	}
}

func TestWithUserID_EmptyString(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "")
	if got := UserIDFromContext(ctx); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}
