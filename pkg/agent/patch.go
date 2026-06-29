package agent

import (
	"fmt"
	"strings"
)

// ApplyDiffBlock applies a SEARCH/REPLACE diff block to a file's content
func ApplyDiffBlock(content string, searchBlock string, replaceBlock string) (string, error) {
	// Simple string replacement for now, mimicking Aider's format exactly
	if !strings.Contains(content, searchBlock) {
		return "", fmt.Errorf("search block not found in file content")
	}

	newContent := strings.Replace(content, searchBlock, replaceBlock, 1)
	return newContent, nil
}
