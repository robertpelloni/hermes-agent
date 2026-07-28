package ai

import "fmt"

// NewTextMessage creates a simple text message.
func NewTextMessage(role MessageRole, text string) Message {
	return Message{
		Role:    role,
		Content: []Content{{Type: ContentTypeText, Text: text}},
	}
}

// NewToolCallMessage creates an assistant message with tool calls.
func NewToolCallMessage(toolCalls []ToolCallData) Message {
	content := make([]Content, 0, len(toolCalls))
	for _, tc := range toolCalls {
		content = append(content, Content{
			Type:       ContentTypeToolCall,
			ToolCallID: tc.ID,
			Name:       tc.Name,
			Arguments:  tc.Arguments,
		})
	}
	return Message{
		Role:    RoleAssistant,
		Content: content,
	}
}

// NewToolResultMessage creates a tool result message.
func NewToolResultMessage(toolCallID, toolName string, result any, isError bool) Message {
	var text string
	if isError {
		if err, ok := result.(error); ok {
			text = "Error: " + err.Error()
		} else {
			text = "Error: " + fmt.Sprintf("%v", result)
		}
	} else {
		switch v := result.(type) {
		case string:
			text = v
		case []byte:
			text = string(v)
		case fmt.Stringer:
			text = v.String()
		default:
			text = fmt.Sprintf("%+v", v)
		}
	}
	return Message{
		Role:       RoleTool,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Content:    []Content{{Type: ContentTypeText, Text: text}},
	}
}

// EstimateTokenCount provides a rough estimate of tokens in a message list.
func EstimateTokenCount(messages []Message) int {
	total := 0
	for _, msg := range messages {
		for _, c := range msg.Content {
			total += len(c.Text) / 4 // approx 1 token per 4 chars
		}
	}
	return total
}
