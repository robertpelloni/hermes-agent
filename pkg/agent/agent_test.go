package agent_test

import (
	"os"
	"testing"

	"github.com/robertpelloni/hermes-agent/pkg/agent"
)

func TestDefaultConfigEnvVars(t *testing.T) {
	// Backup original env vars
	origModel := os.Getenv("HERMES_MODEL")
	origProvider := os.Getenv("HERMES_PROVIDER")

	defer func() {
		os.Setenv("HERMES_MODEL", origModel)
		os.Setenv("HERMES_PROVIDER", origProvider)
	}()

	// Test default values
	os.Setenv("HERMES_MODEL", "")
	os.Setenv("HERMES_PROVIDER", "")

	cfg := agent.DefaultConfig()
	if cfg.Model != "free-llm" {
		t.Errorf("expected default model 'free-llm', got %q", cfg.Model)
	}
	if cfg.Provider != "local-llm" {
		t.Errorf("expected default provider 'local-llm', got %q", cfg.Provider)
	}

	// Test custom values
	os.Setenv("HERMES_MODEL", "gpt-4o")
	os.Setenv("HERMES_PROVIDER", "openai")

	cfg2 := agent.DefaultConfig()
	if cfg2.Model != "gpt-4o" {
		t.Errorf("expected custom model 'gpt-4o', got %q", cfg2.Model)
	}
	if cfg2.Provider != "openai" {
		t.Errorf("expected custom provider 'openai', got %q", cfg2.Provider)
	}
}
