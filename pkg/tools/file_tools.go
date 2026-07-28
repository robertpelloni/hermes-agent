package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

func init() {
	// Register read_file
	toolregistry.Global().Register(&toolregistry.Tool{
		Name:        "read_file",
		Description: "Read the entire contents of a file. Returns a string. If the file does not exist an error is returned.",
		Category:    "file",
		Parameters: map[string]any{
			"path": map[string]string{"type": "string", "description": "Path to the file to read"},
		},
		Handler: readFileHandler,
		Native:  true,
	})

	// Register write_file
	toolregistry.Global().Register(&toolregistry.Tool{
		Name:        "write_file",
		Description: "Write a string to a file, optionally creating parent directories.",
		Category:    "file",
		Parameters: map[string]any{
			"path":    map[string]string{"type": "string", "description": "Path to write"},
			"content": map[string]string{"type": "string", "description": "File content"},
			"mode":    map[string]string{"type": "string", "description": "File mode (e.g., '0644'), optional"},
		},
		Handler: writeFileHandler,
		Native:  true,
	})
}

func readFileHandler(args map[string]any, _ map[string]any) (any, error) {
	pRaw, ok := args["path"]
	if !ok {
		return nil, fmt.Errorf("missing 'path' argument")
	}
	pathStr, ok := pRaw.(string)
	if !ok {
		return nil, fmt.Errorf("'path' must be a string")
	}
	absPath := filepath.Clean(pathStr)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func writeFileHandler(args map[string]any, _ map[string]any) (any, error) {
	pRaw, ok := args["path"]
	if !ok {
		return nil, fmt.Errorf("missing 'path' argument")
	}
	pathStr, ok := pRaw.(string)
	if !ok {
		return nil, fmt.Errorf("'path' must be a string")
	}
	cRaw, ok := args["content"]
	if !ok {
		return nil, fmt.Errorf("missing 'content' argument")
	}
	content, ok := cRaw.(string)
	if !ok {
		return nil, fmt.Errorf("'content' must be a string")
	}

	mode := os.FileMode(0644)
	if modeRaw, ok := args["mode"]; ok {
		if modeStr, ok := modeRaw.(string); ok {
			parsed, err := parseIntMode(modeStr)
			if err == nil {
				mode = os.FileMode(parsed)
			}
		}
	}

	absPath := filepath.Clean(pathStr)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return nil, err
	}

	if err := os.WriteFile(absPath, []byte(content), mode); err != nil {
		return nil, err
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), absPath), nil
}

// parseIntMode parses a permission string like "0644" or "755" into a uint32.
func parseIntMode(s string) (uint32, error) {
	s = strings.TrimPrefix(s, "0")
	if s == "" {
		return 0, fmt.Errorf("empty mode string")
	}
	var val uint32
	for _, c := range s {
		if c < '0' || c > '7' {
			return 0, fmt.Errorf("invalid octal digit '%c' in mode %q", c, s)
		}
		val = (val << 3) | uint32(c-'0')
	}
	return val, nil
}
