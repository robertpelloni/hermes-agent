package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/robertpelloni/hermes-agent/pkg/ai"
	"github.com/robertpelloni/hermes-agent/pkg/aiclient"
	"github.com/robertpelloni/hermes-agent/pkg/prompttemplates"
	"github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

// Agent is the core Hermes AI agent.
type Agent struct {
	config   Config
	running  bool
}

// Config holds agent configuration.
type Config struct {
	Model         string
	Provider      string
	APIKey        string
	BaseURL       string
	MaxTokens     int
	Temperature   float32
	TopP          float32
	MaxIterations int
	SystemPrompt  string
	SessionTTL    time.Duration
}

// DefaultConfig returns a default agent configuration.
func DefaultConfig() Config {
	apiKey := os.Getenv("HERMES_FREE_LLM_API_KEY")
	if apiKey == "" {
		apiKey = "not-needed"
	}

	// Load system prompt from template (configurable via HERMES_PROMPT_TEMPLATE env var)
	promptTemplate := getEnvDefault("HERMES_PROMPT_TEMPLATE", "default")
	systemPrompt := prompttemplates.Default()
	if loaded, err := prompttemplates.Load(promptTemplate); err == nil {
		systemPrompt = loaded
	} else if promptTemplate != "default" {
		fmt.Fprintf(os.Stderr, "[hermes] warning: prompt template %q not found, using default\n", promptTemplate)
	}

	return Config{
		Model:         "free-llm",
		Provider:      "local-llm",
		APIKey:        apiKey,
		BaseURL:       fmt.Sprintf("http://127.0.0.1:%s/v1", getEnvDefault("HERMES_FREE_LLM_PORT", "4000")),
		MaxTokens:     4096,
		Temperature:   0.7,
		TopP:          0.9,
		MaxIterations: 10,
		SystemPrompt:  systemPrompt,
		SessionTTL:    24 * time.Hour,
	}
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}

// New creates a new Hermes Agent.
func New(cfg Config) *Agent {
	return &Agent{config: cfg, running: false}
}

// Run starts the agent's main loop (no-op for now; the agent is driven per-message).
func (a *Agent) Run(ctx context.Context) error {
	fmt.Printf("[hermes] Agent ready: model=%s provider=%s baseURL=%s\n", a.config.Model, a.config.Provider, a.config.BaseURL)
	a.running = true
	return nil
}

// SystemPrompt returns the configured system prompt.
func (a *Agent) SystemPrompt() string {
	return a.config.SystemPrompt
}

// HandleMessage processes an incoming user message through the full agent loop.
func (a *Agent) HandleMessage(ctx context.Context, platform, userID, text string) (string, error) {
	fmt.Printf("[hermes] Message from %s/%s: %s\n", platform, userID, text)
	return a.runConversation(ctx, text)
}

// HandleMessageStream processes a user message and streams events on the returned channel.
// The caller must read from the channel until it is closed.
func (a *Agent) HandleMessageStream(ctx context.Context, platform, userID, text string) <-chan StreamEvent {
	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		a.handleMessageStreamImpl(ctx, text, ch)
	}()
	return ch
}

