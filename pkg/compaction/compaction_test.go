package compaction

import (
	"strings"
	"testing"

	"github.com/robertpelloni/hermes-agent/pkg/ai"
)

func TestCompactor_Lifecycle(t *testing.T) {
	comp := New(100)
	if comp.ThresholdTokens != 100 {
		t.Errorf("expected threshold 100, got %d", comp.ThresholdTokens)
	}

	messages := []ai.Message{
		ai.NewTextMessage("system", "You are Hermes."),
		ai.NewTextMessage(ai.RoleUser, "Message 1"),
		ai.NewTextMessage(ai.RoleAssistant, "Response 1"),
		ai.NewTextMessage(ai.RoleUser, "Message 2"),
		ai.NewTextMessage(ai.RoleAssistant, "Response 2"),
		ai.NewTextMessage(ai.RoleUser, "Message 3"),
		ai.NewTextMessage(ai.RoleAssistant, "Response 3"),
		ai.NewTextMessage(ai.RoleUser, "Message 4"),
		ai.NewTextMessage(ai.RoleAssistant, "Response 4"),
		ai.NewTextMessage(ai.RoleUser, "Message 5"),
		ai.NewTextMessage(ai.RoleAssistant, "Response 5"),
	}

	tokenCount := comp.TokenCount(messages)
	if tokenCount <= 0 {
		t.Errorf("expected positive token count, got %d", tokenCount)
	}

	// Compact messages
	compacted := comp.Compact(messages)
	if len(compacted) >= len(messages) {
		t.Errorf("expected compacted messages to be fewer than original (%d vs %d)", len(compacted), len(messages))
	}

	// First message should remain system prompt
	if compacted[0].Role != "system" || compacted[0].Content[0].Text != "You are Hermes." {
		t.Errorf("expected system prompt preserved, got: %+v", compacted[0])
	}

	// Middle message should be compressed history
	if !strings.Contains(compacted[1].Content[0].Text, "Compressed history") {
		t.Errorf("expected compressed history in compacted[1], got: %s", compacted[1].Content[0].Text)
	}
}
