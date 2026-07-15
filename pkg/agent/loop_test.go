package agent

import (
	"context"
	"strings"
	"testing"
	"github.com/robertpelloni/hermes-agent/pkg/memory"
)

func TestRunConversationLoopEcho(t *testing.T) {
	agent := &Agent{}
	store := memory.NewStore()

	ctx := context.Background()
	response, err := agent.RunConversationLoop(ctx, "cli", "user123", "hello hermes", store)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(response, "hello hermes") {
		t.Fatalf("Expected response to echo the input, got: %s", response)
	}
}

func TestAPIDispatcher(t *testing.T) {
	agent := &Agent{}
	store := memory.NewStore()

	ctx := context.Background()
	args := map[string]string{
		"platform": "dashboard",
		"userID":   "admin",
		"text":     "dashboard test",
	}

	response, err := agent.APIDispatcher(ctx, "converse", args, store)
	if err != nil {
		t.Fatalf("Expected no error for converse action, got %v", err)
	}
	if !strings.Contains(response, "dashboard test") {
		t.Fatalf("Expected response to contain 'dashboard test', got: %s", response)
	}

	_, err = agent.APIDispatcher(ctx, "unknown_action", args, store)
	if err == nil {
		t.Fatalf("Expected error for unknown action, got nil")
	}
}
