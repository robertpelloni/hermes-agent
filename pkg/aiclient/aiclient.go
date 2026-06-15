package aiclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/robertpelloni/hermes-agent/pkg/ai"
)

// Client is a multi-provider AI client that streams responses.
type Client struct {
	provider string
	apiKey   string
	baseURL  string
}

// NewClient creates a new AI client for the specified provider.
func NewClient(provider, apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = getDefaultBaseURL(provider)
	}
	return &Client{
		provider: provider,
		apiKey:   apiKey,
		baseURL:  baseURL,
	}
}

func getDefaultBaseURL(provider string) string {
	switch provider {
	case "openai", "openai-compatible":
		return "https://api.openai.com/v1"
	case "anthropic":
		return "https://api.anthropic.com"
	case "google":
		return "https://generativelanguage.googleapis.com/v1beta"
	case "groq":
		return "https://api.groq.com/openai/v1"
	case "xai":
		return "https://api.x.ai/v1"
	case "mistral":
		return "https://api.mistral.ai/v1"
	case "openrouter":
		return "https://openrouter.ai/api/v1"
	case "deepseek":
		return "https://api.deepseek.com"
	case "local-llm":
		return "http://127.0.0.1:4000/v1"
	default:
		return "https://api.openai.com/v1"
	}
}