// handleMessageStreamImpl is the goroutine that emits StreamEvents.
func (a *Agent) handleMessageStreamImpl(ctx context.Context, userText string, ch chan<- StreamEvent) {
	messages := a.buildInitialMessages(userText)

	for iter := 0; iter < a.config.MaxIterations; iter++ {
		messages = a.maybeCompact(messages)
		toolDefs := a.buildToolDefinitions()
		client := aiclient.NewClient(a.config.Provider, a.config.APIKey, a.config.BaseURL)

		var assistantText string
		var pendingToolCalls []ai.ToolCallData
		var usage *ai.Usage

		stream, err := client.Stream(ctx, a.config.Model, messages, aiclient.StreamOptions{
			Temperature: a.config.Temperature,
			MaxTokens:   a.config.MaxTokens,
			TopP:        a.config.TopP,
			Tools:       toolDefs,
		})
		if err != nil {
			ch <- StreamEvent{Type: EventError, Error: fmt.Errorf("ai stream error: %w", err)}
			ch <- StreamEvent{Type: EventDone}
			return
		}

		type partialTool struct {
			ID        string
			Name      string
			Arguments string
		}
		partials := make(map[int]*partialTool)

		for event := range stream {
			switch event.Type {
			case "error":
				ch <- StreamEvent{Type: EventError, Error: event.Error}
				ch <- StreamEvent{Type: EventDone}
				return

			case "text_delta":
				assistantText += event.Text
				ch <- StreamEvent{Type: EventText, Text: event.Text}

			case "thinking_delta":
				if event.Thinking != "" {
					// thinking is included in the final response
				}

			case "tool_call_delta":
				if event.ToolCall != nil {
					idx := -1
					for i, pt := range partials {
						if pt.ID == "" || pt.ID == event.ToolCall.ID {
							idx = i
							break
						}
					}
					if idx < 0 {
						idx = len(partials)
					}
					if _, exists := partials[idx]; !exists {
						partials[idx] = &partialTool{ID: event.ToolCall.ID, Name: event.ToolCall.Name}
						// Send tool start event
						ch <- StreamEvent{
							Type:     EventToolStart,
							ToolName: event.ToolCall.Name,
							ToolID:   event.ToolCall.ID,
							ToolArgs: event.ToolCall.Arguments,
						}
					}
				}

			case "tool_call_complete":
				if event.ToolCall != nil {
					existing := false
					for _, pt := range pendingToolCalls {
						if pt.ID == event.ToolCall.ID {
							existing = true
							break
						}
					}
					if !existing {
						pendingToolCalls = append(pendingToolCalls, *event.ToolCall)
					}
				}

			case "usage":
				usage = event.Usage

			case "done":
			}
		}

		if usage != nil {
			fmt.Printf("[hermes] LLM usage: %d in / %d out\n", usage.Input, usage.Output)
		}

		if assistantText != "" {
			messages = append(messages, ai.NewTextMessage(ai.RoleAssistant, assistantText))
		}

		// Detect text-based tool calls
		if len(pendingToolCalls) == 0 && assistantText != "" {
			if detected := detectTextToolCall(assistantText); detected != nil {
				assistantText = "[tool_call: " + detected.Name + "]"
				pendingToolCalls = append(pendingToolCalls, *detected)
				fmt.Printf("[hermes] Detected text-based tool call: %s\n", detected.Name)
			}
		}

		// Build remaining partial tool calls
		if len(pendingToolCalls) > 0 {
			for _, pt := range partials {
				alreadyIn := false
				for _, tc := range pendingToolCalls {
					if tc.ID == pt.ID {
						alreadyIn = true
						break
					}
				}
				if !alreadyIn && pt.Name != "" {
					args := make(map[string]any)
					if pt.Arguments != "" {
						if err := parseJSONArgs(pt.Arguments, &args); err != nil {
							args["raw"] = pt.Arguments
						}
					}
					pendingToolCalls = append(pendingToolCalls, ai.ToolCallData{
						ID:        pt.ID,
						Name:      pt.Name,
						Arguments: args,
					})
				}
			}

			// Execute each tool call, emitting events
			for _, tc := range pendingToolCalls {
				// Emit tool start if not already emitted
				// (partial tool might not have emitted it via tool_call_delta)
				ch <- StreamEvent{
					Type:     EventToolStart,
					ToolName: tc.Name,
					ToolID:   tc.ID,
					ToolArgs: tc.Arguments,
				}

				result, err := toolregistry.Global().Call(tc.Name, tc.Arguments, map[string]any{})
				if err != nil {
					ch <- StreamEvent{
						Type:       EventToolDone,
						ToolName:   tc.Name,
						ToolID:     tc.ID,
						ToolResult: "",
						ToolError:  err,
					}
					messages = append(messages, ai.NewToolResultMessage(tc.ID, tc.Name, err, true))
				} else {
					// If result is a map or any type, convert to string
					resultStr := fmt.Sprintf("%v", result)
					ch <- StreamEvent{
						Type:       EventToolDone,
						ToolName:   tc.Name,
						ToolID:     tc.ID,
						ToolResult: resultStr,
					}
					messages = append(messages, ai.NewToolResultMessage(tc.ID, tc.Name, result, false))
				}
			}

			continue // next iteration
		}

		// No tool calls – this is the final response
		ch <- StreamEvent{Type: EventOutcome, Outcome: assistantText}
		ch <- StreamEvent{Type: EventDone}
		return
	}

	ch <- StreamEvent{Type: EventError, Error: fmt.Errorf("max iterations reached without final response")}
	ch <- StreamEvent{Type: EventDone}
}


