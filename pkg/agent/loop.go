package agent

import (
	"context"
	"fmt"
	"os"
	"time"
	"github.com/robertpelloni/hermes-agent/pkg/memory"
)

// RunConversationLoop manages a single conversation turn
func (a *Agent) RunConversationLoop(ctx context.Context, platform, userID, text string, memStore *memory.Store) (string, error) {
	start := time.Now()
	fmt.Printf("[hermes] processing msg from %s/%s: %s\n", platform, userID, text)

	model := os.Getenv("HERMES_MODEL")
	if model == "" {
		model = "hermes-default-model"
	}
	provider := os.Getenv("HERMES_PROVIDER")

	// Simulate conversation logic and model execution
	response := fmt.Sprintf("I received your message on %s (via %s/%s): %s", platform, provider, model, text)

	// Track token usage logic per-model (e.g. simulated tokens here)
	tokensUsed := len(text) + len(response)
	durationMs := int(time.Since(start).Milliseconds())

	if memStore != nil {
		memStore.StoreUsageAttribution(fmt.Sprintf("%s-%s", platform, userID), model, tokensUsed, durationMs)
	}

	return response, nil
}

// HandleMessageStreamLoop is a stub for stream processing
func (a *Agent) HandleMessageStreamLoop(ctx context.Context, platform, userID, text string, ch chan<- StreamEvent, memStore *memory.Store) {
	start := time.Now()
	ch <- StreamEvent{Type: EventOutcome, Outcome: fmt.Sprintf("Stream response to %s: %s", platform, text)}
	ch <- StreamEvent{Type: EventDone}

	// Simulate usage tracking for streams
	tokensUsed := len(text) + 20 // arbitrary streaming overhead
	durationMs := int(time.Since(start).Milliseconds())
	if memStore != nil {
		memStore.StoreUsageAttribution(fmt.Sprintf("%s-%s", platform, userID), "hermes-stream-model", tokensUsed, durationMs)
	}
}

// APIDispatcher is a minimal API for the Python dashboard to call
func (a *Agent) APIDispatcher(ctx context.Context, action string, args map[string]string, memStore *memory.Store) (string, error) {
	if action == "converse" {
		return a.RunConversationLoop(ctx, args["platform"], args["userID"], args["text"], memStore)
	}
	return "", fmt.Errorf("unknown action: %s", action)
}
