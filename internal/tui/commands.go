package tui

import (
	"fmt"
	"strings"
)

// CommandDef defines a slash command available in the TUI.
type CommandDef struct {
	Name        string
	Description string
	Category    string
	Aliases     []string
	ArgsHint    string
	Handler     func(args string) string
}

// CommandRegistry holds all registered slash commands.
var CommandRegistry = []CommandDef{
	{
		Name:        "help",
		Description: "Show available commands and usage",
		Category:    "Info",
		Aliases:     []string{"?", "h"},
		Handler: nil, // Initialized in init()
	},
	{
		Name:        "clear",
		Description: "Clear the transcript and start fresh",
		Category:    "Session",
		Aliases:     []string{"cls"},
		Handler: func(args string) string {
			return "_Transcript cleared. Starting fresh._"
		},
	},
	{
		Name:        "new",
		Description: "Start a new conversation session",
		Category:    "Session",
		Handler: func(args string) string {
			return "_Starting a new conversation session. Previous context has been cleared._"
		},
	},
	{
		Name:        "quit",
		Description: "Exit the TUI",
		Category:    "Exit",
		Aliases:     []string{"exit", "q"},
		Handler: func(args string) string {
			return "_Goodbye!_"
		},
	},
	{
		Name:        "status",
		Description: "Show current session status",
		Category:    "Info",
		Handler: func(args string) string {
			return "**Session Status:** Active\n\n- **Model:** free-llm\n- **Provider:** local-llm\n- **Mode:** Interactive TUI"
		},
	},
	{
		Name:        "tools",
		Description: "List available tools",
		Category:    "Tools & Skills",
		Handler: func(args string) string {
			// This would ideally query the tool registry
			return "**Available Tools:**\n\n- `terminal` - Execute shell commands\n- `read_file` - Read file contents\n- `write_file` - Write file contents\n- `search_files` - Search for files\n- `delegate_task` - Delegate to a subagent\n\nUse `hermes tools` in the CLI for full management."
		},
	},
	{
		Name:        "skills",
		Description: "List loaded skills",
		Category:    "Tools & Skills",
		Handler: func(args string) string {
			return "**Loaded Skills:**\n\n_No skills currently loaded._\n\nUse `hermes skills list` in the CLI to see available skills."
		},
	},
}

// FindCommand looks up a command by name or alias.
func FindCommand(name string) *CommandDef {
	name = strings.TrimPrefix(name, "/")
	for _, cmd := range CommandRegistry {
		if strings.EqualFold(cmd.Name, name) {
			return &cmd
		}
		for _, alias := range cmd.Aliases {
			if strings.EqualFold(alias, name) {
				return &cmd
			}
		}
	}
	return nil
}

// FilterCommands returns commands matching the query (fuzzy-ish match).
func FilterCommands(query string) []CommandDef {
	if query == "" {
		return CommandRegistry
	}
	query = strings.ToLower(query)
	var results []CommandDef
	for _, cmd := range CommandRegistry {
		if strings.Contains(strings.ToLower(cmd.Name), query) ||
			strings.Contains(strings.ToLower(cmd.Description), query) ||
			strings.Contains(strings.ToLower(cmd.Category), query) {
			results = append(results, cmd)
		}
	}
	return results
}
func init() {
	CommandRegistry[0].Handler = func(args string) string {
		var sb strings.Builder
		sb.WriteString("## Hermes TUI - Available Commands\n\n")
		sb.WriteString("| Command | Description |\n")
		sb.WriteString("|---------|-------------|\n")
		for _, cmd := range CommandRegistry {
			sb.WriteString(fmt.Sprintf("| `/%s` | %s |\n", cmd.Name, cmd.Description))
		}
		sb.WriteString("\n**Usage:** Type `/` followed by a command name, or press `/` to open the command palette.\n")
		return sb.String()
	}
}
