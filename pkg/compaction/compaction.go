package compaction

import (
	"fmt"
	"strings"

	"github.com/robertpelloni/hermes-agent/pkg/ai"
)

// Compactor provides context compaction for long conversations.
type Compactor struct {
	ThresholdTokens int // Token count threshold to trigger compaction
	KeepRecent      int // Number of recent messages to keep after compaction
}

// New creates a new Compactor with default settings.
func New(threshold int) *Compactor {
	if threshold <= 0 {
		threshold = 8000
	}
	keep := 6
	return &Compactor{ThresholdTokens: threshold, KeepRecent: keep}
}

// ShouldCompact checks if the message list exceeds the token threshold.
func (c *Compactor) ShouldCompact(messages []ai.Message) (bool, int) {
	tokens := ai.EstimateTokenCount(messages)
	return tokens > c.ThresholdTokens, tokens
}

// Compact performs lossy compression on the message list.
// Always keeps the system prompt (first message) and the N most recent messages.
func (c *Compactor) Compact(messages []ai.Message) []ai.Message {
	if len(messages) <= c.KeepRecent+2 {
		return messages
	}

	compacted := make([]ai.Message, 0, c.KeepRecent+2)

	// Keep system prompt (index 0)
	if len(messages) > 0 {
		compacted = append(compacted, messages[0])
	}

	// Compress middle messages into a summary
	compressed := c.summarizeMessages(messages[1 : len(messages)-c.KeepRecent])

	if compressed != "" {
		compacted = append(compacted, ai.NewTextMessage(ai.RoleUser, compressed))
	}

	// Add recent messages
	compacted = append(compacted, messages[len(messages)-c.KeepRecent:]...)

	return compacted
}

// summarizeMessages creates a compressed text summary of the given messages.
func (c *Compactor) summarizeMessages(messages []ai.Message) string {
	if len(messages) == 0 {
		return ""
	}

	var lines []string
	for _, msg := range messages {
		for _, content := range msg.Content {
			if content.Text == "" {
				continue
			}
			prefix := ""
			switch msg.Role {
			case ai.RoleUser:
				prefix = "User: "
			case ai.RoleAssistant:
				prefix = "Assistant: "
			case ai.RoleTool:
				prefix = fmt.Sprintf("Tool(%s): ", msg.ToolName)
			default:
				prefix = string(msg.Role) + ": "
			}
			text := content.Text
			if len(text) > 300 {
				text = text[:300] + "..."
			}
			lines = append(lines, prefix+text)
		}
	}

	if len(lines) == 0 {
		return ""
	}

	return "[Compressed history: " + strings.Join(lines, " | ") + "]"
}

// TokenCount returns the estimated token count for the message list.
func (c *Compactor) TokenCount(messages []ai.Message) int {
	return ai.EstimateTokenCount(messages)
}
