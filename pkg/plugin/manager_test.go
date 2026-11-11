package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

func TestPluginManager_Lifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hermes-plugin-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mgr := NewManager(tempDir)

	// Create a mock plugin directory with plugin.yaml
	mockPluginDir := filepath.Join(tempDir, "sample-plugin")
	_ = os.MkdirAll(mockPluginDir, 0755)
	_ = os.WriteFile(filepath.Join(mockPluginDir, "plugin.yaml"), []byte("name: sample-plugin\nversion: 1.0.0\n"), 0644)

	discovered, err := mgr.Discover()
	if err != nil {
		t.Fatalf("unexpected discovery error: %v", err)
	}
	if len(discovered) != 1 || discovered[0] != "sample-plugin" {
		t.Fatalf("expected to discover sample-plugin, got: %v", discovered)
	}

	// Register a plugin programmatically with a tool and lifecycle hooks
	loaded := false
	unloaded := false
	testTool := &toolregistry.Tool{
		Name:        "plugin_test_tool",
		Description: "A test tool from plugin",
		Handler: func(args, ctx map[string]any) (any, error) {
			return "plugin_ok", nil
		},
	}

	p := &Plugin{
		Name:        "test-plugin",
		Version:     "0.1.0",
		Description: "Test plugin",
		Tools:       []*toolregistry.Tool{testTool},
		OnLoad: func() error {
			loaded = true
			return nil
		},
		OnUnload: func() {
			unloaded = true
		},
	}

	if err := mgr.Register(p); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if !loaded {
		t.Errorf("expected OnLoad to be invoked")
	}

	// Verify plugin exists in manager
	got, ok := mgr.Get("test-plugin")
	if !ok || got.Version != "0.1.0" {
		t.Errorf("expected to get test-plugin with version 0.1.0")
	}

	// Verify tool was registered in global tool registry
	foundTool, ok := toolregistry.Global().Get("plugin_test_tool")
	if !ok || foundTool.Name != "plugin_test_tool" {
		t.Errorf("expected plugin_test_tool to be registered")
	}

	// Verify FormatStatus
	status := mgr.FormatStatus()
	if !strings.Contains(status, "test-plugin") || !strings.Contains(status, "0.1.0") {
		t.Errorf("unexpected status format: %s", status)
	}

	// Unload plugin
	if err := mgr.Unload("test-plugin"); err != nil {
		t.Fatalf("unexpected unload error: %v", err)
	}
	if !unloaded {
		t.Errorf("expected OnUnload to be invoked")
	}
	if _, ok := toolregistry.Global().Get("plugin_test_tool"); ok {
		t.Errorf("expected plugin_test_tool to be removed after unload")
	}
}
