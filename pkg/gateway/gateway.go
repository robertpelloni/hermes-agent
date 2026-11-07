package gateway

import (
	"context"
	"fmt"

	"github.com/robertpelloni/hermes-agent/pkg/agent"
)

// Gateway provides multi-platform connectivity (Telegram, Discord, CLI, etc.)
type Gateway struct {
	agent    *agent.Agent
	platforms []Platform
}

// Platform represents a communication platform.
type Platform interface {
	Name() string
	Start(ctx context.Context) error
	Stop()
}

// CLIPlatform is the built-in terminal interface.
type CLIPlatform struct{}

func (c *CLIPlatform) Name() string { return "cli" }
func (c *CLIPlatform) Start(ctx context.Context) error {
	fmt.Println("[hermes:gateway] CLI platform ready")
	return nil
}
func (c *CLIPlatform) Stop() {}

// New creates a new gateway.
func New(ag *agent.Agent) *Gateway {
	return &Gateway{
		agent:     ag,
		platforms: []Platform{&CLIPlatform{}},
	}
}

// RegisterPlatform dynamically registers a communication platform.
func (g *Gateway) RegisterPlatform(p Platform) {
	g.platforms = append(g.platforms, p)
	fmt.Printf("[hermes:gateway] Registered platform: %s\n", p.Name())
}

// Platforms returns the names of all registered communication platforms.
func (g *Gateway) Platforms() []string {
	names := make([]string, 0, len(g.platforms))
	for _, p := range g.platforms {
		names = append(names, p.Name())
	}
	return names
}

// Start launches all platform listeners.
func (g *Gateway) Start(ctx context.Context) error {
	for _, p := range g.platforms {
		fmt.Printf("[hermes:gateway] Starting platform: %s\n", p.Name())
		if err := p.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Stop shuts down all platforms.
func (g *Gateway) Stop() {
	for _, p := range g.platforms {
		p.Stop()
	}
}
