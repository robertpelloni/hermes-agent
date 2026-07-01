package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"time"

	"github.com/robertpelloni/hermes-agent/pkg/ai"

	"github.com/robertpelloni/hermes-agent/pkg/memory"
	"github.com/robertpelloni/hermes-agent/pkg/prompttemplates"

)

// Agent is the core Hermes AI agent.
type Agent struct {
	config   Config
	running  bool
	store    *memory.Store
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
		Model:         getEnvDefault("HERMES_MODEL", "free-llm"),
		Provider:      getEnvDefault("HERMES_PROVIDER", "local-llm"),
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
func New(cfg Config, store *memory.Store) *Agent {
	return &Agent{config: cfg, running: false, store: store}
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
	return a.runConversation(ctx, platform, userID, text)
}

// HandleMessageStream processes a user message and streams events on the returned channel.
// The caller must read from the channel until it is closed.
func (a *Agent) HandleMessageStream(ctx context.Context, platform, userID, text string) <-chan StreamEvent {
	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		a.handleMessageStreamImpl(ctx, platform, userID, text, ch)
	}()
	return ch
}

func (a *Agent) loadHistory(sessionID string) []ai.Message {
	if a.store == nil {
		return []ai.Message{}
	}

	msgs, err := a.store.LoadMessages(sessionID)
	if err != nil || len(msgs) == 0 {
		return []ai.Message{}
	}

	var aiMsgs []ai.Message
	for _, m := range msgs {
		var content []ai.Content
		if err := json.Unmarshal([]byte(m.Content), &content); err != nil {
			content = []ai.Content{{Type: "text", Text: m.Content}}
		}

		role := ai.MessageRole(m.Role)
		aiMsgs = append(aiMsgs, ai.Message{
			Role:    role,
			Content: content,
		})
	}

	return aiMsgs
}

func (a *Agent) saveHistory(sessionID string, msg ai.Message) {
	if a.store == nil {
		return
	}

	data, err := json.Marshal(msg.Content)
	if err != nil {
		return
	}

	_ = a.store.SaveMessage(sessionID, string(msg.Role), string(data))
}

