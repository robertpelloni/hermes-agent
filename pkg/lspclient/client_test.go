package lspclient

import (
	"testing"
)

func TestGetSingletonAndClose(t *testing.T) {
	client := Get()
	if client == nil {
		t.Fatalf("expected non-nil LSP client singleton")
	}

	// Double call should return the exact same instance
	client2 := Get()
	if client != client2 {
		t.Errorf("expected Get() to return the singleton instance")
	}

	// Close safely
	if err := client.Close(); err != nil {
		t.Errorf("unexpected error closing unstarted client: %v", err)
	}

	// After closing, ensureRunning should report client is shut down
	if err := client.ensureRunning(); err == nil {
		t.Errorf("expected error from ensureRunning after Close(), got nil")
	}
}
