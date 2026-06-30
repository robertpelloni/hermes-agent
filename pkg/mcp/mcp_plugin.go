package mcp

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// Plugin represents a loaded MCP plugin wrapper.
type Plugin struct {
	Name    string
	Command *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
}

// PluginManager manages multiple MCP plugins.
type PluginManager struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*Plugin),
	}
}

// Load loads an external MCP server plugin over standard IO.
func (pm *PluginManager) Load(name, command string, args ...string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	plugin := &Plugin{
		Name:    name,
		Command: cmd,
		Stdin:   stdin,
		Stdout:  stdout,
	}

	pm.plugins[name] = plugin
	fmt.Printf("[hermes:mcp] Loaded plugin: %s\n", name)

	return nil
}

// Unload unloads an MCP plugin.
func (pm *PluginManager) Unload(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, ok := pm.plugins[name]
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if err := plugin.Command.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill plugin process: %w", err)
	}

	delete(pm.plugins, name)
	fmt.Printf("[hermes:mcp] Unloaded plugin: %s\n", name)

	return nil
}
