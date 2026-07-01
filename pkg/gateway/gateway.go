package gateway

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/robertpelloni/hermes-agent/pkg/agent"
)

// Gateway provides multi-platform connectivity (Telegram, Discord, CLI, etc.)
type Gateway struct {
	agent     *agent.Agent
	platforms []Platform
}

// Platform represents a communication platform.
type Platform interface {
	Name() string
	Start(ctx context.Context, ag *agent.Agent) error
	Stop()
}

// CLIPlatform is the built-in terminal interface.
type CLIPlatform struct {
	stopCh chan struct{}
}

func (c *CLIPlatform) Name() string { return "cli" }

func (c *CLIPlatform) Start(ctx context.Context, ag *agent.Agent) error {
	c.stopCh = make(chan struct{})
	fmt.Println("[hermes:gateway] CLI platform ready. Type '/quit' or '/exit' to stop.")

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.stopCh:
				return
			default:
				fmt.Print("\n> ")
				if !scanner.Scan() {
					return
				}
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				if line == "/quit" || line == "/exit" {
					fmt.Println("exiting CLI platform")
					// Send a signal to shut down or just return
					return
				}
				resp, err := ag.HandleMessage(ctx, "cli", "local", line)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Println(resp)
				}
			}
		}
	}()

	return nil
}

func (c *CLIPlatform) Stop() {
	if c.stopCh != nil {
		close(c.stopCh)
	}
}

// New creates a new gateway.
func New(ag *agent.Agent) *Gateway {
	return &Gateway{
		agent:     ag,
		platforms: []Platform{&CLIPlatform{}},
	}
}

// Start launches all platform listeners.
func (g *Gateway) Start(ctx context.Context) error {
	for _, p := range g.platforms {
		fmt.Printf("[hermes:gateway] Starting platform: %s\n", p.Name())
		if err := p.Start(ctx, g.agent); err != nil {
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