// runConversation runs the full agent loop:
// 1. Manage message history
// 2. Call AI for streaming response
// 3. Handle tool calls
// 4. Repeat until text response
func (a *Agent) runConversation(ctx context.Context, userText string) (string, error) {
	// Build initial message list
	messages := a.buildInitialMessages(userText)

	// Main conversation loop
	var finalResponse string

	for iter := 0; iter < a.config.MaxIterations; iter++ {
		// Check compaction before each LLM call
		messages = a.maybeCompact(messages)

		// Build tool definitions from registry
		toolDefs := a.buildToolDefinitions()

		// Call AI client
		client := aiclient.NewClient(a.config.Provider, a.config.APIKey, a.config.BaseURL)

		var assistantText string
		var pendingToolCalls []ai.ToolCallData
		var usage *ai.Usage

		stream, err := client.Stream(ctx, a.config.Model, messages, aiclient.StreamOptions{
			Temperature: a.config.Temperature,
			MaxTokens:   a.config.MaxTokens,
			TopP:        a.config.TopP,
			Tools:       toolDefs,
		})
		if err != nil {
			return "", fmt.Errorf("ai stream error: %w", err)
		}

		// Accumulate tool call arguments by index
		type partialTool struct {
			ID        string
			Name      string
			Arguments string
		}
		partials := make(map[int]*partialTool)

		for event := range stream {
			switch event.Type {
			case "error":
				return "", fmt.Errorf("ai streaming error: %w", event.Error)

			case "text_delta":
				assistantText += event.Text

			case "tool_call_delta":
				if event.ToolCall != nil {
					idx := -1
					for i, pt := range partials {
						if pt.ID == "" || pt.ID == event.ToolCall.ID {
							idx = i
							break
						}
					}
					if idx < 0 {
						idx = len(partials)
					}
					if _, exists := partials[idx]; !exists {
						partials[idx] = &partialTool{ID: event.ToolCall.ID, Name: event.ToolCall.Name}
					}
				}

			case "tool_call_complete":
				if event.ToolCall != nil {
					existing := false
					for _, pt := range pendingToolCalls {
						if pt.ID == event.ToolCall.ID {
							existing = true
							break
						}
					}
					if !existing {
						pendingToolCalls = append(pendingToolCalls, *event.ToolCall)
					}
				}

			case "usage":
				usage = event.Usage

			case "done":
			}
		}

		// Log usage if available
		if usage != nil {
			fmt.Printf("[hermes] LLM usage: %d in / %d out\n", usage.Input, usage.Output)
		}

		// Add assistant message to history
		if assistantText != "" {
			messages = append(messages, ai.NewTextMessage(ai.RoleAssistant, assistantText))
		}

		// Handle tool calls – check both structured tool_calls and text-based tool calls
		if len(pendingToolCalls) == 0 && assistantText != "" {
			// Some servers (e.g. FreeLLM) return tool calls as text JSON instead of structured tool_calls
			if detected := detectTextToolCall(assistantText); detected != nil {
				// Clear the text since we're treating this as a tool call, not a text response
				assistantText = "[tool_call: " + detected.Name + "]"
				pendingToolCalls = append(pendingToolCalls, *detected)
				fmt.Printf("[hermes] Detected text-based tool call: %s\n", detected.Name)
			}
		}

		if len(pendingToolCalls) > 0 {
			// Also check for tool calls accumulated from delta events that weren't emitted as complete
			for _, pt := range partials {
				alreadyIn := false
				for _, tc := range pendingToolCalls {
					if tc.ID == pt.ID {
						alreadyIn = true
						break
					}
				}
				if !alreadyIn && pt.Name != "" {
					args := make(map[string]any)
					if pt.Arguments != "" {
						// Arguments are JSON; accept simple string maps
						if err := parseJSONArgs(pt.Arguments, &args); err != nil {
							args["raw"] = pt.Arguments
						}
					}
					pendingToolCalls = append(pendingToolCalls, ai.ToolCallData{
						ID:        pt.ID,
						Name:      pt.Name,
						Arguments: args,
					})
				}
			}

			fmt.Printf("[hermes] Tool calls: %d\n", len(pendingToolCalls))

			// Add tool call message to history
			for _, tc := range pendingToolCalls {
				messages = append(messages, ai.NewTextMessage(
					ai.RoleAssistant,
					fmt.Sprintf("[Using tool: %s]", tc.Name),
				))
			}

			// Execute each tool call
			for _, tc := range pendingToolCalls {
				result, err := toolregistry.Global().Call(tc.Name, tc.Arguments, map[string]any{})
				if err != nil {
					messages = append(messages, ai.NewToolResultMessage(tc.ID, tc.Name, err, true))
				} else {
					messages = append(messages, ai.NewToolResultMessage(tc.ID, tc.Name, result, false))
				}
			}

			continue
		}

		// No tool calls - this is the final response
		finalResponse = assistantText
		break
	}

	return finalResponse, nil
}

