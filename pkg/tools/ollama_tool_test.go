package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOllamaRunWrapper(t *testing.T) {
	mockBin := buildMockBinary(t, "ollama")
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", filepath.Dir(mockBin)+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	args := map[string]any{
		"model":  "llama3",
		"prompt": "write a poem",
		"json":   false,
	}
	out, err := ollamaRunHandler(args, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	var captured []string
	if err := json.Unmarshal([]byte(out.(string)), &captured); err != nil {
		t.Fatalf("failed to parse mock output: %v", out)
	}
	// Expected: ["run","llama3","write a poem"]
	expectedCmd := []string{"run", "llama3", "write a poem"}
	if len(captured) != len(expectedCmd) {
		t.Fatalf("unexpected args length: got %d, want %d (captured: %v)", len(captured), len(expectedCmd), captured)
	}
	for i := range expectedCmd {
		if captured[i] != expectedCmd[i] {
			t.Errorf("arg[%d]: got %q, want %q", i, captured[i], expectedCmd[i])
		}
	}
}

func TestOllamaListWrapper(t *testing.T) {
	mockBin := buildMockBinary(t, "ollama")
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", filepath.Dir(mockBin)+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	args := map[string]any{}
	out, err := ollamaListHandler(args, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	var captured []string
	if err := json.Unmarshal([]byte(out.(string)), &captured); err != nil {
		t.Fatalf("failed to parse mock output: %v", out)
	}
	// Expected: ["list"]
	expectedCmd := []string{"list"}
	if len(captured) != len(expectedCmd) {
		t.Fatalf("unexpected args length: got %d, want %d (captured: %v)", len(captured), len(expectedCmd), captured)
	}
	for i := range expectedCmd {
		if captured[i] != expectedCmd[i] {
			t.Errorf("arg[%d]: got %q, want %q", i, captured[i], expectedCmd[i])
		}
	}
}

func TestOllamaRunMissingModel(t *testing.T) {
	args := map[string]any{"prompt": "hello"}
	_, err := ollamaRunHandler(args, nil)
	if err == nil {
		t.Fatalf("expected error for missing model, got nil")
	}
}