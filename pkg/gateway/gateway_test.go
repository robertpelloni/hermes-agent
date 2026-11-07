package gateway

import (
	"context"
	"testing"

	"github.com/robertpelloni/hermes-agent/pkg/agent"
)

type mockPlatform struct {
	name    string
	started bool
	stopped bool
}

func (m *mockPlatform) Name() string { return m.name }
func (m *mockPlatform) Start(ctx context.Context) error {
	m.started = true
	return nil
}
func (m *mockPlatform) Stop() {
	m.stopped = true
}

func TestGatewayLifecycle(t *testing.T) {
	ag := agent.New(agent.DefaultConfig())
	gw := New(ag)

	// Built-in should include "cli"
	platforms := gw.Platforms()
	if len(platforms) != 1 || platforms[0] != "cli" {
		t.Fatalf("expected initial platform ['cli'], got %v", platforms)
	}

	// Register custom mock platform
	mock := &mockPlatform{name: "telegram"}
	gw.RegisterPlatform(mock)

	platforms = gw.Platforms()
	if len(platforms) != 2 || platforms[1] != "telegram" {
		t.Fatalf("expected platforms ['cli', 'telegram'], got %v", platforms)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := gw.Start(ctx); err != nil {
		t.Fatalf("gw.Start failed: %v", err)
	}

	if !mock.started {
		t.Errorf("expected mock platform to be started")
	}

	gw.Stop()
	if !mock.stopped {
		t.Errorf("expected mock platform to be stopped")
	}
}
