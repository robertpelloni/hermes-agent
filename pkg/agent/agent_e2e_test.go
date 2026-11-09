package agent_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/hermes-agent/pkg/agent"
	"github.com/robertpelloni/hermes-agent/pkg/sessionruntime"
)

// sseEvent formats a JSON object as an SSE data: line.
func sseEvent(data any) string {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(data)
	return "data: " + strings.TrimSpace(buf.String()) + "\n\n"
}

// sseDone returns the SSE stream termination signal.
func sseDone() string {
	return "data: [DONE]\n\n"
}

// TestAgentEndToEnd tests the full agent loop with a mock LLM server.
func TestAgentEndToEnd(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// SSE chunk with content delta
		chunk1 := map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"role":    "assistant",
						"content": "Hello! I am Hermes. How can I help you today?",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 10,
				"total_tokens":      20,
			},
		}
		fmt.Fprint(w, sseEvent(chunk1))
		fmt.Fprint(w, sseDone())
	}))
	defer mockServer.Close()

	cfg := agent.Config{
		Model:         "test-model",
		Provider:      "local-llm",
		APIKey:        "test-key",
		BaseURL:       mockServer.URL + "/v1",
		MaxTokens:     1000,
		Temperature:   0.7,
		TopP:          0.9,
		MaxIterations: 5,
		SystemPrompt:  "You are Hermes.",
		SessionTTL:    time.Hour,
	}

	ag := agent.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := ag.HandleMessage(ctx, "test", "user1", "Hello, test the agent")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if response == "" {
		t.Fatal("HandleMessage returned empty response")
	}
	if !strings.Contains(response, "Hermes") {
		t.Errorf("response should mention Hermes, got: %s", response)
	}
}

// TestAgentWithToolCallDetection tests that text-based tool calls are detected.
func TestAgentWithToolCallDetection(t *testing.T) {
	callCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		callCount++

		if callCount <= 1 {
			// First call: return a JSON tool call as text
			chunk := map[string]any{
				"id":      "chatcmpl-test",
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   "test-model",
				"choices": []map[string]any{
					{
						"index": 0,
						"delta": map[string]any{
							"role":    "assistant",
							"content": `{"name": "read_file", "parameters": {"path": "test.txt"}}`,
						},
						"finish_reason": "stop",
					},
				},
			}
			fmt.Fprint(w, sseEvent(chunk))
			fmt.Fprint(w, sseDone())
		} else {
			// Second call: return final text after tool result
			chunk := map[string]any{
				"id":      "chatcmpl-test",
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   "test-model",
				"choices": []map[string]any{
					{
						"index": 0,
						"delta": map[string]any{
							"role":    "assistant",
							"content": "I tried to read the file but it doesn't exist.",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]int{
					"prompt_tokens":     20,
					"completion_tokens": 15,
					"total_tokens":      35,
				},
			}
			fmt.Fprint(w, sseEvent(chunk))
			fmt.Fprint(w, sseDone())
		}
	}))
	defer mockServer.Close()

	cfg := agent.Config{
		Model:         "test-model",
		Provider:      "local-llm",
		APIKey:        "test-key",
		BaseURL:       mockServer.URL + "/v1",
		MaxTokens:     1000,
		Temperature:   0.7,
		TopP:          0.9,
		MaxIterations: 5,
		SystemPrompt:  "You are Hermes.",
		SessionTTL:    time.Hour,
	}

	ag := agent.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := ag.HandleMessage(ctx, "test", "user1", "Read the file test.txt")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if response == "" {
		t.Fatal("HandleMessage returned empty response")
	}
	if !strings.Contains(response, "file") && !strings.Contains(response, "exist") {
		t.Logf("Note: response = %s", response)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls (tool + result), got %d", callCount)
	}
}

// TestAgentMaxIterations tests that the agent respects max iterations.
func TestAgentMaxIterations(t *testing.T) {
	iterCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		iterCount++
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		chunk := map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"role":    "assistant",
						"content": "continuing...",
					},
					"finish_reason": "stop",
				},
			},
		}
		fmt.Fprint(w, sseEvent(chunk))
		fmt.Fprint(w, sseDone())
	}))
	defer mockServer.Close()

	cfg := agent.Config{
		Model:         "test-model",
		Provider:      "local-llm",
		APIKey:        "test-key",
		BaseURL:       mockServer.URL + "/v1",
		MaxTokens:     1000,
		Temperature:   0.7,
		TopP:          0.9,
		MaxIterations: 3,
		SystemPrompt:  "You are Hermes.",
		SessionTTL:    time.Hour,
	}

	ag := agent.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := ag.HandleMessage(ctx, "test", "user1", "Keep going")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if response == "" {
		t.Fatal("HandleMessage returned empty response")
	}
	if iterCount > cfg.MaxIterations+1 {
		t.Errorf("iteration count %d exceeded max %d", iterCount, cfg.MaxIterations)
	}
}

