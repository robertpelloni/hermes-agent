package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	pkgagent "github.com/robertpelloni/hermes-agent/pkg/agent"
)

// StartRepl starts the interactive Read-Eval-Print Loop for the agent.
func StartRepl() {
	fmt.Println("Starting Jules Agent REPL...")
	fmt.Println("Type 'exit' or 'quit' to terminate.")

	agent := pkgagent.New(pkgagent.DefaultConfig())
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("\nError reading input: %v\n", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Shutting down agent loop.")
			break
		}

		// Process Input using the actual LLM agent from pkg/agent
		response, err := agent.HandleMessage(context.Background(), "cli", "repl-session-1", input)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		fmt.Printf("Jules: %s\n", response)
	}
}