// StreamOptions holds streaming configuration.
type StreamOptions struct {
	Temperature float32 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"maxTokens,omitempty"`
	TopP        float32 `json:"topP,omitempty"`
	Tools       []ai.ToolDefinition `json:"tools,omitempty"`
	SystemPrompt string `json:"systemPrompt,omitempty"`
}

// StreamEvent represents a streaming event from the AI client.
type StreamEvent struct {
	Type       string          // "start", "text_delta", "thinking_delta", "tool_call_delta", "tool_call_complete", "usage", "done", "error"
	Text       string          `json:"text,omitempty"`
	Thinking   string          `json:"thinking,omitempty"`
	ToolCall   *ai.ToolCallData `json:"toolCall,omitempty"`
	Usage      *ai.Usage       `json:"usage,omitempty"`
	Error      error           `json:"error"`
	IsComplete bool            `json:"isComplete"`
	StopReason string          `json:"stopReason,omitempty"`
}

// Stream starts a streaming conversation with the AI.
func (c *Client) Stream(ctx context.Context, modelID string, messages []ai.Message, opts StreamOptions) (<-chan StreamEvent, error) {
	eventChan := make(chan StreamEvent, 100)

	// Route to provider-specific implementation
	switch c.provider {
	case "openai", "openrouter", "groq", "xai", "minimax", "deepseek", "local-llm", "openai-compatible":
		go c.streamOpenAICompatible(ctx, modelID, messages, opts, eventChan)
	case "anthropic":
		go c.streamAnthropic(ctx, modelID, messages, opts, eventChan)
	case "google":
		go c.streamGoogle(ctx, modelID, messages, opts, eventChan)
	case "mistral":
		go c.streamMistral(ctx, modelID, messages, opts, eventChan)
	default:
		go c.streamOpenAICompatible(ctx, modelID, messages, opts, eventChan)
	}

	return eventChan, nil
}

// streamOpenAICompatible handles OpenAI-compatible providers with SSE streaming.
func (c *Client) streamOpenAICompatible(ctx context.Context, modelID string, messages []ai.Message, opts StreamOptions, eventChan chan<- StreamEvent) {
	defer close(eventChan)

	// Convert messages to OpenAI format
	openAIMsgs := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		openAIMsg := map[string]any{"role": string(msg.Role)}
		if msg.Role == ai.RoleTool {
			// Tool result messages use content as string, not array
			content := ""
			for _, c := range msg.Content {
				content += c.Text
			}
			openAIMsg["content"] = content
			openAIMsg["tool_call_id"] = msg.ToolCallID
		} else {
			content := make([]map[string]any, 0, len(msg.Content))
			for _, c := range msg.Content {
				item := map[string]any{"type": string(c.Type)}
				switch c.Type {
				case ai.ContentTypeText:
					item["text"] = c.Text
				case ai.ContentTypeToolCall:
					item["tool_call_id"] = c.ToolCallID
					item["name"] = c.Name
					item["arguments"] = c.Arguments
				case ai.ContentTypeImage:
					item["image_url"] = c.Image
					item["mime_type"] = c.MimeType
				case ai.ContentTypeThinking:
					item["thinking"] = c.Thinking
				}
				content = append(content, item)
			}
			openAIMsg["content"] = content
		}
		openAIMsgs = append(openAIMsgs, openAIMsg)
	}

	// Build request body
	body := map[string]any{
		"model":    modelID,
		"messages": openAIMsgs,
		"stream":   true,
	}

	if opts.Temperature != 0 {
		body["temperature"] = opts.Temperature
	}
	if opts.MaxTokens > 0 {
		body["max_tokens"] = opts.MaxTokens
	}
	if opts.TopP > 0 {
		body["top_p"] = opts.TopP
	}
	if len(opts.Tools) > 0 {
		tools := make([]map[string]any, 0, len(opts.Tools))
		for _, t := range opts.Tools {
			toolDef := map[string]any{
				"type": "function",
				"function": map[string]any{
					"name":        t.Name,
					"description": t.Description,
				},
			}
			if len(t.Parameters) > 0 {
				props := make(map[string]any)
				required := make([]string, 0)
				for _, p := range t.Parameters {
					props[p.Name] = map[string]string{"type": p.Type, "description": p.Description}
					if p.Required {
						required = append(required, p.Name)
					}
				}
				toolDef["function"].(map[string]any)["parameters"] = map[string]any{
					"type":       "object",
					"properties": props,
					"required":   required,
				}
			}
			tools = append(tools, toolDef)
		}
		body["tools"] = tools
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		eventChan <- StreamEvent{Type: "error", Error: fmt.Errorf("failed to marshal request: %w", err)}
		return
	}

	// Build and send request
	endpoint := c.baseURL + "/chat/completions"
	if !strings.Contains(endpoint, "/chat/completions") && strings.HasSuffix(c.baseURL, "/") {
		endpoint = c.baseURL + "chat/completions"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		eventChan <- StreamEvent{Type: "error", Error: fmt.Errorf("failed to create request: %w", err)}
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		eventChan <- StreamEvent{Type: "error", Error: fmt.Errorf("request failed: %w", err)}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		eventChan <- StreamEvent{Type: "error", Error: fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))}
		return
	}

	// Parse SSE stream
	eventChan <- StreamEvent{Type: "start"}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	// Accumulate tool calls by index
	type partialToolCall struct {
		ID        string
		Name      string
		Arguments string
	}
	pendingCalls := make(map[int]*partialToolCall)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		// Parse the chunk
		var chunk struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			Model   string `json:"model"`
			Choices []struct {
				Delta struct {
					Content   *string `json:"content"`
					Role      *string `json:"role"`
					ToolCalls []struct {
						Index    int     `json:"index"`
						ID       *string `json:"id"`
						Type     *string `json:"type"`
						Function *struct {
							Name      *string `json:"name"`
							Arguments *string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
				Index        int     `json:"index"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // skip malformed chunks
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		delta := choice.Delta

		// Text content
		if delta.Content != nil && *delta.Content != "" {
			eventChan <- StreamEvent{Type: "text_delta", Text: *delta.Content}
		}

		// Tool calls
		for _, tc := range delta.ToolCalls {
			idx := tc.Index
			if _, exists := pendingCalls[idx]; !exists {
				tcID := ""
				if tc.ID != nil {
					tcID = *tc.ID
				}
				name := ""
				if tc.Function != nil && tc.Function.Name != nil {
					name = *tc.Function.Name
				}
				pendingCalls[idx] = &partialToolCall{
					ID:   tcID,
					Name: name,
				}
				if tcID != "" {
					eventChan <- StreamEvent{
						Type: "tool_call_delta",
						ToolCall: &ai.ToolCallData{
							ID:   tcID,
							Name: name,
						},
					}
				}
			}

			ptc := pendingCalls[idx]
			if tc.Function != nil && tc.Function.Name != nil {
				ptc.Name = *tc.Function.Name
			}
			if tc.Function != nil && tc.Function.Arguments != nil {
				ptc.Arguments += *tc.Function.Arguments
			}
		}

		// Check finish reason
		if choice.FinishReason != nil && *choice.FinishReason != "" {
			// Emit completed tool calls
			for _, ptc := range pendingCalls {
				args := make(map[string]any)
				if ptc.Arguments != "" {
					if err := json.Unmarshal([]byte(ptc.Arguments), &args); err != nil {
						// Invalid JSON arguments; use as-is
						args["raw"] = ptc.Arguments
					}
				}
				eventChan <- StreamEvent{
					Type: "tool_call_complete",
					ToolCall: &ai.ToolCallData{
						ID:        ptc.ID,
						Name:      ptc.Name,
						Arguments: args,
					},
				}
			}

			// Usage
			if chunk.Usage != nil {
				eventChan <- StreamEvent{
					Type: "usage",
					Usage: &ai.Usage{
						Input:       chunk.Usage.PromptTokens,
						Output:      chunk.Usage.CompletionTokens,
						TotalTokens: chunk.Usage.TotalTokens,
					},
				}
			}

			eventChan <- StreamEvent{
				Type:       "done",
				IsComplete: true,
				StopReason: *choice.FinishReason,
			}
			return
		}
	}

	if err := scanner.Err(); err != nil {
		eventChan <- StreamEvent{Type: "error", Error: fmt.Errorf("stream read error: %w", err)}
		return
	}

	// If we fell through (no finish reason), send done anyway
	eventChan <- StreamEvent{Type: "done", IsComplete: true}
}

