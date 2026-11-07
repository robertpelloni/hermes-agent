package modelregistry

import (
	"os"
	"strings"
	"testing"
)

func TestModelRegistry_ProvidersAndModels(t *testing.T) {
	reg := NewModelRegistry()

	// Verify default providers exist
	providers := reg.GetAllProviders()
	if len(providers) == 0 {
		t.Fatalf("expected non-empty providers list")
	}

	p, ok := reg.GetProviderInfo("openai")
	if !ok {
		t.Fatalf("expected openai provider to be registered")
	}
	if p.Label != "OpenAI" {
		t.Errorf("expected label OpenAI, got %s", p.Label)
	}

	// Verify default models list
	models := reg.GetAllModels()
	if len(models) == 0 {
		t.Fatalf("expected non-empty models list")
	}

	// Search models by pattern
	gptModels := reg.SearchModels("gpt")
	if len(gptModels) == 0 {
		t.Errorf("expected at least one gpt model")
	}

	// Resolve API key with dummy environment variable
	_ = os.Setenv("TEST_REG_KEY", "dummy-secret-key")
	defer os.Unsetenv("TEST_REG_KEY")

	reg.RegisterProvider(ProviderInfo{
		Name:    "test-provider",
		EnvVars: []string{"TEST_REG_KEY"},
		Label:   "Test Provider",
	})

	if !reg.HasAPIKey("test-provider") {
		t.Errorf("expected test-provider to have API key")
	}

	key, err := reg.GetAPIKey("test-provider")
	if err != nil {
		t.Fatalf("unexpected error getting API key: %v", err)
	}
	if key != "dummy-secret-key" {
		t.Errorf("expected dummy-secret-key, got %s", key)
	}

	// Format provider status
	status := reg.FormatProviderStatus()
	if !strings.Contains(status, "Test Provider") {
		t.Errorf("expected status to mention Test Provider: %s", status)
	}
}
