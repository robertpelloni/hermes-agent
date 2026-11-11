package agent_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/hermes-agent/pkg/agent"
	"github.com/robertpelloni/hermes-agent/pkg/sessionruntime"
)

func TestStandaloneLoopSlashCommands(t *testing.T) {
	cfg := agent.Config{
		Model:         "hermes-test",
		Provider:      "local-llm",
		BaseURL:       "http://127.0.0.1:4000/v1",
		MaxTokens:     512,
		Temperature:   0.7,
		TopP:          0.9,
		MaxIterations: 2,
		SystemPrompt:  "Test",
		SessionTTL:    1 * time.Hour,
	}

	ag := agent.New(cfg)
	tempDir := t.TempDir()
	sessRuntime := sessionruntime.NewRuntime(tempDir, false)
	ag.SetSessionRuntime(sessRuntime)

	// Feed input containing slash commands followed by /quit
	input := strings.NewReader("/help\n/model\n/sessions\n/clear\n/quit\n")
	var output bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ag.RunStandaloneLoop(ctx, agent.StandaloneOptions{
		In:        input,
		Out:       &output,
		SessionID: "cli:test_user",
	})
	if err != nil {
		t.Fatalf("unexpected error running standalone loop: %v", err)
	}

	outStr := output.String()
	if !strings.Contains(outStr, "Hermes Agent Go CLI") {
		t.Errorf("expected banner in output, got: %s", outStr)
	}
	if !strings.Contains(outStr, "/help") || !strings.Contains(outStr, "/model") {
		t.Errorf("expected slash command help in output, got: %s", outStr)
	}
	if !strings.Contains(outStr, "hermes-test") {
		t.Errorf("expected model name in output, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Exiting Hermes agent.") {
		t.Errorf("expected exit message in output, got: %s", outStr)
	}
}