// streamAnthropic handles Anthropic's streaming API.
func (c *Client) streamAnthropic(ctx context.Context, modelID string, messages []ai.Message, opts StreamOptions, eventChan chan<- StreamEvent) {
	defer close(eventChan)
	eventChan <- StreamEvent{Type: "start"}

	// Convert messages to Anthropic format
	anthropicMsgs := make([]map[string]any, 0, len(messages))
	systemPrompt := ""
	for _, msg := range messages {
		if msg.Role == ai.RoleAssistant {
			content := make([]map[string]any, 0, len(msg.Content))
			for _, c := range msg.Content {
				switch c.Type {
				case ai.ContentTypeText:
					content = append(content, map[string]any{"type": "text", "text": c.Text})
				case ai.ContentTypeThinking:
					content = append(content, map[string]any{"type": "thinking", "thinking": c.Thinking})
				case ai.ContentTypeToolCall:
					content = append(content, map[string]any{
						"type": "tool_use",
						"id":   c.ToolCallID,
						"name": c.Name,
						"input": c.Arguments,
					})
				}
			}
			anthropicMsgs = append(anthropicMsgs, map[string]any{"role": "assistant", "content": content})
		} else if msg.Role == ai.RoleUser {
			content := make([]map[string]any, 0, len(msg.Content))
			for _, c := range msg.Content {
				item := map[string]any{"type": "text", "text": c.Text}
				content = append(content, item)
			}
			anthropicMsgs = append(anthropicMsgs, map[string]any{"role": "user", "content": content})
		} else if msg.Role == ai.RoleTool {
			anthropicMsgs = append(anthropicMsgs, map[string]any{
				"role":         "user",
				"content":      []map[string]any{{"type": "tool_result", "tool_use_id": msg.ToolCallID, "content": textOf(msg)}},
			})
		} else if msg.Role == "system" {
			systemPrompt += textOf(msg) + "\n"
		}
	}

	body := map[string]any{
		"model":      modelID,
		"messages":   anthropicMsgs,
		"max_tokens": opts.MaxTokens,
		"stream":     true,
	}
	if systemPrompt != "" {
		body["system"] = systemPrompt
	}

	jsonBody, _ := json.Marshal(body)
	endpoint := c.baseURL + "/messages"
	req, _ := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	if c.apiKey != "" {
		req.Header.Set("anthropic-version", "2023-06-01")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		eventChan <- StreamEvent{Type: "error", Error: fmt.Errorf("request failed: %w", err)}
		return
	}
	defer resp.Body.Close()

	_ = resp // Parse SSE similarly to OpenAI
	eventChan <- StreamEvent{Type: "done", IsComplete: true}
}

// streamGoogle handles Google's Generative AI streaming API.
func (c *Client) streamGoogle(ctx context.Context, modelID string, messages []ai.Message, opts StreamOptions, eventChan chan<- StreamEvent) {
	defer close(eventChan)
	eventChan <- StreamEvent{Type: "start"}

	// TODO: Implement Google streaming
	eventChan <- StreamEvent{Type: "done", IsComplete: true}
}

// streamMistral handles Mistral AI's streaming API.
func (c *Client) streamMistral(ctx context.Context, modelID string, messages []ai.Message, opts StreamOptions, eventChan chan<- StreamEvent) {
	defer close(eventChan)
	eventChan <- StreamEvent{Type: "start"}

	// Mistral uses OpenAI-compatible API
	c.streamOpenAICompatible(ctx, modelID, messages, opts, eventChan)
}

// Complete is a convenience method that waits for the stream to complete and returns the final response.
func (c *Client) Complete(ctx context.Context, modelID string, messages []ai.Message, opts StreamOptions) (string, *ai.Usage, error) {
	stream, err := c.Stream(ctx, modelID, messages, opts)
	if err != nil {
		return "", nil, err
	}

	var fullResponse string
	var usage *ai.Usage

	for event := range stream {
		switch event.Type {
		case "error":
			return "", usage, event.Error
		case "text_delta":
			fullResponse += event.Text
		case "usage":
			usage = event.Usage
		}
	}

	return fullResponse, usage, nil
}

// textOf extracts text from a message's content blocks.
func textOf(msg ai.Message) string {
	var text string
	for _, c := range msg.Content {
		if c.Text != "" {
			text += c.Text + " "
		}
	}
	return strings.TrimSpace(text)
}
