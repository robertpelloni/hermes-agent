package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectProvider(t *testing.T) {
	// Test default / coding
	route, err := SelectProvider(ProviderRouteRequest{TaskType: "coding"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route.Provider == "" || route.Model == "" {
		t.Errorf("expected non-empty provider and model, got %+v", route)
	}

	// Test budget preference
	budgetRoute, err := SelectProvider(ProviderRouteRequest{TaskType: "coding", CostPreference: "budget"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if budgetRoute.Cost != "budget" {
		t.Errorf("expected budget cost, got %s", budgetRoute.Cost)
	}

	// Test require local
	localRoute, err := SelectProvider(ProviderRouteRequest{RequireLocal: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if localRoute.Provider != "local" || localRoute.Cost != "free" {
		t.Errorf("expected local free route, got %+v", localRoute)
	}
}

func TestPrepareExecutionAndFormatting(t *testing.T) {
	route := ProviderRoute{
		Provider: "anthropic",
		Model:    "claude-3-7-sonnet",
		Cost:     "quality",
		Reason:   "advanced reasoning",
	}
	execMap := PrepareExecution("Write a go function", route)
	if execMap["provider"] != "anthropic" || execMap["prompt"] != "Write a go function" {
		t.Errorf("unexpected execution map: %+v", execMap)
	}

	formatted := FormatProviderInfo()
	if !strings.Contains(formatted, "PROVIDER") || !strings.Contains(formatted, "anthropic") {
		t.Errorf("unexpected formatted output: %s", formatted)
	}
}

func TestDiscoverMCPConfigs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mcpJSON := `{"name": "filesystem", "command": "npx", "args": ["-y", "@modelcontextprotocol/server-filesystem"]}`
	if err := os.WriteFile(filepath.Join(tempDir, ".mcp.json"), []byte(mcpJSON), 0644); err != nil {
		t.Fatalf("failed to write .mcp.json: %v", err)
	}

	configs, err := DiscoverMCPConfigs(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Name != "filesystem" || !configs[0].Enabled {
		t.Errorf("unexpected config: %+v", configs[0])
	}
}
