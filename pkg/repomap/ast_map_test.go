package repomap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateASTMap(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repomap-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.go")
	testContent := []byte(`
package test

func MyFunction() {
}

type MyStruct struct {
}
`)
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	res, err := GenerateASTMap(tempDir, []string{"test.go"})
	if err != nil {
		t.Fatalf("GenerateASTMap returned error: %v", err)
	}

	if !strings.Contains(res, "func MyFunction") {
		t.Errorf("Expected output to contain 'func MyFunction', got: %s", res)
	}
	if !strings.Contains(res, "type MyStruct") {
		t.Errorf("Expected output to contain 'type MyStruct', got: %s", res)
	}
}
