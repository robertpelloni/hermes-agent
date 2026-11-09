package subcmd

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunBinary(t *testing.T) {
	ctx := context.Background()

	// Execute a known basic command (e.g., echo)
	out, err := RunBinary(ctx, "echo", []string{"hello", "hermes"}, 3*time.Second)
	if err != nil {
		t.Fatalf("unexpected error running binary: %v", err)
	}
	if !strings.Contains(out, "hello hermes") {
		t.Errorf("expected output to contain 'hello hermes', got: %s", out)
	}

	// Test command timeout
	_, errTimeout := RunBinary(ctx, "sleep", []string{"2"}, 10*time.Millisecond)
	if errTimeout == nil {
		t.Errorf("expected timeout error for sleep command, got nil")
	}
}
