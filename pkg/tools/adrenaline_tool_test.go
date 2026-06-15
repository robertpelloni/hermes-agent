package tools

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// buildMockBinary creates a temporary binary that echoes its arguments as JSON.
// The binary name matches the first argument string (e.g. "adrenaline").
func buildMockBinary(t *testing.T, name string) string {
	t.Helper()
	src := `package main
import (
	"encoding/json"
	"os"
)
func main() {
	// Skip program name in args
	json.NewEncoder(os.Stdout).Encode(os.Args[1:])
}
`
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(src), 0o644); err != nil {
		t.Fatalf("failed to write mock source: %v", err)
	}
	binPath := filepath.Join(tmpDir, name)
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build mock binary %s: %v\n%s", name, err, out)
	}
	return binPath
}

func TestAdrenalineChatWrapper(t *testing.T) {
	mockBin := buildMockBinary(t, "adrenaline")
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", filepath.Dir(mockBin)+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	args := map[string]any{
		"prompt": "hello world",
		"model":  "claude-3",
	}
	out, err := adrenalineChatHandler(args, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	// The mock prints the args as JSON, e.g. ["chat","hello world","--model","claude-3"]
	var captured []string
	if err := json.Unmarshal([]byte(out.(string)), &captured); err != nil {
		t.Fatalf("failed to parse mock output: %v", out)
	}
	expectedCmd := []string{"chat", "hello world", "--model", "claude-3"}
	if len(captured) != len(expectedCmd) {
		t.Fatalf("unexpected args length: got %d, want %d (captured: %v)", len(captured), len(expectedCmd), captured)
	}
	for i := range expectedCmd {
		if captured[i] != expectedCmd[i] {
			t.Errorf("arg[%d]: got %q, want %q", i, captured[i], expectedCmd[i])
		}
	}
}

func TestAdrenalineChatMissingPrompt(t *testing.T) {
	args := map[string]any{"model": "claude-3"}
	_, err := adrenalineChatHandler(args, nil)
	if err == nil {
		t.Fatalf("expected error for missing prompt, got nil")
	}
	if _, ok := err.(*json.SyntaxError); ok {
		t.Fatalf("unexpected json error: %v", err)
	}
}