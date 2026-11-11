package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestGatewayStatusHandler(t *testing.T) {
	ag := agent.New(agent.DefaultConfig())
	gw := New(ag)
	gw.RegisterPlatform(&mockPlatform{name: "discord"})

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	gw.StatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Status    string   `json:"status"`
		Platforms []string `json:"platforms"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse status json: %v", err)
	}

	if resp.Status != "running" {
		t.Errorf("expected status 'running', got %s", resp.Status)
	}
	if len(resp.Platforms) != 2 || resp.Platforms[1] != "discord" {
		t.Errorf("expected platforms ['cli', 'discord'], got %v", resp.Platforms)
	}
}
