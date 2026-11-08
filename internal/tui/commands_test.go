package tui

import (
	"strings"
	"testing"
)

func TestFindCommand(t *testing.T) {
	// Look up by direct name
	cmd := FindCommand("help")
	if cmd == nil {
		t.Fatalf("expected to find /help command")
	}
	if cmd.Name != "help" {
		t.Errorf("expected name 'help', got %s", cmd.Name)
	}

	// Look up with slash prefix
	cmdSlash := FindCommand("/help")
	if cmdSlash == nil || cmdSlash.Name != "help" {
		t.Errorf("expected /help lookup to match 'help'")
	}

	// Look up by alias
	cmdAlias := FindCommand("?")
	if cmdAlias == nil || cmdAlias.Name != "help" {
		t.Errorf("expected '?' lookup to match 'help'")
	}

	cmdQuit := FindCommand("q")
	if cmdQuit == nil || cmdQuit.Name != "quit" {
		t.Errorf("expected 'q' lookup to match 'quit'")
	}

	// Unknown command
	if unknown := FindCommand("nonexistent"); unknown != nil {
		t.Errorf("expected nil for nonexistent command, got %+v", unknown)
	}
}

func TestCommandHandlers(t *testing.T) {
	helpCmd := FindCommand("help")
	out := helpCmd.Handler("")
	if !strings.Contains(out, "Available Commands") {
		t.Errorf("expected help output, got: %s", out)
	}

	clearCmd := FindCommand("clear")
	if !strings.Contains(clearCmd.Handler(""), "Transcript cleared") {
		t.Errorf("unexpected clear output")
	}

	statusCmd := FindCommand("status")
	if !strings.Contains(statusCmd.Handler(""), "Active") {
		t.Errorf("unexpected status output")
	}
}
