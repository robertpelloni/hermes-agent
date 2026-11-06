package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillRepository(t *testing.T) {
	repo := NewRepository()

	s := &Skill{
		Name:        "test-skill",
		Description: "A unit test skill",
		Code:        "print('test')",
	}
	repo.Register(s)

	found, ok := repo.Lookup("test-skill")
	if !ok || found.Description != "A unit test skill" {
		t.Fatalf("failed to lookup registered skill")
	}

	all := repo.List()
	if len(all) != 1 {
		t.Fatalf("expected 1 skill in repo, got %d", len(all))
	}
}

func TestSkillLoadFromDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hermes_skills_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	skillMd := `---
name: my-skill
description: "Demonstration skill markdown"
---
# Details
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMd), 0644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	pySkill := `"""Custom calculator tool."""
def calc(a, b):
    return a + b
`
	if err := os.WriteFile(filepath.Join(tmpDir, "calculator.py"), []byte(pySkill), 0644); err != nil {
		t.Fatalf("failed to write calculator.py: %v", err)
	}

	repo := NewRepository()
	loaded, err := repo.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir failed: %v", err)
	}
	if loaded != 2 {
		t.Fatalf("expected 2 loaded skills, got %d", loaded)
	}

	s1, ok := repo.Lookup("my-skill")
	if !ok || s1.Description != "Demonstration skill markdown" {
		t.Errorf("expected my-skill with description, got %+v", s1)
	}

	s2, ok := repo.Lookup("calculator")
	if !ok || s2.Description != "Custom calculator tool." {
		t.Errorf("expected calculator with description, got %+v", s2)
	}
}