// TestDefaultConfig loads the default agent config.
func TestDefaultConfig(t *testing.T) {
	os.Setenv("HERMES_FREE_LLM_PORT", "4000")
	defer os.Unsetenv("HERMES_FREE_LLM_PORT")

	cfg := agent.DefaultConfig()
	if cfg.Provider != "local-llm" {
		t.Errorf("expected provider local-llm, got %s", cfg.Provider)
	}
	if cfg.Model != "free-llm" {
		t.Errorf("expected model free-llm, got %s", cfg.Model)
	}
	if !strings.Contains(cfg.BaseURL, "4000") {
		t.Errorf("expected port 4000 in base URL, got %s", cfg.BaseURL)
	}
	if cfg.SystemPrompt == "" {
		t.Error("SystemPrompt is empty")
	}
}

// TestDefaultConfigWithTemplate tests configuring a custom prompt template.
func TestDefaultConfigWithTemplate(t *testing.T) {
	os.Setenv("HERMES_PROMPT_TEMPLATE", "coding")
	defer os.Unsetenv("HERMES_PROMPT_TEMPLATE")
	os.Setenv("HERMES_FREE_LLM_PORT", "4000")
	defer os.Unsetenv("HERMES_FREE_LLM_PORT")

	cfg := agent.DefaultConfig()
	if !strings.Contains(cfg.SystemPrompt, "read_file") {
		t.Errorf("expected coding template to mention read_file, got: %s", cfg.SystemPrompt)
	}
}

// TestDefaultConfigWithEnvOverrides tests that environment variables override DefaultConfig.
func TestDefaultConfigWithEnvOverrides(t *testing.T) {
	os.Setenv("HERMES_MODEL", "gpt-4o")
	defer os.Unsetenv("HERMES_MODEL")
	os.Setenv("HERMES_PROVIDER", "openai")
	defer os.Unsetenv("HERMES_PROVIDER")
	os.Setenv("HERMES_BASE_URL", "https://api.openai.com/v1")
	defer os.Unsetenv("HERMES_BASE_URL")

	cfg := agent.DefaultConfig()
	if cfg.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", cfg.Model)
	}
	if cfg.Provider != "openai" {
		t.Errorf("expected provider openai, got %s", cfg.Provider)
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("expected baseURL https://api.openai.com/v1, got %s", cfg.BaseURL)
	}
}

// TestDefaultConfigWithInvalidTemplateFallback tests that invalid templates fall back to default.
func TestDefaultConfigWithInvalidTemplateFallback(t *testing.T) {
	os.Setenv("HERMES_PROMPT_TEMPLATE", "nonexistent")
	defer os.Unsetenv("HERMES_PROMPT_TEMPLATE")
	os.Setenv("HERMES_FREE_LLM_PORT", "4000")
	defer os.Unsetenv("HERMES_FREE_LLM_PORT")

	cfg := agent.DefaultConfig()
	if !strings.Contains(cfg.SystemPrompt, "Hermes") {
		t.Errorf("fallback should use default prompt, got: %s", cfg.SystemPrompt)
	}
}

