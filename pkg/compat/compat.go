package compat

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ParityLevel tracks how close an implementation is to an upstream tool contract.
type ParityLevel string

const (
	ParityPlanned  ParityLevel = "planned"
	ParityBridged  ParityLevel = "bridged"
	ParitySpeced   ParityLevel = "speced"
	ParityNative   ParityLevel = "native"
	ParityVerified ParityLevel = "verified"
)

// ToolContract describes a model-facing tool surface that must remain stable.
type ToolContract struct {
	Source            string          `json:"source"`
	Name              string          `json:"name"`
	Description       string          `json:"description,omitempty"`
	Parameters        json.RawMessage `json:"parameters,omitempty"`
	ResultFormat      string          `json:"resultFormat"`
	ExactName         bool            `json:"exactName"`
	ExactParameters   bool            `json:"exactParameters"`
	ExactResultShape  bool            `json:"exactResultShape"`
	Status            ParityLevel     `json:"status"`
	ImplementationRef string          `json:"implementationRef,omitempty"`
	Notes             []string        `json:"notes,omitempty"`
}

// Clone returns a safe copy.
func (c ToolContract) Clone() ToolContract {
	out := c
	if c.Parameters != nil {
		out.Parameters = append(json.RawMessage(nil), c.Parameters...)
	}
	if c.Notes != nil {
		out.Notes = append([]string(nil), c.Notes...)
	}
	return out
}

// Catalog stores exact model-facing tool contracts.
type Catalog struct {
	mu       sync.RWMutex
	bySource map[string][]ToolContract
	byName   map[string][]ToolContract
}

// NewCatalog creates an empty tool contract catalog.
func NewCatalog() *Catalog {
	return &Catalog{
		bySource: map[string][]ToolContract{},
		byName:   map[string][]ToolContract{},
	}
}

// Register adds a tool contract to the catalog.
func (c *Catalog) Register(contract ToolContract) error {
	if strings.TrimSpace(contract.Source) == "" {
		return fmt.Errorf("tool contract source is required")
	}
	if strings.TrimSpace(contract.Name) == "" {
		return fmt.Errorf("tool contract name is required")
	}
	clone := contract.Clone()
	clone.Source = strings.TrimSpace(clone.Source)
	clone.Name = strings.TrimSpace(clone.Name)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.bySource[clone.Source] = append(c.bySource[clone.Source], clone)
	c.byName[clone.Name] = append(c.byName[clone.Name], clone)
	return nil
}

// MustRegister panics on error.
func (c *Catalog) MustRegister(contract ToolContract) {
	if err := c.Register(contract); err != nil {
		panic(err)
	}
}

// Count returns total registered contracts.
func (c *Catalog) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	total := 0
	for _, contracts := range c.bySource {
		total += len(contracts)
	}
	return total
}

// Sources returns sorted list of registered source names.
func (c *Catalog) Sources() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sources := make([]string, 0, len(c.bySource))
	for source := range c.bySource {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	return sources
}

