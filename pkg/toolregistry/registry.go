package toolregistry

import (
	"fmt"
	"sort"
	"sync"
)

// Tool is a registered tool that can be called by the agent.
type Tool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Parameters  map[string]any   `json:"parameters"`
	Handler     func(args map[string]any, context map[string]any) (any, error) `json:"-"`
	Native      bool              `json:"native"`
}

// Registry holds all registered tools.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]*Tool
}

var (
	globalRegistry *Registry
	once           sync.Once
)

// Global returns the global tool registry (singleton).
func Global() *Registry {
	once.Do(func() {
		globalRegistry = New()
	})
	return globalRegistry
}

// New creates a new tool registry.
func New() *Registry {
	return &Registry{
		tools: make(map[string]*Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool *Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tool == nil {
		return fmt.Errorf("cannot register nil tool")
	}
	if tool.Name == "" {
		return fmt.Errorf("tool must have a name")
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool %s must have a handler", tool.Name)
	}

	r.tools[tool.Name] = tool
	return nil
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (*Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools.
func (r *Registry) List() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Call invokes a tool by name with the given arguments and context.
func (r *Registry) Call(name string, args map[string]any, context map[string]any) (any, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Handler(args, context)
}

// Remove removes a tool from the registry.
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
}

// Count returns the number of registered tools.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}