// buildInitialMessages creates the initial message list with system prompt and user message.
func (a *Agent) buildInitialMessages(userText string) []ai.Message {
	messages := make([]ai.Message, 0, 4)

	// Add system prompt
	messages = append(messages, ai.NewTextMessage("system", a.config.SystemPrompt))

	// Add user message
	messages = append(messages, ai.NewTextMessage(ai.RoleUser, userText))

	return messages
}

// buildToolDefinitions converts registered tools into AI-compatible definitions.
func (a *Agent) buildToolDefinitions() []ai.ToolDefinition {
	registered := toolregistry.Global().List()
	defs := make([]ai.ToolDefinition, 0, len(registered))

	for _, t := range registered {
		def := ai.ToolDefinition{
			Name: t.Name,
			Description: t.Description,
			Parameters: make([]ai.ToolParameter, 0),
		}

		// Convert from map[string]any parameters to []ToolParameter
		for paramName, paramInfo := range t.Parameters {
			param := ai.ToolParameter{
				Name: paramName,
				Type: "string",
				Required: true,
			}

			if info, ok := paramInfo.(map[string]any); ok {
				if t, ok := info["type"].(string); ok {
					param.Type = t
				}
				if desc, ok := info["description"].(string); ok {
					param.Description = desc
				}
				if req, ok := info["required"]; ok {
					if b, ok := req.(bool); ok {
						param.Required = b
					}
				}
			}

			def.Parameters = append(def.Parameters, param)
		}

		defs = append(defs, def)
	}

	return defs
}

// maybeCompact checks if the message list exceeds the token budget and compacts if needed.
func (a *Agent) maybeCompact(messages []ai.Message) []ai.Message {
	threshold := getEnvDefaultInt("HERMES_COMPACT_THRESHOLD", 8000)
	tokens := ai.EstimateTokenCount(messages)

	if tokens <= threshold {
		return messages
	}

	fmt.Printf("[hermes] Compacting context: ~%d tokens (threshold %d)\n", tokens, threshold)

	return compactMessages(messages)
}

// compactMessages performs lossy context compression by summarizing older messages.
func compactMessages(messages []ai.Message) []ai.Message {
	if len(messages) <= 4 {
		return messages
	}

	// Keep system prompt + last N recent messages
	keepRecent := 6 // system + user + assistant = at least 3 pairs
	if len(messages) <= keepRecent+2 {
		return messages
	}

	// Keep system prompt (index 0)
	compacted := []ai.Message{messages[0]}

	// Compress middle messages
	compressed := make([]string, 0, len(messages)-keepRecent)
	for i := 1; i < len(messages)-keepRecent; i++ {
		msg := messages[i]
		for _, c := range msg.Content {
			if c.Text != "" {
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
				// Truncate long content
				content := c.Text
				if len(content) > 200 {
					content = content[:200] + "..."
				}
				compressed = append(compressed, prefix+content)
			}
		}
	}

	if len(compressed) > 0 {
		summary := "[Compressed history: " + strings.Join(compressed, "; ") + "]"
		compacted = append(compacted, ai.NewTextMessage(ai.RoleUser, summary))
	}

	// Add recent messages
	for i := len(messages) - keepRecent; i < len(messages); i++ {
		compacted = append(compacted, messages[i])
	}

	newTokens := ai.EstimateTokenCount(compacted)
	fmt.Printf("[hermes] Compacted: %d messages -> %d messages, ~%d tokens\n", len(messages), len(compacted), newTokens)

	return compacted
}

