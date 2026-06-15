package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/robertpelloni/hermes-agent/internal/tui"
)

func main() {
    // Create the model as a pointer so we can attach the program.
    m := tui.NewModel()
    p := tea.NewProgram(m, tea.WithAltScreen())
    // Attach the running program to the model – this lets the agent's
    // streaming goroutine push events back into the Bubble Tea loop.
    m.SetProgram(p)
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "error running hermes‑tui: %v\n", err)
        os.Exit(1)
    }
}
