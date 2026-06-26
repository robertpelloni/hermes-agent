package agent

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// AgentContext holds the state for an active agent session.
type AgentContext struct {
	SessionID string
	IsActive  bool
	Memory    MemoryStore
}

// MemoryStore defines the interface for persistent memory storage.
// This lays the groundwork for the SQLite/file-based Phase 2 memory store.
type MemoryStore interface {
	SaveInteraction(role, message string) error
	GetHistory() ([]Interaction, error)
}

// Interaction represents a single message in the history.
type Interaction struct {
	Role      string
	Message   string
	Timestamp time.Time
}

// DummyMemory implements a temporary in-memory store before SQLite is wired.
type DummyMemory struct {
	history []Interaction
}

func (m *DummyMemory) SaveInteraction(role, message string) error {
	m.history = append(m.history, Interaction{
		Role:      role,
		Message:   message,
		Timestamp: time.Now(),
	})
	return nil
}

func (m *DummyMemory) GetHistory() ([]Interaction, error) {
	return m.history, nil
}

// StartRepl starts the interactive Read-Eval-Print Loop for the agent.
func StartRepl() {
	ctx := &AgentContext{
		SessionID: "repl-session-1",
		IsActive:  true,
		Memory:    &DummyMemory{},
	}

	fmt.Println("Starting Jules Agent REPL...")
	fmt.Println("Type 'exit' or 'quit' to terminate.")

	reader := bufio.NewReader(os.Stdin)

	for ctx.IsActive {
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
			ctx.IsActive = false
			break
		}

		// Save user input to memory
		ctx.Memory.SaveInteraction("user", input)

		// Process Input (Mock LLM Response for now)
		response := processInput(input)

		// Save agent response to memory
		ctx.Memory.SaveInteraction("agent", response)

		fmt.Printf("Jules: %s\n", response)
	}
}

// processInput acts as the orchestrator/dispatcher for the agent.
func processInput(input string) string {
	// In the final implementation, this routes to the reasoning engine,
	// evaluates tool calls, or interacts with the MCP protocol.
	return fmt.Sprintf("Acknowledged: '%s'. Awaiting connection to reasoning engine.", input)
}
