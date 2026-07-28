package agent

// StreamEventType categorizes events from HandleMessageStream.
type StreamEventType string

const (
	EventText      StreamEventType = "text"
	EventToolStart StreamEventType = "tool_start"
	EventToolChunk StreamEventType = "tool_chunk"
	EventToolDone  StreamEventType = "tool_done"
	EventOutcome   StreamEventType = "outcome"
	EventDone      StreamEventType = "done"
	EventError     StreamEventType = "error"
)

// StreamEvent represents a single event emitted during streaming message handling.
type StreamEvent struct {
	Type     StreamEventType
	Text     string
	ToolName string
	ToolID   string
	ToolArgs map[string]any

	// ToolResult is set on EventToolDone.
	ToolResult string
	ToolError  error

	// Outcome is the final text of the conversation, set on EventOutcome.
	Outcome string

	// Error is set on EventError.
	Error error
}
