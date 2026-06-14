package agent

import (
	"context"
	"fmt"

	"github.com/robertpelloni/hermes-agent/pkg/mcp"
	"github.com/robertpelloni/hermes-agent/pkg/memory"
	"github.com/robertpelloni/hermes-agent/pkg/scheduler"
	"github.com/robertpelloni/hermes-agent/pkg/skill"
)

// Agent is the core Hermes AI agent.
type Agent struct {
	config   Config
	running  bool
}

// Config holds agent configuration.
type Config struct {
	Model     string
	Provider  string
	Memory    *memory.Store
	Skills    *skill.Repository
	MCPServer *mcp.Server
	Scheduler *scheduler.Scheduler
}

// New creates a new Hermes Agent.
func New(cfg Config) *Agent {
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-20250514"
	}
	if cfg.Provider == "" {
		cfg.Provider = "openrouter"
	}
	return &Agent{config: cfg, running: false}
}

// Run starts the agent's main loop.
func (a *Agent) Run(ctx context.Context) error {
	fmt.Printf("[hermes] Agent running: model=%s provider=%s\n", a.config.Model, a.config.Provider)
	a.running = true
	return nil
}

// HandleMessage processes an incoming user message.
func (a *Agent) HandleMessage(ctx context.Context, platform, userID, text string) (string, error) {
	fmt.Printf("[hermes] Message from %s/%s: %s\n", platform, userID, text)
	// TODO: Implement full agent loop (context compression, tool calling, skill lookup)
	return fmt.Sprintf("[hermes-go stub] Received: %s", text), nil
}