// detectTextToolCall checks if the assistant's text response contains a tool call.
// This handles servers that return tool calls as text instead of structured tool_calls.
func detectTextToolCall(text string) *ai.ToolCallData {
	if text == "" {
		return nil
	}

	// Pattern 1: JSON with "name" and "parameters" keys
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(text), &parsed); err == nil {
			name, hasName := parsed["name"].(string)
			if hasName && name != "" {
				params, hasParams := parsed["parameters"].(map[string]any)
				if !hasParams {
					// Try function or param
					if p, ok := parsed["function"]; ok {
						if pm, ok := p.(string); ok {
							params = map[string]any{"raw": pm}
						}
					}
					// Use the rest of the JSON as arguments
					if params == nil {
						params = make(map[string]any)
						for k, v := range parsed {
							if k != "name" && k != "id" {
								params[k] = v
							}
						}
					}
				}
				if params != nil {
					return &ai.ToolCallData{
						Name:      name,
						Arguments: params,
					}
				}
			}
		}
	}

	// Pattern 2: XML-like <tool_call> block
	if strings.Contains(text, "<tool_call>") {
		name := extractXMLTag(text, "function")
		if name == "" {
			name = extractXMLTag(text, "tool")
		}
		if name != "" {
			params := make(map[string]any)
			// Extract parameter tags
			paramPatterns := extractXMLParams(text)
			for k, v := range paramPatterns {
				params[k] = v
			}
			return &ai.ToolCallData{
				Name:      name,
				Arguments: params,
			}
		}
	}

	// Pattern 3: function=name in angle brackets
	if strings.Contains(text, "<function=") {
		parts := strings.Split(text, "<function=")
		if len(parts) >= 2 {
			name := strings.TrimSuffix(parts[1], ">")
			name = strings.Split(name, ">")[0]
			if name != "" {
				params := make(map[string]any)
				// Look for argument pattern
				paramPatterns := extractXMLParams(parts[1])
				for k, v := range paramPatterns {
					params[k] = v
				}
				return &ai.ToolCallData{
					Name:      name,
					Arguments: params,
				}
			}
		}
	}

	return nil
}

// extractXMLTag extracts the content of an XML tag.
func extractXMLTag(text, tag string) string {
	prefix := "<" + tag + "="
	idx := strings.Index(text, prefix)
	if idx < 0 {
		// Try <tag>...</tag>
		open := "<" + tag + ">"
		close := "</" + tag + ">"
		start := strings.Index(text, open)
		end := strings.Index(text, close)
		if start >= 0 && end > start {
			return strings.TrimSpace(text[start+len(open) : end])
		}
		return ""
	}
	rest := text[idx+len(prefix):]
	end := strings.IndexAny(rest, "> \n\r")
	if end < 0 {
		return ""
	}
	return rest[:end]
}

// extractXMLParams extracts named parameters from XML-style tag content.
func extractXMLParams(text string) map[string]any {
	params := make(map[string]any)
	for _, prefix := range []string{"<parameter=", "<param=", "<arg="} {
		idx := 0
		for {
			pos := strings.Index(text[idx:], prefix)
			if pos < 0 {
				break
			}
			pos += idx
			rest := text[pos+len(prefix):]
			end := strings.IndexAny(rest, "> \n\r")
			if end > 0 {
				paramName := rest[:end]
				// Look for the value after closing tag
				valStart := strings.Index(rest[end:], "</"+strings.TrimPrefix(prefix, "<")[:len(strings.TrimPrefix(prefix, "<"))-1]+">")
				if valStart > 0 && strings.HasPrefix(rest[end:], "\n") {
					// Newline after >, value is on next lines
					valLines := strings.Split(rest[end+1:], "</")
					if len(valLines) > 0 {
						params[paramName] = strings.TrimSpace(valLines[0])
					}
				} else {
					// Try <parameter=name>value</parameter>
					closeTag := "</" + strings.TrimSuffix(strings.TrimPrefix(prefix, "<"), "=") + ">"
					if rest[end:] != "\n" && !strings.HasPrefix(rest[end:], "\n") {
						closeIdx := strings.Index(rest[end:], closeTag)
						if closeIdx > 0 {
							params[paramName] = strings.TrimSpace(rest[end : end+closeIdx])
						}
					}
				}
			}
			idx = pos + 1
		}
	}
	return params
}

// parseJSONArgs attempts to parse a JSON string into a map.
func parseJSONArgs(s string, out *map[string]any) error {
	s = strings.TrimSpace(s)
	if s == "" {
		*out = make(map[string]any)
		return nil
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		// Fallback: store raw string
		*out = make(map[string]any)
		(*out)["raw"] = s
		return nil
	}
	*out = parsed
	return nil
}
