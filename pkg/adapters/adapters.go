package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ProviderRoute describes a provider selection with cost/preference awareness.
type ProviderRoute struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Reason   string `json:"reason"`
	Cost     string `json:"cost"`
}

// ProviderRouteRequest holds selection criteria.
type ProviderRouteRequest struct {
	TaskType       string `json:"taskType"`
	CostPreference string `json:"costPreference"`
	RequireLocal   bool   `json:"requireLocal"`
}

// MCPAdapter bridges Model Context Protocol servers.
type MCPAdapter struct {
	rootDir string
	servers map[string]MCPConfig
	mu      sync.RWMutex
}

// MCPConfig describes an MCP server configuration.
type MCPConfig struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	Args     []string `json:"args,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// HyperCodeAdapter bridges HyperCode/Borg workflows.
type HyperCodeAdapter struct {
	rootDir string
	running bool
}

// NewMCPAdapter creates a new MCP adapter.
func NewMCPAdapter(rootDir string) *MCPAdapter {
	return &MCPAdapter{
		rootDir: rootDir,
		servers: make(map[string]MCPConfig),
	}
}

// status returns a status payload for the MCP adapter (for diagnostics).
func (a *MCPAdapter) status() map[string]any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	enabled := 0
	for _, s := range a.servers {
		if s.Enabled {
			enabled++
		}
	}
	return map[string]any{
		"servers":      len(a.servers),
		"enabled":      enabled,
		"rootDir":      a.rootDir,
		"hasConfig":    a.loadConfig() != nil,
	}
}

// loadConfig attempts to read MCP server configuration from the project.
func (a *MCPAdapter) loadConfig() *os.File {
	cfgPath := filepath.Join(a.rootDir, ".mcp.json")
	f, err := os.Open(cfgPath)
	if err != nil {
		return nil
	}
	return f
}

// NewHyperCodeAdapter creates a new HyperCode adapter.
func NewHyperCodeAdapter(rootDir string) *HyperCodeAdapter {
	return &HyperCodeAdapter{rootDir: rootDir}
}

// status returns diagnostic info about HyperCode integration.
func (a *HyperCodeAdapter) status() map[string]any {
	return map[string]any{
		"rootDir": a.rootDir,
		"running": a.running,
	}
}

// Status returns full diagnostic info.
func (a *MCPAdapter) Status() map[string]any    { return a.status() }
func (a *HyperCodeAdapter) Status() map[string]any { return a.status() }

// DefaultProviderRoutes returns built-in provider routes for common tasks.
func DefaultProviderRoutes() map[string][]ProviderRoute {
	return map[string][]ProviderRoute{
		"coding": {
			{Provider: "openai", Model: "gpt-4o-mini", Reason: "fast and cost-effective for code", Cost: "budget"},
			{Provider: "anthropic", Model: "claude-sonnet-4-20250514", Reason: "strong coding capabilities", Cost: "quality"},
			{Provider: "openrouter", Model: "openai/gpt-4o-mini", Reason: "routed through openrouter", Cost: "budget"},
		},
		"analysis": {
			{Provider: "anthropic", Model: "claude-sonnet-4-20250514", Reason: "strong reasoning", Cost: "quality"},
			{Provider: "openai", Model: "gpt-5-mini", Reason: "strong reasoning with thinking", Cost: "quality"},
			{Provider: "google", Model: "gemini-2.5-flash", Reason: "fast analysis via gemini", Cost: "budget"},
		},
		"chat": {
			{Provider: "openai", Model: "gpt-4o-mini", Reason: "fast and cheap for conversation", Cost: "budget"},
			{Provider: "openrouter", Model: "openai/gpt-4o-mini", Reason: "cheap routing", Cost: "budget"},
		},
	}
}

// SelectProvider selects the best provider route based on request criteria.
func SelectProvider(req ProviderRouteRequest) (ProviderRoute, error) {
	routes := DefaultProviderRoutes()
	taskType := req.TaskType
	if taskType == "" {
		taskType = "coding"
	}
	candidates, ok := routes[taskType]
	if !ok {
		candidates = routes["coding"]
	}
	if req.RequireLocal {
		return ProviderRoute{Provider: "local", Model: "ollama", Reason: "local execution required", Cost: "free"}, nil
	}
	if req.CostPreference == "budget" {
		for _, r := range candidates {
			if r.Cost == "budget" {
				return r, nil
			}
		}
	}
	return candidates[0], nil
}

// PrepareExecution creates an execution plan for a given prompt and route.
func PrepareExecution(prompt string, route ProviderRoute) map[string]any {
	return map[string]any{
		"provider":    route.Provider,
		"model":       route.Model,
		"prompt":      prompt,
		"maxTokens":   4096,
		"temperature": 0.7,
	}
}

// DiscoverMCPConfigs scans common paths for MCP server configurations.
func DiscoverMCPConfigs(rootDir string) ([]MCPConfig, error) {
	var configs []MCPConfig
	candidates := []string{
		filepath.Join(rootDir, ".mcp.json"),
		filepath.Join(rootDir, "mcp.json"),
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg MCPConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		cfg.Enabled = true
		configs = append(configs, cfg)
	}
	if configs == nil {
		configs = []MCPConfig{}
	}
	return configs, nil
}

// FormatProviderInfo returns a human-readable summary of available providers.
func FormatProviderInfo() string {
	out := fmt.Sprintf("%-20s %-20s %-10s %s\n", "PROVIDER", "MODEL", "COST", "USE CASE")
	out += fmt.Sprintf("%-20s %-20s %-10s %s\n", "--------", "-----", "----", "--------")
	for task, routes := range DefaultProviderRoutes() {
		for _, r := range routes {
			out += fmt.Sprintf("%-20s %-20s %-10s %s\n", r.Provider, r.Model, r.Cost, task)
		}
	}
	return out
}
