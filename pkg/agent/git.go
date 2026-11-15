package agent

import (
	"fmt"
	"os/exec"
)

// AutoCommit commits all changes in the repo
func AutoCommit(message string) error {
	cmd := exec.Command("git", "add", "-A")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	fmt.Printf("[hermes:git] Auto-committed: %s\n", message)
	return nil
}
