package skill_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robertpelloni/hermes-agent/pkg/skill"
)

func TestDiscoverAndLoad(t *testing.T) {
	// Setup dummy skills dir
	dir := t.TempDir()

	validPy := filepath.Join(dir, "my_skill.py")
	os.WriteFile(validPy, []byte(`"""
This is a test skill.
"""
def hello():
    pass
`), 0644)

	invalidPy := filepath.Join(dir, "not_a_skill.txt")
	os.WriteFile(invalidPy, []byte(`hello`), 0644)

	initPy := filepath.Join(dir, "__init__.py")
	os.WriteFile(initPy, []byte(``), 0644)

	repo := skill.NewRepository()
	err := repo.DiscoverAndLoad(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	skills := repo.List()
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	if skills[0].Name != "my_skill" {
		t.Errorf("expected skill name 'my_skill', got %s", skills[0].Name)
	}

	if skills[0].Description != "This is a test skill." {
		t.Errorf("expected description 'This is a test skill.', got %s", skills[0].Description)
	}
}
