package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/robertpelloni/hermes-agent/pkg/ai"
)

// StandaloneOptions configures interactive standalone agent loop execution.
type StandaloneOptions struct {
	In          io.Reader
	Out         io.Writer
	SessionID   string
	PromptPrefix string
}

// RunStandaloneLoop runs an interactive conversation loop over In/Out streams.
// It intercepts slash commands (/help, /model, /skills, /sessions, /clear, /quit, /exit)
// and passes user prompts to the agent streaming pipeline.
func (a *Agent) RunStandaloneLoop(ctx context.Context, opts StandaloneOptions) error {
	in := opts.In
	out := opts.Out
	if in == nil {
		return fmt.Errorf("input reader cannot be nil")
	}
	if out == nil {
		return fmt.Errorf("output writer cannot be nil")
	}

	prefix := opts.PromptPrefix
	if prefix == "" {
		prefix = "> "
	}

	sessID := opts.SessionID
	if sessID == "" {
		sessID = "cli:standalone"
	}

	fmt.Fprintf(out, "Hermes Agent Go CLI (model: %s, provider: %s)\n", a.config.Model, a.config.Provider)
	fmt.Fprintf(out, "Type /help for available commands, or /quit to exit.\n\n")

	scanner := bufio.NewScanner(in)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Fprint(out, prefix)
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Handle slash commands
		if strings.HasPrefix(line, "/") {
			handled, shouldExit := a.handleSlashCommand(out, line, sessID)
			if shouldExit {
				fmt.Fprintln(out, "Exiting Hermes agent.")
				return nil
			}
			if handled {
				continue
			}
		}

		// Stream response from agent
		eventCh := a.HandleMessageStream(ctx, "cli", strings.TrimPrefix(sessID, "cli:"), line)
		var thinkingActive bool

		for event := range eventCh {
			switch event.Type {
			case EventThinking:
				if !thinkingActive {
					fmt.Fprint(out, "\n[thinking] ")
					thinkingActive = true
				}
				fmt.Fprint(out, event.Text)

			case EventText:
				if thinkingActive {
					fmt.Fprint(out, "\n[/thinking]\n\n")
					thinkingActive = false
				}
				fmt.Fprint(out, event.Text)

			case EventToolStart:
				fmt.Fprintf(out, "\n[tool: %s]\n", event.ToolName)

			case EventToolDone:
				if event.ToolError != nil {
					fmt.Fprintf(out, "[tool error: %v]\n", event.ToolError)
				} else {
					fmt.Fprintln(out, "[tool completed]")
				}

			case EventError:
				fmt.Fprintf(out, "\nError: %v\n", event.Error)

			case EventDone:
				if thinkingActive {
					fmt.Fprintln(out, "\n[/thinking]")
					thinkingActive = false
				}
				fmt.Fprintln(out)
			}
		}
	}

	return scanner.Err()
}

// handleSlashCommand handles built-in slash commands and returns (handled, shouldExit).
func (a *Agent) handleSlashCommand(out io.Writer, cmd, sessID string) (bool, bool) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false, false
	}

	switch parts[0] {
	case "/quit", "/exit":
		return true, true

	case "/help":
		fmt.Fprintln(out, "Available slash commands:")
		fmt.Fprintln(out, "  /help         - Show this help message")
		fmt.Fprintln(out, "  /model        - Show current model and provider")
		fmt.Fprintln(out, "  /sessions     - List active sessions")
		fmt.Fprintln(out, "  /clear        - Clear current session conversation turns")
		fmt.Fprintln(out, "  /quit, /exit  - Exit agent session")
		return true, false

	case "/model":
		fmt.Fprintf(out, "Model: %s\nProvider: %s\nBaseURL: %s\n", a.config.Model, a.config.Provider, a.config.BaseURL)
		return true, false

	case "/sessions":
		if a.sessions == nil {
			fmt.Fprintln(out, "Session runtime not attached.")
			return true, false
		}
		sessions := a.sessions.List()
		fmt.Fprintf(out, "Active sessions (%d):\n", len(sessions))
		for _, s := range sessions {
			fmt.Fprintf(out, "  - %s (%d messages, label: %q)\n", s.ID, len(s.Messages), s.Label)
		}
		return true, false

	case "/clear":
		if a.sessions != nil {
			_ = a.sessions.Delete(sessID)
			created, err := a.sessions.Create(sessID)
			if err == nil && created != nil {
				created.Messages = []ai.Message{}
			}
		}
		fmt.Fprintln(out, "Conversation cleared.")
		return true, false

	default:
		fmt.Fprintf(out, "Unknown command %q. Type /help for available commands.\n", parts[0])
		return true, false
	}
}
