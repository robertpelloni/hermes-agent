package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/robertpelloni/hermes-agent/pkg/memory"
)

// StartRepl starts the interactive agent REPL.
func StartRepl() {
	fmt.Println("Starting Hermes Agent REPL...")
	fmt.Println("Type 'exit' or 'quit' to exit.")

	memStore := memory.NewStore()
	agent := New(DefaultConfig())

	// Optionally wire up VFS or other systems here...

	scanner := bufio.NewScanner(os.Stdin)
	sessionID := "repl-session"

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}

		// Store user input in memory
		memStore.SaveMessage(sessionID, "user", input)

		// Process input via the agent. We inject the memory interaction directly here for the REPL loop.
		response, err := agent.HandleMessage(context.Background(), "cli", sessionID, input)

		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		// Store assistant response in memory
		memStore.SaveMessage(sessionID, "assistant", response)

		fmt.Printf("\nAgent: %s\n", response)
	}
}
