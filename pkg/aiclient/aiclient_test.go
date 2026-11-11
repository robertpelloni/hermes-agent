package aiclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/hermes-agent/pkg/ai"
	"github.com/robertpelloni/hermes-agent/pkg/aiclient"
)

func sseEvent(data any) string {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(data)
	return "data: " + strings.TrimSpace(buf.String()) + "\n\n"
}

func TestNewClient(t *testing.T) {
	client := aiclient.NewClient("openai", "test-key", "https://api.openai.com/v1")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClientCompleteOpenAICompatible(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		chunk := map[string]any{
			"id":      "chatcmpl-aiclient-1",
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"role":    "assistant",
						"content": "Hello from aiclient!",
					},
					"finish_reason": "stop",
				},
			},
		}
		fmt.Fprint(w, sseEvent(chunk))
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer mockServer.Close()

	client := aiclient.NewClient("local-llm", "test-key", mockServer.URL+"/v1")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, _, err := client.Complete(ctx, "test-model", []ai.Message{
		ai.NewTextMessage(ai.RoleUser, "Hi"),
	}, aiclient.StreamOptions{
		Temperature: 0.7,
		MaxTokens:   512,
	})

	if err != nil {
		t.Fatalf("unexpected Complete error: %v", err)
	}
	if resp != "Hello from aiclient!" {
		t.Errorf("expected 'Hello from aiclient!', got: %s", resp)
	}
}
