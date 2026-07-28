package prompttemplates

import (
	"strings"
	"testing"
)

func TestDefaultTemplateLoads(t *testing.T) {
	prompt, err := Load("default")
	if err != nil {
		t.Fatalf("Load('default') returned error: %v", err)
	}
	if prompt == "" {
		t.Fatal("default prompt is empty")
	}
	if !strings.Contains(prompt, "Hermes") {
		t.Errorf("default prompt should mention 'Hermes', got: %s", prompt)
	}
}

func TestCodingTemplateLoads(t *testing.T) {
	prompt, err := Load("coding")
	if err != nil {
		t.Fatalf("Load('coding') returned error: %v", err)
	}
	if prompt == "" {
		t.Fatal("coding prompt is empty")
	}
	if !strings.Contains(prompt, "read_file") {
		t.Errorf("coding prompt should mention tools, got: %s", prompt)
	}
}

func TestQATemplateLoads(t *testing.T) {
	prompt, err := Load("qa")
	if err != nil {
		t.Fatalf("Load('qa') returned error: %v", err)
	}
	if prompt == "" {
		t.Fatal("qa prompt is empty")
	}
	if !strings.Contains(prompt, "web_search") {
		t.Errorf("qa prompt should mention web_search, got: %s", prompt)
	}
}

func TestListTemplates(t *testing.T) {
	templates := List()
	if len(templates) == 0 {
		t.Fatal("List() returned empty list")
	}
	expected := map[string]bool{"default": true, "coding": true, "qa": true}
	for _, name := range templates {
		if !expected[name] {
			t.Errorf("unexpected template name: %s", name)
		}
		delete(expected, name)
	}
	if len(expected) > 0 {
		t.Errorf("missing templates: %v", expected)
	}
}

func TestLoadNonExistentReturnsError(t *testing.T) {
	_, err := Load("nonexistent")
	if err == nil {
		t.Fatal("Load('nonexistent') should return an error")
	}
}

func TestDefaultFunction(t *testing.T) {
	prompt := Default()
	if prompt == "" {
		t.Fatal("Default() returned empty prompt")
	}
	if !strings.Contains(prompt, "Hermes") {
		t.Errorf("Default() should mention 'Hermes', got: %s", prompt)
	}
}
