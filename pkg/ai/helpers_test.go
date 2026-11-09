package ai

import (
	"errors"
	"testing"
)

func TestAiHelpers(t *testing.T) {
	// Test NewTextMessage
	userMsg := NewTextMessage(RoleUser, "Hello Hermes")
	if userMsg.Role != RoleUser || len(userMsg.Content) != 1 || userMsg.Content[0].Text != "Hello Hermes" {
		t.Errorf("unexpected text message: %+v", userMsg)
	}

	// Test NewToolCallMessage
	tcData := []ToolCallData{
		{
			ID:        "call_123",
			Name:      "read_file",
			Arguments: map[string]any{"path": "test.txt"},
		},
	}
	tcMsg := NewToolCallMessage(tcData)
	if tcMsg.Role != RoleAssistant || len(tcMsg.Content) != 1 || tcMsg.Content[0].ToolCallID != "call_123" {
		t.Errorf("unexpected tool call message: %+v", tcMsg)
	}

	// Test NewToolResultMessage success
	resMsg := NewToolResultMessage("call_123", "read_file", "file contents", false)
	if resMsg.Role != RoleTool || resMsg.ToolCallID != "call_123" || resMsg.Content[0].Text != "file contents" {
		t.Errorf("unexpected tool result message: %+v", resMsg)
	}

	// Test NewToolResultMessage error
	errMsg := NewToolResultMessage("call_123", "read_file", errors.New("file not found"), true)
	if errMsg.Role != RoleTool || errMsg.Content[0].Text != "Error: file not found" {
		t.Errorf("unexpected tool error message: %+v", errMsg)
	}

	// Test EstimateTokenCount
	messages := []Message{userMsg, resMsg}
	tokens := EstimateTokenCount(messages)
	if tokens <= 0 {
		t.Errorf("expected positive token count, got %d", tokens)
	}
}