// ContractsBySource returns all contracts from a source, sorted by name.
func (c *Catalog) ContractsBySource(source string) []ToolContract {
	c.mu.RLock()
	defer c.mu.RUnlock()
	contracts := c.bySource[strings.TrimSpace(source)]
	out := make([]ToolContract, 0, len(contracts))
	for _, contract := range contracts {
		out = append(out, contract.Clone())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Lookup finds all contracts with the given name across sources.
func (c *Catalog) Lookup(name string) []ToolContract {
	c.mu.RLock()
	defer c.mu.RUnlock()
	contracts := c.byName[strings.TrimSpace(name)]
	out := make([]ToolContract, 0, len(contracts))
	for _, contract := range contracts {
		out = append(out, contract.Clone())
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source == out[j].Source {
			return out[i].Name < out[j].Name
		}
		return out[i].Source < out[j].Source
	})
	return out
}

// DefaultCatalog returns the catalog with built-in tool contracts.
// These provide tool parity across popular agentic harnesses.
func DefaultCatalog() *Catalog {
	c := NewCatalog()

	// Register tools from popular harnesses
	registerCommonTools(c)

	return c
}

func registerCommonTools(c *Catalog) {
	// Claude Code tools
	c.MustRegister(ToolContract{
		Source: "claude-code", Name: "Bash", Description: "Execute shell commands",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "claude-code", Name: "Read", Description: "Read file contents",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "claude-code", Name: "Write", Description: "Write content to a file",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "claude-code", Name: "Edit", Description: "Edit existing file content",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "claude-code", Name: "Glob", Description: "Find files matching a pattern",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "claude-code", Name: "Grep", Description: "Search file contents",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "claude-code", Name: "WebSearch", Description: "Search the web",
		ResultFormat: "json", ExactName: true, Status: ParitySpeced,
	})

	// Codex tools
	c.MustRegister(ToolContract{
		Source: "codex", Name: "execute_bash", Description: "Execute bash command",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "codex", Name: "read_file", Description: "Read file contents",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "codex", Name: "write_file", Description: "Write to a file",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "codex", Name: "search_files", Description: "Search files by pattern",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "codex", Name: "web_search", Description: "Search the web",
		ResultFormat: "json", ExactName: true, Status: ParitySpeced,
	})

	// Warp tools
	c.MustRegister(ToolContract{
		Source: "warp", Name: "warp_execute", Description: "Execute a command in warp terminal",
		ResultFormat: "text", ExactName: true, Status: ParityBridged,
	})
	c.MustRegister(ToolContract{
		Source: "warp", Name: "warp_read_file", Description: "Read a file in warp",
		ResultFormat: "text", ExactName: true, Status: ParityBridged,
	})

	// Aider tools
	c.MustRegister(ToolContract{
		Source: "aider", Name: "aider_commit", Description: "Commit changes via aider",
		ResultFormat: "text", ExactName: true, Status: ParityBridged,
	})
	c.MustRegister(ToolContract{
		Source: "aider", Name: "aider_read_only", Description: "Add file as read-only context",
		ResultFormat: "text", ExactName: true, Status: ParityBridged,
	})

	// pi coding agent tools
	c.MustRegister(ToolContract{
		Source: "pi", Name: "read_file", Description: "Read file contents",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "pi", Name: "write_file", Description: "Write to a file",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "pi", Name: "edit", Description: "Edit file content",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "pi", Name: "glob", Description: "Find files matching a pattern",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "pi", Name: "search_files", Description: "Search in files",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "pi", Name: "execute_command", Description: "Execute a command",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "pi", Name: "web_search", Description: "Search the web",
		ResultFormat: "json", ExactName: true, Status: ParitySpeced,
	})

	// Hermes Agent tools (self-registration)
	c.MustRegister(ToolContract{
		Source: "hermes", Name: "terminal", Description: "Execute terminal commands",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "hermes", Name: "read_file", Description: "Read file contents",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "hermes", Name: "write_file", Description: "Write content to a file",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "hermes", Name: "patch", Description: "Apply a patch to a file",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "hermes", Name: "search_files", Description: "Search files by pattern",
		ResultFormat: "text", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "hermes", Name: "web_search", Description: "Search the web",
		ResultFormat: "json", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "hermes", Name: "delegate_task", Description: "Delegate work to a sub-agent",
		ResultFormat: "json", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "hermes", Name: "memory_search", Description: "Search agent memory",
		ResultFormat: "json", ExactName: true, Status: ParityNative,
	})
	c.MustRegister(ToolContract{
		Source: "hermes", Name: "memory_store", Description: "Store agent memory",
		ResultFormat: "json", ExactName: true, Status: ParityNative,
	})
}

// LookupEquivalent finds tools across harnesses with the same semantic purpose.
func (c *Catalog) LookupEquivalent(source, toolName string) []ToolContract {
	all := c.Lookup(toolName)
	results := make([]ToolContract, 0)
	for _, contract := range all {
		if contract.Source != source {
			results = append(results, contract)
		}
	}
	return results
}

// Summary returns a human-readable summary of the catalog.
func (c *Catalog) Summary() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var out strings.Builder
	out.WriteString("Tool Parity Catalog\n")
	out.WriteString("===================\n\n")

	sources := c.Sources()
	for _, src := range sources {
		contracts := c.ContractsBySource(src)
		out.WriteString(fmt.Sprintf("Source: %s (%d tools)\n", src, len(contracts)))
		for _, ct := range contracts {
			status := string(ct.Status)
			nameLabel := ct.Name
			if ct.ExactName {
				nameLabel += " (exact)"
			}
			out.WriteString(fmt.Sprintf("  %-25s [%s]\n", nameLabel, status))
		}
		out.WriteString("\n")
	}
	out.WriteString(fmt.Sprintf("\nTotal: %d tools across %d harnesses\n", c.Count(), len(sources)))
	return out.String()
}