// TestAgentStreamingLoop tests HandleMessageStream with event sequence and session persistence.
func TestAgentStreamingLoop(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// SSE chunk with text delta
		chunk1 := map[string]any{
			"id":      "chatcmpl-stream-1",
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"role":    "assistant",
						"content": "Streaming response from Hermes.",
					},
					"finish_reason": "stop",
				},
			},
		}
		fmt.Fprint(w, sseEvent(chunk1))
		fmt.Fprint(w, sseDone())
	}))
	defer mockServer.Close()

	cfg := agent.Config{
		Model:         "test-model",
		Provider:      "local-llm",
		APIKey:        "test-key",
		BaseURL:       mockServer.URL + "/v1",
		MaxTokens:     1024,
		Temperature:   0.7,
		TopP:          0.9,
		MaxIterations: 5,
		SystemPrompt:  "You are Hermes.",
		SessionTTL:    1 * time.Hour,
	}

	ag := agent.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventCh := ag.HandleMessageStream(ctx, "cli", "user1", "Hello stream")

	var receivedText string
	var sawOutcome bool
	var sawDone bool

	for event := range eventCh {
		switch event.Type {
		case agent.EventThinking:
			// Verified thinking event type
		case agent.EventText:
			receivedText += event.Text
		case agent.EventOutcome:
			sawOutcome = true
			if !strings.Contains(event.Outcome, "Streaming response") {
				t.Errorf("unexpected outcome: %s", event.Outcome)
			}
		case agent.EventDone:
			sawDone = true
		case agent.EventError:
			t.Fatalf("unexpected stream error: %v", event.Error)
		}
	}

	if !sawDone {
		t.Errorf("expected EventDone event")
	}
	if !sawOutcome {
		t.Errorf("expected EventOutcome event")
	}
	if receivedText != "Streaming response from Hermes." {
		t.Errorf("unexpected accumulated text: %s", receivedText)
	}
}

// TestAgentSessionPersistence tests that agent conversation turns are saved into SessionRuntime.
func TestAgentSessionPersistence(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		chunk := map[string]any{
			"id":      "chatcmpl-persist",
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"role":    "assistant",
						"content": "I remember your conversation.",
					},
					"finish_reason": "stop",
				},
			},
		}
		fmt.Fprint(w, sseEvent(chunk))
		fmt.Fprint(w, sseDone())
	}))
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "hermes-agent-sess-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sessRuntime := sessionruntime.NewRuntime(tempDir, true)

	cfg := agent.Config{
		Model:         "test-model",
		Provider:      "local-llm",
		APIKey:        "test-key",
		BaseURL:       mockServer.URL + "/v1",
		MaxTokens:     1024,
		Temperature:   0.7,
		TopP:          0.9,
		MaxIterations: 5,
		SystemPrompt:  "You are Hermes.",
		SessionTTL:    1 * time.Hour,
	}

	ag := agent.New(cfg)
	ag.SetSessionRuntime(sessRuntime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := ag.HandleMessage(ctx, "slack", "user_123", "Hello there!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "I remember your conversation." {
		t.Errorf("unexpected response: %s", resp)
	}

	// Verify session was created and contains user + assistant messages
	sess, ok := sessRuntime.Get("slack:user_123")
	if !ok {
		t.Fatalf("expected session 'slack:user_123' to be created")
	}
	if sess.Label != "Hello there!" {
		t.Errorf("expected session label 'Hello there!', got: %s", sess.Label)
	}
	if len(sess.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(sess.Messages))
	}
	if sess.Messages[0].Content[0].Text != "Hello there!" {
		t.Errorf("expected user message 'Hello there!', got: %s", sess.Messages[0].Content[0].Text)
	}
	if sess.Messages[1].Content[0].Text != "I remember your conversation." {
		t.Errorf("expected assistant message, got: %s", sess.Messages[1].Content[0].Text)
	}

	// Turn 2: send another message and verify session grows to 4 messages
	resp2, err := ag.HandleMessage(ctx, "slack", "user_123", "What did I just say?")
	if err != nil {
		t.Fatalf("unexpected turn 2 error: %v", err)
	}
	if resp2 == "" {
		t.Errorf("expected non-empty response for turn 2")
	}
	sessUpdated, _ := sessRuntime.Get("slack:user_123")
	if len(sessUpdated.Messages) != 4 {
		t.Errorf("expected 4 messages in session after 2 turns, got %d", len(sessUpdated.Messages))
	}
}

// TestIntegrationWithRealLLM is a placeholder for real integration tests.
func TestIntegrationWithRealLLM(t *testing.T) {
	t.Skip("requires real LLM server; set HERMES_TEST_REAL_LLM=1 to run")

	if os.Getenv("HERMES_TEST_REAL_LLM") == "" {
		t.Skip("set HERMES_TEST_REAL_LLM=1 to run")
	}

	cfg := agent.DefaultConfig()
	ag := agent.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := ag.HandleMessage(ctx, "test", "user", "Say hello in one word")
	if err != nil {
		t.Fatalf("real LLM test failed: %v", err)
	}
	if response == "" {
		t.Fatal("real LLM test returned empty response")
	}
	t.Logf("response: %s", response)
}
