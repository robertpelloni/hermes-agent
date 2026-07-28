package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

// Plugin represents a loaded plugin with lifecycle hooks.
type Plugin struct {
	Name        string
	Version     string
	Description string
	Directory   string
	OnLoad      func() error
	OnUnload    func()
	Tools       []*toolregistry.Tool
}

// Manager discovers, loads, and manages plugins.
type Manager struct {
	mu          sync.RWMutex
	plugins     map[string]*Plugin
	pluginDirs  []string
}

// NewManager creates a plugin manager with search directories.
func NewManager(dirs ...string) *Manager {
	if len(dirs) == 0 {
		home, _ := os.UserHomeDir()
		dirs = []string{
			filepath.Join(home, ".hermes", "plugins"),
			filepath.Join(home, ".hermes", "go-plugins"),
		}
	}

	m := &Manager{
		plugins:    make(map[string]*Plugin),
		pluginDirs: dirs,
	}

	// Ensure directories exist
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}

	return m
}

// Discover scans plugin directories and loads any found plugins.
func (m *Manager) Discover() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	discovered := make([]string, 0)

	for _, dir := range m.pluginDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()

			// Skip non-plugin directories/files
			if entry.IsDir() {
				// Directory-based plugin (contains plugin.yaml or similar)
				pluginPath := filepath.Join(dir, name, "plugin.yaml")
				if _, err := os.Stat(pluginPath); err == nil {
					p := &Plugin{
						Name:      name,
						Directory: filepath.Join(dir, name),
					}
					m.plugins[name] = p
					discovered = append(discovered, name)
				}
			} else if filepath.Ext(name) == ".so" {
				// Go plugin shared object
				pluginName := name[:len(name)-3]
				p := &Plugin{
					Name:      pluginName,
					Directory: filepath.Join(dir, name),
				}
				m.plugins[pluginName] = p
				discovered = append(discovered, pluginName)
			}
		}
	}

	return discovered, nil
}

// Register adds a plugin programmatically.
func (m *Manager) Register(p *Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p.Name == "" {
		return fmt.Errorf("plugin must have a name")
	}

	m.plugins[p.Name] = p

	// Register tools if provided
	for _, tool := range p.Tools {
		if err := toolregistry.Global().Register(tool); err != nil {
			return fmt.Errorf("plugin %s: failed to register tool %s: %w", p.Name, tool.Name, err)
		}
	}

	// Call OnLoad hook
	if p.OnLoad != nil {
		if err := p.OnLoad(); err != nil {
			return fmt.Errorf("plugin %s: OnLoad failed: %w", p.Name, err)
		}
	}

	return nil
}

// Get returns a plugin by name.
func (m *Manager) Get(name string) (*Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[name]
	return p, ok
}

// List returns all managed plugins.
func (m *Manager) List() []*Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		result = append(result, p)
	}
	return result
}

// Unload unloads a plugin and calls its OnUnload hook.
func (m *Manager) Unload(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.plugins[name]
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if p.OnUnload != nil {
		p.OnUnload()
	}

	// Remove plugin tools
	for _, tool := range p.Tools {
		toolregistry.Global().Remove(tool.Name)
	}

	delete(m.plugins, name)
	return nil
}

// FormatStatus returns a human-readable status of all plugins.
func (m *Manager) FormatStatus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := fmt.Sprintf("Plugins (%d total):\n", len(m.plugins))
	result += fmt.Sprintf("%-25s %-15s %s\n", "NAME", "VERSION", "TOOLS")
	result += fmt.Sprintf("%-25s %-15s %s\n", "----", "-------", "-----")

	for _, p := range m.plugins {
		toolCount := len(p.Tools)
		version := p.Version
		if version == "" {
			version = "0.0.0"
		}
		result += fmt.Sprintf("%-25s %-15s %d\n", p.Name, version, toolCount)
	}

	return result
}
