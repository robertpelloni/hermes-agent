package agent

import (
	"context"
	"fmt"
)

// RunConversationLoop manages a single conversation turn
func (a *Agent) RunConversationLoop(ctx context.Context, platform, userID, text string) (string, error) {
	fmt.Printf("[hermes] processing msg from %s/%s: %s\n", platform, userID, text)
	return fmt.Sprintf("I received your message on %s: %s", platform, text), nil
}

// HandleMessageStreamLoop is a stub for stream processing
func (a *Agent) HandleMessageStreamLoop(ctx context.Context, platform, userID, text string, ch chan<- StreamEvent) {
	ch <- StreamEvent{Type: EventOutcome, Outcome: fmt.Sprintf("Stream response to %s: %s", platform, text)}
	ch <- StreamEvent{Type: EventDone}
}
package agent

import (
	"context"
	"fmt"
)

// RunConversationLoop manages a single conversation turn
func (a *Agent) RunConversationLoop(ctx context.Context, platform, userID, text string) (string, error) {
	fmt.Printf("[hermes] processing msg from %s/%s: %s\n", platform, userID, text)
	return fmt.Sprintf("I received your message on %s: %s", platform, text), nil
}

// HandleMessageStreamLoop is a stub for stream processing
func (a *Agent) HandleMessageStreamLoop(ctx context.Context, platform, userID, text string, ch chan<- StreamEvent) {
	ch <- StreamEvent{Type: EventOutcome, Outcome: fmt.Sprintf("Stream response to %s: %s", platform, text)}
	ch <- StreamEvent{Type: EventDone}
}
