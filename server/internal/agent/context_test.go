package agent

import (
	"context"
	"testing"
)

func TestWithAgentID_Roundtrip(t *testing.T) {
	ctx := context.Background()
	ctx = WithAgentID(ctx, "agent-abc")
	if got := AgentIDFromContext(ctx); got != "agent-abc" {
		t.Errorf("expected agent-abc, got %s", got)
	}
}

func TestAgentIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	if got := AgentIDFromContext(ctx); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestWithAgentID_Overwrite(t *testing.T) {
	ctx := context.Background()
	ctx = WithAgentID(ctx, "agent-first")
	ctx = WithAgentID(ctx, "agent-second")
	if got := AgentIDFromContext(ctx); got != "agent-second" {
		t.Errorf("expected agent-second, got %s", got)
	}
}

func TestWithAgentID_EmptyString(t *testing.T) {
	ctx := context.Background()
	ctx = WithAgentID(ctx, "")
	if got := AgentIDFromContext(ctx); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}
