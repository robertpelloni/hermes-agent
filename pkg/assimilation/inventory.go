package assimilation

import (
	"fmt"
	"sort"
	"strings"
)

// SourceToolchain describes an upstream system targeted for native assimilation.
type SourceToolchain struct {
	ID               string   `json:"id"`
	DisplayName      string   `json:"displayName"`
	Language         string   `json:"language"`
	Category         string   `json:"category"`
	UpstreamPath     string   `json:"upstreamPath"`
	Strengths        []string `json:"strengths"`
	Noteworthy       []string `json:"noteworthy"`
	AssimilationPlan []string `json:"assimilationPlan"`
}

// Inventory returns the complete list of source toolchains to assimilate.
func Inventory() []SourceToolchain {
	items := []SourceToolchain{
		{
			ID:           "hypercode",
			DisplayName:  "HyperCode",
			Language:     "TypeScript + Go",
			Category:     "control-plane",
			UpstreamPath: "../hypercode",
			Strengths:    []string{"MCP aggregation", "provider routing", "memory and session substrate", "operator observability"},
			Noteworthy:   []string{"Best existing fit for Borg-native provider routing and MCP management.", "Should remain the substrate instead of being reimplemented in the new harness."},
			AssimilationPlan: []string{"Consume as the default local control plane.", "Expose HyperCode routes and tool inventories as native Go adapters.", "Let the new harness own the agent UX and exact tool contracts while HyperCode owns orchestration truth."},
		},
		{
			ID:           "pi-mono",
			DisplayName:  "Pi Mono",
			Language:     "TypeScript + Go",
			Category:     "foundation",
			UpstreamPath: "../pi-mono",
			Strengths:    []string{"minimal harness", "excellent extension seams", "interactive/json/rpc modes", "tool-first design", "multi-provider LLM client"},
			Noteworthy:   []string{"The cleanest conceptual base for a Go rewrite.", "Good default philosophy: small core, user-extensible edges.", "Comprehensive pi-ai TypeScript/Go multi-provider client."},
			AssimilationPlan: []string{"Port agent runtime contracts first.", "Mirror session model, settings, command surface, and extension seams.", "Retain exact default tool names and event vocabulary.", "Port pi-ai multi-provider client patterns."},
		},
		{
			ID:           "aider",
			DisplayName:  "Aider",
			Language:     "Python",
			Category:     "coding-agent",
			UpstreamPath: "../hyperharness/aider",
			Strengths:    []string{"repo map", "edit strategies", "git-aware workflow"},
			Noteworthy:   []string{"Best-in-class context condensation for large repos.", "Multiple edit coders are a proven pattern worth preserving."},
			AssimilationPlan: []string{"Port repomap and edit strategies as pluggable Go context engines.", "Keep git-native UX and commit-oriented workflows."},
		},
		{
			ID:           "claude-code",
			DisplayName:  "Claude Code",
			Language:     "TypeScript",
			Category:     "coding-agent",
			UpstreamPath: "../hyperharness/claude-code",
			Strengths:    []string{"popular agent ergonomics", "tool-based coding workflow"},
			Noteworthy:   []string{"Use only as behavior reference; keep implementation clean-room and lawful."},
			AssimilationPlan: []string{"Observe public UX patterns and command conventions only."},
		},
		{
			ID:           "copilot-cli",
			DisplayName:  "GitHub Copilot CLI",
			Language:     "TypeScript",
			Category:     "coding-agent",
			UpstreamPath: "../hyperharness/copilot-cli",
			Strengths:    []string{"official subscription integration", "shell ergonomics"},
			Noteworthy:   []string{"Important for OAuth/subscription bridging and enterprise familiarity."},
			AssimilationPlan: []string{"Bridge auth and import session history, emulate stable shell-oriented affordances."},
		},
		{
			ID:           "gemini-cli",
			DisplayName:  "Gemini CLI",
			Language:     "TypeScript",
			Category:     "provider",
			UpstreamPath: "../hyperharness/gemini-cli",
			Strengths:    []string{"Google auth integration", "strong coding models"},
			Noteworthy:   []string{"Important subscription/provider compatibility target."},
			AssimilationPlan: []string{"Bridge auth/session import via HyperCode; mirror public model/provider affordances."},
		},
		{
			ID:           "grok-cli",
			DisplayName:  "Grok CLI",
			Language:     "TypeScript",
			Category:     "agent-core",
			UpstreamPath: "../hyperharness/grok-cli",
			Strengths:    []string{"sub-agents", "batch mode", "remote control", "verify"},
			Noteworthy:   []string{"Unique remote-control and multi-agent delegation model."},
			AssimilationPlan: []string{"Assimilate delegation and remote-control abstractions behind native Go command and session services."},
		},
		{
			ID:           "kilocode",
			DisplayName:  "Kilo Code",
			Language:     "TypeScript",
			Category:     "coding-agent",
			UpstreamPath: "../hyperharness/kilocode",
			Strengths:    []string{"high-volume autonomous engineering", "popular open source coding agent"},
			Noteworthy:   []string{"Good benchmark for task decomposition and throughput."},
			AssimilationPlan: []string{"Mine orchestration and model defaults for autonomous modes."},
		},
		{
			ID:           "litellm",
			DisplayName:  "LiteLLM",
			Language:     "Python",
			Category:     "provider-router",
			UpstreamPath: "../hyperharness/litellm",
			Strengths:    []string{"provider normalization", "OpenAI-compatible proxy", "routing"},
			Noteworthy:   []string{"HyperCode already overlaps heavily here."},
			AssimilationPlan: []string{"Use HyperCode routing rather than reimplement; assimilate only missing provider-normalization details."},
		},
		{
			ID:           "ollama",
			DisplayName:  "Ollama",
			Language:     "Go",
			Category:     "runtime",
			UpstreamPath: "../hyperharness/ollama",
			Strengths:    []string{"local model serving", "single binary distribution", "provider API"},
			Noteworthy:   []string{"Natural local runtime complement for a Go-native harness."},
			AssimilationPlan: []string{"Integrate as first-class local provider target through HyperCode/provider abstractions."},
		},
		{
			ID:           "open-interpreter",
			DisplayName:  "Open Interpreter",
			Language:     "Python",
			Category:     "computer-use",
			UpstreamPath: "../hyperharness/open-interpreter",
			Strengths:    []string{"computer control", "desktop automation", "OS integration"},
			Noteworthy:   []string{"Best reference for computer-use and local action layers."},
			AssimilationPlan: []string{"Assimilate computer-use capabilities behind explicit trust boundaries and sandbox policies."},
		},
		{
			ID:           "opencode",
			DisplayName:  "OpenCode",
			Language:     "TypeScript/Bun",
			Category:     "agent-platform",
			UpstreamPath: "../hyperharness/opencode",
			Strengths:    []string{"rich TUI", "client/server split", "provider agnosticism"},
			Noteworthy:   []string{"Primary competitive benchmark for UX, extensibility, and reach."},
			AssimilationPlan: []string{"Beat on startup speed, context accuracy, and trust while matching core UX expectations."},
		},
		{
			ID:           "crush",
			DisplayName:  "Crush",
			Language:     "Go",
			Category:     "tui",
			UpstreamPath: "../hyperharness/crush",
			Strengths:    []string{"Charm-powered UX", "fluid terminal interaction"},
			Noteworthy:   []string{"Best aesthetic benchmark for a Go-native TUI."},
			AssimilationPlan: []string{"Borrow TUI patterns and interaction density, not wholesale code."},
		},
		{
			ID:           "goose",
			DisplayName:  "Goose",
			Language:     "Rust",
			Category:     "agent-core",
			UpstreamPath: "../hyperharness/goose",
			Strengths:    []string{"clean architecture", "protocol orientation", "performance"},
			Noteworthy:   []string{"One of the best references for a serious agent core, despite different language choice."},
			AssimilationPlan: []string{"Adopt its protocol and layering discipline in Go."},
		},
		{
			ID:           "factory",
			DisplayName:  "Factory Droid",
			Language:     "TypeScript",
			Category:     "orchestrator",
			UpstreamPath: "../hyperharness/factory-cli",
			Strengths:    []string{"persistent autonomous workflows", "verification", "multi-surface sessions"},
			Noteworthy:   []string{"Strongest reference for long-running background work and verify loops."},
			AssimilationPlan: []string{"Assimilate persistence, background sessions, and verification workflows into the Go orchestration layer."},
		},
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

// FormatInventory returns a human-readable summary of the assimilation inventory.
func FormatInventory() string {
	items := Inventory()
	var out strings.Builder
	out.WriteString("Assimilation Inventory\n")
	out.WriteString("======================\n\n")
	out.WriteString(fmt.Sprintf("Total systems: %d\n\n", len(items)))

	byCategory := make(map[string][]SourceToolchain)
	for _, item := range items {
		byCategory[item.Category] = append(byCategory[item.Category], item)
	}

	categories := []string{"control-plane", "foundation", "coding-agent", "provider", "agent-core", "runtime", "computer-use", "agent-platform", "tui", "orchestrator"}
	for _, cat := range categories {
		items, ok := byCategory[cat]
		if !ok || len(items) == 0 {
			continue
		}
		out.WriteString(fmt.Sprintf("### %s (%d systems)\n", strings.Title(cat), len(items)))
		for _, item := range items {
			out.WriteString(fmt.Sprintf("  %-20s %-25s %s\n", item.ID, item.DisplayName, item.Language))
		}
		out.WriteString("\n")
	}

	return out.String()
}