package repomap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateRepoMap(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hermes-repomap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	_ = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Title\nDocumentation\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "helper.go"), []byte("package main\n\nfunc HelperFunc() {}\n"), 0644)

	res, err := Generate(Options{
		BaseDir:        tempDir,
		MaxFiles:       5,
		MentionedFiles: []string{"helper.go"},
		MentionedIdents: []string{"HelperFunc"},
	})
	if err != nil {
		t.Fatalf("unexpected error generating repo map: %v", err)
	}

	if res.Stats.Included != 3 {
		t.Errorf("expected 3 files included, got %d", res.Stats.Included)
	}

	if len(res.Files) == 0 {
		t.Fatalf("expected non-empty files")
	}

	// helper.go should be top ranked because it was explicitly mentioned
	if res.Files[0].Path != "helper.go" {
		t.Errorf("expected helper.go to be highest ranked, got %s", res.Files[0].Path)
	}

	if !strings.Contains(res.Map, "helper.go") || !strings.Contains(res.Map, "main.go") {
		t.Errorf("expected map to contain helper.go and main.go: %s", res.Map)
	}
}
